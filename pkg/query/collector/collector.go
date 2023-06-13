package collector

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/client-go/rest"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/clusters"
)

const (
	ClusterWatchingStarting = "starting"
	ClusterWatchingStarted  = "started"
	ClusterWatchingStopped  = "stopped"
	ClusterWatchingFailed   = "failed"
	ClusterWatchingErrored  = "error"
)

const (
	ClusterWatchingStarted = "started"
	ClusterWatchingStopped = "stopped"
	ClusterWatchingFailed  = "failed"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// ClusterWatcher is for managing the lifecycle of watchers.
type ClusterWatcher interface {
	// Status return the watcher status for the cluster identified as clusterName.
	Status(clusterName string) (string, error)
}

// Starter is the expected return value of NewWatcherFunc.
type Starter interface {
	Start(context.Context) error
}

// Function to create a watcher for a set of kinds. Operations target an store.
type NewWatcherFunc = func(clusterName string, config *rest.Config) (Starter, error)

// StopWatcherFunc represents a hook to call when a watcher is
// stopped. This is used, for instance, to delete the records of a
// cluster that goes away.
type StopWatcherFunc = func(clusterName string) error

//counterfeiter:generate . Collector

// Collector is ClusterWatcher that has its own lifecycle (i.e.,
// Start() and Stop()).
type Collector interface {
	ClusterWatcher
	Starter
}

type ImpersonateServiceAccount struct {
	Name      string
	Namespace string
}

type ImpersonateServiceAccount struct {
	Name      string
	Namespace string
}

type CollectorOpts struct {
	Name            string // for metrics, logs
	Log             logr.Logger
	Clusters        clusters.Subscriber
	NewWatcherFunc  NewWatcherFunc
	StopWatcherFunc StopWatcherFunc
	ServiceAccount  ImpersonateServiceAccount // this gives the service account to impersonate when watching each cluster
}

func (o *CollectorOpts) Validate() error {
	if o.Name == "" {
		return fmt.Errorf("name should be non-empty, to distinguish this collector in metrics, logs, etc.")
	}
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
