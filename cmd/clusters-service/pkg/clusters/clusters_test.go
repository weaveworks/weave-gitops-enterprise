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
	c1 := makeTestCluster("weave-cluster", "foo")
	c2 := makeTestCluster("weave-cluster2", "bar")
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
	c1 := makeTestCluster("weave-cluster", "foo")
	c2 := makeTestCluster("weave-cluster2", "bar")
	lib := CRDLibrary{Log: logr.Discard(), ClientGetter: kubefakes.NewFakeClientGetter(makeClient(t, c1, c2))}
	result, err := lib.List(context.Background())
	if err != nil {
		t.Fatalf("On no, error: %v", err)
	}
	want := map[string]*capiv1.WeaveCluster{
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

func makeTestCluster(name, label string, opts ...func(*capiv1.WeaveCluster)) *capiv1.WeaveCluster {
	c := &capiv1.WeaveCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "capi.weave.works/v1alpha1",
			Kind:       "WeaveCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: capiv1.WeaveClusterSpec{
			Label: label,
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}
