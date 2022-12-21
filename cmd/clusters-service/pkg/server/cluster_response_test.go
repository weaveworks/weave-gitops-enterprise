package server

import (
	"context"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	"google.golang.org/protobuf/testing/protocmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	capiv1 "github.com/weaveworks/templates-controller/apis/capi/v1alpha2"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func TestAddCAPIClusters(t *testing.T) {
	testCases := []struct {
		name           string
		gitopsClusters []*capiv1_proto.GitopsCluster
		capiClusters   []clusterv1.Cluster
		clusterObjects []runtime.Object
		expected       []*capiv1_proto.GitopsCluster
		err            error
	}{
		{
			name:           "empty",
			gitopsClusters: []*capiv1_proto.GitopsCluster{},
			clusterObjects: []runtime.Object{},
			expected:       []*capiv1_proto.GitopsCluster{},
		},
		{
			name: "CapiClusterRef exists",
			gitopsClusters: []*capiv1_proto.GitopsCluster{
				{
					Name:        "capi-cluster",
					Namespace:   "default",
					Annotations: map[string]string{},
					Labels:      map[string]string{},
					CapiClusterRef: &capiv1_proto.GitopsClusterRef{
						Name: "dev",
					},
				},
			},
			clusterObjects: []runtime.Object{
				makeTestCluster(func(o *clusterv1.Cluster) {
					o.ObjectMeta.Name = "capi-cluster"
					o.ObjectMeta.Namespace = "default"
					o.ObjectMeta.Annotations = map[string]string{
						"cni": "calico",
					}
					o.Status.Phase = "Provisioned"
					o.Status.Conditions = clusterv1.Conditions{
						clusterv1.Condition{
							Type:   clusterv1.ControlPlaneInitializedCondition,
							Status: corev1.ConditionStatus(strconv.FormatBool(true)),
						},
					}
				}),
			},
			expected: []*capiv1_proto.GitopsCluster{
				{
					Name:      "capi-cluster",
					Namespace: "default",
					CapiClusterRef: &capiv1_proto.GitopsClusterRef{
						Name: "dev",
					},
					CapiCluster: &capiv1_proto.CapiCluster{
						Name:      "capi-cluster",
						Namespace: "default",
						Annotations: map[string]string{
							"cni": "calico",
						},
						Status: &capiv1_proto.CapiClusterStatus{
							Phase:                   "Provisioned",
							ControlPlaneInitialized: true,
							Conditions: []*capiv1_proto.Condition{
								{
									Type:      string(clusterv1.ControlPlaneInitializedCondition),
									Status:    "true",
									Timestamp: "0001-01-01 00:00:00 +0000 UTC",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "CapiCluster Infrastructure ref exists",
			gitopsClusters: []*capiv1_proto.GitopsCluster{
				{
					Name:        "capi-cluster",
					Namespace:   "default",
					Annotations: map[string]string{},
					Labels:      map[string]string{},
					CapiClusterRef: &capiv1_proto.GitopsClusterRef{
						Name: "dev",
					},
				},
			},
			clusterObjects: []runtime.Object{
				makeTestCluster(func(o *clusterv1.Cluster) {
					o.ObjectMeta.Name = "capi-cluster"
					o.ObjectMeta.Namespace = "default"
					o.ObjectMeta.Annotations = map[string]string{
						"cni": "calico",
					}
					o.Status.Phase = "Provisioned"
					o.Status.Conditions = clusterv1.Conditions{
						clusterv1.Condition{
							Type:   clusterv1.ControlPlaneInitializedCondition,
							Status: corev1.ConditionStatus(strconv.FormatBool(true)),
						},
					}
					o.Spec.InfrastructureRef = &corev1.ObjectReference{
						APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
						Name:       "capi-cluster",
						Kind:       "DockerCluster",
					}
				}),
			},
			expected: []*capiv1_proto.GitopsCluster{
				{
					Name:      "capi-cluster",
					Namespace: "default",
					CapiClusterRef: &capiv1_proto.GitopsClusterRef{
						Name: "dev",
					},
					CapiCluster: &capiv1_proto.CapiCluster{
						Name:      "capi-cluster",
						Namespace: "default",
						Annotations: map[string]string{
							"cni": "calico",
						},
						Status: &capiv1_proto.CapiClusterStatus{
							Phase:                   "Provisioned",
							ControlPlaneInitialized: true,
							Conditions: []*capiv1_proto.Condition{
								{
									Type:      string(clusterv1.ControlPlaneInitializedCondition),
									Status:    "true",
									Timestamp: "0001-01-01 00:00:00 +0000 UTC",
								},
							},
						},
						InfrastructureRef: &capiv1_proto.CapiClusterInfrastructureRef{
							ApiVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
							Name:       "capi-cluster",
							Kind:       "DockerCluster",
						},
					},
				},
			},
		},
		{
			name: "CapiClusterRef exists but cluster isn't ready",
			gitopsClusters: []*capiv1_proto.GitopsCluster{
				{
					Name:        "capi-cluster",
					Namespace:   "default",
					Annotations: map[string]string{},
					Labels:      map[string]string{},
					CapiClusterRef: &capiv1_proto.GitopsClusterRef{
						Name: "dev",
					},
				},
			},
			clusterObjects: []runtime.Object{
				makeTestCluster(func(o *clusterv1.Cluster) {
					o.ObjectMeta.Name = "capi-cluster"
					o.ObjectMeta.Namespace = "default"
					o.ObjectMeta.Annotations = map[string]string{
						"cni": "calico",
					}
					o.Status.Phase = string(clusterv1.ClusterPhasePending)
				}),
			},
			expected: []*capiv1_proto.GitopsCluster{
				{
					Name:      "capi-cluster",
					Namespace: "default",
					CapiClusterRef: &capiv1_proto.GitopsClusterRef{
						Name: "dev",
					},
					CapiCluster: &capiv1_proto.CapiCluster{
						Name:      "capi-cluster",
						Namespace: "default",
						Annotations: map[string]string{
							"cni": "calico",
						},
						Status: &capiv1_proto.CapiClusterStatus{
							Phase:                   string(clusterv1.ClusterPhasePending),
							ControlPlaneInitialized: false,
						},
					},
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
			expected: []*capiv1_proto.GitopsCluster{
				{
					Name:      "gitops-cluster",
					Namespace: "default",
					SecretRef: &capiv1_proto.GitopsClusterRef{
						Name: "dev",
					},
				},
			},
		},
		{
			name: "CapiClusterRef exists but no cluster present",
			gitopsClusters: []*capiv1_proto.GitopsCluster{
				{
					Name:        "capi-cluster",
					Namespace:   "default",
					Annotations: map[string]string{},
					Labels:      map[string]string{},
					CapiClusterRef: &capiv1_proto.GitopsClusterRef{
						Name: "dev",
					},
				},
			},
			expected: []*capiv1_proto.GitopsCluster{
				{
					Name:      "capi-cluster",
					Namespace: "default",
					CapiClusterRef: &capiv1_proto.GitopsClusterRef{
						Name: "dev",
					},
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			c := makeTestClient(t, tt.clusterObjects...)
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
