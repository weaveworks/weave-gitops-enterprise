package collector

import (
	"fmt"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"

	"github.com/go-logr/logr"
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

// ClustersSubscriber represents the requirement for a value that can
// notify about clusters being added, updated, and removed.
type ClustersSubscriber interface {
	Subscribe() *clustersmngr.ClustersWatcher // NB distinct from ClusterWatcher here
	GetClusters() []cluster.Cluster
}

type CollectorOpts struct {
	Log            logr.Logger
	Clusters       ClustersSubscriber
	NewWatcherFunc NewWatcherFunc
}

func (o *CollectorOpts) Validate() error {
	if o.Clusters == nil {
		return fmt.Errorf("invalid cluster subscriber")
	}
	if o.NewWatcherFunc == nil {
		return fmt.Errorf("NewWatcherFunc must be supplied")
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
