package indexer

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/fetcher"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/multiwatcher"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

type ClusterHelmIndexerTracker struct {
	Cache                 helm.ChartsCacherWriter
	ManagementClusterName string
	ClusterWatchers       map[string]*multiwatcher.Watcher
}

func NewClusterHelmIndexerTracker(c helm.ChartsCacherWriter, managementClusterName string) *ClusterHelmIndexerTracker {
	return &ClusterHelmIndexerTracker{
		Cache:                 c,
		ManagementClusterName: managementClusterName,
		ClusterWatchers:       make(map[string]*multiwatcher.Watcher),
	}
}

// Start the indexer and wait for cluster updates notifications.
func (i *ClusterHelmIndexerTracker) Start(ctx context.Context, cm clustersmngr.ClustersManager, log logr.Logger) error {

	cw := cm.Subscribe()

	err := i.addClusters(ctx, cm.GetClusters(), log)
	if err != nil {
		return fmt.Errorf("failed to add clusters: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case updates := <-cw.Updates:
			err := i.addClusters(ctx, updates.Added, log)
			if err != nil {
				log.Error(err, "unable to create indexer")
			}

			for _, removed := range updates.Removed {
				watcher, ok := i.ClusterWatchers[removed.Name]
				if ok {
					watcher.Stop()
					// TODO
					// Remove all the helm releases from the cache
					// cache.DeleteCluster(types.NamespacedName{Name: removed.Name, Namespace: i.Namespace})
				} else {
					log.Info("cluster not found in indexer", "cluster", removed.Name)
				}
			}
		}
	}
}

func (i *ClusterHelmIndexerTracker) addClusters(ctx context.Context, clusters []clustersmngr.Cluster, log logr.Logger) error {
	for _, cl := range clusters {
		_, ok := i.ClusterWatchers[cl.Name]
		if !ok {
			log.Info("adding indexer for cluster", "cluster", cl.Name)
			clientConfig, err := clustersmngr.ClientConfigAsServer()(cl)
			if err != nil {
				return fmt.Errorf("failed to get client config for cluster %s: %w", cl.Name, err)
			}
			watcher, err := i.newIndexer(ctx, clientConfig, fetcher.FromClusterName(cl.Name), log)
			if err != nil {
				return fmt.Errorf("failed to create indexer for cluster %s: %w", cl.Name, err)
			}
			i.ClusterWatchers[cl.Name] = watcher

			go func() {
				err = watcher.StartWatcher(ctx, log)
				if err != nil {
					log.Error(err, "failed to start indexer")
				}
			}()
		} else {
			log.Info("indexer already exists for cluster", "cluster", cl.Name)
		}
	}

	return nil
}

func (i *ClusterHelmIndexerTracker) newIndexer(ctx context.Context, config *rest.Config, cluster types.NamespacedName, log logr.Logger) (*multiwatcher.Watcher, error) {
	w, err := multiwatcher.NewWatcher(multiwatcher.Options{
		ClusterRef:    cluster,
		ClientConfig:  config,
		Cache:         i.Cache,
		UseProxy:      !fetcher.IsManagementCluster(i.ManagementClusterName, cluster),
		ValuesFetcher: helm.NewValuesFetcher(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}

	return w, nil
}
