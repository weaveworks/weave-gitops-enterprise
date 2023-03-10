package collector

import (
	"context"
	"fmt"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

// Interface for watching clusters via kuberentes api
// https://kubernetes.io/docs/reference/using-api/api-concepts/#semantics-for-watch
type ClustersWatcher interface {
	AddCluster(cluster types.NamespacedName, config *rest.Config, ctx context.Context, log logr.Logger) error
	RemoveCluster(cluster types.NamespacedName) error
	StatusCluster(cluster types.NamespacedName) (string, error)
}

// Cluster watcher for watching flux applications kinds helm releases and kustomizations
type watchingCollector struct {
	store           store.Store
	clusterWatchers map[string]Watcher
	newWatcherFunc  NewWatcherFunc
	kinds           []string
	msg             chan []ObjectRecord
}

// Collector factory method. It creates a collection with cluster watching strategy by default.
func newWatchingCollector(opts CollectorOpts, store store.Store, newWatcherFunc NewWatcherFunc) (*watchingCollector, error) {
	if opts.Log.GetSink() == nil {
		return &watchingCollector{}, fmt.Errorf("invalid log")
	}
	log := opts.Log

	if store == nil {
		return nil, fmt.Errorf("invalid store")
	}

	if newWatcherFunc == nil {
		newWatcherFunc = defaultNewWatcher
		log.Info("using default watcher function")
	}
	kinds := []string{
		v2beta1.HelmReleaseKind,
		v1beta2.KustomizationKind,
	}
	return &watchingCollector{
		clusterWatchers: make(map[string]Watcher),
		newWatcherFunc:  newWatcherFunc,
		store:           store,
		kinds:           kinds,
	}, nil
}

func (c *watchingCollector) Start() (<-chan []ObjectRecord, error) {
	return c.msg, fmt.Errorf("not implemented yet")
}

func (c *watchingCollector) Stop() error {
	return fmt.Errorf("not implemented yet")
}

// Function to create a watcher for a set of kinds. Operations target an store.
type NewWatcherFunc = func(config *rest.Config, cluster types.NamespacedName, store store.Store, kind []string, log logr.Logger) (Watcher, error)

// TODO add unit tests
func defaultNewWatcher(config *rest.Config, cluster types.NamespacedName, store store.Store, kinds []string, log logr.Logger) (Watcher, error) {
	if store == nil {
		return nil, fmt.Errorf("invalid store")
	}

	if len(kinds) == 0 {
		return nil, fmt.Errorf("at least one kind to watch is required")
	}

	w, err := NewWatcher(WatcherOptions{
		ClusterRef:   cluster,
		ClientConfig: config,
		Kinds:        kinds,
	}, nil, store, log)

	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	return w, nil
}

// TODO make me compatible with gitopscluster
func (w *watchingCollector) AddCluster(cluster types.NamespacedName, config *rest.Config, ctx context.Context, log logr.Logger) error {
	if config == nil {
		return fmt.Errorf("config not found")
	}

	if cluster.Name == "" || cluster.Namespace == "" {
		return fmt.Errorf("cluster name or namespace is empty")
	}

	log.Info("adding cluster", cluster)
	watcher, err := w.newWatcherFunc(config, cluster, w.store, w.kinds, log)
	if err != nil {
		return fmt.Errorf("failed to create watcher for cluster %s: %w", cluster.String(), err)
	}
	w.clusterWatchers[cluster.String()] = watcher
	log.Info("watcher created")
	go func() {
		err = watcher.Start(ctx, log)
		if err != nil {
			log.Error(err, "failed to start watcher", "cluster", cluster.Name)
		}
		log.Info("watcher started")
	}()
	return nil
}

func (w *watchingCollector) RemoveCluster(cluster types.NamespacedName) error {
	return fmt.Errorf("not yet implemented")
}

func (w *watchingCollector) StatusCluster(cluster types.NamespacedName) (string, error) {
	if cluster.Name == "" || cluster.Namespace == "" {
		return "", fmt.Errorf("cluster name or namespace is empty")
	}

	watcher := w.clusterWatchers[cluster.String()]
	if watcher == nil {
		return "", fmt.Errorf("cluster not found")
	}

	//TODO review whether we need a new layer here
	return watcher.Status()
}
