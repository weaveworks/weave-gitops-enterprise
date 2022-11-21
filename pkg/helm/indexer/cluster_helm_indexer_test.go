package indexer_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/helmfakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/indexer"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/multiwatcher"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster/clusterfakes"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/core/nsaccess/nsaccessfakes"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

func TestClusterHelmIndexerTracker(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := logr.Discard()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nsChecker := &nsaccessfakes.FakeChecker{}
	clustersFetcher := new(clustersmngrfakes.FakeClusterFetcher)
	clientsFactory := clustersmngr.NewClustersManager(clustersFetcher, nsChecker, logger)
	err := clientsFactory.UpdateClusters(ctx)
	g.Expect(err).To(BeNil())

	managementClusterName := "cluster1-management"
	cluster1 := types.NamespacedName{
		Name: managementClusterName,
	}
	clusterName2 := "default/cluster2"

	c1 := makeLeafCluster(managementClusterName)
	c2 := makeLeafCluster(clusterName2)

	fakeCache := helmfakes.NewFakeChartCache(helmfakes.WithCharts(
		helmfakes.ClusterRefToString(
			helm.ObjectReference{
				Namespace: "test-namespace",
				Name:      "test-name",
			},
			cluster1,
		),
		[]helm.Chart{
			{
				Name:    "test-profiles-1",
				Version: "0.0.1",
				Layer:   "test-layer",
			},
		},
	))

	ind := indexer.NewClusterHelmIndexerTracker(fakeCache, managementClusterName, newFakeWatcher)
	g.Expect(ind.ClusterWatchers).ToNot(BeNil())
	go func() {
		err := ind.Start(ctx, clientsFactory, logger)
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	}()

	t.Run("should be notified with two clusters added", func(t *testing.T) {
		g := NewGomegaWithT(t)
		clustersFetcher.FetchReturns([]cluster.Cluster{c1, c2}, nil)

		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())

		g.Eventually(func() []string {
			return clusterNames(ind.ClusterWatchers)
		}).Should(ConsistOf(managementClusterName, clusterName2))

		// cast to fakeWatcher to check if it is a management cluster
		fakeWatcher1 := ind.ClusterWatchers[managementClusterName].(*fakeWatcher)
		g.Expect(fakeWatcher1.isManagementCluster).To(BeTrue())

		fakeWatcher2 := ind.ClusterWatchers[clusterName2].(*fakeWatcher)
		g.Expect(fakeWatcher2.isManagementCluster).To(BeFalse())
	})

	t.Run("should remove items when cluster is removed", func(t *testing.T) {
		g := NewGomegaWithT(t)
		clustersFetcher.FetchReturns([]cluster.Cluster{c1, c2}, nil)
		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())
		clustersFetcher.FetchReturns([]cluster.Cluster{c2}, nil)
		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())

		// only cluster 2 is left
		g.Eventually(func() []string {
			return clusterNames(ind.ClusterWatchers)
		}).Should(ConsistOf(clusterName2))

		// cache should be empty
		g.Expect(fakeCache.Charts).To(BeEmpty())
	})
}

func TestErroringCache(t *testing.T) {
	g := NewGomegaWithT(t)

	// create a new testing logger
	logger := testr.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nsChecker := &nsaccessfakes.FakeChecker{}
	clustersFetcher := new(clustersmngrfakes.FakeClusterFetcher)
	clientsFactory := clustersmngr.NewClustersManager(clustersFetcher, nsChecker, logger)
	err := clientsFactory.UpdateClusters(ctx)
	g.Expect(err).To(BeNil())

	clusterName1 := "default/cluster1"
	clusterName2 := "default/cluster2"

	c1 := makeLeafCluster(clusterName1)
	c2 := makeLeafCluster(clusterName2)

	ind := indexer.NewClusterHelmIndexerTracker(helmfakes.NewFakeChartCache(func(fc *helmfakes.FakeChartCache) {
		fc.DeleteAllChartsForClusterError = errors.New("oh no")
	}), "", newFakeWatcher)
	g.Expect(ind.ClusterWatchers).ToNot(BeNil())
	go func() {
		err := ind.Start(ctx, clientsFactory, logger)
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	}()

	t.Run("should still remove the watcher if the cache has some issues cleaning up", func(t *testing.T) {
		g := NewGomegaWithT(t)

		clustersFetcher.FetchReturns([]cluster.Cluster{c1, c2}, nil)
		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())
		g.Eventually(func() []string {
			return clusterNames(ind.ClusterWatchers)
		}).Should(ConsistOf(clusterName1, clusterName2))

		clustersFetcher.FetchReturns([]cluster.Cluster{c2}, nil)
		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())
		g.Eventually(func() []string {
			return clusterNames(ind.ClusterWatchers)
		}).Should(ConsistOf(clusterName2))
	})
}

func TestNewIndexerSetsUserProxy(t *testing.T) {
	g := NewGomegaWithT(t)

	// Manamgent cluster should not use proxy
	isManagementCluster := true
	ind, err := indexer.NewIndexer(&rest.Config{}, types.NamespacedName{Name: "foo", Namespace: "bar"}, isManagementCluster, nil)
	g.Expect(err).To(BeNil())
	watcher := ind.(*multiwatcher.Watcher)
	g.Expect(watcher.UseProxy).To(BeFalse())

	// other clusters should use proxy
	isManagementCluster = false
	ind, err = indexer.NewIndexer(&rest.Config{}, types.NamespacedName{Name: "foo", Namespace: "bar"}, isManagementCluster, nil)
	g.Expect(err).To(BeNil())
	watcher = ind.(*multiwatcher.Watcher)
	g.Expect(watcher.UseProxy).To(BeTrue())
}

func makeLeafCluster(name string) cluster.Cluster {
	cluster := clusterfakes.FakeCluster{}
	cluster.GetNameReturns(name)
	return &cluster
}

func clusterNames(c map[string]indexer.Watcher) []string {
	names := []string{}
	for name := range c {
		names = append(names, name)
	}

	return names
}

func newFakeWatcher(config *rest.Config, cluster types.NamespacedName, isManagementCluster bool, cache helm.ChartsCacherWriter) (indexer.Watcher, error) {
	return &fakeWatcher{isManagementCluster: isManagementCluster}, nil
}

type fakeWatcher struct {
	isManagementCluster bool
}

func (f fakeWatcher) StartWatcher(ctx context.Context, logr logr.Logger) error {
	return nil
}

func (f fakeWatcher) Stop() {
}
