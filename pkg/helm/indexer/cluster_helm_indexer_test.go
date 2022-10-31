package indexer_test

import (
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/indexer"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/multiwatcher"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/core/nsaccess/nsaccessfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"golang.org/x/net/context"
)

func TestClusterHelmIndexerTracker(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := logr.Discard()
	ctx := context.Background()

	nsChecker := &nsaccessfakes.FakeChecker{}

	clustersFetcher := new(clustersmngrfakes.FakeClusterFetcher)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	clientsFactory := clustersmngr.NewClustersManager(
		clustersFetcher, nsChecker, logger, scheme, clustersmngr.ClientFactory, clustersmngr.DefaultKubeConfigOptions)
	err = clientsFactory.UpdateClusters(ctx)
	g.Expect(err).To(BeNil())

	clusterName1 := "bar"
	clusterName2 := "foo"

	c1 := makeLeafCluster(clusterName1)
	c2 := makeLeafCluster(clusterName2)

	indexer := indexer.NewClusterHelmIndexerTracker(nil, "")
	g.Expect(indexer.ClusterWatchers).ToNot(BeNil())

	go func() {
		err := indexer.Start(ctx, clientsFactory, logger)
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	}()

	clusterNames := func(c map[string]*multiwatcher.Watcher) []string {
		names := []string{}
		for name := range c {
			names = append(names, name)
		}

		return names
	}

	t.Run("indexer should be notified with two clusters added", func(t *testing.T) {
		g := NewGomegaWithT(t)
		clustersFetcher.FetchReturns([]clustersmngr.Cluster{c1, c2}, nil)

		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())

		g.Eventually(func() []string {
			return clusterNames(indexer.ClusterWatchers)
		}).Should(ConsistOf(clusterName1, clusterName2))
	})
}

func makeLeafCluster(name string) clustersmngr.Cluster {
	return clustersmngr.Cluster{
		Name: name,
	}
}
