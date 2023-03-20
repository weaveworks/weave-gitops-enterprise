package indexer

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/fetcher"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/multiwatcher"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

// ClusterHelmIndexerTracker tracks the helm indexers for each cluster.
// We subscribe to cluster updates and start/stop indexers as needed.
type ClusterHelmIndexerTracker struct {
	Cache                 helm.ChartsCacheWriter
	ManagementClusterName string
	ClusterWatchers       map[string]Watcher
	newWatcherFunc        NewWatcherFunc
}

type Watcher interface {
	StartWatcher(ctx context.Context, log logr.Logger) error
	Stop()
}

type NewWatcherFunc = func(config *rest.Config, cluster types.NamespacedName, isManagementCluster bool, cache helm.ChartsCacheWriter) (Watcher, error)

// NewClusterHelmIndexerTracker creates a new ClusterHelmIndexerTracker.
// Pass in the management cluster name so we can determine if we need to use the proxy.
func NewClusterHelmIndexerTracker(c helm.ChartsCacheWriter, managementClusterName string, newWatcherFunc NewWatcherFunc) *ClusterHelmIndexerTracker {
	return &ClusterHelmIndexerTracker{
		Cache:                 c,
		ManagementClusterName: managementClusterName,
		ClusterWatchers:       make(map[string]Watcher),
		newWatcherFunc:        newWatcherFunc,
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
				log.Error(err, "failed to add clusters")
			}

			for _, removed := range updates.Removed {
				watcher, ok := i.ClusterWatchers[removed.GetName()]
				if ok {
					watcher.Stop()
					// remove the watcher from the map
					delete(i.ClusterWatchers, removed.GetName())
					err := i.Cache.DeleteAllChartsForCluster(ctx, fetcher.FromClusterName(removed.GetName()))
					if err != nil {
						log.Error(err, "unable to delete charts for cluster")
					}
				} else {
					log.Info("cluster not found in indexer", "cluster", removed.GetName())
				}
			}
		}
	}
}

func (i *ClusterHelmIndexerTracker) addClusters(ctx context.Context, clusters []cluster.Cluster, log logr.Logger) error {
	for _, cl := range clusters {
		clusterName := cl.GetName()
		_, ok := i.ClusterWatchers[clusterName]
		if !ok {
			log.Info("adding indexer for cluster", "cluster", clusterName)
			clientConfig, err := cl.GetServerConfig()
			if err != nil {
				return fmt.Errorf("failed to get client config for cluster %s: %w", clusterName, err)
			}

			cluster := fetcher.FromClusterName(clusterName)
			isManagementCluster := fetcher.IsManagementCluster(i.ManagementClusterName, cluster)
			watcher, err := i.newWatcherFunc(clientConfig, cluster, isManagementCluster, i.Cache)
			if err != nil {
				return fmt.Errorf("failed to create indexer for cluster %s: %w", clusterName, err)
			}
			i.ClusterWatchers[clusterName] = watcher

			go func() {
				err = watcher.StartWatcher(ctx, log)
				if err != nil {
					log.Error(err, "failed to start indexer", "cluster", clusterName)
				}
			}()
		} else {
			log.Info("indexer already exists for cluster", "cluster", clusterName)
		}
	}

	return nil
}

func NewIndexer(config *rest.Config, cluster types.NamespacedName, isManagementCluster bool, cache helm.ChartsCacheWriter) (Watcher, error) {
	w, err := multiwatcher.NewWatcher(multiwatcher.Options{
		ClusterRef:    cluster,
		ClientConfig:  config,
		Cache:         cache,
		UseProxy:      !isManagementCluster,
		ValuesFetcher: helm.NewValuesFetcher(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}

	return w, nil
}
