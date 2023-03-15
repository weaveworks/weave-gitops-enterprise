package collector

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// Interface for watching clusters via kuberentes api
// https://kubernetes.io/docs/reference/using-api/api-concepts/#semantics-for-watch
type ClusterWatcher interface {
	Watch(cluster cluster.Cluster, objectsChannel chan []models.ObjectRecord, ctx context.Context, log logr.Logger) error
	Unwatch(cluster cluster.Cluster) error
	Status(cluster cluster.Cluster) (string, error)
}

//counterfeiter:generate . Collector
type Collector interface {
	ClusterWatcher
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type CollectorOpts struct {
	Log            logr.Logger
	ObjectKinds    []schema.GroupVersionKind
	ClusterManager clustersmngr.ClustersManager
	Clusters       []cluster.Cluster
}

// Collector factory method. It creates a collection with clusterName watching strategy by default.
func NewCollector(opts CollectorOpts, store store.Store, processRecordsFunc ProcessRecordsFunc, newWatcherFunc NewWatcherFunc) (Collector, error) {
	return newWatchingCollector(opts, store, newWatcherFunc, processRecordsFunc)
}
