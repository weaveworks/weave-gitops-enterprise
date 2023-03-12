package collector

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . ObjectRecord
type ObjectRecord interface {
	ClusterName() string
	Object() client.Object
}

//counterfeiter:generate . Collector
type Collector interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type CollectorOpts struct {
	Log            logr.Logger
	ObjectKinds    []schema.GroupVersionKind
	ClusterManager clustersmngr.ClustersManager
	Clusters       []cluster.Cluster
	PollInterval   time.Duration
}

// Collector factory method. It creates a collection with cluster watching strategy by default.
func NewCollector(opts CollectorOpts, store store.Store, newWatcherFunc NewWatcherFunc) (Collector, error) {
	return newWatchingCollector(opts, store, newWatcherFunc)
}
