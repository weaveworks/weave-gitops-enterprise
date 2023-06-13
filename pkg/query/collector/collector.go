package collector

import (
	"fmt"

	"github.com/go-logr/logr"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/clusters"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
)

const (
	ClusterWatchingStarted = "started"
	ClusterWatchingStopped = "stopped"
	ClusterWatchingFailed  = "failed"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// ClusterWatcher is for managing the lifecycle of watchers.
type ClusterWatcher interface {
	// Watch starts watching the cluster passed as input
	Watch(cluster cluster.Cluster) error
	// Unwatch stops watching the cluster identified by input clusterName.
	Unwatch(clusterName string) error
	// Status return the watcher status for the cluster identified as clusterName.
	Status(clusterName string) (string, error)
}

//counterfeiter:generate . Collector

// Collector is ClusterWatcher that has its own lifecycle (i.e.,
// Start() and Stop()).
type Collector interface {
	ClusterWatcher
	Start() error
	Stop() error
}

type ImpersonateServiceAccount struct {
	Name      string
	Namespace string
}

type CollectorOpts struct {
	Log            logr.Logger
	Clusters       clusters.Subscriber
	NewWatcherFunc NewWatcherFunc
	ServiceAccount ImpersonateServiceAccount // this gives the service account to impersonate when watching each cluster
}

func (o *CollectorOpts) Validate() error {
	if o.Clusters == nil {
		return fmt.Errorf("invalid cluster subscriber")
	}
	if o.NewWatcherFunc == nil {
		return fmt.Errorf("NewWatcherFunc must be supplied")
	}
	if o.ServiceAccount.Name == "" || o.ServiceAccount.Namespace == "" {
		return fmt.Errorf("ImpersonateServiceAccount name and namespace must be non-empty")
	}
	return nil
}

// Collector factory method. It creates a collection with clusterName watching strategy by default.
func NewCollector(opts CollectorOpts) (Collector, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid collector options: %w", err)
	}
	return newWatchingCollector(opts)
}
