package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// pollingCollector is a collector that polls the cluster for objects at a regular interval.
// This is meant to be a temporary solution until we can implement a `watch` pattern.
type pollingCollector struct {
	mgr                  clustersmngr.ClustersManager
	log                  logr.Logger
	kinds                []schema.GroupVersionKind
	ticker               *time.Ticker
	quit                 chan bool
	msg                  chan []ObjectRecord
	additionalNamespaces []string
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

	cl, err := c.mgr.GetServerClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get server client: %w", err)
	}

	for _, clus := range clusters {
		clusterName := clus.GetName()

		namespaces := namespaces[clusterName]

		if len(c.additionalNamespaces) > 0 {
			for _, n := range c.additionalNamespaces {
				add := v1.Namespace{}
				add.Name = n
				namespaces = append(namespaces, add)
			}

		}

		for _, ns := range namespaces {
			for _, kind := range c.kinds {

				list := unstructured.UnstructuredList{}

				list.SetGroupVersionKind(kind)

				err := cl.List(ctx, clusterName, &list, client.InNamespace(ns.Name))

				if err != nil {
					if apierrors.IsNotFound(err) {
						// fmt.Printf("not found: %s/%s in %s/%s\n", kind.Group, kind.Kind, clusterName, ns.Name)
						continue
					}

					if apierrors.IsForbidden(err) {
						fmt.Printf("forbidden: %s/%s in %s\n", kind.Group, kind.Kind, clusterName)
						continue
					}

					c.log.Error(err, "failed to list objects", "cluster", clusterName, "namespace", ns.Name, "kind", kind.Kind)
					continue
				}

				if len(list.Items) > 0 {
					fmt.Printf("found %d %s/%s in %s/%s\n", len(list.Items), kind.Group, kind.Kind, clusterName, ns.Name)
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
