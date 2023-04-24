//go:build integration
// +build integration

package objectscollector_test

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/objectscollector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/core/nsaccess/nsaccessfakes"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	typedauth "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
	"os"
	"testing"
	"time"
)

const (
	defaultTimeout  = time.Second * 5
	defaultInterval = time.Second
)

// TestObjectsCollector is an integration test for testing integration of a collector with a kubernetes cluster
func TestObjectsCollector(t *testing.T) {
	g := NewGomegaWithT(t)
	g.SetDefaultEventuallyTimeout(defaultTimeout)
	g.SetDefaultEventuallyPollingInterval(defaultInterval)

	testLog := testr.New(t)
	tests := []struct {
		name string
	}{
		{
			name: "should support apps (using helm releases)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given a collector
			// and a cluster without any particular configuration
			oc, store, err := makeObjectsCollector(t, cfg, testLog)
			g.Expect(err).To(BeNil())
			// when collected
			g.Expect(oc.Start()).To(Succeed())
			// then data has been collected and stored
			querySucceeded := g.Eventually(func() bool {
				objectsIterator, err := store.GetObjects(ctx, nil, nil)
				g.Expect(err).To(BeNil())
				all, err := objectsIterator.All()
				g.Expect(err).To(BeNil())
				return len(all) > 0
			}).Should(BeTrue())
			//Then query is successfully executed
			g.Expect(querySucceeded).To(BeTrue())

		})
	}
}

func makeObjectsCollector(t *testing.T, cfg *rest.Config, testLog logr.Logger) (*objectscollector.ObjectsCollector, store.Store, error) {

	fetcher := &clustersmngrfakes.FakeClusterFetcher{}

	fakeCluster, err := cluster.NewSingleCluster("envtest", cfg, scheme.Scheme)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create cluster:%w", err)
	}

	fetcher.FetchReturns([]cluster.Cluster{fakeCluster}, nil)

	nsChecker := nsaccessfakes.FakeChecker{}
	nsChecker.FilterAccessibleNamespacesStub = func(ctx context.Context, client typedauth.AuthorizationV1Interface, n []v1.Namespace) ([]v1.Namespace, error) {
		// Pretend the user has access to everything
		return n, nil
	}

	clustersManager := clustersmngr.NewClustersManager(
		[]clustersmngr.ClusterFetcher{fetcher},
		&nsChecker,
		testLog,
	)

	dbDir, err := os.MkdirTemp("", "db")
	if err != nil {
		return nil, nil, err
	}

	s, err := store.NewStore(store.StorageBackendSQLite, dbDir, testLog)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create store:%w", err)
	}

	opts := collector.CollectorOpts{
		ObjectKinds:    configuration.SupportedObjectKinds,
		ClusterManager: clustersManager,
		Log:            testLog,
	}

	oc, err := objectscollector.NewObjectsCollector(s, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create collector:%w", err)
	}

	err = oc.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create start collector:%w", err)
	}

	t.Cleanup(func() {
		oc.Stop()
	})

	return oc, s, nil
}
