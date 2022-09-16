package server

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/testing/protocmp"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func TestCreateAutomationsPullRequest(t *testing.T) {
	viper.SetDefault("capi-repository-path", "clusters/my-cluster/clusters")
	viper.SetDefault("capi-repository-clusters-path", "clusters")
	viper.SetDefault("add-bases-kustomization", "enabled")
	testCases := []struct {
		name           string
		clusterState   []runtime.Object
		provider       git.Provider
		pruneEnvVar    string
		req            *capiv1_protos.CreateAutomationsPullRequestRequest
		expected       string
		committedFiles []*capiv1_protos.CommitFile
		err            error
	}{
		{
			name: "validation errors",
			req:  &capiv1_protos.CreateAutomationsPullRequestRequest{},
			err:  errors.New("at least one cluster automation must be specified"),
		},
		{
			name:     "pull request failed",
			provider: NewFakeGitProvider("", nil, errors.New("oops"), nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster: testNewClusterNamespacedName(t, "billing", "dev"),
						Kustomization: &capiv1_protos.Kustomization{
							Metadata: testNewMetadata(t, "apps-billing", "flux-system"),
							Spec: &capiv1_protos.KustomizationSpec{
								Path:      "./apps/billing",
								SourceRef: testNewSourceRef(t, "flux-system", "flux-system"),
							},
						},
					},
				},
			},
			err: errors.New(`rpc error: code = Unauthenticated desc = failed to access repo https://github.com/org/repo.git: oops`),
		},
		{
			name:     "create pull request",
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						Kustomization: &capiv1_protos.Kustomization{
							Metadata: testNewMetadata(t, "apps-capi", "flux-system"),
							Spec: &capiv1_protos.KustomizationSpec{
								Path:            "./apps/capi",
								SourceRef:       testNewSourceRef(t, "flux-system", "flux-system"),
								TargetNamespace: "foo-ns",
							},
						},
					},
					{
						Cluster: testNewClusterNamespacedName(t, "billing", "dev"),
						HelmRelease: &capiv1_protos.HelmRelease{
							Metadata: testNewMetadata(t, "apps-billing", "flux-system"),
							Spec: &capiv1_protos.HelmReleaseSpec{
								Chart:  testNewChart(t, "test-chart", testNewSourceRef(t, "test", "test-ns")),
								Values: "",
							},
						},
					},
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "committed files for kustomization",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster Kustomization",
				Description:   "Creates cluster Kustomizations",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						Kustomization: &capiv1_protos.Kustomization{
							Metadata: testNewMetadata(t, "apps-capi", "flux-system"),
							Spec: &capiv1_protos.KustomizationSpec{
								Path:            "./apps/capi",
								SourceRef:       testNewSourceRef(t, "flux-system", "flux-system"),
								TargetNamespace: "foo-ns",
							},
						},
					},
					{
						Cluster: testNewClusterNamespacedName(t, "billing", "dev"),
						Kustomization: &capiv1_protos.Kustomization{
							Metadata: testNewMetadata(t, "apps-billing", "flux-system"),
							Spec: &capiv1_protos.KustomizationSpec{
								Path:      "./apps/billing",
								SourceRef: testNewSourceRef(t, "flux-system", "flux-system"),
							},
						},
					},
				},
			},
			committedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/management/apps-capi-flux-system-kustomization.yaml",
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
  targetNamespace: foo-ns
status: {}
`,
				},
				{
					Path: "clusters/dev/billing/apps-billing-flux-system-kustomization.yaml",
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
		}, {
			name: "default values for namespace in kustomizations and helm releases",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster Kustomization",
				Description:   "Creates cluster Kustomizations",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						Kustomization: &capiv1_protos.Kustomization{
							Metadata: testNewMetadata(t, "apps-capi", ""),
							Spec: &capiv1_protos.KustomizationSpec{
								Path:      "./apps/capi",
								SourceRef: testNewSourceRef(t, "flux-system", "flux-system"),
							},
						},
					},
					{
						Cluster: testNewClusterNamespacedName(t, "billing", "dev"),
						HelmRelease: &capiv1_protos.HelmRelease{
							Metadata: testNewMetadata(t, "test-profile", ""),
							Spec: &capiv1_protos.HelmReleaseSpec{
								Chart: testNewChart(t, "test-chart", testNewSourceRef(t, "weaveworks-charts", "default")),
							},
						},
					},
				},
			},
			committedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/management/apps-capi-flux-system-kustomization.yaml",
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
					Path: "clusters/dev/billing/test-profile-flux-system-helmrelease.yaml",
					Content: `apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: test-profile
  namespace: flux-system
spec:
  chart:
    spec:
      chart: test-chart
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: weaveworks-charts
        namespace: default
  interval: 10m0s
  values: null
status: {}
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "committed files for helm release",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster HelmRelease",
				Description:   "Creates cluster HelmReleases",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						HelmRelease: &capiv1_protos.HelmRelease{
							Metadata: testNewMetadata(t, "first-profile", "flux-system"),
							Spec: &capiv1_protos.HelmReleaseSpec{
								Chart:  testNewChart(t, "test-chart", testNewSourceRef(t, "weaveworks-charts", "default")),
								Values: "foo: bar",
							},
						},
					},
					{
						Cluster: testNewClusterNamespacedName(t, "billing", "dev"),
						HelmRelease: &capiv1_protos.HelmRelease{
							Metadata: testNewMetadata(t, "second-profile", "flux-system"),
							Spec: &capiv1_protos.HelmReleaseSpec{
								Chart:  testNewChart(t, "test-chart", testNewSourceRef(t, "weaveworks-charts", "default")),
								Values: "",
							},
						},
					},
				},
			},
			committedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/management/first-profile-flux-system-helmrelease.yaml",
					Content: `apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: first-profile
  namespace: flux-system
spec:
  chart:
    spec:
      chart: test-chart
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: weaveworks-charts
        namespace: default
  interval: 10m0s
  values:
    foo: bar
status: {}
`,
				},
				{
					Path: "clusters/dev/billing/second-profile-flux-system-helmrelease.yaml",
					Content: `apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: second-profile
  namespace: flux-system
spec:
  chart:
    spec:
      chart: test-chart
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: weaveworks-charts
        namespace: default
  interval: 10m0s
  values: null
status: {}
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "helm release validation errors",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster HelmRelease",
				Description:   "Creates cluster HelmReleases",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster: testNewClusterNamespacedName(t, "management", "default"),
						HelmRelease: &capiv1_protos.HelmRelease{
							Metadata: testNewMetadata(t, "", "@helmrelease"),
							Spec: &capiv1_protos.HelmReleaseSpec{
								Chart: testNewChart(t, "test-chart", testNewSourceRef(t, "weaveworks-charts", "default")),
							},
						},
					},
					{
						Cluster: testNewClusterNamespacedName(t, "billing", "dev"),
						HelmRelease: &capiv1_protos.HelmRelease{
							Metadata: testNewMetadata(t, "test-profile", "flux-system"),
							Spec: &capiv1_protos.HelmReleaseSpec{
								Chart: testNewChart(t, "test-chart", testNewSourceRef(t, "", "")),
							},
						},
					},
				},
			},
			err: errors.New("3 errors occurred:\nhelmrelease name must be specified\ninvalid namespace: @helmrelease, a lowercase RFC 1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')\nsourceRef name must be specified in chart test-chart in HelmRelease test-profile"),
		},
		{
			name: "chart validation errors",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster HelmRelease",
				Description:   "Creates cluster HelmReleases",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster: testNewClusterNamespacedName(t, "management", "default"),
						HelmRelease: &capiv1_protos.HelmRelease{
							Metadata: testNewMetadata(t, "foo-hr", "flux-system"),
							Spec:     &capiv1_protos.HelmReleaseSpec{},
						},
					},
					{
						Cluster: testNewClusterNamespacedName(t, "billing", "dev"),
						HelmRelease: &capiv1_protos.HelmRelease{
							Metadata: testNewMetadata(t, "bar-hr", "flux-system"),
							Spec: &capiv1_protos.HelmReleaseSpec{
								Chart: testNewChart(t, "", testNewSourceRef(t, "weaveworks-charts", "default")),
							},
						},
					},
				},
			},
			err: errors.New("2 errors occurred:\nchart must be specified in HelmRelease foo-hr\nchart name must be specified in HelmRelease bar-hr"),
		},
		{
			name: "chart values decoding errors, gotta be an object",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster HelmRelease",
				Description:   "Creates cluster HelmReleases",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster: testNewClusterNamespacedName(t, "billing", "dev"),
						HelmRelease: &capiv1_protos.HelmRelease{
							Metadata: testNewMetadata(t, "bar-hr", "flux-system"),
							Spec: &capiv1_protos.HelmReleaseSpec{
								Chart:  testNewChart(t, "foo", testNewSourceRef(t, "weaveworks-charts", "default")),
								Values: "bar",
							},
						},
					},
				},
			},
			err: errors.New("failed to create Helm Release object: flux-system/bar-hr: failed to yaml-unmarshal values: failed to parse values from JSON: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type map[string]interface {}"),
		},
		{
			name: "helmrelease with metadata is nil",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster HelmRelease",
				Description:   "Creates cluster HelmReleases",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster: testNewClusterNamespacedName(t, "management", "default"),
						HelmRelease: &capiv1_protos.HelmRelease{
							Spec: &capiv1_protos.HelmReleaseSpec{
								Chart: testNewChart(t, "test-chart", testNewSourceRef(t, "weaveworks-charts", "default")),
							},
						},
					},
				},
			},
			err: errors.New("helmrelease metadata must be specified"),
		},
		{
			name: "ClusterAutomation with Cluster is nil",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster HelmRelease",
				Description:   "Creates cluster HelmReleases",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						HelmRelease: &capiv1_protos.HelmRelease{
							Metadata: testNewMetadata(t, "test-profile", "flux-system"),
							Spec: &capiv1_protos.HelmReleaseSpec{
								Chart: testNewChart(t, "test-chart", testNewSourceRef(t, "weaveworks-charts", "default")),
							},
						},
					},
				},
			},
			err: errors.New("cluster object must be specified"),
		},
		{
			name: "custom filepath",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster Kustomization",
				Description:   "Creates cluster Kustomizations",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster: testNewClusterNamespacedName(t, "billing", "dev"),
						Kustomization: &capiv1_protos.Kustomization{
							Metadata: testNewMetadata(t, "apps-billing", "flux-system"),
							Spec: &capiv1_protos.KustomizationSpec{
								Path:      "./apps/billing",
								SourceRef: testNewSourceRef(t, "flux-system", "flux-system"),
							},
						},
						FilePath: "clusters/dev/test-kustomization.yaml",
					},
					{
						Cluster: testNewClusterNamespacedName(t, "billing", "dev"),
						HelmRelease: &capiv1_protos.HelmRelease{
							Metadata: testNewMetadata(t, "test-profile", "flux-system"),
							Spec: &capiv1_protos.HelmReleaseSpec{
								Chart:  testNewChart(t, "test-chart", testNewSourceRef(t, "weaveworks-charts", "default")),
								Values: "",
							},
						},
						FilePath: "clusters/prod/test-hr.yaml",
					},
				},
			},
			committedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/dev/test-kustomization.yaml",
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
				{
					Path: "clusters/prod/test-hr.yaml",
					Content: `apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: test-profile
  namespace: flux-system
spec:
  chart:
    spec:
      chart: test-chart
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: weaveworks-charts
        namespace: default
  interval: 10m0s
  values: null
status: {}
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
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
				clusterState: tt.clusterState,
				namespace:    "default",
				provider:     tt.provider,
				hr:           hr,
			})

			// request
			createPullRequestResponse, err := s.CreateAutomationsPullRequest(context.Background(), tt.req)

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
				if diff := cmp.Diff(prepCommitedFiles(t, ts.URL, tt.committedFiles), fakeGitProvider.GetCommittedFiles(), protocmp.Transform()); len(tt.committedFiles) > 0 && diff != "" {
					t.Fatalf("committed files do not match expected committed files:\n%s", diff)
				}
			}
		})
	}
}

func TestRenderAutomation(t *testing.T) {
	viper.SetDefault("capi-repository-path", "clusters/my-cluster/clusters")
	viper.SetDefault("capi-repository-clusters-path", "clusters")
	viper.SetDefault("add-bases-kustomization", "enabled")

	testCases := []struct {
		name           string
		clusterState   []runtime.Object
		pruneEnvVar    string
		req            *capiv1_protos.RenderAutomationRequest
		expected       string
		committedFiles []*capiv1_protos.CommitFile
		err            error
	}{
		{
			name: "render automations",
			req: &capiv1_protos.RenderAutomationRequest{
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						Kustomization: &capiv1_protos.Kustomization{
							Metadata: testNewMetadata(t, "apps-capi", ""),
							Spec: &capiv1_protos.KustomizationSpec{
								Path:      "./apps/capi",
								SourceRef: testNewSourceRef(t, "flux-system", "flux-system"),
							},
						},
					},
					{
						Cluster: testNewClusterNamespacedName(t, "billing", "dev"),
						HelmRelease: &capiv1_protos.HelmRelease{
							Metadata: testNewMetadata(t, "test-profile", ""),
							Spec: &capiv1_protos.HelmReleaseSpec{
								Chart: testNewChart(t, "test-chart", testNewSourceRef(t, "weaveworks-charts", "default")),
							},
						},
					},
				},
			},
			committedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/management/apps-capi-flux-system-kustomization.yaml",
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
					Path: "clusters/dev/billing/test-profile-flux-system-helmrelease.yaml",
					Content: `apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: test-profile
  namespace: flux-system
spec:
  chart:
    spec:
      chart: test-chart
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta2
        kind: HelmRepository
        name: weaveworks-charts
        namespace: default
  interval: 10m0s
  values: null
status: {}
`,
				},
			},
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
				clusterState: tt.clusterState,
				namespace:    "default",
				hr:           hr,
			})

			// request
			renderAutomationResponse, err := s.RenderAutomation(context.Background(), tt.req)

			// Check the response looks good
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to render automations:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("got the wrong error:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.committedFiles, renderAutomationResponse.KustomizationFiles, protocmp.Transform()); diff != "" {
					t.Fatalf("committed files do not match expected committed files:\n%s", diff)
				}
			}
		})
	}
}
