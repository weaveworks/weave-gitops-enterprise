//go:build integration
// +build integration

package objectscollector_test

import (
	"fmt"
	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	flaggerv1beta1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	gitopssetsv1alpha1 "github.com/weaveworks/gitopssets-controller/api/v1alpha1"
	pipelinev1alpha1 "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	pacv2beta1 "github.com/weaveworks/policy-agent/api/v2beta1"
	pacv2beta2 "github.com/weaveworks/policy-agent/api/v2beta2"
	capiv1 "github.com/weaveworks/templates-controller/apis/capi/v1alpha2"
	gapiv1 "github.com/weaveworks/templates-controller/apis/gitops/v1alpha2"
	tfctrl "github.com/weaveworks/tf-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/fetcher"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/objectscollector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/logger"
	"github.com/weaveworks/weave-gitops/core/nsaccess"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"os"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"testing"
	"time"
)

const (
	defaultTimeout  = time.Second * 30
	defaultInterval = time.Second
)

// TestObjectsCollector is an integration test for testing integration of a collector with a kubernetes cluster
func TestObjectsCollector(t *testing.T) {
	g := NewGomegaWithT(t)
	g.SetDefaultEventuallyTimeout(defaultTimeout)
	g.SetDefaultEventuallyPollingInterval(defaultInterval)

	testLog, err := logger.New("debug", false)
	g.Expect(err).To(BeNil())

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
				testLog.Info("num objects:", "numObjects", len(all))
				g.Expect(err).To(BeNil())
				return len(all) > 0
			}).Should(BeTrue())
			//Then query is successfully executed
			g.Expect(querySucceeded).To(BeTrue())

		})
	}
}

func makeObjectsCollector(t *testing.T, cfg *rest.Config, testLog logr.Logger) (*objectscollector.ObjectsCollector, store.Store, error) {
	clustersManagerScheme := runtime.NewScheme()
	builder := runtime.NewSchemeBuilder(
		capiv1.AddToScheme,
		pacv2beta1.AddToScheme,
		pacv2beta2.AddToScheme,
		esv1beta1.AddToScheme,
		flaggerv1beta1.AddToScheme,
		pipelinev1alpha1.AddToScheme,
		tfctrl.AddToScheme,
		gitopsv1alpha1.AddToScheme,
		gitopssetsv1alpha1.AddToScheme,
		clusterv1.AddToScheme,
		gapiv1.AddToScheme,
		v1.AddToScheme,
	)
	if err := builder.AddToScheme(clustersManagerScheme); err != nil {
		return nil, nil, err
	}

	mgmtCluster, err := cluster.NewSingleCluster("management", cfg, clustersManagerScheme, cluster.DefaultKubeConfigOptions...)

	gcf := fetcher.NewGitopsClusterFetcher(testLog, mgmtCluster, "flux-system",
		clustersManagerScheme, false, cluster.DefaultKubeConfigOptions...)
	fetchers := []clustersmngr.ClusterFetcher{gcf}

	clustersManager := clustersmngr.NewClustersManager(
		fetchers,
		nsaccess.NewChecker(nsaccess.DefautltWegoAppRules),
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

	//watch object events
	go func() {
		clustersManager.Start(ctx)
	}()

	t.Cleanup(func() {
		oc.Stop()
	})

	return oc, s, nil
}
