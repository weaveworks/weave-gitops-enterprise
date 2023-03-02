package query

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Collector interface {
	CollectAccessRules() ([]models.AccessRule, error)
	CollectObjects() ([]models.Object, error)
}

type baseCollector struct {
	mgr   clustersmngr.ClustersManager
	log   logr.Logger
	kinds []schema.GroupVersionKind
}

type CollectorOpts struct {
	ObjectKinds []schema.GroupVersionKind
}

func NewCollector(l logr.Logger, fetcher clustersmngr.ClustersManager, opts CollectorOpts) Collector {
	return &baseCollector{
		mgr:   fetcher,
		log:   l,
		kinds: opts.ObjectKinds,
	}
}

func (c *baseCollector) CollectAccessRules() ([]models.AccessRule, error) {
	result := []models.AccessRule{}

	clusters := c.mgr.GetClusters()

	fmt.Println(len(clusters))

	for _, clus := range clusters {
		clusterName := clus.GetName()
		fmt.Println(clusterName)
		cl, err := clus.GetServerClient()
		if err != nil {
			c.log.Error(err, "failed to get client for cluster")
			continue
		}

		cRoles := v1.ClusterRoleList{}
		if err := cl.List(context.Background(), &cRoles); err != nil {
			c.log.Error(err, "failed to list cluster roles")
		}

		roles := v1.RoleList{}
		if err := cl.List(context.Background(), &roles); err != nil {
			c.log.Error(err, "failed to list roles")
		}

		for _, cRole := range cRoles.Items {
			result = append(result, models.AccessRule{
				Cluster:         clusterName,
				Role:            cRole.Name,
				Namespace:       cRole.Namespace,
				AccessibleKinds: cRole.Rules[0].Resources,
			})
		}

		for _, r := range roles.Items {
			result = append(result, models.AccessRule{
				Cluster:         clusterName,
				Role:            r.Name,
				Namespace:       r.Namespace,
				AccessibleKinds: r.Rules[0].Resources,
			})
		}

	}

	return result, nil
}

func (c *baseCollector) CollectObjects() ([]models.Object, error) {

	result := []models.Object{}

	nsLookup := c.mgr.GetClustersNamespaces()

	for _, cluster := range c.mgr.GetClusters() {
		cl, err := cluster.GetServerClient()
		if err != nil {
			c.log.Error(err, "failed to get client for cluster")
			continue
		}

		for _, kind := range c.kinds {
			namespaces := nsLookup[cluster.GetName()]

			for _, ns := range namespaces {
				listResult := unstructured.UnstructuredList{}

				listResult.SetGroupVersionKind(kind)

				if err := cl.List(context.Background(), &listResult, client.InNamespace(ns.Name)); err != nil {
					c.log.Error(err, "failed to list objects")
					continue
				}

				for _, o := range listResult.Items {
					r := convertK8sToModelObject(o)
					r.Cluster = cluster.GetName()
					result = append(result, r)
				}
			}

		}

	}
	return result, nil
}

func convertK8sToModelObject(obj unstructured.Unstructured) models.Object {
	return models.Object{
		Kind:      obj.GetKind(),
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
}
