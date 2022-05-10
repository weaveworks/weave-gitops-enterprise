package clusters

import (
	"context"
	"fmt"
	"testing"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetClusterFromCRDs(t *testing.T) {
	c1 := makeTestCluster(func(o *gitopsv1alpha1.GitopsCluster) {
		o.ObjectMeta.Name = "gitops-cluster"
		o.Spec.CAPIClusterRef = &meta.LocalObjectReference{
			Name: "dev",
		}
	})
	c2 := makeTestCluster(func(o *gitopsv1alpha1.GitopsCluster) {
		o.ObjectMeta.Name = "gitops-cluster2"
		o.Spec.SecretRef = &meta.LocalObjectReference{
			Name: "dev",
		}
	})
	lib := CRDLibrary{Log: logr.Discard(), ClientGetter: kubefakes.NewFakeClientGetter(makeClient(t, c1, c2))}
	result, err := lib.Get(context.Background(), "gitops-cluster2")
	if err != nil {
		t.Fatalf("On no, error: %v", err)
	}
	if diff := cmp.Diff(c2, result); diff != "" {
		t.Fatalf("On no, diff clusters: %v", diff)
	}
}

func TestListClusterFromCRDs(t *testing.T) {
	c1 := makeTestCluster(func(o *gitopsv1alpha1.GitopsCluster) {
		o.ObjectMeta.Name = "gitops-cluster"
		o.ObjectMeta.Namespace = "default"
		o.Spec.CAPIClusterRef = &meta.LocalObjectReference{
			Name: "dev",
		}
	})
	c2 := makeTestCluster(func(o *gitopsv1alpha1.GitopsCluster) {
		o.ObjectMeta.Name = "gitops-cluster2"
		o.ObjectMeta.Namespace = "default"
		o.Spec.SecretRef = &meta.LocalObjectReference{
			Name: "dev",
		}
	})
	lib := CRDLibrary{Log: logr.Discard(), ClientGetter: kubefakes.NewFakeClientGetter(makeClient(t, c1, c2))}
	result, _, err := lib.List(context.Background(), client.ListOptions{})
	if err != nil {
		t.Fatalf("On no, error: %v", err)
	}
	want := map[string]*gitopsv1alpha1.GitopsCluster{
		"gitops-cluster":  c1,
		"gitops-cluster2": c2,
	}
	if diff := cmp.Diff(want, result); diff != "" {
		t.Fatalf("On no, diff clusters: %v", diff)
	}
}

func TestListClusterFromCRDs_Pagination(t *testing.T) {
	clusters := []runtime.Object{}
	for i := 1; i <= 25; i++ {
		c1 := makeTestCluster(func(o *gitopsv1alpha1.GitopsCluster) {
			o.ObjectMeta.Name = fmt.Sprintf("gitops-cluster-%d", i)
			o.ObjectMeta.Namespace = "default"
			o.Spec.CAPIClusterRef = &meta.LocalObjectReference{
				Name: "dev",
			}
		})
		clusters = append(clusters, c1)
	}

	lib := CRDLibrary{Log: logr.Discard(), ClientGetter: kubefakes.NewFakeClientGetter(makeClient(t, clusters...))}
	opts := client.ListOptions{
		Limit: 10,
	}
	result, nextPageToken, err := lib.List(context.Background(), opts)
	if err != nil {
		t.Fatalf("On no, error: %v", err)
	}
	fmt.Println(nextPageToken)
	fmt.Println(result["gitops-cluster-13"].GetObjectMeta())
	assert.Equal(t, 10, len(result))
}

func makeClient(t *testing.T, clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		capiv1.AddToScheme,
		gitopsv1alpha1.AddToScheme,
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

func makeTestCluster(opts ...func(*gitopsv1alpha1.GitopsCluster)) *gitopsv1alpha1.GitopsCluster {
	c := &gitopsv1alpha1.GitopsCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gitops.weave.works/v1alpha1",
			Kind:       "GitopsCluster",
		},
		Spec: gitopsv1alpha1.GitopsClusterSpec{},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}
