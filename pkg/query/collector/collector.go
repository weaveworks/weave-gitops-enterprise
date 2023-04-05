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

// Interface for watching clusters via kuberentes api
// https://kubernetes.io/docs/reference/using-api/api-concepts/#semantics-for-watch
type ClusterWatcher interface {
	Watch(ctx context.Context, cluster cluster.Cluster) error
	Unwatch(clusterName string) error
	Status(cluster cluster.Cluster) (string, error)
}

//counterfeiter:generate . Collector
type Collector interface {
	ClusterWatcher
	Start() error
	Stop() error
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
