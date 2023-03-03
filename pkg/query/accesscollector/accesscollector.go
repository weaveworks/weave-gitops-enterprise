package collector

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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// accessRulesCollector is responsible for collecting access rules from all clusters.
// It is a wrapper around a Collector that converts the received objects to AccessRules.
// It writes the received rules to a StoreWriter.
type accessRulesCollector struct {
	col       collector.Collector
	log       logr.Logger
	converter runtime.UnstructuredConverter
	w         store.StoreWriter
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

				rules, err := handleRulesReceived(a.converter, objects)
				if err != nil {
					a.log.Error(err, "failed to handle rules received")
					continue
				}

				if err := a.w.StoreAccessRules(rules); err != nil {
					a.log.Error(err, "failed to store access rules")
					continue
				}
			}
		}
	}()
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
	}
}

func handleRulesReceived(converter runtime.UnstructuredConverter, objects []collector.ObjectRecord) ([]models.AccessRule, error) {
	result := []models.AccessRule{}

	for _, o := range objects {
		c := o.ClusterName()
		obj := o.Object()
		var r models.AccessRule

		adapter, err := newAdapter(obj)
		if err != nil {
			return result, fmt.Errorf("failed to create adapter for object: %w", err)
		}

		r = convertToAccessRule(c, adapter)

		result = append(result, r)

	}

	return result, nil
}

// func (c *pollingCollector) CollectAccessRules() ([]models.AccessRule, error) {
// 	result := []models.AccessRule{}

// 	clusters := c.mgr.GetClusters()

// 	for _, clus := range clusters {
// 		clusterName := clus.GetName()
// 		cl, err := clus.GetServerClient()
// 		if err != nil {
// 			c.log.Error(err, "failed to get client for cluster")
// 			continue
// 		}

// 		cRoles := v1.ClusterRoleList{}
// 		if err := cl.List(context.Background(), &cRoles); err != nil {
// 			c.log.Error(err, "failed to list cluster roles")
// 		}

// 		roles := v1.RoleList{}
// 		if err := cl.List(context.Background(), &roles); err != nil {
// 			c.log.Error(err, "failed to list roles")
// 		}

// 		for _, cRole := range cRoles.Items {
// 			result = append(result, models.AccessRule{
// 				Cluster:         clusterName,
// 				Role:            cRole.Name,
// 				Namespace:       cRole.Namespace,
// 				AccessibleKinds: cRole.Rules[0].Resources,
// 			})
// 		}

// 		for _, r := range roles.Items {
// 			result = append(result, models.AccessRule{
// 				Cluster:         clusterName,
// 				Role:            r.Name,
// 				Namespace:       r.Namespace,
// 				AccessibleKinds: r.Rules[0].Resources,
// 			})
// 		}

// 	}

// 	return result, nil
// }

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

func convertToAccessRule(clusterName string, obj RoleLike) models.AccessRule {
	rules := obj.GetRules()

	return models.AccessRule{
		Cluster:         clusterName,
		Role:            obj.GetName(),
		Namespace:       obj.GetNamespace(),
		AccessibleKinds: rules[0].Resources,
	}
}
