package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"text/template"
	"time"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/testing/protocmp"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	templatesv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/charts"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func TestListGitopsClusters(t *testing.T) {
	testCases := []struct {
		name         string
		clusterState []runtime.Object
		refType      string
		expected     []*capiv1_protos.GitopsCluster
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
			expected: []*capiv1_protos.GitopsCluster{
				{
					Name:      "gitops-cluster",
					Namespace: "default",
					CapiClusterRef: &capiv1_protos.GitopsClusterRef{
						Name: "dev",
					},
				},
				{
					Name:      "gitops-cluster2",
					Namespace: "default",
					SecretRef: &capiv1_protos.GitopsClusterRef{
						Name: "dev",
					},
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
			refType: "Secret",
			expected: []*capiv1_protos.GitopsCluster{
				{
					Name:      "gitops-cluster2",
					Namespace: "default",
					SecretRef: &capiv1_protos.GitopsClusterRef{
						Name: "dev",
					},
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
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			viper.SetDefault("runtime-namespace", "default")

			// setup
			s := createServer(t, serverOptions{
				clusterState: tt.clusterState,
				namespace:    "default",
			})

			// request
			listGitopsClustersRequest := new(capiv1_protos.ListGitopsClustersRequest)
			listGitopsClustersRequest.RefType = tt.refType
			listGitopsClustersResponse, err := s.ListGitopsClusters(context.Background(), listGitopsClustersRequest)

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

func TestCreatePullRequest(t *testing.T) {
	viper.SetDefault("capi-repository-path", "clusters/my-cluster/clusters")
	viper.SetDefault("capi-repository-clusters-path", "clusters")
	viper.SetDefault("add-bases-kustomization", "enabled")
	testCases := []struct {
		name           string
		clusterState   []runtime.Object
		provider       git.Provider
		pruneEnvVar    string
		req            *capiv1_protos.CreatePullRequestRequest
		expected       string
		committedFiles []CommittedFile
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
				makeTemplateConfigMap("capi-templates", "template1", makeCAPITemplate(t)),
			},
			req: &capiv1_protos.CreatePullRequestRequest{
				TemplateName: "cluster-template-1",
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
			},
			err: errors.New(`validation error rendering template cluster-template-1, invalid value for metadata.name: "foo bar bad name", a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`),
		},
		{
			name: "namespace validation errors",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("capi-templates", "template1", makeCAPITemplate(t)),
			},
			req: &capiv1_protos.CreatePullRequestRequest{
				TemplateName: "cluster-template-1",
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
				makeTemplateConfigMap("capi-templates", "template1", makeCAPITemplate(t)),
			},
			provider: NewFakeGitProvider("", nil, errors.New("oops")),
			req: &capiv1_protos.CreatePullRequestRequest{
				TemplateName: "cluster-template-1",
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
			err: errors.New(`rpc error: code = Unauthenticated desc = failed to access repo https://github.com/org/repo.git: oops`),
		},
		{
			name: "create pull request",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("capi-templates", "template1", makeCAPITemplate(t)),
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				TemplateName: "cluster-template-1",
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
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "default profile values",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("capi-templates", "template1", makeCAPITemplate(t)),
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				TemplateName: "cluster-template-1",
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
				Values: []*capiv1_protos.ProfileValues{
					{
						Name:    "demo-profile",
						Version: "0.0.1",
						Values:  base64.StdEncoding.EncodeToString([]byte(``)),
					},
				},
			},
			committedFiles: []CommittedFile{
				{
					Path: "clusters/my-cluster/clusters/default/dev.yaml",
					Content: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    kustomize.toolkit.fluxcd.io/prune: disabled
  name: dev
  namespace: default
`,
				},
				{
					Path: "clusters/default/dev/clusters-bases-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1beta2
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
  name: dev-demo-profile
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
  values:
    favoriteDrink: coffee
status: {}
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "specify profile namespace and cluster namespace",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("capi-templates", "template1", makeCAPITemplate(t)),
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				TemplateName: "cluster-template-1",
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
				Values: []*capiv1_protos.ProfileValues{
					{
						Name:      "demo-profile",
						Version:   "0.0.1",
						Values:    base64.StdEncoding.EncodeToString([]byte(``)),
						Namespace: "test-system",
					},
				},
			},
			committedFiles: []CommittedFile{
				{
					Path: "clusters/my-cluster/clusters/clusters-namespace/dev.yaml",
					Content: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    kustomize.toolkit.fluxcd.io/prune: disabled
  name: dev
  namespace: clusters-namespace
`,
				},
				{
					Path: "clusters/clusters-namespace/dev/clusters-bases-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1beta2
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
  name: dev-demo-profile
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
  values:
    favoriteDrink: coffee
status: {}
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "create kustomizations",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("capi-templates", "template1", makeCAPITemplate(t)),
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				TemplateName: "cluster-template-1",
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
				Kustomizations: []*capiv1_protos.Kustomization{
					{
						Metadata: &capiv1_protos.Metadata{
							Name:      "apps-capi",
							Namespace: "flux-system",
						},
						Spec: &capiv1_protos.Spec{
							Path: "./apps/capi",
							SourceRef: &capiv1_protos.SourceRef{
								Name:      "flux-system",
								Namespace: "flux-system",
							},
						},
					},
					{
						Metadata: &capiv1_protos.Metadata{
							Name:      "apps-billing",
							Namespace: "flux-system",
						},
						Spec: &capiv1_protos.Spec{
							Path: "./apps/billing",
							SourceRef: &capiv1_protos.SourceRef{
								Name:      "flux-system",
								Namespace: "flux-system",
							},
						},
					},
				},
			},
			committedFiles: []CommittedFile{
				{
					Path: "clusters/my-cluster/clusters/clusters-namespace/dev.yaml",
					Content: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    kustomize.toolkit.fluxcd.io/prune: disabled
  name: dev
  namespace: clusters-namespace
`,
				},
				{
					Path: "clusters/clusters-namespace/dev/clusters-bases-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1beta2
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
					Path: "clusters/clusters-namespace/dev/flux-system/apps-capi-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1beta2
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
status: {}
`,
				},
				{
					Path: "clusters/clusters-namespace/dev/flux-system/apps-billing-kustomization.yaml",
					Content: `apiVersion: kustomize.toolkit.fluxcd.io/v1beta2
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
				makeTemplateConfigMap("capi-templates", "template1", makeCAPITemplate(t)),
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				TemplateName: "cluster-template-1",
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
				Kustomizations: []*capiv1_protos.Kustomization{
					{
						Metadata: &capiv1_protos.Metadata{
							Name:      "",
							Namespace: "@kustomization",
						},
						Spec: &capiv1_protos.Spec{
							Path: "./apps/capi",
							SourceRef: &capiv1_protos.SourceRef{
								Name:      "flux-system",
								Namespace: "flux-system",
							},
						},
					},
					{
						Metadata: &capiv1_protos.Metadata{
							Name:      "apps-capi",
							Namespace: "flux-system",
						},
						Spec: &capiv1_protos.Spec{
							Path: "./apps/capi",
							SourceRef: &capiv1_protos.SourceRef{
								Name:      "",
								Namespace: "",
							},
						},
					},
				},
			},
			err: errors.New("3 errors occurred:\nkustomization name must be specified\ninvalid namespace: @kustomization, a lowercase RFC 1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')\nsourceRef name must be specified in Kustomization apps-capi"),
		},
		{
			name: "kustomization with metadata is nil",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("capi-templates", "template1", makeCAPITemplate(t)),
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				TemplateName: "cluster-template-1",
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
				Kustomizations: []*capiv1_protos.Kustomization{
					{
						Spec: &capiv1_protos.Spec{
							Path: "./apps/capi",
							SourceRef: &capiv1_protos.SourceRef{
								Name:      "flux-system",
								Namespace: "flux-system",
							},
						},
					},
				},
			},
			err: errors.New("kustomization metadata must be specified"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			viper.SetDefault("runtime-namespace", "default")
			// setup
			ts := httptest.NewServer(makeServeMux(t))
			hr := makeTestHelmRepository(ts.URL, func(hr *sourcev1.HelmRepository) {
				hr.Name = "weaveworks-charts"
				hr.Namespace = "default"
			})
			tt.clusterState = append(tt.clusterState, hr)
			s := createServer(t, serverOptions{
				clusterState:  tt.clusterState,
				configMapName: "capi-templates",
				namespace:     "default",
				provider:      tt.provider,
				hr:            hr,
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
				fakeGitProvider := (tt.provider).(*FakeGitProvider)
				if diff := cmp.Diff(prepCommitedFiles(t, ts.URL, tt.committedFiles), fakeGitProvider.GetCommittedFiles()); len(tt.committedFiles) > 0 && diff != "" {
					t.Fatalf("committed files do not match expected committed files:\n%s", diff)
				}
			}
		})
	}
}

func prepCommitedFiles(t *testing.T, serverUrl string, files []CommittedFile) []CommittedFile {
	parsedURL, err := url.Parse(serverUrl)
	if err != nil {
		t.Fatalf("failed to parse URL %s", err)
	}
	newFiles := []CommittedFile{}
	for _, f := range files {
		newFiles = append(newFiles, CommittedFile{
			Path:    f.Path,
			Content: simpleTemplate(t, f.Content, struct{ Port string }{Port: parsedURL.Port()}),
		})
	}
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
		clusterState            []runtime.Object
		clusterObjectsNamespace string // Namespace that cluster objects are created in
		req                     *capiv1_protos.GetKubeconfigRequest
		ctx                     context.Context
		expected                []byte
		err                     error
	}{
		{
			name: "get kubeconfig as JSON",
			clusterState: []runtime.Object{
				makeSecret("dev-kubeconfig", "default", "value", "foo"),
			},
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.MD{}),
			expected: []byte(fmt.Sprintf(`{"kubeconfig":"%s"}`, base64.StdEncoding.EncodeToString([]byte("foo")))),
		},
		{
			name: "get kubeconfig as binary",
			clusterState: []runtime.Object{
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
			name:                    "secret not found",
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			err: errors.New("unable to get secret \"dev-kubeconfig\" for Kubeconfig: secrets \"dev-kubeconfig\" not found"),
		},
		{
			name: "secret found but is missing key",
			clusterState: []runtime.Object{
				makeSecret("dev-kubeconfig", "default", "val", "foo"),
			},
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			err: errors.New("secret \"default/dev-kubeconfig\" was found but is missing key \"value\""),
		},
		{
			name: "use cluster_namespace to get secret",
			clusterState: []runtime.Object{
				makeSecret("dev-kubeconfig", "kube-system", "value", "foo"),
			},
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName:      "dev",
				ClusterNamespace: "kube-system",
			},
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.MD{}),
			expected: []byte(fmt.Sprintf(`{"kubeconfig":"%s"}`, base64.StdEncoding.EncodeToString([]byte("foo")))),
		},
		{
			name: "no namespace and lookup across namespaces, use default namespace",
			clusterState: []runtime.Object{
				makeSecret("dev-kubeconfig", "default", "value", "foo"),
			},
			clusterObjectsNamespace: "",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.MD{}),
			expected: []byte(fmt.Sprintf(`{"kubeconfig":"%s"}`, base64.StdEncoding.EncodeToString([]byte("foo")))),
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			viper.SetDefault("capi-clusters-namespace", tt.clusterObjectsNamespace)

			s := createServer(t, serverOptions{
				clusterState:  tt.clusterState,
				configMapName: "capi-templates",
				namespace:     "default",
				ns:            tt.clusterObjectsNamespace,
			})

			res, err := s.GetKubeconfig(tt.ctx, tt.req)

			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to get the kubeconfig:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("got the wrong error:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, res.Data, protocmp.Transform()); diff != "" {
					t.Fatalf("kubeconfig didn't match expected:\n%s", diff)
				}
			}
		})
	}
}

func TestDeleteClustersPullRequest(t *testing.T) {
	testCases := []struct {
		name     string
		provider git.Provider
		req      *capiv1_protos.DeleteClustersPullRequestRequest
		expected string
		err      error
	}{
		{
			name: "validation errors",
			req:  &capiv1_protos.DeleteClustersPullRequestRequest{},
			err:  errors.New(deleteClustersRequiredErr),
		},
		//
		// -- FIXME: consider checking the contents of git before trying to delete
		//
		// {
		// 	name:     "cluster does not exist",
		// 	provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil),
		// 	req: &capiv1_protos.DeleteClustersPullRequestRequest{
		// 		ClusterNames:  []string{"foo"},
		// 		RepositoryUrl: "https://github.com/org/repo.git",
		// 		HeadBranch:    "feature-02",
		// 		BaseBranch:    "feature-01",
		// 		Title:         "Delete Cluster",
		// 		Description:   "Deletes a cluster",
		// 		CommitMessage: "Remove cluster manifest",
		// 	},
		// },
		//
		{
			name:     "create delete pull request",
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil),
			req: &capiv1_protos.DeleteClustersPullRequestRequest{
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
			name:     "create delete pull request with namespaced cluster names",
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil),
			req: &capiv1_protos.DeleteClustersPullRequestRequest{
				ClusterNamespacedNames: []*capiv1_protos.ClusterNamespacedName{
					{
						Name:      "foo",
						Namespace: "ns-foo",
					},
					{
						Name:      "bar",
						Namespace: "ns-bar",
					},
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
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			s := createServer(t, serverOptions{
				configMapName: "capi-templates",
				namespace:     "default",
				provider:      tt.provider,
			})

			// delete request
			deletePullRequestResponse, err := s.DeleteClustersPullRequest(context.Background(), tt.req)

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

func NewFakeGitProvider(url string, repo *git.GitRepo, err error) git.Provider {
	return &FakeGitProvider{
		url:  url,
		repo: repo,
		err:  err,
	}
}

type FakeGitProvider struct {
	url            string
	repo           *git.GitRepo
	err            error
	committedFiles []gitprovider.CommitFile
}

func (p *FakeGitProvider) WriteFilesToBranchAndCreatePullRequest(ctx context.Context, req git.WriteFilesToBranchAndCreatePullRequestRequest) (*git.WriteFilesToBranchAndCreatePullRequestResponse, error) {
	if p.err != nil {
		return nil, p.err
	}
	p.committedFiles = append(p.committedFiles, req.Files...)
	return &git.WriteFilesToBranchAndCreatePullRequestResponse{WebURL: p.url}, nil
}

func (p *FakeGitProvider) CloneRepoToTempDir(req git.CloneRepoToTempDirRequest) (*git.CloneRepoToTempDirResponse, error) {
	if p.err != nil {
		return nil, p.err
	}
	return &git.CloneRepoToTempDirResponse{Repo: p.repo}, nil
}

func (p *FakeGitProvider) GetRepository(ctx context.Context, gp git.GitProvider, url string) (gitprovider.OrgRepository, error) {
	if p.err != nil {
		return nil, p.err
	}
	return nil, nil
}

func (p *FakeGitProvider) GetCommittedFiles() []CommittedFile {
	var committedFiles []CommittedFile
	for _, f := range p.committedFiles {
		committedFiles = append(committedFiles, CommittedFile{
			Path:    *f.Path,
			Content: *f.Content,
		})
	}
	return committedFiles
}

type CommittedFile struct {
	Path    string
	Content string
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
	c := createClient(t, makeTestHelmRepository("base"))
	file, err := generateProfileFiles(
		context.TODO(),
		makeTestTemplate(templatesv1.RenderTypeEnvsubst),
		types.NamespacedName{
			Name:      "cluster-foo",
			Namespace: "ns-foo",
		},
		c,
		generateProfileFilesParams{
			helmRepository: types.NamespacedName{
				Name:      "testing",
				Namespace: "test-ns",
			},
			profileValues: []*capiv1_protos.ProfileValues{
				{
					Name:    "foo",
					Version: "0.0.1",
					Values:  base64.StdEncoding.EncodeToString([]byte("foo: bar")),
				},
			},
			parameterValues: map[string]string{},
		},
	)
	assert.NoError(t, err)
	expected := `apiVersion: source.toolkit.fluxcd.io/v1beta2
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
  name: cluster-foo-foo
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
`
	assert.Equal(t, expected, *file.Content)
}

func TestGenerateProfileFiles_with_templates(t *testing.T) {
	c := createClient(t, makeTestHelmRepository("base"))
	params := map[string]string{
		"CLUSTER_NAME": "test-cluster-name",
		"NAMESPACE":    "default",
	}

	file, err := generateProfileFiles(
		context.TODO(),
		makeTestTemplate(templatesv1.RenderTypeEnvsubst),
		types.NamespacedName{
			Name:      "cluster-foo",
			Namespace: "ns-foo",
		},
		c,
		generateProfileFilesParams{
			helmRepository: types.NamespacedName{
				Name:      "testing",
				Namespace: "test-ns",
			},
			profileValues: []*capiv1_protos.ProfileValues{
				{
					Name:    "foo",
					Version: "0.0.1",
					Values:  base64.StdEncoding.EncodeToString([]byte("foo: ${CLUSTER_NAME}")),
				},
			},
			parameterValues: params,
		},
	)
	assert.NoError(t, err)
	expected := `apiVersion: source.toolkit.fluxcd.io/v1beta2
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
  name: cluster-foo-foo
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
`
	assert.Equal(t, expected, *file.Content)
}

func TestGenerateProfileFilesWithLayers(t *testing.T) {
	c := createClient(t, makeTestHelmRepository("base"))
	file, err := generateProfileFiles(
		context.TODO(),
		makeTestTemplate(templatesv1.RenderTypeEnvsubst),
		types.NamespacedName{
			Name:      "cluster-foo",
			Namespace: "ns-foo",
		},
		c,
		generateProfileFilesParams{
			helmRepository: types.NamespacedName{
				Name:      "testing",
				Namespace: "test-ns",
			},
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
		},
	)
	assert.NoError(t, err)
	expected := `apiVersion: source.toolkit.fluxcd.io/v1beta2
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
  name: cluster-foo-bar
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
  name: cluster-foo-foo
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
  - name: cluster-foo-bar
  install:
    crds: CreateReplace
  interval: 1m0s
  upgrade:
    crds: CreateReplace
  values:
    foo: bar
status: {}
`
	assert.Equal(t, expected, *file.Content)
}

func TestGenerateProfileFiles_with_text_templates(t *testing.T) {
	c := createClient(t, makeTestHelmRepository("base"))
	params := map[string]string{
		"CLUSTER_NAME": "test-cluster-name",
		"NAMESPACE":    "default",
		"TEST_PARAM":   "this-is-a-test",
	}

	file, err := generateProfileFiles(
		context.TODO(),
		makeTestTemplate(templatesv1.RenderTypeTemplating),
		types.NamespacedName{
			Name:      "cluster-foo",
			Namespace: "ns-foo",
		},
		c,
		generateProfileFilesParams{
			helmRepository: types.NamespacedName{
				Name:      "testing",
				Namespace: "test-ns",
			},
			profileValues: []*capiv1_protos.ProfileValues{
				{
					Name:    "foo",
					Version: "0.0.1",
					Values:  base64.StdEncoding.EncodeToString([]byte("foo: '{{ .params.TEST_PARAM }}'")),
				},
			},
			parameterValues: params,
		},
	)
	assert.NoError(t, err)
	expected := `apiVersion: source.toolkit.fluxcd.io/v1beta2
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
  name: cluster-foo-foo
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
`
	assert.Equal(t, expected, *file.Content)
}

func makeTestTemplate(renderType string) *templatesv1.Template {
	return &templatesv1.Template{
		Spec: templatesv1.TemplateSpec{
			RenderType: renderType,
		},
	}
}
