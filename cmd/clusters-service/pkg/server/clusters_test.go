package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/testing/protocmp"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"

	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	capiv1 "github.com/weaveworks/templates-controller/apis/capi/v1alpha2"
	templatesv1 "github.com/weaveworks/templates-controller/apis/core"
	gapiv1 "github.com/weaveworks/templates-controller/apis/gitops/v1alpha2"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/charts"
	csgit "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git/gitfakes"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func TestListGitopsClusters(t *testing.T) {
	testCases := []struct {
		name         string
		clusterState []runtime.Object
		refType      string
		expected     []*capiv1_protos.GitopsCluster
		capiEnabled  bool
		err          error
	}{
		{
			name: "no clusters",
			clusterState: []runtime.Object{
				makeTestGitopsCluster(),
			},
			expected: []*capiv1_protos.GitopsCluster{
				{
					Name: "management",
					Conditions: []*capiv1_protos.Condition{
						{
							Type:   "Ready",
							Status: "True",
						},
					},
					ControlPlane: true,
				},
			},
		},
		{
			name: "1 cluster",
			clusterState: []runtime.Object{
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "gitops-cluster"
					o.ObjectMeta.Namespace = "default"
					o.TypeMeta.Kind = "GitopsCluster"
					o.Spec.CAPIClusterRef = &meta.LocalObjectReference{
						Name: "dev",
					}
				}),
			},
			expected: []*capiv1_protos.GitopsCluster{
				{
					Name:      "gitops-cluster",
					Namespace: "default",
					CapiClusterRef: &capiv1_protos.GitopsClusterRef{
						Name: "dev",
					},
					Type: "GitopsCluster",
				},
				{
					Name: "management",
					Conditions: []*capiv1_protos.Condition{
						{
							Type:   "Ready",
							Status: "True",
						},
					},
					ControlPlane: true,
				},
			},
		},
		{
			name: "2 clusters",
			clusterState: []runtime.Object{
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "gitops-cluster"
					o.ObjectMeta.Namespace = "default"
					o.TypeMeta.Kind = "GitopsCluster"
					o.Spec.CAPIClusterRef = &meta.LocalObjectReference{
						Name: "dev",
					}
				}),
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "gitops-cluster2"
					o.ObjectMeta.Namespace = "default"
					o.TypeMeta.Kind = "GitopsCluster"
					o.Spec.SecretRef = &meta.LocalObjectReference{
						Name: "dev",
					}
				}),
			},
			expected: []*capiv1_protos.GitopsCluster{
				{
					Name:      "gitops-cluster",
					Namespace: "default",
					CapiClusterRef: &capiv1_protos.GitopsClusterRef{
						Name: "dev",
					},
					Type: "GitopsCluster",
				},
				{
					Name:      "gitops-cluster2",
					Namespace: "default",
					SecretRef: &capiv1_protos.GitopsClusterRef{
						Name: "dev",
					},
					Type: "GitopsCluster",
				},
				{
					Name: "management",
					Conditions: []*capiv1_protos.Condition{
						{
							Type:   "Ready",
							Status: "True",
						},
					},
					ControlPlane: true,
				},
			},
		},
		{
			name: "filter by reference type",
			clusterState: []runtime.Object{
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "gitops-cluster"
					o.ObjectMeta.Namespace = "default"
					o.TypeMeta.Kind = "GitopsCluster"
					o.Spec.CAPIClusterRef = &meta.LocalObjectReference{
						Name: "dev",
					}
				}),
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "gitops-cluster2"
					o.ObjectMeta.Namespace = "default"
					o.TypeMeta.Kind = "GitopsCluster"
					o.Spec.SecretRef = &meta.LocalObjectReference{
						Name: "dev",
					}
				}),
			},
			refType: "Secret",
			expected: []*capiv1_protos.GitopsCluster{
				{
					Name:      "gitops-cluster2",
					Namespace: "default",
					SecretRef: &capiv1_protos.GitopsClusterRef{
						Name: "dev",
					},
					Type: "GitopsCluster",
				},
				{
					Name: "management",
					Conditions: []*capiv1_protos.Condition{
						{
							Type:   "Ready",
							Status: "True",
						},
					},
					ControlPlane: true,
				},
			},
		},
		{
			name: "invalid refType for filtering",
			clusterState: []runtime.Object{
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "gitops-cluster"
					o.ObjectMeta.Namespace = "default"
					o.Spec.CAPIClusterRef = &meta.LocalObjectReference{
						Name: "dev",
					}
				}),
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "gitops-cluster2"
					o.ObjectMeta.Namespace = "default"
					o.Spec.SecretRef = &meta.LocalObjectReference{
						Name: "dev",
					}
				}),
			},
			refType: "foo",
			err:     errors.New(`reference type "foo" is not recognised`),
		},
		{
			name: "capi-enabled is true",
			clusterState: []runtime.Object{
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "gitops-cluster"
					o.ObjectMeta.Namespace = "default"
					o.TypeMeta.Kind = "GitopsCluster"
					o.Spec.CAPIClusterRef = &meta.LocalObjectReference{
						Name: "dev",
					}
				}),
				makeTestCluster(func(o *clusterv1.Cluster) {
					o.ObjectMeta.Name = "gitops-cluster"
					o.ObjectMeta.Namespace = "default"
					o.TypeMeta.Kind = "GitopsCluster"
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
			expected: []*capiv1_protos.GitopsCluster{
				{
					Name:      "gitops-cluster",
					Namespace: "default",
					CapiClusterRef: &capiv1_protos.GitopsClusterRef{
						Name: "dev",
					},
					CapiCluster: &capiv1_protos.CapiCluster{
						Name:      "gitops-cluster",
						Namespace: "default",
						Annotations: map[string]string{
							"cni": "calico",
						},
						Status: &capiv1_protos.CapiClusterStatus{
							Phase:                   "Provisioned",
							ControlPlaneInitialized: true,
							Conditions: []*capiv1_protos.Condition{
								{
									Type:      string(clusterv1.ControlPlaneInitializedCondition),
									Status:    "true",
									Timestamp: "0001-01-01 00:00:00 +0000 UTC",
								},
							},
						},
					},
					Type: "GitopsCluster",
				},
				{
					Name: "management",
					Conditions: []*capiv1_protos.Condition{
						{
							Type:   "Ready",
							Status: "True",
						},
					},
					ControlPlane: true,
				},
			},
			capiEnabled: true,
		},
		{
			name: "capi-enabled is false",
			clusterState: []runtime.Object{
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "gitops-cluster"
					o.ObjectMeta.Namespace = "default"
					o.TypeMeta.Kind = "GitopsCluster"
					o.Spec.CAPIClusterRef = &meta.LocalObjectReference{
						Name: "dev",
					}
				}),
				makeTestCluster(func(o *clusterv1.Cluster) {
					o.ObjectMeta.Name = "gitops-cluster"
					o.ObjectMeta.Namespace = "default"
					o.TypeMeta.Kind = "GitopsCluster"
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
			expected: []*capiv1_protos.GitopsCluster{
				{
					Name:      "gitops-cluster",
					Namespace: "default",
					CapiClusterRef: &capiv1_protos.GitopsClusterRef{
						Name: "dev",
					},
					Type: "GitopsCluster",
				},
				{
					Name: "management",
					Conditions: []*capiv1_protos.Condition{
						{
							Type:   "Ready",
							Status: "True",
						},
					},
					ControlPlane: true,
				},
			},
			capiEnabled: false,
		},
	}

	ctx := auth.WithPrincipal(context.Background(), &auth.UserPrincipal{ID: "userID"})
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			setViperWithTestCleanup(t, map[string]string{
				"runtime-namespace": "default",
			})

			// setup
			s := createServer(t, serverOptions{
				clusterState: tt.clusterState,
				namespace:    "default",
				capiEnabled:  tt.capiEnabled,
				cluster:      "management",
			})

			// request
			listGitopsClustersRequest := new(capiv1_protos.ListGitopsClustersRequest)
			listGitopsClustersRequest.RefType = tt.refType
			listGitopsClustersResponse, err := s.ListGitopsClusters(ctx, listGitopsClustersRequest)

			// check response
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to list gitops clusters:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("got the wrong error:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, listGitopsClustersResponse.GitopsClusters, protocmp.Transform()); diff != "" {
					t.Fatalf("gitops clusters list didn't match expected:\n%s", diff)
				}
			}

		})
	}
}

func setViperWithTestCleanup(t *testing.T, m map[string]string) {
	t.Helper()
	for k, v := range m {
		viper.SetDefault(k, v)
	}
	t.Cleanup(func() {
		viper.Reset()
	})
}

func TestCreatePullRequest(t *testing.T) {
	// Lets make the above calls here instead
	setViperWithTestCleanup(t, map[string]string{
		"capi-repository-path":          "clusters/my-cluster/clusters",
		"capi-repository-clusters-path": "clusters",
		"add-bases-kustomization":       "enabled",
		"capi-templates-namespace":      "default",
	})

	testCases := []struct {
		name           string
		clusterState   []runtime.Object
		provider       csgit.Provider
		pruneEnvVar    string
		req            *capiv1_protos.CreatePullRequestRequest
		expected       string
		CommittedFiles []*capiv1_protos.CommitFile
		err            error
	}{
		{
			name: "validation errors",
			req:  &capiv1_protos.CreatePullRequestRequest{},
			err:  errors.New("2 errors occurred:\ntemplate name must be specified\nparameter values must be specified"),
		},
		{
			name: "name validation errors",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			req: &capiv1_protos.CreatePullRequestRequest{
				Name: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "foo bar bad name",
					"NAMESPACE":    "default",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
				Namespace:     "default",
			},
			err: errors.New(`validation error rendering template cluster-template-1, invalid value for metadata.name: "foo bar bad name", a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`),
		},
		{
			name: "namespace validation errors",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			req: &capiv1_protos.CreatePullRequestRequest{
				Name: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "foo bar bad name",
					"NAMESPACE":    "default-",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
				Namespace:     "default",
				Values: []*capiv1_protos.ProfileValues{
					{
						Namespace: "bad_namespace",
					},
				},
			},
			err: errors.New("2 errors occurred:\ninvalid namespace: default-, a lowercase RFC 1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')\ninvalid namespace: bad_namespace, a lowercase RFC 1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')"),
		},
		{
			name: "pull request failed",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("", nil, errors.New("oops"), nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				Name: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "foo",
					"NAMESPACE":    "default",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
				Namespace:     "default",
			},
			err: errors.New(`rpc error: code = Unauthenticated desc = failed to access repo https://github.com/org/repo.git: oops`),
		},
		{
			name: "create pull request",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				Name: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "foo",
					"NAMESPACE":    "default",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
				Namespace:     "default",
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "create pull request with template with comments",
			clusterState: []runtime.Object{
				readCAPITemplateFixture(t, "testdata/template-with-comments.yaml"),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				Name: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "foo",
					"NAMESPACE":    "default",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
				Namespace:     "default",
			},
			expected: "https://github.com/org/repo/pull/1",
			CommittedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/default/foo/clusters-bases-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  creationTimestamp: null
  name: clusters-bases-kustomization
  namespace: flux-system
spec:
  interval: 10m0s
  path: clusters/bases
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
status: {}
`},
				{
					Path: "clusters/my-cluster/clusters/default/foo.yaml",
					Content: `apiVersion: controlplane.cluster.x-k8s.io/v1beta1
kind: KubeadmControlPlane
metadata:
  name: "foo-control-plane"
  namespace: "default"
  annotations:
    templates.weave.works/create-request: "{\"repository_url\":\"https://github.com/org/repo.git\",\"head_branch\":\"feature-01\",\"base_branch\":\"main\",\"title\":\"New Cluster\",\"description\":\"Creates a cluster through a CAPI template\",\"name\":\"cluster-template-1\",\"parameter_values\":{\"CLUSTER_NAME\":\"foo\",\"NAMESPACE\":\"default\"},\"commit_message\":\"Add cluster manifest\",\"namespace\":\"default\",\"template_kind\":\"CAPITemplate\"}"
    templates.weave.works/created-files: "{\"files\":[\"clusters/my-cluster/clusters/default/foo.yaml\"]}"
spec:
  machineTemplate:
    infrastructureRef:
      apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
      kind: DockerMachineTemplate # {"testing": "field"}
      name: "foo-control-plane"
      namespace: "default"
  version: "1.26.1"
`,
				},
			},
		},
		{
			name: "create pull request with template with sops enabled",
			clusterState: []runtime.Object{
				readCAPITemplateFixture(t, "testdata/template-with-sops.yaml"),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				Name: "cluster-template-sops",
				ParameterValues: map[string]string{
					"CLUSTER_NAME":              "foo",
					"NAMESPACE":                 "default",
					"SOPS_KUSTOMIZATION_NAME":   "my-secrets",
					"SOPS_SECRET_REF":           "sops-gpg",
					"SOPS_SECRET_REF_NAMESPACE": "flux-system",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
				Namespace:     "default",
			},
			expected: "https://github.com/org/repo/pull/1",
			CommittedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/default/foo/sops-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  annotations:
    sops-public-key/name: sops-gpg-pub
    sops-public-key/namespace: flux-system
  creationTimestamp: null
  name: my-secrets
  namespace: flux-system
spec:
  decryption:
    provider: sops
    secretRef:
      name: sops-gpg
  interval: 10m0s
  path: clusters/default/foo/sops
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
status: {}
`,
				},
				{
					Path: "clusters/default/foo/clusters-bases-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  creationTimestamp: null
  name: clusters-bases-kustomization
  namespace: flux-system
spec:
  interval: 10m0s
  path: clusters/bases
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
status: {}
`},
				{
					Path: "clusters/my-cluster/clusters/default/foo.yaml",
					Content: `apiVersion: controlplane.cluster.x-k8s.io/v1beta1
kind: KubeadmControlPlane
metadata:
  name: "foo-control-plane"
  namespace: "default"
  annotations:
    templates.weave.works/create-request: "{\"repository_url\":\"https://github.com/org/repo.git\",\"head_branch\":\"feature-01\",\"base_branch\":\"main\",\"title\":\"New Cluster\",\"description\":\"Creates a cluster through a CAPI template\",\"name\":\"cluster-template-sops\",\"parameter_values\":{\"CLUSTER_NAME\":\"foo\",\"NAMESPACE\":\"default\",\"SOPS_KUSTOMIZATION_NAME\":\"my-secrets\",\"SOPS_SECRET_REF\":\"sops-gpg\",\"SOPS_SECRET_REF_NAMESPACE\":\"flux-system\"},\"commit_message\":\"Add cluster manifest\",\"namespace\":\"default\",\"template_kind\":\"CAPITemplate\"}"
    templates.weave.works/created-files: "{\"files\":[\"clusters/my-cluster/clusters/default/foo.yaml\"]}"
spec:
  machineTemplate:
    infrastructureRef:
      apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
      kind: DockerMachineTemplate # {"testing": "field"}
      name: "foo-control-plane"
      namespace: "default"
  version: "1.26.1"
`,
				},
			},
		},
		{
			name: "default profile values",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				Name: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "dev",
					"NAMESPACE":    "default",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
				Namespace:     "default",
				Values: []*capiv1_protos.ProfileValues{
					{
						Name:    "demo-profile",
						Version: "0.0.1",
						Values:  base64.StdEncoding.EncodeToString([]byte(``)),
					},
				},
			},
			CommittedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/my-cluster/clusters/default/dev.yaml",
					Content: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    kustomize.toolkit.fluxcd.io/prune: disabled
    templates.weave.works/create-request: "{\"repository_url\":\"https://github.com/org/repo.git\",\"head_branch\":\"feature-01\",\"base_branch\":\"main\",\"title\":\"New Cluster\",\"description\":\"Creates a cluster through a CAPI template\",\"name\":\"cluster-template-1\",\"parameter_values\":{\"CLUSTER_NAME\":\"dev\",\"NAMESPACE\":\"default\"},\"commit_message\":\"Add cluster manifest\",\"values\":[{\"name\":\"demo-profile\",\"version\":\"0.0.1\"}],\"namespace\":\"default\",\"template_kind\":\"CAPITemplate\"}"
    templates.weave.works/created-files: "{\"files\":[\"clusters/my-cluster/clusters/default/dev.yaml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: default
  name: dev
  namespace: default
`,
				},
				{
					Path: "clusters/default/dev/clusters-bases-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  creationTimestamp: null
  name: clusters-bases-kustomization
  namespace: flux-system
spec:
  interval: 10m0s
  path: clusters/bases
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
status: {}
`,
				},
				{
					Path: "clusters/default/dev/profiles.yaml",
					Content: `apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: weaveworks-charts
  namespace: default
spec:
  interval: 10m0s
  url: http://127.0.0.1:{{ .Port }}/charts
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: demo-profile
  namespace: flux-system
spec:
  chart:
    spec:
      chart: demo-profile
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: weaveworks-charts
        namespace: default
      version: 0.0.1
  install:
    crds: CreateReplace
  interval: 1m0s
  upgrade:
    crds: CreateReplace
  values: {}
status: {}
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "create pull request with path configured in template",
			clusterState: []runtime.Object{
				makeCAPITemplate(t, func(c *capiv1.CAPITemplate) {
					c.Spec.ResourceTemplates[0].Path = "clusters/foo.yml"
				}),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				Name:      "cluster-template-1",
				Namespace: "default",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "foo",
					"NAMESPACE":    "default",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
			},
			CommittedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/foo.yml",
					Content: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    kustomize.toolkit.fluxcd.io/prune: disabled
    templates.weave.works/create-request: "{\"repository_url\":\"https://github.com/org/repo.git\",\"head_branch\":\"feature-01\",\"base_branch\":\"main\",\"title\":\"New Cluster\",\"description\":\"Creates a cluster through a CAPI template\",\"name\":\"cluster-template-1\",\"parameter_values\":{\"CLUSTER_NAME\":\"foo\",\"NAMESPACE\":\"default\"},\"commit_message\":\"Add cluster manifest\",\"namespace\":\"default\",\"template_kind\":\"CAPITemplate\"}"
    templates.weave.works/created-files: "{\"files\":[\"clusters/foo.yml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: default
  name: foo
  namespace: default
`,
				},
				{
					Path: "clusters/default/foo/clusters-bases-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  creationTimestamp: null
  name: clusters-bases-kustomization
  namespace: flux-system
spec:
  interval: 10m0s
  path: clusters/bases
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
status: {}
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "specify profile namespace and cluster namespace",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				Name: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "dev",
					"NAMESPACE":    "clusters-namespace",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
				Namespace:     "default",
				Values: []*capiv1_protos.ProfileValues{
					{
						Name:      "demo-profile",
						Version:   "0.0.1",
						Values:    base64.StdEncoding.EncodeToString([]byte(``)),
						Namespace: "test-system",
					},
				},
			},
			CommittedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/my-cluster/clusters/clusters-namespace/dev.yaml",
					Content: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    kustomize.toolkit.fluxcd.io/prune: disabled
    templates.weave.works/create-request: "{\"repository_url\":\"https://github.com/org/repo.git\",\"head_branch\":\"feature-01\",\"base_branch\":\"main\",\"title\":\"New Cluster\",\"description\":\"Creates a cluster through a CAPI template\",\"name\":\"cluster-template-1\",\"parameter_values\":{\"CLUSTER_NAME\":\"dev\",\"NAMESPACE\":\"clusters-namespace\"},\"commit_message\":\"Add cluster manifest\",\"values\":[{\"name\":\"demo-profile\",\"version\":\"0.0.1\",\"namespace\":\"test-system\"}],\"namespace\":\"default\",\"template_kind\":\"CAPITemplate\"}"
    templates.weave.works/created-files: "{\"files\":[\"clusters/my-cluster/clusters/clusters-namespace/dev.yaml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: default
  name: dev
  namespace: clusters-namespace
`,
				},
				{
					Path: "clusters/clusters-namespace/dev/clusters-bases-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  creationTimestamp: null
  name: clusters-bases-kustomization
  namespace: flux-system
spec:
  interval: 10m0s
  path: clusters/bases
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
status: {}
`,
				},
				{
					Path: "clusters/clusters-namespace/dev/profiles.yaml",
					Content: `apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: weaveworks-charts
  namespace: default
spec:
  interval: 10m0s
  url: http://127.0.0.1:{{ .Port }}/charts
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: demo-profile
  namespace: flux-system
spec:
  chart:
    spec:
      chart: demo-profile
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: weaveworks-charts
        namespace: default
      version: 0.0.1
  install:
    crds: CreateReplace
    createNamespace: true
  interval: 1m0s
  targetNamespace: test-system
  upgrade:
    crds: CreateReplace
  values: {}
status: {}
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "create kustomizations",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				Name: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "dev",
					"NAMESPACE":    "clusters-namespace",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
				Namespace:     "default",
				Kustomizations: []*capiv1_protos.Kustomization{
					{
						Metadata: testNewMetadata(t, "apps-capi", "flux-system"),
						Spec: &capiv1_protos.KustomizationSpec{
							Path:            "./apps/capi",
							SourceRef:       testNewSourceRef(t, "flux-system", "flux-system"),
							TargetNamespace: "foo-ns",
						},
					},
					{
						Metadata: testNewMetadata(t, "apps-billing", "flux-system"),
						Spec: &capiv1_protos.KustomizationSpec{
							Path:      "./apps/billing",
							SourceRef: testNewSourceRef(t, "flux-system", "flux-system"),
						},
					},
				},
			},
			CommittedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/my-cluster/clusters/clusters-namespace/dev.yaml",
					Content: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    kustomize.toolkit.fluxcd.io/prune: disabled
    templates.weave.works/create-request: "{\"repository_url\":\"https://github.com/org/repo.git\",\"head_branch\":\"feature-01\",\"base_branch\":\"main\",\"title\":\"New Cluster\",\"description\":\"Creates a cluster through a CAPI template\",\"name\":\"cluster-template-1\",\"parameter_values\":{\"CLUSTER_NAME\":\"dev\",\"NAMESPACE\":\"clusters-namespace\"},\"commit_message\":\"Add cluster manifest\",\"kustomizations\":[{\"metadata\":{\"name\":\"apps-capi\",\"namespace\":\"flux-system\"},\"spec\":{\"path\":\"./apps/capi\",\"source_ref\":{\"name\":\"flux-system\",\"namespace\":\"flux-system\"},\"target_namespace\":\"foo-ns\"}},{\"metadata\":{\"name\":\"apps-billing\",\"namespace\":\"flux-system\"},\"spec\":{\"path\":\"./apps/billing\",\"source_ref\":{\"name\":\"flux-system\",\"namespace\":\"flux-system\"}}}],\"namespace\":\"default\",\"template_kind\":\"CAPITemplate\"}"
    templates.weave.works/created-files: "{\"files\":[\"clusters/my-cluster/clusters/clusters-namespace/dev.yaml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: default
  name: dev
  namespace: clusters-namespace
`,
				},
				{
					Path: "clusters/clusters-namespace/dev/clusters-bases-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  creationTimestamp: null
  name: clusters-bases-kustomization
  namespace: flux-system
spec:
  interval: 10m0s
  path: clusters/bases
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
status: {}
`,
				},
				{
					Path: "clusters/clusters-namespace/dev/apps-capi-flux-system-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  creationTimestamp: null
  name: apps-capi
  namespace: flux-system
spec:
  interval: 10m0s
  path: ./apps/capi
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
    namespace: flux-system
  targetNamespace: foo-ns
status: {}
`,
				},
				{
					Path: "clusters/clusters-namespace/dev/apps-billing-flux-system-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  creationTimestamp: null
  name: apps-billing
  namespace: flux-system
spec:
  interval: 10m0s
  path: ./apps/billing
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
    namespace: flux-system
status: {}
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "kustomizations validation errors",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				Name: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "dev",
					"NAMESPACE":    "clusters-namespace",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
				Namespace:     "default",
				Kustomizations: []*capiv1_protos.Kustomization{
					{
						Metadata: testNewMetadata(t, "", "@kustomization"),
						Spec: &capiv1_protos.KustomizationSpec{
							Path:      "./apps/capi",
							SourceRef: testNewSourceRef(t, "flux-system", "flux-system"),
						},
					},
					{
						Metadata: testNewMetadata(t, "apps-capi", "flux-system"),
						Spec: &capiv1_protos.KustomizationSpec{
							Path:      "./apps/capi",
							SourceRef: testNewSourceRef(t, "", ""),
						},
					},
				},
			},
			err: errors.New("3 errors occurred:\nkustomization name must be specified\ninvalid namespace: @kustomization, a lowercase RFC 1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')\nsourceRef name must be specified in Kustomization apps-capi"),
		},
		{
			name: "kustomization with metadata is nil",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				Name: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "dev",
					"NAMESPACE":    "clusters-namespace",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
				Namespace:     "default",
				Kustomizations: []*capiv1_protos.Kustomization{
					{
						Spec: &capiv1_protos.KustomizationSpec{
							Path:      "./apps/capi",
							SourceRef: testNewSourceRef(t, "flux-system", "flux-system"),
						},
					},
				},
			},
			err: errors.New("kustomization metadata must be specified"),
		},
		{
			name: "Edit cluster, remove kustomization",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				Name: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "dev",
					"NAMESPACE":    "clusters-namespace",
				},
				RepositoryUrl:  "https://github.com/org/repo.git",
				HeadBranch:     "feature-01",
				BaseBranch:     "main",
				Title:          "Edit Cluster",
				Description:    "Delete kustomization from cluster",
				CommitMessage:  "Edits dev",
				Namespace:      "default",
				Kustomizations: []*capiv1_protos.Kustomization{},

				PreviousValues: &capiv1_protos.PreviousValues{
					ParameterValues: map[string]string{
						"CLUSTER_NAME": "dev",
						"NAMESPACE":    "clusters-namespace",
					},
					Kustomizations: []*capiv1_protos.Kustomization{
						{
							Metadata: testNewMetadata(t, "apps-capi", "flux-system"),
							Spec: &capiv1_protos.KustomizationSpec{
								Path:            "./apps/capi",
								SourceRef:       testNewSourceRef(t, "flux-system", "flux-system"),
								TargetNamespace: "foo-ns",
							},
						},
					},
					Credentials: &capiv1_protos.Credential{},
				},
			},
			CommittedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/my-cluster/clusters/clusters-namespace/dev.yaml",
					Content: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    kustomize.toolkit.fluxcd.io/prune: disabled
    templates.weave.works/create-request: "{\"repository_url\":\"https://github.com/org/repo.git\",\"head_branch\":\"feature-01\",\"base_branch\":\"main\",\"title\":\"Edit Cluster\",\"description\":\"Delete kustomization from cluster\",\"name\":\"cluster-template-1\",\"parameter_values\":{\"CLUSTER_NAME\":\"dev\",\"NAMESPACE\":\"clusters-namespace\"},\"commit_message\":\"Edits dev\",\"namespace\":\"default\",\"template_kind\":\"CAPITemplate\"}"
    templates.weave.works/created-files: "{\"files\":[\"clusters/my-cluster/clusters/clusters-namespace/dev.yaml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: default
  name: dev
  namespace: clusters-namespace
`,
				},
				{
					Path: "clusters/clusters-namespace/dev/clusters-bases-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  creationTimestamp: null
  name: clusters-bases-kustomization
  namespace: flux-system
spec:
  interval: 10m0s
  path: clusters/bases
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
status: {}
`,
				},
				{
					Path:    "clusters/clusters-namespace/dev/apps-capi-flux-system-kustomization.yaml",
					Content: "",
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "Edit cluster namespace",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				Name: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "dev",
					"NAMESPACE":    "clusters-namespace-2",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "Edit Cluster",
				Description:   "Edit namespace",
				CommitMessage: "Edits dev",
				Namespace:     "default",
				Kustomizations: []*capiv1_protos.Kustomization{
					{
						Metadata: testNewMetadata(t, "apps-capi", "flux-system"),
						Spec: &capiv1_protos.KustomizationSpec{
							Path:            "./apps/capi",
							SourceRef:       testNewSourceRef(t, "flux-system", "flux-system"),
							TargetNamespace: "foo-ns",
						},
					},
				},

				PreviousValues: &capiv1_protos.PreviousValues{
					ParameterValues: map[string]string{
						"CLUSTER_NAME": "dev",
						"NAMESPACE":    "clusters-namespace",
					},
					Kustomizations: []*capiv1_protos.Kustomization{
						{
							Metadata: testNewMetadata(t, "apps-capi", "flux-system"),
							Spec: &capiv1_protos.KustomizationSpec{
								Path:            "./apps/capi",
								SourceRef:       testNewSourceRef(t, "flux-system", "flux-system"),
								TargetNamespace: "foo-ns",
							},
						},
					},
					Credentials: &capiv1_protos.Credential{},
				},
			},
			CommittedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/clusters-namespace-2/dev/apps-capi-flux-system-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  creationTimestamp: null
  name: apps-capi
  namespace: flux-system
spec:
  interval: 10m0s
  path: ./apps/capi
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
    namespace: flux-system
  targetNamespace: foo-ns
status: {}
`,
				},
				{
					Path: "clusters/clusters-namespace-2/dev/clusters-bases-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  creationTimestamp: null
  name: clusters-bases-kustomization
  namespace: flux-system
spec:
  interval: 10m0s
  path: clusters/bases
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
status: {}
`,
				},
				{
					Path:    "clusters/clusters-namespace/dev/apps-capi-flux-system-kustomization.yaml",
					Content: "",
				},
				{
					Path:    "clusters/clusters-namespace/dev/clusters-bases-kustomization.yaml",
					Content: "",
				},
				{
					Path:    "clusters/my-cluster/clusters/clusters-namespace/dev.yaml",
					Content: "",
				},
				{
					Path: "clusters/my-cluster/clusters/clusters-namespace-2/dev.yaml",
					Content: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    kustomize.toolkit.fluxcd.io/prune: disabled
    templates.weave.works/create-request: "{\"repository_url\":\"https://github.com/org/repo.git\",\"head_branch\":\"feature-01\",\"base_branch\":\"main\",\"title\":\"Edit Cluster\",\"description\":\"Edit namespace\",\"name\":\"cluster-template-1\",\"parameter_values\":{\"CLUSTER_NAME\":\"dev\",\"NAMESPACE\":\"clusters-namespace-2\"},\"commit_message\":\"Edits dev\",\"kustomizations\":[{\"metadata\":{\"name\":\"apps-capi\",\"namespace\":\"flux-system\"},\"spec\":{\"path\":\"./apps/capi\",\"source_ref\":{\"name\":\"flux-system\",\"namespace\":\"flux-system\"},\"target_namespace\":\"foo-ns\"}}],\"namespace\":\"default\",\"template_kind\":\"CAPITemplate\"}"
    templates.weave.works/created-files: "{\"files\":[\"clusters/my-cluster/clusters/clusters-namespace-2/dev.yaml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: default
  name: dev
  namespace: clusters-namespace-2
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "Edit cluster namespace and kustomization name",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				Name: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "dev",
					"NAMESPACE":    "clusters-namespace-2",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "Edit Cluster",
				Description:   "Edit namespace",
				CommitMessage: "Edits dev",
				Namespace:     "default",
				Kustomizations: []*capiv1_protos.Kustomization{
					{
						Metadata: testNewMetadata(t, "apps-capi-2", "flux-system"),
						Spec: &capiv1_protos.KustomizationSpec{
							Path:            "./apps/capi",
							SourceRef:       testNewSourceRef(t, "flux-system", "flux-system"),
							TargetNamespace: "foo-ns",
						},
					},
				},

				PreviousValues: &capiv1_protos.PreviousValues{
					ParameterValues: map[string]string{
						"CLUSTER_NAME": "dev",
						"NAMESPACE":    "clusters-namespace",
					},
					Kustomizations: []*capiv1_protos.Kustomization{
						{
							Metadata: testNewMetadata(t, "apps-capi", "flux-system"),
							Spec: &capiv1_protos.KustomizationSpec{
								Path:            "./apps/capi",
								SourceRef:       testNewSourceRef(t, "flux-system", "flux-system"),
								TargetNamespace: "foo-ns",
							},
						},
					},
					Credentials: &capiv1_protos.Credential{},
				},
			},
			CommittedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/clusters-namespace-2/dev/apps-capi-2-flux-system-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  creationTimestamp: null
  name: apps-capi-2
  namespace: flux-system
spec:
  interval: 10m0s
  path: ./apps/capi
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
    namespace: flux-system
  targetNamespace: foo-ns
status: {}
`,
				},
				{
					Path: "clusters/clusters-namespace-2/dev/clusters-bases-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  creationTimestamp: null
  name: clusters-bases-kustomization
  namespace: flux-system
spec:
  interval: 10m0s
  path: clusters/bases
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
status: {}
`,
				},
				{
					Path:    "clusters/clusters-namespace/dev/apps-capi-flux-system-kustomization.yaml",
					Content: "",
				},
				{
					Path:    "clusters/clusters-namespace/dev/clusters-bases-kustomization.yaml",
					Content: "",
				},
				{
					Path:    "clusters/my-cluster/clusters/clusters-namespace/dev.yaml",
					Content: "",
				},
				{
					Path: "clusters/my-cluster/clusters/clusters-namespace-2/dev.yaml",
					Content: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    kustomize.toolkit.fluxcd.io/prune: disabled
    templates.weave.works/create-request: "{\"repository_url\":\"https://github.com/org/repo.git\",\"head_branch\":\"feature-01\",\"base_branch\":\"main\",\"title\":\"Edit Cluster\",\"description\":\"Edit namespace\",\"name\":\"cluster-template-1\",\"parameter_values\":{\"CLUSTER_NAME\":\"dev\",\"NAMESPACE\":\"clusters-namespace-2\"},\"commit_message\":\"Edits dev\",\"kustomizations\":[{\"metadata\":{\"name\":\"apps-capi-2\",\"namespace\":\"flux-system\"},\"spec\":{\"path\":\"./apps/capi\",\"source_ref\":{\"name\":\"flux-system\",\"namespace\":\"flux-system\"},\"target_namespace\":\"foo-ns\"}}],\"namespace\":\"default\",\"template_kind\":\"CAPITemplate\"}"
    templates.weave.works/created-files: "{\"files\":[\"clusters/my-cluster/clusters/clusters-namespace-2/dev.yaml\"]}"
  labels:
    templates.weave.works/template-name: cluster-template-1
    templates.weave.works/template-namespace: default
  name: dev
  namespace: clusters-namespace-2
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			ts := httptest.NewServer(makeServeMux(t))
			defer ts.Close()
			hr := makeTestHelmRepository(ts.URL, func(hr *sourcev1.HelmRepository) {
				hr.Name = "weaveworks-charts"
				hr.Namespace = "default"
			})
			tt.clusterState = append(tt.clusterState, hr)
			fakeCache := testNewFakeChartCache(t,
				nsn("management", ""),
				helm.ObjectReference{
					Name:      "weaveworks-charts",
					Namespace: "default",
				},
				[]helm.Chart{})

			s := createServer(t, serverOptions{
				profileHelmRepository: &types.NamespacedName{Name: "weaveworks-charts", Namespace: "default"},
				clusterState:          tt.clusterState,
				namespace:             "default",
				provider:              tt.provider,
				chartsCache:           fakeCache,
			})

			// request
			createPullRequestResponse, err := s.CreatePullRequest(context.Background(), tt.req)

			// Check the response looks good
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to create a pull request:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("got the wrong error:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, createPullRequestResponse.WebUrl, protocmp.Transform()); diff != "" {
					t.Fatalf("pull request url didn't match expected:\n%s", diff)
				}
				fakeGitProvider := (tt.provider).(*gitfakes.FakeGitProvider)

				if diff := cmp.Diff(prepCommitedFiles(t, ts.URL, tt.CommittedFiles), fakeGitProvider.GetCommittedFiles(), protocmp.Transform()); len(tt.CommittedFiles) > 0 && diff != "" {
					t.Fatalf("committed files do not match expected committed files:\n%s", diff)
				}
			}
		})
	}
}

func sortCommitFiles(files []*capiv1_protos.CommitFile) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
}

func prepCommitedFiles(t *testing.T, serverUrl string, files []*capiv1_protos.CommitFile) []*capiv1_protos.CommitFile {
	parsedURL, err := url.Parse(serverUrl)
	if err != nil {
		t.Fatalf("failed to parse URL %s", err)
	}
	newFiles := []*capiv1_protos.CommitFile{}
	for _, f := range files {
		newFiles = append(newFiles, &capiv1_protos.CommitFile{
			Path:    f.Path,
			Content: simpleTemplate(t, f.Content, struct{ Port string }{Port: parsedURL.Port()}),
		})
	}
	sortCommitFiles(newFiles)
	return newFiles
}

func simpleTemplate(t *testing.T, templateString string, data interface{}) string {
	tm := template.Must(template.New("its-a-template").Parse(templateString))
	var tpl bytes.Buffer
	if err := tm.Execute(&tpl, data); err != nil {
		t.Fatalf("failed to template %s", err)
	}
	return tpl.String()
}

func TestGetKubeconfig(t *testing.T) {
	testCases := []struct {
		name                    string
		client                  client.Client
		clusterState            []runtime.Object
		clusterObjectsNamespace string // Namespace that cluster objects are created in
		req                     *capiv1_protos.GetKubeconfigRequest
		ctx                     context.Context
		expected                []byte
		wantErr                 string
	}{
		{
			name: "get kubeconfig as JSON",
			clusterState: []runtime.Object{
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "dev"
					o.ObjectMeta.Namespace = "default"
				}),
				makeSecret("dev-kubeconfig", "default", "value.yaml", "foo"),
			},
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.MD{}),
			expected: []byte("foo"),
		},
		{
			name: "get kubeconfig as binary",
			clusterState: []runtime.Object{
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "dev"
					o.ObjectMeta.Namespace = "default"
				}),
				makeSecret("dev-kubeconfig", "default", "value", "foo"),
			},
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.Pairs("accept", "application/octet-stream")),
			expected: []byte("foo"),
		},
		{
			name: "secret not found",
			clusterState: []runtime.Object{
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "dev"
					o.ObjectMeta.Namespace = "testing"
				}),
			},
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName:      "dev",
				ClusterNamespace: "testing",
			},
			wantErr: "unable to get kubeconfig secret for cluster testing/dev",
		},
		{
			name: "secret found but is missing key",
			clusterState: []runtime.Object{
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "dev"
					o.ObjectMeta.Namespace = "default"
				}),
				makeSecret("dev-kubeconfig", "default", "val", "foo"),
			},
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			wantErr: `secret "default/dev-kubeconfig" was found but is missing key "value"`,
		},
		{
			name: "use cluster_namespace to get secret",
			clusterState: []runtime.Object{
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "dev"
					o.ObjectMeta.Namespace = "kube-system"
				}),
				makeSecret("dev-kubeconfig", "kube-system", "value", "foo"),
			},
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName:      "dev",
				ClusterNamespace: "kube-system",
			},
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.MD{}),
			expected: []byte("foo"),
		},
		{
			name: "no namespace and lookup across namespaces, use default namespace",
			clusterState: []runtime.Object{
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "dev"
					o.ObjectMeta.Namespace = "default"
				}),
				makeSecret("dev-kubeconfig", "default", "value", "foo"),
			},
			clusterObjectsNamespace: "",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.MD{}),
			expected: []byte("foo"),
		},
		{
			name: "user kubeconfig exists",
			clusterState: []runtime.Object{
				makeSecret("dev-kubeconfig", "default", "value.yaml", "foo"),
				makeSecret("dev-user-kubeconfig", "default", "value.yaml", "bar"),
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "dev"
					o.ObjectMeta.Namespace = "default"
				}),
			},
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.MD{}),
			expected: []byte("bar"),
		},
		{
			name: "no access to additional secrets",
			client: fake.NewClientBuilder().WithScheme(newTestScheme(t)).WithObjectTracker(&mockTracker{
				getImpl: func(gvr schema.GroupVersionResource, ns, name string) (runtime.Object, error) {
					if name == "dev" && ns == "default" {
						return makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
							o.ObjectMeta.Name = "dev"
							o.ObjectMeta.Namespace = "default"
						}), nil
					}
					if name == "dev-user-kubeconfig" && ns == "default" {
						return makeSecret("dev-user-kubeconfig", "default", "value.yaml", "bar"), nil
					}

					return nil, apierrors.NewForbidden(gvr.GroupResource(), name, errors.New("forbidden"))

				}}).Build(),
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.MD{}),
			expected: []byte("bar"),
		},
		{
			name: "kubeconfig override exists merges with existing secret",
			clusterState: []runtime.Object{
				makeSecret("dev-kubeconfig", "default", "value.yaml", `{"apiVersion":"v1","clusters":[{"cluster":{"certificate-authority-data":"anVzdCB0ZXN0aW5n","server":"https://example.com"},"name":"example-com"}],"contexts":[{"context":{"cluster":"example-com","user":"example-com-admin"},"name":"example-com"}],"current-context":"example-com","kind":"Config","preferences":{},"users":[{"name":"example-com-admin","user":{"token":"token"}}]}`),
				makeSecret("cluster-kubeconfig-override", "default", "value.yaml", `
apiVersion: v1
kind: Config
users:
- name: example-com-overridden-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      args:
      - oidc-login
      - get-token
      - --oidc-issuer-url=https://auth.w3ops.net
      - --oidc-client-id=k8s-leaf-cluster-auth
      - --oidc-client-secret=<redacted>
      - --oidc-extra-scope=groups
      - --oidc-extra-scope=profile
      - --oidc-extra-scope=email
      command: kubectl
`),
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "dev"
					o.ObjectMeta.Namespace = "default"
				}),
			},
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			ctx: metadata.NewIncomingContext(context.Background(), metadata.MD{}),
			expected: []byte(`apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: anVzdCB0ZXN0aW5n
    server: https://example.com
  name: example-com
contexts:
- context:
    cluster: example-com
    user: example-com-overridden-user
  name: example-com
current-context: example-com
kind: Config
preferences: {}
users:
- name: example-com-overridden-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      args:
      - oidc-login
      - get-token
      - --oidc-issuer-url=https://auth.w3ops.net
      - --oidc-client-id=k8s-leaf-cluster-auth
      - --oidc-client-secret=<redacted>
      - --oidc-extra-scope=groups
      - --oidc-extra-scope=profile
      - --oidc-extra-scope=email
      command: kubectl
      env: null
      interactiveMode: IfAvailable
      provideClusterInfo: false
`),
		},
		{
			name: "gitops cluster references secret",
			clusterState: []runtime.Object{
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "gitops-cluster"
					o.ObjectMeta.Namespace = "default"
					o.Spec.SecretRef = &meta.LocalObjectReference{
						Name: "just-a-test-config",
					}
				}),
				makeSecret("just-a-test-config", "default", "value.yaml", "foo"),
			},
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "gitops-cluster",
			},
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.MD{}),
			expected: []byte("foo"),
		},
		{
			name: "gitops cluster references non-existent secret",
			clusterState: []runtime.Object{
				makeTestGitopsCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = "gitops-cluster"
					o.ObjectMeta.Namespace = "default"
					o.Spec.SecretRef = &meta.LocalObjectReference{
						Name: "just-a-test-config",
					}
				}),
			},
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "gitops-cluster",
			},
			ctx:     metadata.NewIncomingContext(context.Background(), metadata.MD{}),
			wantErr: "failed to load referenced secret default/just-a-test-config for cluster default/gitops-cluster",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			setViperWithTestCleanup(t, map[string]string{"capi-clusters-namespace": tt.clusterObjectsNamespace})

			s := createServer(t, serverOptions{
				clusterState: tt.clusterState,
				client:       tt.client,
				namespace:    "default",
				ns:           tt.clusterObjectsNamespace,
			})

			res, err := s.GetKubeconfig(tt.ctx, tt.req)

			if err != nil {
				if tt.wantErr == "" {
					t.Fatalf("failed to get the kubeconfig: %s", err)
				}
				if msg := err.Error(); msg != tt.wantErr {
					t.Fatalf("got error %q, want %q", msg, tt.wantErr)
				}
				return
			}

			var data []byte
			if res != nil {
				data = res.Data
				// If we received a JSON response, parse it and extract the
				// kubeconfig for comparison with the desired result.
				if res.ContentType == "application/json" {
					var body map[string]any

					if err := json.Unmarshal(res.Data, &body); err != nil {
						t.Fatal(err)
					}

					if data, err = base64.StdEncoding.DecodeString(body["kubeconfig"].(string)); err != nil {
						t.Fatal(err)
					}
				}
			}

			if diff := cmp.Diff(string(tt.expected), string(data)); diff != "" {
				t.Fatalf("kubeconfig didn't match expected:\n%s", diff)
			}
		})
	}
}

func TestDeleteClustersPullRequest(t *testing.T) {
	setViperWithTestCleanup(t, map[string]string{
		"capi-repository-path":          "clusters/management/clusters",
		"capi-repository-clusters-path": "clusters/",
	})

	testCases := []struct {
		name           string
		provider       csgit.Provider
		req            *capiv1_protos.CreateDeletionPullRequestRequest
		CommittedFiles []*capiv1_protos.CommitFile
		expected       string
		err            error
	}{
		{
			name: "validation errors",
			req:  &capiv1_protos.CreateDeletionPullRequestRequest{},
			err:  errors.New(deleteClustersRequiredErr),
		},

		{
			name:     "create delete pull request",
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateDeletionPullRequestRequest{
				ClusterNames:  []string{"foo", "bar"},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-02",
				BaseBranch:    "feature-01",
				Title:         "Delete Cluster",
				Description:   "Deletes a cluster",
				CommitMessage: "Remove cluster manifest",
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "create delete pull request including multiple files in tree",
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, []string{
				"clusters/default/foo/kustomization.yaml",
				"clusters/management/clusters/default/foo.yaml",
			}, nil),
			req: &capiv1_protos.CreateDeletionPullRequestRequest{
				ClusterNames:  []string{"foo"},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-02",
				BaseBranch:    "feature-01",
				Title:         "Delete Cluster",
				Description:   "Deletes a cluster",
				CommitMessage: "Remove cluster manifest",
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name:     "create delete pull request with namespaced cluster names",
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateDeletionPullRequestRequest{
				ClusterNamespacedNames: []*capiv1_protos.ClusterNamespacedName{
					testNewClusterNamespacedName(t, "foo", "ns-foo"),
					testNewClusterNamespacedName(t, "bar", "ns-bar"),
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-02",
				BaseBranch:    "feature-01",
				Title:         "Delete Cluster",
				Description:   "Deletes a cluster",
				CommitMessage: "Remove cluster manifest",
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "create delete pull request with namespaced cluster names including multiple files in tree",
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, []string{
				"clusters/ns-foo/foo/kustomization.yaml",
				"clusters/management/clusters/ns-foo/foo.yaml",
			}, nil),
			req: &capiv1_protos.CreateDeletionPullRequestRequest{
				ClusterNamespacedNames: []*capiv1_protos.ClusterNamespacedName{
					{
						Name:      "foo",
						Namespace: "ns-foo",
					},
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-02",
				BaseBranch:    "feature-01",
				Title:         "Delete Cluster",
				Description:   "Deletes a cluster",
				CommitMessage: "Remove cluster files",
			},
			expected: "https://github.com/org/repo/pull/1",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			s := createServer(t, serverOptions{
				namespace: "default",
				provider:  tt.provider,
			})

			// delete request
			deletePullRequestResponse, err := s.CreateDeletionPullRequest(context.Background(), tt.req)

			// Check the response looks good
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to create a pull request:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("got the wrong error:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, deletePullRequestResponse.WebUrl, protocmp.Transform()); diff != "" {
					t.Fatalf("pull request url didn't match expected:\n%s", diff)
				}
				fakeGitProvider := (tt.provider).(*gitfakes.FakeGitProvider)

				if fakeGitProvider.OriginalFiles != nil {
					// sort CommittedFiles and OriginalFiles for comparison
					sort.Slice(fakeGitProvider.CommittedFiles[:], func(i, j int) bool {
						currFile := fakeGitProvider.CommittedFiles[i].Path
						nextFile := fakeGitProvider.CommittedFiles[j].Path
						return currFile < nextFile
					})
					sort.Strings(fakeGitProvider.OriginalFiles)

					if len(fakeGitProvider.CommittedFiles) != len(fakeGitProvider.OriginalFiles) {
						t.Fatalf("number of committed files (%d) do not match number of expected files (%d)\n", len(fakeGitProvider.CommittedFiles), len(fakeGitProvider.OriginalFiles))
					}
					for ind, committedFile := range fakeGitProvider.CommittedFiles {
						if committedFile.Path != fakeGitProvider.OriginalFiles[ind] {
							t.Fatalf("committed file does not match expected file\n%v\n%v", committedFile.Path, fakeGitProvider.OriginalFiles[ind])

						}
					}
				}
			}
		})
	}
}

func makeNamespace(n string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: n,
		},
	}
}

func makeSecret(n string, ns string, s ...string) *corev1.Secret {
	data := make(map[string][]byte)
	for i := 0; i < len(s); i += 2 {
		data[s[i]] = []byte(s[i+1])
	}

	nsObj := makeNamespace(ns)

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: nsObj.GetName(),
		},
		Data: data,
	}
}

func makeTestGitopsCluster(opts ...func(*gitopsv1alpha1.GitopsCluster)) *gitopsv1alpha1.GitopsCluster {
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

func makeServeMux(t *testing.T, opts ...func(*repo.IndexFile)) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/charts/index.yaml", func(w http.ResponseWriter, req *http.Request) {
		b, err := yaml.Marshal(makeTestChartIndex(opts...))
		if err != nil {
			t.Fatal(err)
		}

		_, err = w.Write(b)
		if err != nil {
			t.Fatal(err)
		}
	})
	mux.Handle("/", http.FileServer(http.Dir("../charts/testdata")))
	return mux
}

func makeTestChartIndex(opts ...func(*repo.IndexFile)) *repo.IndexFile {
	ri := &repo.IndexFile{
		APIVersion: "v1",
		Entries: map[string]repo.ChartVersions{
			"demo-profile": {
				{
					Metadata: &chart.Metadata{
						Annotations: map[string]string{
							charts.ProfileAnnotation: "demo-profile",
						},
						Description: "Simple demo profile",
						Home:        "https://example.com/testing",
						Name:        "demo-profile",
						Sources: []string{
							"https://example.com/testing",
						},
						Version: "0.0.1",
					},
					Created: time.Now(),
					Digest:  "aaff4545f79d8b2913a10cb400ebb6fa9c77fe813287afbacf1a0b897cdffffff",
					URLs: []string{
						"/charts/demo-profile-0.1.0.tgz",
					},
				},
			},
		},
	}
	for _, o := range opts {
		o(ri)
	}
	return ri
}

func TestGenerateProfileFiles(t *testing.T) {
	fakeCache := testNewFakeChartCache(t,
		nsn("cluster-foo", "ns-foo"),
		helm.ObjectReference{
			Name:      "testing",
			Namespace: "test-ns",
		},
		[]helm.Chart{})
	files, err := generateProfileFiles(
		context.TODO(),
		makeTestTemplate(templatesv1.RenderTypeEnvsubst),
		nsn("cluster-foo", "ns-foo"),
		makeTestHelmRepositoryTemplate("base"),
		generateProfileFilesParams{
			helmRepositoryCluster: types.NamespacedName{Name: "cluster-foo", Namespace: "ns-foo"},
			helmRepository:        nsn("testing", "test-ns"),
			profileValues: []*capiv1_protos.ProfileValues{
				{
					Name:    "foo",
					Version: "0.0.1",
					Values:  base64.StdEncoding.EncodeToString([]byte("foo: bar")),
				},
			},
			parameterValues: map[string]string{},
			chartsCache:     fakeCache,
		},
	)
	assert.NoError(t, err)
	expected := []git.CommitFile{
		makeCommitFile(
			"ns-foo/cluster-foo/profiles.yaml",
			`apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: testing
  namespace: test-ns
spec:
  interval: 10m0s
  url: base/charts
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: foo
  namespace: flux-system
spec:
  chart:
    spec:
      chart: foo
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: testing
        namespace: test-ns
      version: 0.0.1
  install:
    crds: CreateReplace
  interval: 1m0s
  upgrade:
    crds: CreateReplace
  values:
    foo: bar
status: {}
`,
		),
	}
	assert.Equal(t, expected, files)
}

func TestGenerateProfileFiles_without_editable_flag(t *testing.T) {
	fakeCache := testNewFakeChartCache(t,
		nsn("management", ""),
		helm.ObjectReference{
			Name:      "testing",
			Namespace: "test-ns",
		},
		[]helm.Chart{})
	files, err := generateProfileFiles(
		context.TODO(),
		makeTestTemplateWithProfileAnnotation(
			templatesv1.RenderTypeEnvsubst,
			"capi.weave.works/profile-0",
			"{\"name\": \"foo\", \"version\": \"0.0.1\", \"values\": \"foo: defaultFoo\" }",
		),
		nsn("cluster-foo", "ns-foo"),
		makeTestHelmRepositoryTemplate("base"),
		generateProfileFilesParams{
			helmRepository:        nsn("testing", "test-ns"),
			helmRepositoryCluster: types.NamespacedName{Name: "management"},
			profileValues: []*capiv1_protos.ProfileValues{
				{
					Name:    "foo",
					Version: "0.0.1",
					Values:  base64.StdEncoding.EncodeToString([]byte("foo: bar")),
				},
			},
			parameterValues: map[string]string{},
			chartsCache:     fakeCache,
		},
	)
	require.NoError(t, err)
	expected := []git.CommitFile{
		makeCommitFile(
			"ns-foo/cluster-foo/profiles.yaml",
			`apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: testing
  namespace: test-ns
spec:
  interval: 10m0s
  url: base/charts
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: foo
  namespace: flux-system
spec:
  chart:
    spec:
      chart: foo
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: testing
        namespace: test-ns
      version: 0.0.1
  install:
    crds: CreateReplace
  interval: 1m0s
  upgrade:
    crds: CreateReplace
  values:
    foo: defaultFoo
status: {}
`),
	}
	assert.Equal(t, expected, files)
}

func TestGenerateProfileFiles_with_editable_flag(t *testing.T) {
	fakeCache := testNewFakeChartCache(t,
		nsn("management", ""),
		helm.ObjectReference{
			Name:      "testing",
			Namespace: "test-ns",
		},
		[]helm.Chart{})
	files, err := generateProfileFiles(
		context.TODO(),
		makeTestTemplateWithProfileAnnotation(
			templatesv1.RenderTypeEnvsubst,
			"capi.weave.works/profile-0",
			"{\"name\": \"foo\", \"version\": \"0.0.1\", \"values\": \"foo: defaultFoo\", \"editable\": true }",
		),
		nsn("management", ""),
		makeTestHelmRepositoryTemplate("base"),
		generateProfileFilesParams{
			helmRepository:        nsn("testing", "test-ns"),
			helmRepositoryCluster: types.NamespacedName{Name: "management"},
			profileValues: []*capiv1_protos.ProfileValues{
				{
					Name:    "foo",
					Version: "0.0.1",
					Values:  base64.StdEncoding.EncodeToString([]byte("foo: bar")),
				},
			},
			parameterValues: map[string]string{},
			chartsCache:     fakeCache,
		},
	)
	require.NoError(t, err)
	expected := []git.CommitFile{
		makeCommitFile(
			"management/profiles.yaml",
			`apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: testing
  namespace: test-ns
spec:
  interval: 10m0s
  url: base/charts
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: foo
  namespace: flux-system
spec:
  chart:
    spec:
      chart: foo
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: testing
        namespace: test-ns
      version: 0.0.1
  install:
    crds: CreateReplace
  interval: 1m0s
  upgrade:
    crds: CreateReplace
  values:
    foo: bar
status: {}
`),
	}

	assert.Equal(t, expected, files)
}
func TestGenerateProfileFiles_with_templates(t *testing.T) {
	fakeCache := testNewFakeChartCache(t,
		nsn("management", ""),
		helm.ObjectReference{
			Name:      "testing",
			Namespace: "test-ns",
		},
		[]helm.Chart{})
	params := map[string]string{
		"CLUSTER_NAME": "test-cluster-name",
		"NAMESPACE":    "default",
	}

	files, err := generateProfileFiles(
		context.TODO(),
		makeTestTemplate(templatesv1.RenderTypeEnvsubst),
		nsn("cluster-foo", "ns-foo"),
		makeTestHelmRepositoryTemplate("base"),
		generateProfileFilesParams{
			helmRepository:        nsn("testing", "test-ns"),
			helmRepositoryCluster: types.NamespacedName{Name: "management"},
			profileValues: []*capiv1_protos.ProfileValues{
				{
					Name:    "foo",
					Version: "0.0.1",
					Values:  base64.StdEncoding.EncodeToString([]byte("foo: ${CLUSTER_NAME}")),
				},
			},
			parameterValues: params,
			chartsCache:     fakeCache,
		},
	)
	assert.NoError(t, err)
	expected := []git.CommitFile{
		makeCommitFile(
			"ns-foo/cluster-foo/profiles.yaml",
			`apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: testing
  namespace: test-ns
spec:
  interval: 10m0s
  url: base/charts
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: foo
  namespace: flux-system
spec:
  chart:
    spec:
      chart: foo
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: testing
        namespace: test-ns
      version: 0.0.1
  install:
    crds: CreateReplace
  interval: 1m0s
  upgrade:
    crds: CreateReplace
  values:
    foo: test-cluster-name
status: {}
`),
	}
	assert.Equal(t, expected, files)
}

func TestGenerateProfileFilesWithLayers(t *testing.T) {
	fakeCache := testNewFakeChartCache(t,
		nsn("management", ""),
		helm.ObjectReference{
			Name:      "testing",
			Namespace: "test-ns",
		},
		[]helm.Chart{})
	files, err := generateProfileFiles(
		context.TODO(),
		makeTestTemplate(templatesv1.RenderTypeEnvsubst),
		nsn("cluster-foo", "ns-foo"),
		makeTestHelmRepositoryTemplate("base"),
		generateProfileFilesParams{
			helmRepository:        nsn("testing", "test-ns"),
			helmRepositoryCluster: types.NamespacedName{Name: "management"},
			profileValues: []*capiv1_protos.ProfileValues{
				{
					Name:    "foo",
					Version: "0.0.1",
					Values:  base64.StdEncoding.EncodeToString([]byte("foo: bar")),
				},
				{
					Name:    "bar",
					Version: "0.0.1",
					Layer:   "testing",
					Values:  base64.StdEncoding.EncodeToString([]byte("foo: bar")),
				},
			},
			parameterValues: map[string]string{},
			chartsCache:     fakeCache,
		},
	)
	assert.NoError(t, err)
	expected := []git.CommitFile{
		makeCommitFile(
			"ns-foo/cluster-foo/profiles.yaml",
			`apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: testing
  namespace: test-ns
spec:
  interval: 10m0s
  url: base/charts
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  labels:
    weave.works/applied-layer: testing
  name: bar
  namespace: flux-system
spec:
  chart:
    spec:
      chart: bar
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: testing
        namespace: test-ns
      version: 0.0.1
  install:
    crds: CreateReplace
  interval: 1m0s
  upgrade:
    crds: CreateReplace
  values:
    foo: bar
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: foo
  namespace: flux-system
spec:
  chart:
    spec:
      chart: foo
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: testing
        namespace: test-ns
      version: 0.0.1
  dependsOn:
  - name: bar
  install:
    crds: CreateReplace
  interval: 1m0s
  upgrade:
    crds: CreateReplace
  values:
    foo: bar
status: {}
`),
	}
	assert.Equal(t, expected, files)
}

func TestGenerateProfileFiles_with_text_templates(t *testing.T) {
	fakeCache := testNewFakeChartCache(t,
		nsn("management", ""),
		helm.ObjectReference{
			Name:      "testing",
			Namespace: "test-ns",
		},
		[]helm.Chart{})
	params := map[string]string{
		"CLUSTER_NAME": "test-cluster-name",
		"NAMESPACE":    "default",
		"TEST_PARAM":   "this-is-a-test",
	}

	files, err := generateProfileFiles(
		context.TODO(),
		makeTestTemplate(templatesv1.RenderTypeTemplating),
		nsn("cluster-foo", "ns-foo"),
		makeTestHelmRepositoryTemplate("base"),
		generateProfileFilesParams{
			helmRepository:        nsn("testing", "test-ns"),
			helmRepositoryCluster: types.NamespacedName{Name: "management"},
			profileValues: []*capiv1_protos.ProfileValues{
				{
					Name:    "foo",
					Version: "0.0.1",
					Values:  base64.StdEncoding.EncodeToString([]byte("foo: '{{ .params.TEST_PARAM }}'")),
				},
			},
			parameterValues: params,
			chartsCache:     fakeCache,
		},
	)
	assert.NoError(t, err)
	expected := []git.CommitFile{
		makeCommitFile(
			"ns-foo/cluster-foo/profiles.yaml",
			`apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: testing
  namespace: test-ns
spec:
  interval: 10m0s
  url: base/charts
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: foo
  namespace: flux-system
spec:
  chart:
    spec:
      chart: foo
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: testing
        namespace: test-ns
      version: 0.0.1
  install:
    crds: CreateReplace
  interval: 1m0s
  upgrade:
    crds: CreateReplace
  values:
    foo: this-is-a-test
status: {}
`),
	}
	assert.Equal(t, expected, files)
}

func TestGenerateProfileFiles_with_required_profiles_only(t *testing.T) {
	fakeCache := testNewFakeChartCache(t,
		nsn("management", ""),
		helm.ObjectReference{
			Name:      "testing",
			Namespace: "test-ns",
		},
		[]helm.Chart{})
	values := []byte("foo: defaultFoo")
	profile := fmt.Sprintf("{\"name\": \"foo\", \"version\": \"0.0.1\", \"values\": \"%s\" }", values)
	files, err := generateProfileFiles(
		context.TODO(),
		makeTestTemplateWithProfileAnnotation(
			templatesv1.RenderTypeEnvsubst,
			"capi.weave.works/profile-0",
			profile,
		),
		nsn("cluster-foo", "ns-foo"),
		makeTestHelmRepositoryTemplate("base"),
		generateProfileFilesParams{
			helmRepository: nsn("testing", "test-ns"),
			helmRepositoryCluster: types.NamespacedName{
				Name: "management",
			},
			profileValues:   []*capiv1_protos.ProfileValues{},
			parameterValues: map[string]string{},
			chartsCache:     fakeCache,
		},
	)
	require.NoError(t, err)
	expected := []git.CommitFile{
		makeCommitFile(
			"ns-foo/cluster-foo/profiles.yaml",
			`apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: testing
  namespace: test-ns
spec:
  interval: 10m0s
  url: base/charts
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: foo
  namespace: flux-system
spec:
  chart:
    spec:
      chart: foo
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: testing
        namespace: test-ns
      version: 0.0.1
  install:
    crds: CreateReplace
  interval: 1m0s
  upgrade:
    crds: CreateReplace
  values:
    foo: defaultFoo
status: {}
`),
	}
	assert.Equal(t, expected, files)
}

func TestGenerateProfileFiles_reading_layer_from_cache(t *testing.T) {
	fakeCache := testNewFakeChartCache(t,
		nsn("management", ""),
		helm.ObjectReference{
			Name:      "testing",
			Namespace: "test-ns",
		},
		[]helm.Chart{
			{
				Name:    "foo",
				Version: "0.0.1",
				Layer:   "layer-1",
			},
		})
	files, err := generateProfileFiles(
		context.TODO(),
		makeTestTemplate(templatesv1.RenderTypeEnvsubst),
		nsn("cluster-foo", "ns-foo"),
		makeTestHelmRepositoryTemplate("base"),
		generateProfileFilesParams{
			helmRepository:        nsn("testing", "test-ns"),
			helmRepositoryCluster: types.NamespacedName{Name: "management"},
			profileValues: []*capiv1_protos.ProfileValues{
				{
					Name:    "foo",
					Version: "0.0.1",
					Values:  base64.StdEncoding.EncodeToString([]byte("foo: bar")),
				},
				{
					Name:    "bar",
					Version: "0.0.1",
					Layer:   "layer-0",
					Values:  base64.StdEncoding.EncodeToString([]byte("foo: bar")),
				},
			},
			parameterValues: map[string]string{},
			chartsCache:     fakeCache,
		},
	)
	assert.NoError(t, err)
	expected := []git.CommitFile{
		makeCommitFile(
			"ns-foo/cluster-foo/profiles.yaml",
			`apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: testing
  namespace: test-ns
spec:
  interval: 10m0s
  url: base/charts
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  labels:
    weave.works/applied-layer: layer-0
  name: bar
  namespace: flux-system
spec:
  chart:
    spec:
      chart: bar
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: testing
        namespace: test-ns
      version: 0.0.1
  install:
    crds: CreateReplace
  interval: 1m0s
  upgrade:
    crds: CreateReplace
  values:
    foo: bar
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  labels:
    weave.works/applied-layer: layer-1
  name: foo
  namespace: flux-system
spec:
  chart:
    spec:
      chart: foo
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: testing
        namespace: test-ns
      version: 0.0.1
  dependsOn:
  - name: bar
  install:
    crds: CreateReplace
  interval: 1m0s
  upgrade:
    crds: CreateReplace
  values:
    foo: bar
status: {}
`),
	}
	assert.Equal(t, expected, files)
}

func TestGenerateProfilePaths(t *testing.T) {
	fakeCache := testNewFakeChartCache(t,
		nsn("management", ""),
		helm.ObjectReference{
			Name:      "testing",
			Namespace: "test-ns",
		},
		[]helm.Chart{
			{
				Name:    "foo",
				Version: "0.0.1",
				Layer:   "layer-1",
			},
		})

	expectedHelmRelease := `apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: testing
  namespace: test-ns
spec:
  interval: 10m0s
  url: base/charts
status: {}
`

	expectedBarHelmRelease := `apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  labels:
    weave.works/applied-layer: layer-0
  name: bar
  namespace: flux-system
spec:
  chart:
    spec:
      chart: bar
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: testing
        namespace: test-ns
      version: 0.0.1
  install:
    crds: CreateReplace
  interval: 1m0s
  upgrade:
    crds: CreateReplace
  values:
    foo: bar
status: {}
`

	expectedFooHelmRelease := `apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  labels:
    weave.works/applied-layer: layer-1
  name: foo
  namespace: flux-system
spec:
  chart:
    spec:
      chart: foo
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: testing
        namespace: test-ns
      version: 0.0.1
  dependsOn:
  - name: bar
  install:
    crds: CreateReplace
  interval: 1m0s
  upgrade:
    crds: CreateReplace
  values:
    foo: bar
status: {}
`

	var tests = []struct {
		name     string
		template *gapiv1.GitOpsTemplate
		expected []git.CommitFile
		params   map[string]string
	}{
		{
			name:     "generate profile paths",
			template: makeTestTemplateWithPaths(templatesv1.RenderTypeEnvsubst, "", "", ""),
			expected: []git.CommitFile{
				makeCommitFile(
					"ns-foo/cluster-foo/profiles.yaml",
					concatYaml(expectedHelmRelease, expectedBarHelmRelease, expectedFooHelmRelease),
				),
			},
		},
		{
			name:     "generate profile paths with custom paths",
			template: makeTestTemplateWithPaths(templatesv1.RenderTypeEnvsubst, "repo.yaml", "foo.yaml", "bar.yaml"),
			expected: []git.CommitFile{
				makeCommitFile(
					"bar.yaml",
					concatYaml(expectedBarHelmRelease),
				),
				makeCommitFile(
					"foo.yaml",
					concatYaml(expectedFooHelmRelease),
				),
				makeCommitFile(
					"repo.yaml",
					concatYaml(expectedHelmRelease),
				),
			},
		},
		{
			name:     "generate profile paths with custom paths and params",
			template: makeTestTemplateWithPaths(templatesv1.RenderTypeEnvsubst, "repo.yaml", "foo.yaml", "${BAR_PATH}"),
			params: map[string]string{
				"BAR_PATH": "special-bar.yaml",
			},
			expected: []git.CommitFile{
				makeCommitFile(
					"foo.yaml",
					concatYaml(expectedFooHelmRelease),
				),
				makeCommitFile(
					"repo.yaml",
					concatYaml(expectedHelmRelease),
				),
				makeCommitFile(
					"special-bar.yaml",
					concatYaml(expectedBarHelmRelease),
				),
			},
		},
		{
			name:     "generate profile paths with custom paths and params and render type",
			template: makeTestTemplateWithPaths(templatesv1.RenderTypeTemplating, "repo.yaml", "foo.yaml", "{{ .params.BAR_PATH }}"),
			params: map[string]string{
				"BAR_PATH": "special-bar.yaml",
			},
			expected: []git.CommitFile{
				makeCommitFile(
					"foo.yaml",
					concatYaml(expectedFooHelmRelease),
				),
				makeCommitFile(
					"repo.yaml",
					concatYaml(expectedHelmRelease),
				),
				makeCommitFile(
					"special-bar.yaml",
					concatYaml(expectedBarHelmRelease),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := generateProfileFiles(
				context.TODO(),
				tt.template,
				nsn("cluster-foo", "ns-foo"),
				makeTestHelmRepositoryTemplate("base"),
				generateProfileFilesParams{
					helmRepository:        nsn("testing", "test-ns"),
					helmRepositoryCluster: types.NamespacedName{Name: "management"},
					profileValues: []*capiv1_protos.ProfileValues{
						{
							Name:    "foo",
							Version: "0.0.1",
							Values:  base64.StdEncoding.EncodeToString([]byte("foo: bar")),
						},
						{
							Name:    "bar",
							Version: "0.0.1",
							Layer:   "layer-0",
							Values:  base64.StdEncoding.EncodeToString([]byte("foo: bar")),
						},
					},
					parameterValues: tt.params,
					chartsCache:     fakeCache,
				},
			)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, files)
		})
	}
}

func concatYaml(yamls ...string) string {
	return strings.Join(yamls, "---\n")
}

// generateProfiles takes in a HelmRepo that we are going to write to git,
// it shouldn't have Status etc set
func makeTestHelmRepositoryTemplate(base string) *sourcev1.HelmRepository {
	return makeTestHelmRepository(base, func(hr *sourcev1.HelmRepository) {
		hr.Status = sourcev1.HelmRepositoryStatus{}
	})
}

func makeCommitFile(path, content string) git.CommitFile {
	p := path
	c := content
	return git.CommitFile{
		Path:    p,
		Content: &c,
	}
}

func makeTestTemplateWithPaths(renderType string, helmRepoPath string, fooPath string, barPath string) *gapiv1.GitOpsTemplate {
	chartsSpec := templatesv1.ChartsSpec{
		HelmRepositoryTemplate: templatesv1.HelmRepositoryTemplateSpec{
			Path: helmRepoPath,
		},
		Items: []templatesv1.Chart{
			{
				Chart:    "foo",
				Editable: true,
				HelmReleaseTemplate: templatesv1.HelmReleaseTemplateSpec{
					Path: fooPath,
				},
			},
			{
				Chart:    "bar",
				Editable: true,
				HelmReleaseTemplate: templatesv1.HelmReleaseTemplateSpec{
					Path: barPath,
				},
			},
		},
	}

	return &gapiv1.GitOpsTemplate{
		Spec: templatesv1.TemplateSpec{
			RenderType: renderType,
			Charts:     chartsSpec,
		},
	}
}

func makeTestTemplate(renderType string) *gapiv1.GitOpsTemplate {
	return &gapiv1.GitOpsTemplate{
		Spec: templatesv1.TemplateSpec{
			RenderType: renderType,
		},
	}
}

func makeTestTemplateWithProfileAnnotation(renderType, annotationName, annotationValue string) templatesv1.Template {
	return &capiv1.CAPITemplate{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				annotationName: annotationValue,
			},
		},
		Spec: templatesv1.TemplateSpec{
			RenderType: renderType,
		},
	}
}

func testNewClusterNamespacedName(t *testing.T, name, namespace string) *capiv1_protos.ClusterNamespacedName {
	return &capiv1_protos.ClusterNamespacedName{
		Name:      name,
		Namespace: namespace,
	}
}

func testNewMetadata(t *testing.T, name, namespace string) *capiv1_protos.Metadata {
	return &capiv1_protos.Metadata{
		Name:      name,
		Namespace: namespace,
	}
}

func testNewSourceRef(t *testing.T, name, namespace string) *capiv1_protos.SourceRef {
	return &capiv1_protos.SourceRef{
		Name:      name,
		Namespace: namespace,
	}
}

func testNewChart(t *testing.T, name string, sourceRef *capiv1_protos.SourceRef) *capiv1_protos.Chart {
	return &capiv1_protos.Chart{
		Spec: &capiv1_protos.ChartSpec{
			Chart:     name,
			SourceRef: sourceRef,
		},
	}
}

func readCAPITemplateFixture(t *testing.T, name string) *capiv1.CAPITemplate {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	loaded := &capiv1.CAPITemplate{}
	if err := yaml.Unmarshal(b, &loaded); err != nil {
		t.Fatal(err)
	}

	return loaded
}
