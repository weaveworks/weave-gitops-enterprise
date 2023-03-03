package collector

import (
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster/clusterfakes"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newTestCollector(t *testing.T, kinds []schema.GroupVersionKind, clusterState ...runtime.Object) Collector {
	k8s := createClient(t, clusterState...)
	f := &clusterfakes.FakeCluster{}
	f.GetServerClientReturns(k8s, nil)
	f.GetNameReturns("test-cluster")
	mgr := &clustersmngrfakes.FakeClustersManager{}
	mgr.GetClustersReturns([]cluster.Cluster{f})
	log := logr.Discard()

	collector := NewCollector(CollectorOpts{
		Log:            log,
		ClusterManager: mgr,
		ObjectKinds:    kinds,
		PollInterval:   100 * time.Millisecond,
	})

	return collector

}

func createClient(t *testing.T, clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		appsv1.AddToScheme,
		corev1.AddToScheme,
		rbacv1.AddToScheme,
	}
	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		t.Fatal(err)
	}

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(clusterState...).
		Build()

	return c
}
