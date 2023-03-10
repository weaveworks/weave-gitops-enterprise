package cluster

import (
	"context"
	"fmt"
	"github.com/enekofb/collector/pkg/cluster/store"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

// Cluster watcher for watching flux applications kinds helm releases and kustomizations
type DefaultClustersWatcher struct {
	store           store.Store
	clusterWatchers map[string]Watcher
	newWatcherFunc  NewWatcherFunc
	kinds           []string
}

// Interface for watching clusters via kuberentes api
// https://kubernetes.io/docs/reference/using-api/api-concepts/#semantics-for-watch
type ClustersWatcher interface {
	Start(ctx context.Context, log logr.Logger) error
	Stop() error
	AddCluster(cluster types.NamespacedName, config *rest.Config, ctx context.Context, log logr.Logger) error
	RemoveCluster(cluster types.NamespacedName) error
	StatusCluster(cluster types.NamespacedName) (string, error)
}

// Function to create a watcher for a set of kinds. Operations target an store.
type NewWatcherFunc = func(config *rest.Config, cluster types.NamespacedName, store store.Store, kind []string, log logr.Logger) (Watcher, error)

func NewClustersWatcher(store store.Store, newWatcherFunc NewWatcherFunc, log logr.Logger) (*DefaultClustersWatcher, error) {
	if store == nil {
		return nil, fmt.Errorf("invalid store")
	}

	if newWatcherFunc == nil {
		newWatcherFunc = defaultNewWatcher
		log.V(2).Info("using default watcher function")
	}
	kinds := []string{
		v2beta1.HelmReleaseKind,
		v1beta2.KustomizationKind,
	}
	return &DefaultClustersWatcher{
		clusterWatchers: make(map[string]Watcher),
		newWatcherFunc:  newWatcherFunc,
		store:           store,
		kinds:           kinds,
	}, nil
}

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

func (w *DefaultClustersWatcher) Start(ctx context.Context, log logr.Logger) error {
	return fmt.Errorf("not implemented yet")
}

func (w *DefaultClustersWatcher) Stop() error {
	return fmt.Errorf("not implemented yet")
}

// TODO make me compatible with gitopscluster
func (w *DefaultClustersWatcher) AddCluster(cluster types.NamespacedName, config *rest.Config, ctx context.Context, log logr.Logger) error {
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

func (w *DefaultClustersWatcher) RemoveCluster(cluster types.NamespacedName) error {
	return fmt.Errorf("not yet implemented")
}

func (w *DefaultClustersWatcher) StatusCluster(cluster types.NamespacedName) (string, error) {
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
