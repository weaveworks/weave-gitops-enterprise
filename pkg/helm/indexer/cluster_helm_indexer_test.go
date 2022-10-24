package indexer_test

import (
	"log"
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
		clustersFetcher, nsChecker, logger, scheme, clustersmngr.NewClustersClientsPool, clustersmngr.DefaultKubeConfigOptions)
	err = clientsFactory.UpdateClusters(ctx)
	g.Expect(err).To(BeNil())

	clusterName1 := "bar"
	clusterName2 := "foo"

	c1 := makeLeafCluster(t, clusterName1)
	c2 := makeLeafCluster(t, clusterName2)

	watcher := clientsFactory.Subscribe()
	g.Expect(watcher).ToNot(BeNil())

	indexer := indexer.NewClusterHelmIndexerTracker(nil)
	g.Expect(indexer.ClusterWatchers).ToNot(BeNil())

	indexer.Start(context.TODO(), watcher, logger)

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
		log.Printf("indexer.Added: %+v", indexer)

		<-watcher.Updates
		g.Expect(clusterNames(indexer.ClusterWatchers)).To(Equal([]string{clusterName1, clusterName2}))
	})
}

func makeLeafCluster(t *testing.T, name string) clustersmngr.Cluster {
	t.Helper()

	return clustersmngr.Cluster{
		Name: name,
	}
}
