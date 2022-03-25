package clusters

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetClusterFromCRDs(t *testing.T) {
	c1 := makeTestCluster(func(o *capiv1.Cluster) {
		o.ObjectMeta.Name = "weave-cluster"
		o.ObjectMeta.Namespace = "default"
		o.Spec.CapiClusterRef.Name = "dev"
	})
	c2 := makeTestCluster(func(o *capiv1.Cluster) {
		o.ObjectMeta.Name = "weave-cluster2"
		o.ObjectMeta.Namespace = "default"
		o.Spec.SecretRef.Name = "dev"
	})
	lib := CRDLibrary{Log: logr.Discard(), ClientGetter: kubefakes.NewFakeClientGetter(makeClient(t, c1, c2))}
	result, err := lib.Get(context.Background(), "weave-cluster2")
	if err != nil {
		t.Fatalf("On no, error: %v", err)
	}
	if diff := cmp.Diff(c2, result); diff != "" {
		t.Fatalf("On no, diff clusters: %v", diff)
	}
}

func TestListClusterFromCRDs(t *testing.T) {
	c1 := makeTestCluster(func(o *capiv1.Cluster) {
		o.ObjectMeta.Name = "weave-cluster"
		o.ObjectMeta.Namespace = "default"
		o.Spec.CapiClusterRef.Name = "dev"
	})
	c2 := makeTestCluster(func(o *capiv1.Cluster) {
		o.ObjectMeta.Name = "weave-cluster2"
		o.ObjectMeta.Namespace = "default"
		o.Spec.SecretRef.Name = "dev"
	})
	lib := CRDLibrary{Log: logr.Discard(), ClientGetter: kubefakes.NewFakeClientGetter(makeClient(t, c1, c2))}
	result, err := lib.List(context.Background())
	if err != nil {
		t.Fatalf("On no, error: %v", err)
	}
	want := map[string]*capiv1.Cluster{
		"weave-cluster":  c1,
		"weave-cluster2": c2,
	}
	if diff := cmp.Diff(want, result); diff != "" {
		t.Fatalf("On no, diff clusters: %v", diff)
	}
}

func makeClient(t *testing.T, clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		capiv1.AddToScheme,
	}
	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		t.Fatal(err)
	}

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(clusterState...).
		Build()
}

func makeTestCluster(opts ...func(*capiv1.Cluster)) *capiv1.Cluster {
	c := &capiv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "capi.weave.works/v1alpha1",
			Kind:       "Cluster",
		},
		Spec: capiv1.ClusterSpec{},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}
