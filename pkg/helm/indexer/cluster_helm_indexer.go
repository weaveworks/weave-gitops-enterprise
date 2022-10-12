package indexer

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/watcher"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/watcher/cache"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterHelmIndexerTracker struct {
	Cache           cache.Cache
	ClusterWatchers map[string]*watcher.Watcher
}

var scheme = runtime.NewScheme()

func NewClusterHelmIndexerTracker(c cache.Cache) *ClusterHelmIndexerTracker {
	return &ClusterHelmIndexerTracker{
		Cache:           c,
		ClusterWatchers: make(map[string]*watcher.Watcher),
	}
}

func (i *ClusterHelmIndexerTracker) newIndexer(ctx context.Context, config *rest.Config, cluster types.NamespacedName, log logr.Logger) (*watcher.Watcher, error) {
	client, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}

	w, err := watcher.NewWatcher(watcher.Options{
		Cluster:    cluster,
		KubeClient: client,
		Cache:      i.Cache,
	})
	if err != nil {
		return nil, err
	}

	w.StartWatcher(log)

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
