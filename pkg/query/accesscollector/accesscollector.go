package accesscollector

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubernetes/pkg/apis/rbac"
	"k8s.io/kubernetes/pkg/util/slice"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		ch, error := a.col.Start()
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

	for _, o := range objects {
		c := o.ClusterName()
		obj := o.Object()
		var r models.AccessRule

		adapter, err := newAdapter(obj)
		if err != nil {
			return result, fmt.Errorf("failed to create adapter for object: %w", err)
		}

		r = convertToAccessRule(c, adapter, a.verbs)

		result = append(result, r)

	}

	return result, nil
}

// RoleLike is an interface that represents a role or cluster role
// Tried this with generics but it didn't work out.
type RoleLike interface {
	client.Object

	GetRules() []v1.PolicyRule
}

type cRoleAdapter struct {
	*v1.ClusterRole
}

func (c *cRoleAdapter) GetRules() []v1.PolicyRule {
	return c.Rules
}

type roleAdapter struct {
	*v1.Role
}

func (r *roleAdapter) GetRules() []v1.PolicyRule {
	return r.Rules
}

func newAdapter(obj client.Object) (RoleLike, error) {
	switch o := obj.(type) {
	case *v1.ClusterRole:
		return &cRoleAdapter{o}, nil
	case *v1.Role:
		return &roleAdapter{o}, nil

	default:
		return nil, fmt.Errorf("unknown object type %T", obj)
	}

}

func convertToAccessRule(clusterName string, obj RoleLike, requiredVerbs []string) models.AccessRule {
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
		Role:            obj.GetName(),
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
