package clusters

import (
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Subscription

type Subscription interface {
	Updates() chan clustersmngr.ClusterListUpdate
	Unsubscribe()
}

//counterfeiter:generate . Subscriber

type Subscriber interface {
	Subscribe() Subscription
	GetClusters() []cluster.Cluster
}

type ClustersManagerAsSubscriber struct {
	clustersmngr.ClustersManager
}

func MakeSubscriber(c clustersmngr.ClustersManager) Subscriber {
	return ClustersManagerAsSubscriber{c}
}

func (c ClustersManagerAsSubscriber) Subscribe() Subscription {
	return ClusterWatcherAsSubscription{c.ClustersManager.Subscribe()}
}

type ClusterWatcherAsSubscription struct {
	*clustersmngr.ClustersWatcher
}

func (c ClusterWatcherAsSubscription) Updates() chan clustersmngr.ClusterListUpdate {
	return c.ClustersWatcher.Updates
}
