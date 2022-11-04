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
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/core/nsaccess/nsaccessfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
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
	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())
	clientsFactory := clustersmngr.NewClustersManager(
		clustersFetcher, nsChecker, logger, scheme, clustersmngr.ClientFactory, clustersmngr.DefaultKubeConfigOptions)
	err = clientsFactory.UpdateClusters(ctx)
	g.Expect(err).To(BeNil())

	cluster1 := types.NamespacedName{
		Name:      "cluster1",
		Namespace: "default",
	}
	clusterName1 := cluster1.String()
	clusterName2 := "default/cluster2"

	c1 := makeLeafCluster(clusterName1)
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

	ind := indexer.NewClusterHelmIndexerTracker(fakeCache, "", newFakeWatcher)
	g.Expect(ind.ClusterWatchers).ToNot(BeNil())
	go func() {
		err := ind.Start(ctx, clientsFactory, logger)
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	}()

	t.Run("should be notified with two clusters added", func(t *testing.T) {
		g := NewGomegaWithT(t)
		clustersFetcher.FetchReturns([]clustersmngr.Cluster{c1, c2}, nil)

		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())

		g.Eventually(func() []string {
			return clusterNames(ind.ClusterWatchers)
		}).Should(ConsistOf(clusterName1, clusterName2))
	})

	t.Run("should remove items when cluster is removed", func(t *testing.T) {
		g := NewGomegaWithT(t)
		clustersFetcher.FetchReturns([]clustersmngr.Cluster{c1, c2}, nil)
		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())
		clustersFetcher.FetchReturns([]clustersmngr.Cluster{c2}, nil)
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
	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())
	clientsFactory := clustersmngr.NewClustersManager(
		clustersFetcher, nsChecker, logger, scheme, clustersmngr.ClientFactory, clustersmngr.DefaultKubeConfigOptions)
	err = clientsFactory.UpdateClusters(ctx)
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

		clustersFetcher.FetchReturns([]clustersmngr.Cluster{c1, c2}, nil)
		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())
		g.Eventually(func() []string {
			return clusterNames(ind.ClusterWatchers)
		}).Should(ConsistOf(clusterName1, clusterName2))

		clustersFetcher.FetchReturns([]clustersmngr.Cluster{c2}, nil)
		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())
		g.Eventually(func() []string {
			return clusterNames(ind.ClusterWatchers)
		}).Should(ConsistOf(clusterName2))
	})
}

func makeLeafCluster(name string) clustersmngr.Cluster {
	return clustersmngr.Cluster{
		Name: name,
	}
}

func clusterNames(c map[string]indexer.Watcher) []string {
	names := []string{}
	for name := range c {
		names = append(names, name)
	}

	return names
}

func newFakeWatcher(config *rest.Config, cluster types.NamespacedName, isManagementCluster bool, cache helm.ChartsCacherWriter) (indexer.Watcher, error) {
	return &fakeWatcher{}, nil
}

type fakeWatcher struct{}

func (f fakeWatcher) StartWatcher(ctx context.Context, logr logr.Logger) error {
	return nil
}

func (f fakeWatcher) Stop() {
}
