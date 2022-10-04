package indexer

import (
	"context"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/watcher/cache"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/watcher/controller"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

type ClusterHelmIndexerTracker struct {
	Cache       cache.Cache
	RepoManager helm.RepoManager
	Namespace   string
	// ClusterWatchers map[string]ctrl.Manager
	ClusterWatchers map[string]context.CancelFunc
	clientsPool     clustersmngr.ClientsPool
}

var scheme = runtime.NewScheme()

func NewClusterHelmIndexerTracker(c cache.Cache, rm helm.RepoManager, ns string, clientsPool clustersmngr.ClientsPool) *ClusterHelmIndexerTracker {
	return &ClusterHelmIndexerTracker{
		Cache:           c,
		RepoManager:     rm,
		Namespace:       ns,
		ClusterWatchers: make(map[string]context.CancelFunc),
		clientsPool:     clientsPool,
	}
}

func (i *ClusterHelmIndexerTracker) newIndexer(ctx context.Context, clusterName string) (ctrl.Manager, error) {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		// MetricsBindAddress:     w.metricsBindAddress,
		// HealthProbeBindAddress: w.healthzBindAddress,
		// Port:   w.watcherPort,
		Logger: ctrl.Log,
	})
	if err != nil {
		ctrl.Log.Error(err, "unable to create manager")
		return nil, err
	}

	clientForCluster, err := i.clientsPool.Client(clusterName)
	if err != nil {
		return nil, err
	}

	helmWatcher := &controller.HelmWatcherReconciler{
		Client:      clientForCluster,
		Cache:       i.Cache,
		RepoManager: &i.RepoManager,
		Scheme:      scheme,
		// ExternalEventRecorder: eventRecorder,
	}
	if err = helmWatcher.SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create controller", "controller", "HelmWatcherReconciler")
		return nil, err
	}

	if err := mgr.Start(ctx); err != nil {
		return nil, err
	}

	return mgr, nil
}

// Start the indexer and wait for cluster updates notifications.
func (i *ClusterHelmIndexerTracker) Start(ctx context.Context, cw *clustersmngr.ClustersWatcher) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case updates := <-cw.Updates:
				for _, added := range updates.Added {
					_, ok := i.ClusterWatchers[added.Name]
					if !ok {
						ctx, cancel := context.WithCancel(ctrl.SetupSignalHandler())
						_, err := i.newIndexer(ctx, added.Name)
						if err != nil {
							// FIXME: more info
							ctrl.Log.Error(err, "unable to create indexer")
						}
						i.ClusterWatchers[added.Name] = cancel
					}
				}

				for _, removed := range updates.Removed {
					cancel, ok := i.ClusterWatchers[removed.Name]
					if ok {
						cancel()
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
