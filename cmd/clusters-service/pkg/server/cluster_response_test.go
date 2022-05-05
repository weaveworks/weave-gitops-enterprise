package server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"google.golang.org/protobuf/testing/protocmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestAddCAPIClusters(t *testing.T) {
	testCases := []struct {
		name           string
		gitopsClusters []*capiv1_proto.GitopsCluster
		capiClusters   []clusterv1.Cluster
		expected       []*capiv1_proto.CapiCluster
		err            error
	}{
		{
			name:           "empty",
			gitopsClusters: []*capiv1_proto.GitopsCluster{},
			expected:       []*capiv1_proto.CapiCluster{},
		},
		{
			name: "CapiClusterRef exists",
			gitopsClusters: []*capiv1_proto.GitopsCluster{
				{
					Name:        "gitops-cluster",
					Namespace:   "default",
					Annotations: map[string]string{},
					Labels:      map[string]string{},
					CapiClusterRef: &capiv1_proto.GitopsClusterRef{
						Name: "dev",
					},
				},
			},
			expected: []*capiv1_proto.CapiCluster{
				{
					Name:      "gitops-cluster",
					Namespace: "default",
					Status:    &capiv1_proto.CapiClusterStatus{},
				},
			},
		},
		{
			name: "CapiClusterRef doesn't exist",
			gitopsClusters: []*capiv1_proto.GitopsCluster{
				{
					Name:        "gitops-cluster",
					Namespace:   "default",
					Annotations: map[string]string{},
					Labels:      map[string]string{},
					SecretRef: &capiv1_proto.GitopsClusterRef{
						Name: "dev",
					},
				},
			},
			expected: []*capiv1_proto.CapiCluster{},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			c1 := makeTestCluster(func(o *clusterv1.Cluster) {
				o.ObjectMeta.Name = "gitops-cluster"
				o.ObjectMeta.Namespace = "default"
			})

			c := makeTestClient(t, c1)
			ctx := context.Background()
			result, err := AddCAPIClusters(ctx, c, tt.gitopsClusters)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.expected, result, protocmp.Transform()); diff != "" {
				t.Fatalf("clusters didn't match expected:\n%s", diff)
			}
		})
	}
}

func makeTestClient(t *testing.T, clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		capiv1.AddToScheme,
		gitopsv1alpha1.AddToScheme,
		clusterv1.AddToScheme,
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

func makeTestCluster(opts ...func(*clusterv1.Cluster)) *clusterv1.Cluster {
	c := &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "clusters.cluster.x-k8s.io",
			Kind:       "Cluster",
		},
		Spec: clusterv1.ClusterSpec{},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}
