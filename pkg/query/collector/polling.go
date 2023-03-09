package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// pollingCollector is a collector that polls the cluster for objects at a regular interval.
// This is meant to be a temporary solution until we can implement a `watch` pattern.
type pollingCollector struct {
	mgr    clustersmngr.ClustersManager
	log    logr.Logger
	kinds  []schema.GroupVersionKind
	ticker *time.Ticker
	quit   chan bool
	msg    chan []ObjectRecord
}

func (c *pollingCollector) Start() (<-chan []ObjectRecord, error) {
	c.log.Info("starting polling collector")
	c.quit = make(chan bool)
	go func() {
		for {
			select {
			case <-c.ticker.C:
				objects, err := c.collect(context.Background())
				if err != nil {
					c.log.Error(err, "failed to collect objects")
					continue
				}

				c.msg <- objects

			case <-c.quit:
				return
			}
		}
	}()

	return c.msg, nil
}

func (c *pollingCollector) Stop() error {
	c.quit <- true
	c.ticker.Stop()
	close(c.quit)
	close(c.msg)
	return nil
}

type record struct {
	clusterName string
	object      client.Object
}

func (r record) ClusterName() string {
	return r.clusterName
}

func (r record) Object() client.Object {
	return r.object
}

func (c *pollingCollector) collect(ctx context.Context) ([]ObjectRecord, error) {
	result := []ObjectRecord{}

	clusters := c.mgr.GetClusters()
	namespaces := c.mgr.GetClustersNamespaces()

	for _, clus := range clusters {
		clusterName := clus.GetName()
		// cls, err := clus.GetServerClientset()
		// if err != nil {
		// 	c.log.Error(err, "failed to get client for cluster")
		// 	continue
		// }

		cfg, err := clus.GetServerConfig()
		if err != nil {
			c.log.Error(err, "failed to get config for cluster")
			continue
		}

		dyn, err := dynamic.NewForConfig(cfg)
		if err != nil {
			c.log.Error(err, "failed to get dynamic client for cluster")
			continue
		}

		for _, ns := range namespaces[clusterName] {
			for _, kind := range c.kinds {

				gvr := kind.GroupVersion().WithResource(kind.Kind)

				list, err := dyn.Resource(gvr).Namespace(ns.Name).List(ctx, metav1.ListOptions{})

				if err != nil {
					if apierrors.IsNotFound(err) {
						fmt.Printf("not found: %s/%s in %s\n", gvr.Group, gvr.Resource, clusterName)
						continue
					}

					if apierrors.IsForbidden(err) {
						fmt.Printf("forbidden: %s/%s in %s\n", gvr.Group, gvr.Resource, clusterName)
						continue
					}

					c.log.Error(err, "failed to list objects", "cluster", clusterName, "namespace", ns.Name, "kind", kind.Kind)
					continue
				}

				for _, obj := range list.Items {
					result = append(result, record{
						clusterName: clusterName,
						object:      &obj,
					})
				}
			}
		}

	}

	return result, nil
}
