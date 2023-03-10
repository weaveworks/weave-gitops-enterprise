package collector

import (
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/store"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
	Start() (<-chan []ObjectRecord, error)
	Stop() error
}

type CollectorOpts struct {
	Log            logr.Logger
	ObjectKinds    []schema.GroupVersionKind
	ClusterManager clustersmngr.ClustersManager
}

// Collector factory method. It creates a collection with cluster watching strategy by default.
func NewCollector(opts CollectorOpts, store store.Store, newWatcherFunc NewWatcherFunc) (Collector, error) {
	return newWatchingCollector(opts, store, newWatcherFunc)
}
