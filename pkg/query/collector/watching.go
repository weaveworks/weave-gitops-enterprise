package collector

import (
	"context"
	"fmt"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

func (c *watchingCollector) Start(ctx context.Context) error {
	c.log.Info("starting collector")
	c.objectsChannel = make(chan []models.Object)

	for _, cluster := range c.clusters {
		clusterName := types.NamespacedName{
			Name:      cluster.GetName(),
			Namespace: "default",
		}
		c.log.Info("cluster adding", "name", clusterName.Name)
		config, err := cluster.GetServerConfig()
		if err != nil {
			return err
		}
		err = c.Watch(clusterName, config, c.objectsChannel, ctx, c.log)
		if err != nil {
			return err
		}
		c.log.Info("cluster added", "name", clusterName.Name)
	}

	c.log.Info("watchers created")

	go func() {
		c.log.Info("starting process to watch for records")
		for {
			select {
			case objects := <-c.objectsChannel:
				if err := c.store.StoreObjects(ctx, objects); err != nil {
					c.log.Error(err, "failed to store objects")
					continue
				}
			}
		}
	}()
	c.log.Info("collector started")
	return nil
}

func (c *watchingCollector) Stop(ctx context.Context) error {
	c.log.Info("stopping collector")

	for _, cluster := range c.clusters {
		clusterName := types.NamespacedName{
			Name:      cluster.GetName(),
			Namespace: "default",
		}
		c.log.Info("cluster stopping", "name", clusterName.Name)
		err := c.Unwatch(clusterName)
		if err != nil {
			return err
		}
		c.log.Info("cluster stopped", "name", clusterName.Name)
	}
	c.log.Info("collector stopped")
	return nil
}

// Interface for watching clusters via kuberentes api
// https://kubernetes.io/docs/reference/using-api/api-concepts/#semantics-for-watch
type ClusterWatcher interface {
	Watch(cluster types.NamespacedName, config *rest.Config, objectsChannel chan []models.Object, ctx context.Context, log logr.Logger) error
	Unwatch(cluster types.NamespacedName) error
	Status(cluster types.NamespacedName) (string, error)
}

// Cluster watcher for watching flux applications kinds helm releases and kustomizations
type watchingCollector struct {
	store           store.Store
	clusterWatchers map[string]Watcher
	newWatcherFunc  NewWatcherFunc
	kinds           []string
	msg             chan []models.Object
	log             logr.Logger
	clusters        []cluster.Cluster
	objectsChannel  chan []models.Object
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
		log:             log,
		clusters:        opts.Clusters,
	}, nil
}

// Function to create a watcher for a set of kinds. Operations target an store.
type NewWatcherFunc = func(config *rest.Config, cluster types.NamespacedName, objectsChannel chan []models.Object, kind []string, log logr.Logger) (Watcher, error)

// TODO add unit tests
func defaultNewWatcher(config *rest.Config, cluster types.NamespacedName, objectsChannel chan []models.Object, kinds []string, log logr.Logger) (Watcher, error) {
	if objectsChannel == nil {
		return nil, fmt.Errorf("invalid objects channel")
	}

	if len(kinds) == 0 {
		return nil, fmt.Errorf("at least one kind to watch is required")
	}

	w, err := NewWatcher(WatcherOptions{
		ClusterRef:   cluster,
		ClientConfig: config,
		Kinds:        kinds,
	}, nil, objectsChannel, log)

	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	return w, nil
}

func (w *watchingCollector) Watch(cluster types.NamespacedName, config *rest.Config, objectsChannel chan []models.Object, ctx context.Context, log logr.Logger) error {
	if config == nil {
		return fmt.Errorf("config not found")
	}

	if cluster.Name == "" || cluster.Namespace == "" {
		return fmt.Errorf("cluster name or namespace is empty")
	}

	log.Info("adding cluster", cluster)
	watcher, err := w.newWatcherFunc(config, cluster, w.objectsChannel, w.kinds, log)
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

func (w *watchingCollector) Unwatch(cluster types.NamespacedName) error {
	if cluster.Name == "" || cluster.Namespace == "" {
		return fmt.Errorf("cluster name or namespace is empty")
	}
	w.log.Info("stopping cluster", "cluster", cluster.String())
	clusterWatcher := w.clusterWatchers[cluster.String()]
	if clusterWatcher == nil {
		return fmt.Errorf("cluster watcher not found")
	}
	err := clusterWatcher.Stop()
	if err != nil {
		return err
	}
	w.log.Info("cluster stopped")
	return nil
}

func (w *watchingCollector) Status(cluster types.NamespacedName) (string, error) {
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
