package accesscollector

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/kubectl/pkg/util/slice"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/adapters"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
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
	roles := []models.Role{}
	bindings := []models.RoleBinding{}

	for _, obj := range objectRecords {
		kind := obj.Object().GetObjectKind().GroupVersionKind().Kind

		if kind == "ClusterRole" || kind == "Role" {
			role, err := adapters.NewRoleAdapter(obj.ClusterName(), obj.Object())
			if err != nil {
				return fmt.Errorf("cannot create role: %w", err)
			}
			roles = append(roles, role.ToModel())
		}

		if kind == "ClusterRoleBinding" || kind == "RoleBinding" {
			binding, err := adapters.NewBindingAdapter(obj.ClusterName(), obj.Object())
			if err != nil {
				return fmt.Errorf("cannot create binding: %w", err)
			}
			bindings = append(bindings, binding.ToModel())
		}
	}

	if len(roles) > 0 {
		if err := store.StoreRoles(ctx, roles); err != nil {
			return fmt.Errorf("cannot store roles: %w", err)
		}
	}

	if len(bindings) > 0 {
		if err := store.StoreRoleBindings(ctx, bindings); err != nil {
			return fmt.Errorf("cannot store role bindings: %w", err)
		}
	}

	return nil
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

	ar := models.AccessRule{
		Cluster:         clusterName,
		Namespace:       obj.GetNamespace(),
		AccessibleKinds: accessibleKinds,
		Subjects:        []models.Subject{},
	}

	for _, subject := range obj.GetSubjects() {
		ar.Subjects = append(ar.Subjects, models.Subject{
			Kind:      subject.Kind,
			Name:      subject.Name,
			Namespace: subject.Namespace,
		})
	}

	return ar
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
	if binding.GetClusterName() != role.GetClusterName() {
		return false
	}

	if binding.GetNamespace() != role.GetNamespace() {
		return false
	}

	if binding.GetRoleRef().Kind != role.GetObjectKind().GroupVersionKind().Kind {
		return false
	}

	if binding.GetRoleRef().Name != role.GetName() {
		return false
	}

	return true
}
