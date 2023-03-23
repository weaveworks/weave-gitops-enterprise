package collector

import (
	"context"
	"fmt"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/models"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

func (c *watchingCollector) Start() error {
	//TODO add context
	ctx := context.Background()
	cw := c.clusterManager.Subscribe()
	c.objectsChannel = make(chan []models.ObjectTransaction)

	for _, cluster := range c.clusterManager.GetClusters() {
		err := c.Watch(cluster, c.objectsChannel, context.Background(), c.log)
		if err != nil {
			return fmt.Errorf("cannot watch clusterName: %w", err)
		}
	}

	//watch on clusters
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case updates := <-cw.Updates:

				for _, cluster := range updates.Added {
					err := c.Watch(cluster, c.objectsChannel, context.Background(), c.log)
					if err != nil {
						c.log.Error(err, "cannot watch cluster")
					}
				}

				for _, cluster := range updates.Removed {
					err := c.Unwatch(cluster)
					if err != nil {
						c.log.Error(err, "cannot unwatch cluster")
					}
				}

			}
		}
	}()
	//watch on channels
	go func() {
		for {
			objectTransactions := <-c.objectsChannel
			err := c.processRecordsFunc(ctx, objectTransactions, c.store, c.log)
			if err != nil {
				c.log.Error(err, "cannot process records")
			}
		}
	}()

	return nil
}

func (c *watchingCollector) Stop() error {
	c.log.Info("stopping collector")

	for _, cluster := range c.clusters {
		err := c.Unwatch(cluster)
		if err != nil {
			return err
		}

	}

	return nil
}

// Cluster watcher for watching flux applications kinds helm releases and kustomizations
type watchingCollector struct {
	clusters           []cluster.Cluster
	clusterManager     clustersmngr.ClustersManager
	clusterWatchers    map[string]Watcher
	kinds              []schema.GroupVersionKind
	store              store.StoreWriter
	objectsChannel     chan []models.ObjectTransaction
	newWatcherFunc     NewWatcherFunc
	log                logr.Logger
	processRecordsFunc ProcessRecordsFunc
}

// Collector factory method. It creates a collection with clusterName watching strategy by default.
func newWatchingCollector(opts CollectorOpts, store store.StoreWriter) (*watchingCollector, error) {
	if opts.NewWatcherFunc == nil {
		opts.NewWatcherFunc = defaultNewWatcher
	}

	return &watchingCollector{
		clusters:           opts.Clusters,
		clusterManager:     opts.ClusterManager,
		clusterWatchers:    make(map[string]Watcher),
		newWatcherFunc:     opts.NewWatcherFunc,
		store:              store,
		kinds:              opts.ObjectKinds,
		log:                opts.Log,
		processRecordsFunc: opts.ProcessRecordsFunc,
	}, nil
}

// Function to create a watcher for a set of kinds. Operations target an store.
type NewWatcherFunc = func(config *rest.Config, clusterName string, objectsChannel chan []models.ObjectTransaction, kinds []schema.GroupVersionKind, log logr.Logger) (Watcher, error)

type ProcessRecordsFunc = func(ctx context.Context, objectRecords []models.ObjectTransaction, store store.StoreWriter, log logr.Logger) error

// TODO add unit tests
func defaultNewWatcher(config *rest.Config, clusterName string, objectsChannel chan []models.ObjectTransaction, kinds []schema.GroupVersionKind, log logr.Logger) (Watcher, error) {
	w, err := NewWatcher(WatcherOptions{
		ClusterRef: types.NamespacedName{
			Name:      clusterName,
			Namespace: "default",
		},
		ClientConfig:  config,
		Kinds:         kinds,
		ObjectChannel: objectsChannel,
		Log:           log,
		ManagerFunc:   defaultNewWatcherManager,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	return w, nil
}

func (w *watchingCollector) Watch(cluster cluster.Cluster, objectsChannel chan []models.ObjectTransaction, ctx context.Context, log logr.Logger) error {
	config, err := cluster.GetServerConfig()
	if err != nil {
		return fmt.Errorf("cannot get config: %w", err)
	}

	clusterName := cluster.GetName()
	if clusterName == "" {
		return fmt.Errorf("cluster name is empty")
	}

	watcher, err := w.newWatcherFunc(config, clusterName, w.objectsChannel, w.kinds, log)
	if err != nil {
		return fmt.Errorf("failed to create watcher for clusterName %s: %w", cluster.GetName(), err)
	}
	//TODO it might not be enough to avoid clashes
	w.clusterWatchers[cluster.GetName()] = watcher

	go func() {
		err = watcher.Start(ctx, log)
		if err != nil {
			log.Error(err, "failed to start watcher", "cluster", cluster.GetName())
		}
	}()

	return nil
}

func (w *watchingCollector) Unwatch(cluster cluster.Cluster) error {
	if cluster == nil {
		return fmt.Errorf("invalid clusterName")
	}
	if cluster.GetName() == "" {
		return fmt.Errorf("cluster name is empty")
	}

	clusterWatcher := w.clusterWatchers[cluster.GetName()]
	if clusterWatcher == nil {
		return fmt.Errorf("cluster watcher not found")
	}
	err := clusterWatcher.Stop()
	if err != nil {
		return err
	}
	w.clusterWatchers[cluster.GetName()] = nil

	return nil
}

func (w *watchingCollector) Status(cluster cluster.Cluster) (string, error) {
	if cluster == nil {
		return "", fmt.Errorf("invalid clusterName")
	}
	if cluster.GetName() == "" {
		return "", fmt.Errorf("cluster name is empty")
	}

	watcher := w.clusterWatchers[cluster.GetName()]
	if watcher == nil {
		return "", fmt.Errorf("clusterName not found: %s", cluster.GetName())
	}
	//TODO review whether we need a new layer here
	return watcher.Status()
}
