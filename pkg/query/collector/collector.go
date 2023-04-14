package collector

import (
	"context"
	"fmt"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// ClusterWatcher defines an interface to watch gitops clusters via kubernetes https://kubernetes.io/docs/reference/using-api/api-concepts/#semantics-for-watch
type ClusterWatcher interface {
	// Watch starts watching the cluster passed as input
	Watch(ctx context.Context, cluster cluster.Cluster) error
	// Unwatch stops watching the cluster identified by input clusterName.
	Unwatch(clusterName string) error
	// Status return the watcher status for the cluster identified as clusterName.
	Status(clusterName string) (string, error)
}

//counterfeiter:generate . Collector
type Collector interface {
	ClusterWatcher
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type CollectorOpts struct {
	Log                logr.Logger
	ObjectKinds        []schema.GroupVersionKind
	ClusterManager     clustersmngr.ClustersManager
	ProcessRecordsFunc ProcessRecordsFunc
	NewWatcherFunc     NewWatcherFunc
}

func (o *CollectorOpts) Validate() error {
	if o.ObjectKinds == nil || len(o.ObjectKinds) == 0 {
		return fmt.Errorf("invalid object kinds")
	}
	if o.ClusterManager == nil {
		return fmt.Errorf("invalid cluster manager")
	}
	if o.ProcessRecordsFunc == nil {
		return fmt.Errorf("process records func is nil")
	}

	return nil
}

// Collector factory method. It creates a collection with clusterName watching strategy by default.
func NewCollector(opts CollectorOpts, store store.Store) (Collector, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid collector options: %w", err)
	}
	return newWatchingCollector(opts, store)
}
