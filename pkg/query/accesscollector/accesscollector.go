package accesscollector

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/adapters"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubernetes/pkg/apis/rbac"
	"k8s.io/kubernetes/pkg/util/slice"
)

var DefaultVerbsRequiredForAccess = []string{"list"}

// accessRulesCollector is responsible for collecting access rules from all clusters.
// It is a wrapper around a Collector that converts the received objects to AccessRules.
// It writes the received rules to a StoreWriter.
type accessRulesCollector struct {
	col       collector.Collector
	log       logr.Logger
	converter runtime.UnstructuredConverter
	w         store.StoreWriter
	verbs     []string
	quit      chan struct{}
}

func (a *accessRulesCollector) Start() {
	go func() {
		error := a.col.Start()
		if error != nil {
			a.log.Error(error, "failed to start collector")
			return
		}
		for {
			select {
			case objects := <-ch:
				rules, err := a.handleRulesReceived(objects)
				if err != nil {
					a.log.Error(err, "failed to handle rules received")
					continue
				}

				if err := a.w.StoreAccessRules(rules); err != nil {
					a.log.Error(err, "failed to store access rules")
					continue
				}
			case <-a.quit:
				return
			}
		}
	}()
}

func (a *accessRulesCollector) Stop() {
	a.col.Stop()
	a.quit <- struct{}{}
	close(a.quit)
}

func NewAccessRulesCollector(w store.StoreWriter, opts collector.CollectorOpts) accessRulesCollector {
	opts.ObjectKinds = []schema.GroupVersionKind{
		rbac.SchemeGroupVersion.WithKind("ClusterRole"),
		rbac.SchemeGroupVersion.WithKind("Role"),
		rbac.SchemeGroupVersion.WithKind("ClusterRoleBinding"),
		rbac.SchemeGroupVersion.WithKind("RoleBinding"),
	}
	col := collector.NewCollector(opts)

	return accessRulesCollector{
		col:       col,
		log:       opts.Log,
		converter: runtime.DefaultUnstructuredConverter,
		w:         w,
		verbs:     DefaultVerbsRequiredForAccess,
	}
}

func (a *accessRulesCollector) handleRulesReceived(objects []collector.ObjectRecord) ([]models.AccessRule, error) {
	result := []models.AccessRule{}

	roles := []adapters.RoleLike{}
	bindings := []adapters.BindingLike{}

	for _, obj := range objects {
		kind := obj.Object().GetObjectKind().GroupVersionKind().Kind

		if kind == "ClusterRole" || kind == "Role" {
			adapter, err := adapters.NewRoleAdapter(obj.Object())
			if err != nil {
				return result, fmt.Errorf("failed to create adapter for object: %w", err)
			}
			roles = append(roles, adapter)
		}

		if kind == "ClusterRoleBinding" || kind == "RoleBinding" {
			adapter, err := adapters.NewBindingAdapter(obj.Object())
			if err != nil {
				return result, fmt.Errorf("failed to create binding adapter: %w", err)
			}

			bindings = append(bindings, adapter)
		}
	}

	// Figure out the binding/role pairs

	return result, nil
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

	kinds2 := []string{}
	for group, resources := range derivedAccess {
		for k, v := range resources {
			if v {
				kinds2 = append(kinds2, fmt.Sprintf("%s/%s", group, k))
			}
		}
	}

	return models.AccessRule{
		Cluster:         clusterName,
		Principal:       obj.GetName(),
		Namespace:       obj.GetNamespace(),
		AccessibleKinds: kinds2,
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
