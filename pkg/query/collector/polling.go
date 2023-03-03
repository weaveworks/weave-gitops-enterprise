package collector

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

	for _, clus := range clusters {
		clusterName := clus.GetName()
		cl, err := clus.GetServerClient()
		if err != nil {
			c.log.Error(err, "failed to get client for cluster")
			continue
		}

		for _, kind := range c.kinds {
			list := &unstructured.UnstructuredList{}
			list.SetGroupVersionKind(kind)

			if err := cl.List(ctx, list); err != nil {
				c.log.Error(err, "failed to list objects")
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

	return result, nil
}
