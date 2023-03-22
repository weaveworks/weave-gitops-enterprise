package accesscollector

import (
	"context"
	"fmt"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/models"

	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/kubectl/pkg/util/slice"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/adapters"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var DefaultVerbsRequiredForAccess = []string{"list"}

// AccessRulesCollector is responsible for collecting access rules from all clusters.
// It is a wrapper around a Collector that converts the received objects to AccessRules.
// It writes the received rules to a StoreWriter.
type AccessRulesCollector struct {
	col       collector.Collector
	log       logr.Logger
	converter runtime.UnstructuredConverter
	w         store.StoreWriter
	verbs     []string
	quit      chan struct{}
}

func (a *AccessRulesCollector) Start(ctx context.Context) error {
	err := a.col.Start()
	if err != nil {
		return fmt.Errorf("could not start access collector: %w", err)
	}
	return nil
}

func (a *AccessRulesCollector) Stop() error {
	a.quit <- struct{}{}
	return a.col.Stop()
}

func NewAccessRulesCollector(w store.Store, opts collector.CollectorOpts) (*AccessRulesCollector, error) {
	opts.ObjectKinds = []schema.GroupVersionKind{
		rbacv1.SchemeGroupVersion.WithKind("ClusterRole"),
		rbacv1.SchemeGroupVersion.WithKind("Role"),
		rbacv1.SchemeGroupVersion.WithKind("ClusterRoleBinding"),
		rbacv1.SchemeGroupVersion.WithKind("RoleBinding"),
	}

	opts.ProcessRecordsFunc = defaultProcessRecords

	col, err := collector.NewCollector(opts, w)

	if err != nil {
		return nil, fmt.Errorf("cannot create collector: %w", err)
	}
	return &AccessRulesCollector{
		col:       col,
		log:       opts.Log,
		converter: runtime.DefaultUnstructuredConverter,
		w:         w,
		verbs:     DefaultVerbsRequiredForAccess,
	}, nil
}

func defaultProcessRecords(ctx context.Context, objectRecords []models.ObjectTransaction, store store.Store, log logr.Logger) error {
	upsert, _, err := handleRulesReceived(objectRecords)
	if err != nil {
		return fmt.Errorf("unable to receive rules: %w", err)
	}

	if err := store.StoreAccessRules(ctx, upsert); err != nil {
		return fmt.Errorf("cannot store access rules: %w", err)
	}

	return nil
}

func handleRulesReceived(objects []models.ObjectTransaction) ([]models.AccessRule, []models.AccessRule, error) {
	upsert := []models.AccessRule{}
	// TODO; we will need to figure out how to calculate the removal of an access rule.
	// If a Role gets removed, the RoleBinding might still be there and vice versa.
	// We will need to remove an access rule when EITHER the Role or the Binding is removed.
	remove := []models.AccessRule{}

	roles := []adapters.RoleLike{}
	bindings := []adapters.BindingLike{}

	for _, obj := range objects {
		kind := obj.Object().GetObjectKind().GroupVersionKind().Kind
		if kind == "ClusterRole" || kind == "Role" {
			adapter, err := adapters.NewRoleAdapter(obj.ClusterName(), obj.Object())
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create adapter for object: %w", err)
			}
			roles = append(roles, adapter)
		}

		if kind == "ClusterRoleBinding" || kind == "RoleBinding" {
			adapter, err := adapters.NewBindingAdapter(obj.ClusterName(), obj.Object())
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create binding adapter: %w", err)
			}

			bindings = append(bindings, adapter)
		}
	}

	// Figure out the binding/role pairs
	for _, binding := range bindings {
		for _, role := range roles {
			if bindingRoleMatch(binding, role) {
				upsert = append(upsert, convertToAccessRule(role.GetClusterName(), role, DefaultVerbsRequiredForAccess))
			}
		}

	}

	return upsert, remove, nil
}

func (a *AccessRulesCollector) Watch(cluster cluster.Cluster, objectsChannel chan []models.ObjectTransaction, ctx context.Context, log logr.Logger) error {
	return a.col.Watch(cluster, objectsChannel, ctx, log)
}

func (a *AccessRulesCollector) Status(cluster cluster.Cluster) (string, error) {
	return a.col.Status(cluster)
}

func convertToAccessRule(clusterName string, obj adapters.RoleLike, requiredVerbs []string) models.AccessRule {
	rules := obj.GetRules()

	derivedAccess := map[string]map[string]bool{}

	// {wego.weave.works: {Application: true, Source: true}}
	for _, rule := range rules {
		for _, apiGroup := range rule.APIGroups {
			if _, ok := derivedAccess[apiGroup]; !ok {
				derivedAccess[apiGroup] = map[string]bool{}
			}

			if containsWildcard(rule.Resources) {
				derivedAccess[apiGroup]["*"] = true
			}

			if containsWildcard(rule.Verbs) || hasVerbs(rule.Verbs, requiredVerbs) {
				for _, resource := range rule.Resources {
					derivedAccess[apiGroup][resource] = true
				}
			}
		}
	}

	accessibleKinds := []string{}
	for group, resources := range derivedAccess {
		for k, v := range resources {
			if v {
				accessibleKinds = append(accessibleKinds, fmt.Sprintf("%s/%s", group, k))
			}
		}
	}

	return models.AccessRule{
		Cluster:         clusterName,
		Principal:       obj.GetName(),
		Namespace:       obj.GetNamespace(),
		AccessibleKinds: accessibleKinds,
	}
}

func hasVerbs(a, b []string) bool {
	for _, v := range b {
		if containsWildcard(a) {
			return true
		}
		if slice.ContainsString(a, v, nil) {
			return true
		}
	}

	return false
}

func containsWildcard(permissions []string) bool {
	for _, p := range permissions {
		if p == "*" {
			return true
		}
	}

	return false
}

func bindingRoleMatch(binding adapters.BindingLike, role adapters.RoleLike) bool {
	ref := binding.GetRoleRef()
	roleGroup := role.GetObjectKind().GroupVersionKind().Group
	roleKind := role.GetObjectKind().GroupVersionKind().Kind

	match := ref.APIGroup == roleGroup && binding.GetNamespace() == role.GetNamespace() && ref.Kind == roleKind && ref.Name == role.GetName()
	return match
}
