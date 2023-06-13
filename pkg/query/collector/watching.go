package collector

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/clusters"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"k8s.io/client-go/rest"
)

// Start the collector by creating watchers on existing gitops clusters and managing its lifecycle. Managing
// its lifecycle means responding to the events of adding a new cluster, update an existing cluster or deleting an existing cluster.
// Errors are handled by logging the error and assuming the operation will be retried due to some later event.
func (c *watchingCollector) Start() error {
	c.quit = make(chan struct{})
	c.sub = c.subscriber.Subscribe()

	for _, cluster := range c.subscriber.GetClusters() {
		err := c.watch(cluster)
		if err != nil {
			c.log.Error(err, "cannot watch cluster", "cluster", cluster.GetName())
			continue
		}
		c.log.Info("watching cluster", "cluster", cluster.GetName())
	}

	// watch clusters
	c.done.Add(1)
	go func() {
		defer c.done.Done()
		for {
			select {
			case <-c.quit:
				return
			case updates := <-c.sub.Updates():
				for _, cluster := range updates.Added {
					err := c.watch(cluster)
					if err != nil {
						c.log.Error(err, "cannot watch cluster", "cluster", cluster.GetName())
						continue
					}
					c.log.Info("watching cluster", "cluster", cluster.GetName())
				}

				for _, cluster := range updates.Removed {
					err := c.unwatch(cluster.GetName())
					if err != nil {
						c.log.Error(err, "cannot unwatch cluster", "cluster", cluster.GetName())
						continue
					}
					c.log.Info("unwatched cluster", "cluster", cluster.GetName())
				}
			}
		}
	}()

	return nil
}

// Stop the collector and clean up.
func (c *watchingCollector) Stop() error {
	c.log.Info("stopping collector")
	if c.sub != nil {
		c.sub.Unsubscribe()
	}
	if c.quit != nil {
		close(c.quit)
	}
	c.done.Wait()
	return nil
}

type child struct {
	Starter
	cancel context.CancelFunc
	status string
}

// watchingCollector supervises watchers, starting one per cluster it
// sees from the `Subscriber` and stopping/restarting them as needed.
type watchingCollector struct {
	quit            chan struct{}
	done            sync.WaitGroup
	sub             clusters.Subscription
	subscriber      clusters.Subscriber
	clusterWatchers map[string]*child
	newWatcherFunc  NewWatcherFunc
	stopWatcherFunc StopWatcherFunc
	log             logr.Logger
	sa              ImpersonateServiceAccount
}

// Collector factory method. It creates a collection with clusterName watching strategy by default.
func newWatchingCollector(opts CollectorOpts) (*watchingCollector, error) {
	if opts.StopWatcherFunc == nil {
		opts.StopWatcherFunc = func(string) error {
			return nil
		}
	}
	return &watchingCollector{
		subscriber:      opts.Clusters,
		clusterWatchers: make(map[string]*child),
		newWatcherFunc:  opts.NewWatcherFunc,
		stopWatcherFunc: opts.StopWatcherFunc,
		log:             opts.Log,
		sa:              opts.ServiceAccount,
	}, nil
}

func (w *watchingCollector) watch(cluster cluster.Cluster) error {
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

	saConfig, err := makeServiceAccountImpersonationConfig(config, w.sa.Namespace, w.sa.Name)
	if err != nil {
		return fmt.Errorf("cannot create impersonation config: %w", err)
	}

	watcher, err := w.newWatcherFunc(clusterName, saConfig)
	if err != nil {
		return fmt.Errorf("failed to create watcher for cluster %s: %w", cluster.GetName(), err)
	}

	childctx, cancel := context.WithCancel(context.Background())
	c := &child{
		status:  ClusterWatchingStarted,
		Starter: watcher,
		cancel:  cancel,
	}
	w.clusterWatchers[clusterName] = c
	go func() {
		err := watcher.Start(childctx)
		if err != nil {
			w.log.Error(err, "watcher for cluster failed", "cluster", cluster.GetName())
			c.status = ClusterWatchingFailed
			return
		}
		// TODO remove from map?
		c.status = ClusterWatchingStopped
	}()

	return nil
}

func (w *watchingCollector) unwatch(clusterName string) error {
	if clusterName == "" {
		return fmt.Errorf("cluster name is empty")
	}
	clusterWatcher := w.clusterWatchers[clusterName]
	if clusterWatcher == nil {
		return fmt.Errorf("cluster watcher not found")
	}
	w.clusterWatchers[clusterName] = nil
	clusterWatcher.cancel()
	if err := w.stopWatcherFunc(clusterName); err != nil {
		return fmt.Errorf("stop watcher hook failed: %w", err)
	}
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
	return watcher.status, nil
}

// makeServiceAccountImpersonationConfig when creating a reconciler for watcher we will need to impersonate
// a user to dont use the default one to enhance security. This method creates a new rest.config from the input parameters
// with impersonation configuration pointing to the service account
func makeServiceAccountImpersonationConfig(config *rest.Config, namespace, serviceAccountName string) (*rest.Config, error) {
	if config == nil {
		return nil, fmt.Errorf("invalid rest config")
	}

	if namespace == "" || serviceAccountName == "" {
		return nil, fmt.Errorf("service account cannot be empty")
	}

	copyCfg := rest.CopyConfig(config)
	copyCfg.Impersonate = rest.ImpersonationConfig{
		UserName: fmt.Sprintf("system:serviceaccount:%s:%s", namespace, serviceAccountName),
	}

	return copyCfg, nil
}
