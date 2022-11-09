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

// ClusterHelmIndexerTracker tracks the helm indexers for each cluster.
// We subscribe to cluster updates and start/stop indexers as needed.
type ClusterHelmIndexerTracker struct {
	Cache                 helm.ChartsCacherWriter
	ManagementClusterName string
	ClusterWatchers       map[string]Watcher
	newWatcherFunc        NewWatcherFunc
}

type Watcher interface {
	StartWatcher(ctx context.Context, log logr.Logger) error
	Stop()
}

type NewWatcherFunc = func(config *rest.Config, cluster types.NamespacedName, isManagementCluster bool, cache helm.ChartsCacherWriter) (Watcher, error)

// NewClusterHelmIndexerTracker creates a new ClusterHelmIndexerTracker.
// Pass in the management cluster name so we can determine if we need to use the proxy.
func NewClusterHelmIndexerTracker(c helm.ChartsCacherWriter, managementClusterName string, newWatcherFunc NewWatcherFunc) *ClusterHelmIndexerTracker {
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
				watcher, ok := i.ClusterWatchers[removed.Name]
				if ok {
					watcher.Stop()
					// remove the watcher from the map
					delete(i.ClusterWatchers, removed.Name)
					err := i.Cache.DeleteAllChartsForCluster(ctx, fetcher.FromClusterName(removed.Name))
					if err != nil {
						log.Error(err, "unable to delete charts for cluster")
					}
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

			cluster := fetcher.FromClusterName(cl.Name)
			isManagementCluster := fetcher.IsManagementCluster(i.ManagementClusterName, cluster)
			watcher, err := i.newWatcherFunc(clientConfig, cluster, isManagementCluster, i.Cache)
			if err != nil {
				return fmt.Errorf("failed to create indexer for cluster %s: %w", cl.Name, err)
			}
			i.ClusterWatchers[cl.Name] = watcher

			go func() {
				err = watcher.StartWatcher(ctx, log)
				if err != nil {
					log.Error(err, "failed to start indexer", "cluster", cl.Name)
				}
			}()
		} else {
			log.Info("indexer already exists for cluster", "cluster", cl.Name)
		}
	}

	return nil
}

func NewIndexer(config *rest.Config, cluster types.NamespacedName, isManagementCluster bool, cache helm.ChartsCacherWriter) (Watcher, error) {
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
