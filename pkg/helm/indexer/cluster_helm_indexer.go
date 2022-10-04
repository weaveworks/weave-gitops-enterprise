package indexer

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/watcher"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/watcher/cache"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterHelmIndexerTracker struct {
	Cache           cache.Cache
	RepoManager     helm.RepoManager
	Namespace       string
	ClusterWatchers map[string]*watcher.Watcher
}

var scheme = runtime.NewScheme()

func NewClusterHelmIndexerTracker(c cache.Cache, rm helm.RepoManager, ns string, clientsPool clustersmngr.ClientsPool) *ClusterHelmIndexerTracker {
	return &ClusterHelmIndexerTracker{
		Cache:           c,
		RepoManager:     rm,
		Namespace:       ns,
		ClusterWatchers: make(map[string]*watcher.Watcher),
	}
}

func (i *ClusterHelmIndexerTracker) newIndexer(ctx context.Context, config *rest.Config, log logr.Logger) (*watcher.Watcher, error) {
	client, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}

	w, err := watcher.NewWatcher(watcher.Options{
		KubeClient: client,
		Cache:      i.Cache,
	})
	if err != nil {
		return nil, err
	}

	w.StartWatcher(log)

	return w, nil
}

// Start the indexer and wait for cluster updates notifications.
func (i *ClusterHelmIndexerTracker) Start(ctx context.Context, cw *clustersmngr.ClustersWatcher, log logr.Logger) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case updates := <-cw.Updates:
				for _, added := range updates.Added {
					_, ok := i.ClusterWatchers[added.Name]
					if !ok {
						clientConfig, err := clustersmngr.ClientConfigAsServer()(added)
						if err != nil {
							// FIXME: more info
							ctrl.Log.Error(err, "unable to build client config")
						}
						watcher, err := i.newIndexer(ctx, clientConfig, log)
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
