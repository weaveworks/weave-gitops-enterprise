package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/clusters"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
)

// Start the collector by creating watchers on existing gitops clusters and managing its lifecycle. Managing
// its lifecycle means responding to the events of adding a new cluster, update an existing cluster or deleting an existing cluster.
// Errors are handled by logging the error and assuming the operation will be retried due to some later event.
func (c *watchingCollector) Start(ctx context.Context) error {
	c.sub = c.subscriber.Subscribe()

	ratelimiter := workqueue.DefaultControllerRateLimiter() // TODO make a bespoke one, this retries too fast
	// TODO: check the queue methods work with Cluster, since we get new values each time.
	c.queue = workqueue.NewNamedRateLimitingQueue(ratelimiter, "collector-"+c.name)

	for _, cluster := range c.subscriber.GetClusters() {
		c.queue.Add(cluster)
	}

	// queue.Get() blocks, so we can't do that in a loop with
	// receiving on a channel. It goes in a goroutine, and we'll rely
	// on the queue being shutdown to make sure this exits.
	go func() {
		for {
			obj, shutdown := c.queue.Get()
			if shutdown {
				return
			}
			cluster, ok := obj.(cluster.Cluster)
			if !ok { // not a cluster; skip it
				c.queue.Forget(obj) // tell the rate limiter not to track it
				c.queue.Done(obj)   // dequeue it
				continue
			}
			c.queue.Done(obj)
			if err := c.watch(cluster); err != nil {
				c.log.Error(err, "cannot watch cluster", "cluster", cluster.GetName())
				c.queue.AddRateLimited(cluster)
				continue
			}
		}
	}()

outer:
	for {
		select {
		case <-ctx.Done():
			break outer
		case updates := <-c.sub.Updates():
			for _, cluster := range updates.Added {
				c.queue.Add(cluster)
			}

			// Do unwatches straight away; there's no reason to go
			// through the queue on these.
			for _, cluster := range updates.Removed {
				// remove from the queue (Done) and rate limiter
				// (Forget). If it comes back around, we'll start
				// afresh.
				c.queue.Done(cluster)
				c.queue.Forget(cluster)
				err := c.unwatch(cluster.GetName())
				if err != nil {
					c.log.Error(err, "cannot unwatch cluster", "cluster", cluster.GetName())
					continue
				}
				c.log.Info("unwatched cluster", "cluster", cluster.GetName())
			}
		}
	}

	c.log.Info("stopping collector")
	if c.sub != nil {
		c.sub.Unsubscribe()
	}
	c.queue.ShutDown()
	return nil
}

type child struct {
	Starter
	cancel           context.CancelFunc
	status           string
	lastStatusChange time.Time
}

func (c *child) setStatus(s string) {
	c.lastStatusChange = time.Now()
	c.status = s
}

// watchingCollector supervises watchers, starting one per cluster it
// sees from the `Subscriber` and stopping/restarting them as needed.
type watchingCollector struct {
	name              string
	sub               clusters.Subscription
	subscriber        clusters.Subscriber
	clusterWatchers   map[string]*child
	clusterWatchersMu sync.Mutex
	newWatcherFunc    NewWatcherFunc
	stopWatcherFunc   StopWatcherFunc
	queue             workqueue.RateLimitingInterface
	log               logr.Logger
	sa                ImpersonateServiceAccount
}

// Collector factory method. It creates a collection with clusterName watching strategy by default.
func newWatchingCollector(opts CollectorOpts) (*watchingCollector, error) {
	if opts.StopWatcherFunc == nil {
		opts.StopWatcherFunc = func(string) error {
			return nil
		}
	}
	return &watchingCollector{
		name:            opts.Name,
		subscriber:      opts.Clusters,
		clusterWatchers: make(map[string]*child),
		newWatcherFunc:  opts.NewWatcherFunc,
		stopWatcherFunc: opts.StopWatcherFunc,
		log:             opts.Log,
		sa:              opts.ServiceAccount,
	}, nil
}

func (w *watchingCollector) watch(cluster cluster.Cluster) (reterr error) {
	clusterName := cluster.GetName()
	if clusterName == "" {
		return fmt.Errorf("cluster name is empty")
	}

	// make the record, so status works
	c := &child{}
	c.setStatus(ClusterWatchingStarting)
	childctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	defer func() {
		if reterr != nil {
			w.clusterWatchersMu.Lock()
			c.setStatus(ClusterWatchingFailed)
			cancel := c.cancel
			w.clusterWatchersMu.Unlock()
			cancel()
		}
	}()

	w.clusterWatchersMu.Lock()
	w.clusterWatchers[clusterName] = c
	w.clusterWatchersMu.Unlock()

	config, err := cluster.GetServerConfig()
	if err != nil {
		return fmt.Errorf("cannot get config: %w", err)
	}

	if config == nil {
		return fmt.Errorf("cluster config cannot be nil")
	}

	saConfig, err := makeServiceAccountImpersonationConfig(config, w.sa.Namespace, w.sa.Name)
	if err != nil {
		return fmt.Errorf("cannot create impersonation config: %w", err)
	}

	watcher, err := w.newWatcherFunc(clusterName, saConfig)
	if err != nil {
		return fmt.Errorf("failed to create watcher for cluster %s: %w", cluster.GetName(), err)
	}

	go func() {
		w.clusterWatchersMu.Lock()
		c.setStatus(ClusterWatchingStarted)
		w.clusterWatchersMu.Unlock()
		err := watcher.Start(childctx)
		if err != nil {
			w.log.Error(err, "watcher for cluster failed", "cluster", cluster.GetName())
			w.clusterWatchersMu.Lock()
			c.setStatus(ClusterWatchingErrored)
			w.clusterWatchersMu.Unlock()
			// try again
			w.queue.AddRateLimited(cluster)
			return
		}
		// TODO remove from map?
		w.clusterWatchersMu.Lock()
		c.setStatus(ClusterWatchingStopped)
		w.clusterWatchersMu.Unlock()
	}()

	w.log.Info("watching cluster", "cluster", cluster.GetName())
	return nil
}

func (w *watchingCollector) unwatch(clusterName string) error {
	if clusterName == "" {
		return fmt.Errorf("cluster name is empty")
	}
	w.clusterWatchersMu.Lock()
	clusterWatcher := w.clusterWatchers[clusterName]
	delete(w.clusterWatchers, clusterName)
	w.clusterWatchersMu.Unlock()

	if clusterWatcher == nil {
		return fmt.Errorf("cluster watcher not found")
	}
	if clusterWatcher.cancel != nil {
		clusterWatcher.cancel()
	}
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
	w.clusterWatchersMu.Lock()
	watcher := w.clusterWatchers[clusterName]
	w.clusterWatchersMu.Unlock()
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
