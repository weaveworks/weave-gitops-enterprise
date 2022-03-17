package clusters

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const testdata1 = `
apiVersion: capi.weave.works/v1alpha1
kind: WeaveCluster
metadata:
  name: weave-cluster
spec:
  label: foo
`

const testdata2 = `
apiVersion: capi.weave.works/v1alpha1
kind: WeaveCluster
metadata:
  name: weave-cluster2
spec:
  label: bar
`

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

func TestGetClusterFromCRDs(t *testing.T) {
	c1 := mustParseCluster(t, testdata1)
	c2 := mustParseCluster(t, testdata2)
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
	c1 := mustParseCluster(t, testdata1)
	c2 := mustParseCluster(t, testdata2)
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

func mustParseCluster(t *testing.T, data string) *capiv1.WeaveCluster {
	t.Helper()

	var c capiv1.WeaveCluster
	err := yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(c)
	return &c
}
