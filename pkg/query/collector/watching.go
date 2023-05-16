package collector

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

// Start the collector by creating watchers on existing gitops clusters and managing its lifecycle. Managing
// its lifecycle means responding to the events of adding a new cluster, update an existing cluster or deleting an existing cluster.
// Errors are handled by logging the error and assuming the operation will be retried due to some later event.
func (c *watchingCollector) Start() error {
	cw := c.clusterManager.Subscribe()
	c.objectsChannel = make(chan []models.ObjectTransaction)

	for _, cluster := range c.clusterManager.GetClusters() {
		err := c.Watch(cluster)
		if err != nil {
			c.log.Error(err, "cannot watch cluster", "cluster", cluster.GetName())
			continue
		}
		c.log.Info("watching cluster", "cluster", cluster.GetName())
	}

	//watch clusters
	go func() {
		for updates := range cw.Updates {
			for _, cluster := range updates.Added {
				err := c.Watch(cluster)
				if err != nil {
					c.log.Error(err, "cannot watch cluster", "cluster", cluster.GetName())
					continue
				}
				c.log.Info("watching cluster", "cluster", cluster.GetName())
			}

			for _, cluster := range updates.Removed {
				err := c.Unwatch(cluster.GetName())
				if err != nil {
					c.log.Error(err, "cannot unwatch cluster", "cluster", cluster.GetName())
					continue
				}
				c.log.Info("unwatched cluster", "cluster", cluster.GetName())
			}
		}
	}()

	//watch object events
	go func() {
		for objectTransactions := range c.objectsChannel {
			err := c.processRecordsFunc(objectTransactions, c.store, c.log)
			if err != nil {
				c.log.Error(err, "cannot process records")
			}
		}
	}()

	c.log.Info("watcher started", "kinds", c.kinds)
	return nil
}

// TODO this does nothing?
func (c *watchingCollector) Stop() error {
	c.log.Info("stopping collector")
	return nil
}

// Cluster watcher for watching flux applications kinds helm releases and kustomizations
type watchingCollector struct {
	clusterManager     clustersmngr.ClustersManager
	clusterWatchers    map[string]Watcher
	kinds              []configuration.ObjectKind
	store              store.Store
	objectsChannel     chan []models.ObjectTransaction
	newWatcherFunc     NewWatcherFunc
	log                logr.Logger
	processRecordsFunc ProcessRecordsFunc
	serviceAccount     ImpersonateServiceAccount
}

// Collector factory method. It creates a collection with clusterName watching strategy by default.
func newWatchingCollector(opts CollectorOpts, store store.Store) (*watchingCollector, error) {
	if opts.NewWatcherFunc == nil {
		opts.NewWatcherFunc = defaultNewWatcher
	}

	return &watchingCollector{
		clusterManager:     opts.ClusterManager,
		clusterWatchers:    make(map[string]Watcher),
		newWatcherFunc:     opts.NewWatcherFunc,
		store:              store,
		kinds:              opts.ObjectKinds,
		log:                opts.Log,
		processRecordsFunc: opts.ProcessRecordsFunc,
		serviceAccount:     opts.ServiceAccount,
	}, nil
}

// Function to create a watcher for a set of kinds. Operations target an store.
type NewWatcherFunc = func(config *rest.Config, serviceAccount ImpersonateServiceAccount, clusterName string, objectsChannel chan []models.ObjectTransaction, kinds []configuration.ObjectKind, log logr.Logger) (Watcher, error)

type ProcessRecordsFunc = func(objectRecords []models.ObjectTransaction, store store.Store, log logr.Logger) error

// TODO add unit tests
func defaultNewWatcher(config *rest.Config, serviceAccount ImpersonateServiceAccount, clusterName string, objectsChannel chan []models.ObjectTransaction,
	kinds []configuration.ObjectKind, log logr.Logger) (Watcher, error) {
	w, err := NewWatcher(WatcherOptions{
		ClusterRef: types.NamespacedName{
			Name:      clusterName,
			Namespace: "default",
		},
		ClientConfig:   config,
		Kinds:          kinds,
		ObjectChannel:  objectsChannel,
		Log:            log,
		ManagerFunc:    defaultNewWatcherManager,
		ServiceAccount: serviceAccount,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	return w, nil
}

func (w *watchingCollector) Watch(cluster cluster.Cluster) error {
	config, err := cluster.GetServerConfig()
	if err != nil {
		return fmt.Errorf("cannot get config: %w", err)
	}

	if config == nil {
		return fmt.Errorf("cluster config cannot be nil")
	}

	clusterName := cluster.GetName()
	if clusterName == "" {
		return fmt.Errorf("cluster name is empty")
	}

	watcher, err := w.newWatcherFunc(config, w.serviceAccount, clusterName, w.objectsChannel, w.kinds, w.log)
	if err != nil {
		return fmt.Errorf("failed to create watcher for cluster %s: %w", cluster.GetName(), err)
	}
	w.clusterWatchers[clusterName] = watcher
	err = watcher.Start()
	if err != nil {
		return fmt.Errorf("failed to start watcher for cluster %s: %w", cluster.GetName(), err)
	}

	return nil
}

func (w *watchingCollector) Unwatch(clusterName string) error {
	if clusterName == "" {
		return fmt.Errorf("cluster name is empty")
	}
	clusterWatcher := w.clusterWatchers[clusterName]
	if clusterWatcher == nil {
		return fmt.Errorf("cluster watcher not found")
	}
	err := clusterWatcher.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop watcher for cluster %s: %w", clusterName, err)
	}
	w.clusterWatchers[clusterName] = nil
	return nil
}

// Status returns a cluster watcher status for the cluster named as clusterName.
// It returns an error if empty, cluster does not exist or the status cannot be retrieved.
func (w *watchingCollector) Status(clusterName string) (string, error) {
	if clusterName == "" {
		return "", fmt.Errorf("cluster name is empty")
	}
	watcher := w.clusterWatchers[clusterName]
	if watcher == nil {
		return "", fmt.Errorf("cluster not found: %s", clusterName)
	}
	return watcher.Status()
}
