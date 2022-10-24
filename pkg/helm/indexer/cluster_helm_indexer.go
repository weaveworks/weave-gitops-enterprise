package indexer

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/multiwatcher"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type ClusterHelmIndexerTracker struct {
	Cache           helm.ChartsCacherWriter
	ClusterWatchers map[string]*multiwatcher.Watcher
}

func NewClusterHelmIndexerTracker(c helm.ChartsCacherWriter) *ClusterHelmIndexerTracker {
	return &ClusterHelmIndexerTracker{
		Cache:           c,
		ClusterWatchers: make(map[string]*multiwatcher.Watcher),
	}
}

func (i *ClusterHelmIndexerTracker) newIndexer(ctx context.Context, config *rest.Config, cluster types.NamespacedName, log logr.Logger) (*multiwatcher.Watcher, error) {
	w, err := multiwatcher.NewWatcher(multiwatcher.Options{
		ClusterRef:    cluster,
		ClientConfig:  config,
		Cache:         i.Cache,
		ValuesFetcher: helm.NewValuesFetcher(),
	})
	if err != nil {
		return nil, err
	}

	w.StartWatcher(ctx, log)

	return w, nil
}

func toNamespaceName(name string) types.NamespacedName {
	// use strings.Split to split the name into two parts
	// the first part is the namespace and the second part is the name
	// if the name does not contain a slash, then the namespace is empty

	parts := strings.Split(name, "/")
	if len(parts) == 1 {
		return types.NamespacedName{
			Name:      parts[0],
			Namespace: "default",
		}
	}

	return types.NamespacedName{
		Namespace: parts[0],
		Name:      parts[1],
	}
}

// Start the indexer and wait for cluster updates notifications.
func (i *ClusterHelmIndexerTracker) Start(ctx context.Context, cw *clustersmngr.ClustersWatcher, log logr.Logger) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case updates := <-cw.Updates:
				ctrl.Log.Info("adding indexer for cluster", "update", updates)
				for _, added := range updates.Added {
					_, ok := i.ClusterWatchers[added.Name]
					if !ok {
						ctrl.Log.Info("adding indexer for cluster", "cluster", added.Name)
						clientConfig, err := clustersmngr.ClientConfigAsServer()(added)
						if err != nil {
							// FIXME: more info
							ctrl.Log.Error(err, "unable to build client config")
						}
						watcher, err := i.newIndexer(ctx, clientConfig, toNamespaceName(added.Name), log)
						if err != nil {
							// FIXME: more info
							ctrl.Log.Error(err, "unable to create indexer")
						}
						i.ClusterWatchers[added.Name] = watcher
					}
				}

				for _, removed := range updates.Removed {
					watcher, ok := i.ClusterWatchers[removed.Name]
					if ok {
						watcher.Stop()
						// TODO
						// Remove all the helm releases from the cache
						// cache.DeleteCluster(types.NamespacedName{Name: removed.Name, Namespace: i.Namespace})
					}
				}
			}
		}
	}()

	return nil
}
