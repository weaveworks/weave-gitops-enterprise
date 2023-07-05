package server

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git/gitfakes"
	"google.golang.org/protobuf/testing/protocmp"
	structpb "google.golang.org/protobuf/types/known/structpb"
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
			provider: gitfakes.NewFakeGitProvider("", nil, errors.New("oops"), nil, nil),
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
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
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
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
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
		},
		{
			name: "create target namespace resource",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
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
								CreateNamespace: true,
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
					Path: "clusters/management/foo-ns-namespace.yaml",
					Content: `apiVersion: v1
kind: Namespace
metadata:
  creationTimestamp: null
  name: foo-ns
spec: {}
status: {}
`,
				},
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
		},
		{
			name: "default values for namespace in kustomizations and helm releases",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
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
  targetNamespace: flux-system
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
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
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
  targetNamespace: flux-system
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
  targetNamespace: flux-system
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
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
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
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
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
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
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
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
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
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
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
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
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
  targetNamespace: flux-system
  values: null
status: {}
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "committed files for external secret",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New external secret",
				Description:   "Creates external secret",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						ExternalSecret: &capiv1_protos.ExternalSecret{
							Metadata: testNewMetadata(t, "new-secret", "flux-system"),
							Spec: &capiv1_protos.ExternalSecretSpec{
								RefreshInterval: "1h",
								SecretStoreRef: &capiv1_protos.ExternalSecretStoreRef{
									Name: "testname",
									Kind: "SecretStore",
								},
								Target: &capiv1_protos.ExternalSecretTarget{
									Name: "new-secret",
								},
								Data: &capiv1_protos.ExternalSecretData{
									SecretKey: "test-secret-key",
									RemoteRef: &capiv1_protos.ExternalSecretRemoteRef{
										Key:      "key",
										Property: "property",
									},
								},
							},
						},
					},
				},
			},
			committedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/management/secrets/new-secret-flux-system-externalsecret.yaml",
					Content: `apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  creationTimestamp: null
  name: new-secret
  namespace: flux-system
spec:
  data:
  - remoteRef:
      key: key
      property: property
    secretKey: test-secret-key
  refreshInterval: 1h0m0s
  secretStoreRef:
    kind: SecretStore
    name: testname
  target:
    creationPolicy: Owner
    name: new-secret
status:
  binding: {}
  refreshTime: null
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "validate metadata for external secret",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New external secret",
				Description:   "Creates external secret",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						ExternalSecret: &capiv1_protos.ExternalSecret{
							Spec: &capiv1_protos.ExternalSecretSpec{
								RefreshInterval: "1h",
								SecretStoreRef: &capiv1_protos.ExternalSecretStoreRef{
									Name: "testname",
									Kind: "SecretStore",
								},
								Target: &capiv1_protos.ExternalSecretTarget{
									Name: "new-secret",
								},
								Data: &capiv1_protos.ExternalSecretData{
									SecretKey: "test-secret-key",
									RemoteRef: &capiv1_protos.ExternalSecretRemoteRef{
										Key:      "key",
										Property: "property",
									},
								},
							},
						},
					},
				},
			},
			committedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/management/secrets/new-secret-flux-system-externalsecret.yaml",
					Content: `apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
spec:
  data:
  - remoteRef:
      key: key
      property: property
    secretKey: test-secret-key
  refreshInterval: 1h0m0s
  secretStoreRef:
    kind: SecretStore
    name: testname
  target:
    creationPolicy: Owner
    name: new-secret
status:
  refreshTime: null
`,
				},
			},
			err: errors.New("external secret metadata must be specified"),
		},
		{
			name: "validate name for external secret",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New external secret",
				Description:   "Creates external secret",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						ExternalSecret: &capiv1_protos.ExternalSecret{
							Metadata: testNewMetadata(t, "", "flux-system"),
							Spec: &capiv1_protos.ExternalSecretSpec{
								RefreshInterval: "1h",
								SecretStoreRef: &capiv1_protos.ExternalSecretStoreRef{
									Name: "testname",
									Kind: "SecretStore",
								},
								Target: &capiv1_protos.ExternalSecretTarget{
									Name: "new-secret",
								},
								Data: &capiv1_protos.ExternalSecretData{
									SecretKey: "test-secret-key",
									RemoteRef: &capiv1_protos.ExternalSecretRemoteRef{
										Key:      "key",
										Property: "property",
									},
								},
							},
						},
					},
				},
			},
			committedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/management/secrets/new-secret-flux-system-externalsecret.yaml",
					Content: `apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  creationTimestamp: null
  namespace: flux-system
spec:
  data:
  - remoteRef:
      key: key
      property: property
    secretKey: test-secret-key
  refreshInterval: 1h0m0s
  secretStoreRef:
    kind: SecretStore
    name: testname
  target:
    creationPolicy: Owner
    name: new-secret
status:
  refreshTime: null
`,
				},
			},
			err: errors.New("external secret name must be specified"),
		},
		{
			name: "validate namepace for external secret",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New external secret",
				Description:   "Creates external secret",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						ExternalSecret: &capiv1_protos.ExternalSecret{
							Metadata: testNewMetadata(t, "new-secret", ""),
							Spec: &capiv1_protos.ExternalSecretSpec{
								RefreshInterval: "1h",
								SecretStoreRef: &capiv1_protos.ExternalSecretStoreRef{
									Name: "testname",
									Kind: "SecretStore",
								},
								Target: &capiv1_protos.ExternalSecretTarget{
									Name: "new-secret",
								},
								Data: &capiv1_protos.ExternalSecretData{
									SecretKey: "test-secret-key",
									RemoteRef: &capiv1_protos.ExternalSecretRemoteRef{
										Key:      "key",
										Property: "property",
									},
								},
							},
						},
					},
				},
			},
			committedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/management/secrets/new-secret-flux-system-externalsecret.yaml",
					Content: `apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  creationTimestamp: null
  name: new-secret
spec:
  data:
  - remoteRef:
      key: key
      property: property
    secretKey: test-secret-key
  refreshInterval: 1h0m0s
  secretStoreRef:
    kind: SecretStore
    name: testname
  target:
    creationPolicy: Owner
    name: new-secret
status:
  refreshTime: null
`,
				},
			},
			err: errors.New("external secret namespace must be specified in ExternalSecret new-secret"),
		},
		{
			name: "validate secretstore ref for external secret",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New external secret",
				Description:   "Creates external secret",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						ExternalSecret: &capiv1_protos.ExternalSecret{
							Metadata: testNewMetadata(t, "new-secret", "flux-system"),
							Spec: &capiv1_protos.ExternalSecretSpec{
								RefreshInterval: "1h",
								SecretStoreRef: &capiv1_protos.ExternalSecretStoreRef{
									Kind: "SecretStore",
								},
								Target: &capiv1_protos.ExternalSecretTarget{
									Name: "new-secret",
								},
								Data: &capiv1_protos.ExternalSecretData{
									SecretKey: "test-secret-key",
									RemoteRef: &capiv1_protos.ExternalSecretRemoteRef{
										Key:      "key",
										Property: "property",
									},
								},
							},
						},
					},
				},
			},
			committedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/management/secrets/new-secret-flux-system-externalsecret.yaml",
					Content: `apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  creationTimestamp: null
  name: new-secret
  namespace: flux-system
spec:
  data:
  - remoteRef:
      key: key
      property: property
    secretKey: test-secret-key
  refreshInterval: 1h0m0s
  secretStoreRef:
    kind: SecretStore
  target:
    creationPolicy: Owner
    name: new-secret
status:
  refreshTime: null
`,
				},
			},
			err: errors.New("secretStoreRef name must be specified in ExternalSecret new-secret"),
		},
		{
			name: "validate remote ref for external secret",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New external secret",
				Description:   "Creates external secret",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						ExternalSecret: &capiv1_protos.ExternalSecret{
							Metadata: testNewMetadata(t, "new-secret", "flux-system"),
							Spec: &capiv1_protos.ExternalSecretSpec{
								RefreshInterval: "1h",
								SecretStoreRef: &capiv1_protos.ExternalSecretStoreRef{
									Name: "testname",
									Kind: "SecretStore",
								},
								Target: &capiv1_protos.ExternalSecretTarget{
									Name: "new-secret",
								},
								Data: &capiv1_protos.ExternalSecretData{
									SecretKey: "test-secret-key",
									RemoteRef: &capiv1_protos.ExternalSecretRemoteRef{
										Key: "key",
									},
								},
							},
						},
					},
				},
			},
			committedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/management/secrets/new-secret-flux-system-externalsecret.yaml",
					Content: `apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  creationTimestamp: null
  name: new-secret
  namespace: flux-system
spec:
  data:
  - remoteRef:
      key: key
    secretKey: test-secret-key
  refreshInterval: 1h0m0s
  secretStoreRef:
    kind: SecretStore
    name: testname
  target:
    creationPolicy: Owner
    name: new-secret
status:
  refreshTime: null
`,
				},
			},
			err: errors.New("remoteRef property kind must be specified in ExternalSecret new-secret"),
		},
		{
			name: "committed files for policy config matching workspace",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New policy config",
				Description:   "Creates policy config",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						PolicyConfig: &capiv1_protos.PolicyConfigObject{
							Metadata: testNewMetadata(t, "my-config", ""),
							Spec: &capiv1_protos.PolicyConfigObjectSpec{
								Match: &capiv1_protos.PolicyConfigMatch{
									Workspaces: []string{"devteam"},
								},
								Config: map[string]*capiv1_protos.PolicyConfigConf{
									"policy-1": {
										Parameters: map[string]*structpb.Value{
											"strVal":  structpb.NewStringValue("a"),
											"boolVar": structpb.NewBoolValue(true),
											"intVar":  structpb.NewNumberValue(1),
										},
									},
									"policy-2": {
										Parameters: map[string]*structpb.Value{
											"strVal":  structpb.NewStringValue("b"),
											"boolVar": structpb.NewBoolValue(false),
											"intVar":  structpb.NewNumberValue(2),
										},
									},
								},
							},
						},
					},
				},
			},
			committedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/management/policy-configs/my-config-policy-config.yaml",
					Content: `apiVersion: pac.weave.works/v2beta2
kind: PolicyConfig
metadata:
  creationTimestamp: null
  name: my-config
spec:
  config:
    policy-1:
      parameters:
        boolVar: true
        intVar: 1
        strVal: a
    policy-2:
      parameters:
        boolVar: false
        intVar: 2
        strVal: b
  match:
    workspaces:
    - devteam
status: {}
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "committed files for policy config matching namespace",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New policy config",
				Description:   "Creates policy config",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						PolicyConfig: &capiv1_protos.PolicyConfigObject{
							Metadata: testNewMetadata(t, "my-config", ""),
							Spec: &capiv1_protos.PolicyConfigObjectSpec{
								Match: &capiv1_protos.PolicyConfigMatch{
									Namespaces: []string{"dev"},
								},
								Config: map[string]*capiv1_protos.PolicyConfigConf{
									"policy-1": {
										Parameters: map[string]*structpb.Value{
											"strVal":  structpb.NewStringValue("a"),
											"boolVar": structpb.NewBoolValue(true),
											"intVar":  structpb.NewNumberValue(1),
										},
									},
									"policy-2": {
										Parameters: map[string]*structpb.Value{
											"strVal":  structpb.NewStringValue("b"),
											"boolVar": structpb.NewBoolValue(false),
											"intVar":  structpb.NewNumberValue(2),
										},
									},
								},
							},
						},
					},
				},
			},
			committedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/management/policy-configs/my-config-policy-config.yaml",
					Content: `apiVersion: pac.weave.works/v2beta2
kind: PolicyConfig
metadata:
  creationTimestamp: null
  name: my-config
spec:
  config:
    policy-1:
      parameters:
        boolVar: true
        intVar: 1
        strVal: a
    policy-2:
      parameters:
        boolVar: false
        intVar: 2
        strVal: b
  match:
    namespaces:
    - dev
status: {}
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "committed files for policy config matching application",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New policy config",
				Description:   "Creates policy config",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						PolicyConfig: &capiv1_protos.PolicyConfigObject{
							Metadata: testNewMetadata(t, "my-config", ""),
							Spec: &capiv1_protos.PolicyConfigObjectSpec{
								Match: &capiv1_protos.PolicyConfigMatch{
									Apps: []*capiv1_protos.PolicyConfigApplicationMatch{
										{
											Kind:      "HelmRelease",
											Name:      "my-app",
											Namespace: "test",
										},
									},
								},
								Config: map[string]*capiv1_protos.PolicyConfigConf{
									"policy-1": {
										Parameters: map[string]*structpb.Value{
											"strVal":  structpb.NewStringValue("a"),
											"boolVar": structpb.NewBoolValue(true),
											"intVar":  structpb.NewNumberValue(1),
										},
									},
									"policy-2": {
										Parameters: map[string]*structpb.Value{
											"strVal":  structpb.NewStringValue("b"),
											"boolVar": structpb.NewBoolValue(false),
											"intVar":  structpb.NewNumberValue(2),
										},
									},
								},
							},
						},
					},
				},
			},
			committedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/management/policy-configs/my-config-policy-config.yaml",
					Content: `apiVersion: pac.weave.works/v2beta2
kind: PolicyConfig
metadata:
  creationTimestamp: null
  name: my-config
spec:
  config:
    policy-1:
      parameters:
        boolVar: true
        intVar: 1
        strVal: a
    policy-2:
      parameters:
        boolVar: false
        intVar: 2
        strVal: b
  match:
    apps:
    - kind: HelmRelease
      name: my-app
      namespace: test
status: {}
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "committed files for policy config matching resource",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New policy config",
				Description:   "Creates policy config",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						PolicyConfig: &capiv1_protos.PolicyConfigObject{
							Metadata: testNewMetadata(t, "my-config", ""),
							Spec: &capiv1_protos.PolicyConfigObjectSpec{
								Match: &capiv1_protos.PolicyConfigMatch{
									Resources: []*capiv1_protos.PolicyConfigResourceMatch{
										{
											Kind:      "Deployment",
											Name:      "my-deployment",
											Namespace: "test",
										},
									},
								},
								Config: map[string]*capiv1_protos.PolicyConfigConf{
									"policy-1": {
										Parameters: map[string]*structpb.Value{
											"strVal":  structpb.NewStringValue("a"),
											"boolVar": structpb.NewBoolValue(true),
											"intVar":  structpb.NewNumberValue(1),
										},
									},
									"policy-2": {
										Parameters: map[string]*structpb.Value{
											"strVal":  structpb.NewStringValue("b"),
											"boolVar": structpb.NewBoolValue(false),
											"intVar":  structpb.NewNumberValue(2),
										},
									},
								},
							},
						},
					},
				},
			},
			committedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "clusters/management/policy-configs/my-config-policy-config.yaml",
					Content: `apiVersion: pac.weave.works/v2beta2
kind: PolicyConfig
metadata:
  creationTimestamp: null
  name: my-config
spec:
  config:
    policy-1:
      parameters:
        boolVar: true
        intVar: 1
        strVal: a
    policy-2:
      parameters:
        boolVar: false
        intVar: 2
        strVal: b
  match:
    resources:
    - kind: Deployment
      name: my-deployment
      namespace: test
status: {}
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "invalid policy config missing matches",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New policy config",
				Description:   "Creates policy config",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						PolicyConfig: &capiv1_protos.PolicyConfigObject{
							Metadata: testNewMetadata(t, "my-config", ""),
							Spec: &capiv1_protos.PolicyConfigObjectSpec{
								Match: &capiv1_protos.PolicyConfigMatch{},
								Config: map[string]*capiv1_protos.PolicyConfigConf{
									"policy-1": {
										Parameters: map[string]*structpb.Value{
											"strVal":  structpb.NewStringValue("a"),
											"boolVar": structpb.NewBoolValue(true),
											"intVar":  structpb.NewNumberValue(1),
										},
									},
									"policy-2": {
										Parameters: map[string]*structpb.Value{
											"strVal":  structpb.NewStringValue("b"),
											"boolVar": structpb.NewBoolValue(false),
											"intVar":  structpb.NewNumberValue(2),
										},
									},
								},
							},
						},
					},
				},
			},
			err: errors.New("policy config must target workspaces, namespaces, applications or resources"),
		},
		{
			name: "invalid policy config missing policy configs",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New policy config",
				Description:   "Creates policy config",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						PolicyConfig: &capiv1_protos.PolicyConfigObject{
							Metadata: testNewMetadata(t, "my-config", ""),
							Spec: &capiv1_protos.PolicyConfigObjectSpec{
								Match: &capiv1_protos.PolicyConfigMatch{
									Resources: []*capiv1_protos.PolicyConfigResourceMatch{
										{
											Kind:      "Deployment",
											Name:      "my-deployment",
											Namespace: "test",
										},
									},
								},
							},
						},
					},
				},
			},
			err: errors.New("policy config configuration must be specified"),
		},
		{
			name: "invalid policy config missing parameters",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New policy config",
				Description:   "Creates policy config",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						PolicyConfig: &capiv1_protos.PolicyConfigObject{
							Metadata: testNewMetadata(t, "my-config", ""),
							Spec: &capiv1_protos.PolicyConfigObjectSpec{
								Match: &capiv1_protos.PolicyConfigMatch{
									Resources: []*capiv1_protos.PolicyConfigResourceMatch{
										{
											Kind:      "Deployment",
											Name:      "my-deployment",
											Namespace: "test",
										},
									},
								},
								Config: map[string]*capiv1_protos.PolicyConfigConf{
									"policy-1": {
										Parameters: map[string]*structpb.Value{},
									},
								},
							},
						},
					},
				},
			},
			err: errors.New("policy policy-1 configuration must have at least one parameter"),
		},
		{
			name: "invalid policy config different matches",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New policy config",
				Description:   "Creates policy config",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						PolicyConfig: &capiv1_protos.PolicyConfigObject{
							Metadata: testNewMetadata(t, "my-config", ""),
							Spec: &capiv1_protos.PolicyConfigObjectSpec{
								Match: &capiv1_protos.PolicyConfigMatch{
									Workspaces: []string{"devteam"},
									Namespaces: []string{"dev"},
								},
								Config: map[string]*capiv1_protos.PolicyConfigConf{
									"policy-1": {
										Parameters: map[string]*structpb.Value{
											"strVal": structpb.NewStringValue("a"),
										},
									},
								},
							},
						},
					},
				},
			},
			err: errors.New("cannot target workspaces and namespaces in same policy config"),
		},
		{
			name: "invalid policy config different matches",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New policy config",
				Description:   "Creates policy config",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						PolicyConfig: &capiv1_protos.PolicyConfigObject{
							Metadata: testNewMetadata(t, "my-config", ""),
							Spec: &capiv1_protos.PolicyConfigObjectSpec{
								Match: &capiv1_protos.PolicyConfigMatch{
									Workspaces: []string{"devteam"},
									Namespaces: []string{"dev"},
								},
								Config: map[string]*capiv1_protos.PolicyConfigConf{},
							},
						},
					},
				},
			},
			err: errors.New("policy config configuration must be specified"),
		},
		{
			name: "invalid policy config empty app kind",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New policy config",
				Description:   "Creates policy config",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						PolicyConfig: &capiv1_protos.PolicyConfigObject{
							Metadata: testNewMetadata(t, "my-config", ""),
							Spec: &capiv1_protos.PolicyConfigObjectSpec{
								Match: &capiv1_protos.PolicyConfigMatch{
									Apps: []*capiv1_protos.PolicyConfigApplicationMatch{
										{
											Kind: "",
											Name: "test",
										},
									},
								},
								Config: map[string]*capiv1_protos.PolicyConfigConf{
									"policy-1": {
										Parameters: map[string]*structpb.Value{
											"strVal": structpb.NewStringValue("a"),
										},
									},
								},
							},
						},
					},
				},
			},
			err: errors.New("invalid matches, application kind is required"),
		},
		{
			name: "invalid policy config empty app name",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New policy config",
				Description:   "Creates policy config",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						PolicyConfig: &capiv1_protos.PolicyConfigObject{
							Metadata: testNewMetadata(t, "my-config", ""),
							Spec: &capiv1_protos.PolicyConfigObjectSpec{
								Match: &capiv1_protos.PolicyConfigMatch{
									Apps: []*capiv1_protos.PolicyConfigApplicationMatch{
										{
											Kind: "Deployment",
											Name: "",
										},
									},
								},
								Config: map[string]*capiv1_protos.PolicyConfigConf{
									"policy-1": {
										Parameters: map[string]*structpb.Value{
											"strVal": structpb.NewStringValue("a"),
										},
									},
								},
							},
						},
					},
				},
			},
			err: errors.New("invalid matches, application name is required"),
		},
		{
			name: "committed files for sops secret",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New sops secret",
				Description:   "Creates sops secret",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						FilePath:       "./secrets-age",
						SopsSecret: &capiv1_protos.SopsSecret{
							ApiVersion: "v1",
							Kind:       "Secret",
							Metadata: &capiv1_protos.SopsSecretMetadata{
								Name:      "my-secret",
								Namespace: "my-namepsace",
								Labels: map[string]string{
									"label-1": "value-1",
									"label-2": "value-2",
								},
							},
							Data: map[string]string{
								"username": "admin",
								"password": "password",
							},
							Type:      "Opaque",
							Immutable: true,
						},
					},
				},
			},
			committedFiles: []*capiv1_protos.CommitFile{
				{
					Path: "secrets-age/my-secret-my-namepsace-sops-secret.yaml",
					Content: `apiVersion: v1
data:
  password: password
  username: admin
immutable: true
kind: Secret
metadata:
  labels:
    label-1: value-1
    label-2: value-2
  name: my-secret
  namespace: my-namepsace
type: Opaque
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "committed files for sops secret without name",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New sops secret",
				Description:   "Creates sops secret",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						FilePath:       "./secrets-age",
						SopsSecret: &capiv1_protos.SopsSecret{
							ApiVersion: "v1",
							Kind:       "Secret",
							Metadata: &capiv1_protos.SopsSecretMetadata{
								Namespace: "my-namepsace",
								Labels: map[string]string{
									"label-1": "value-1",
									"label-2": "value-2",
								},
							},
							StringData: map[string]string{
								"username": "admin",
								"password": "password",
							},
							Type:      "Opaque",
							Immutable: true,
						},
					},
				},
			},
			err: errors.New("missing secret name"),
		},
		{
			name: "committed files for sops secret without namespace",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New sops secret",
				Description:   "Creates sops secret",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						FilePath:       "./secrets-age",
						SopsSecret: &capiv1_protos.SopsSecret{
							ApiVersion: "v1",
							Kind:       "Secret",
							Metadata: &capiv1_protos.SopsSecretMetadata{
								Name: "my-secret",
								Labels: map[string]string{
									"label-1": "value-1",
									"label-2": "value-2",
								},
							},
							Data: map[string]string{
								"username": "admin",
								"password": "password",
							},
							Type:      "Opaque",
							Immutable: true,
						},
					},
				},
			},
			err: errors.New("missing secret namespace"),
		},
		{
			name: "committed files for sops secret without data and stringData",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New sops secret",
				Description:   "Creates sops secret",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						FilePath:       "./secrets-age",
						SopsSecret: &capiv1_protos.SopsSecret{
							ApiVersion: "v1",
							Kind:       "Secret",
							Metadata: &capiv1_protos.SopsSecretMetadata{
								Name:      "my-secret",
								Namespace: "my-namepsace",
								Labels: map[string]string{
									"label-1": "value-1",
									"label-2": "value-2",
								},
							},
							Type:      "Opaque",
							Immutable: true,
						},
					},
				},
			},
			err: errors.New("key/value pairs must be set in either data or stringData"),
		},
		{
			name: "committed files for sops secret with data and stringData",
			clusterState: []runtime.Object{
				makeCAPITemplate(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateAutomationsPullRequestRequest{
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New sops secret",
				Description:   "Creates sops secret",
				ClusterAutomations: []*capiv1_protos.ClusterAutomation{
					{
						Cluster:        testNewClusterNamespacedName(t, "management", "default"),
						IsControlPlane: true,
						FilePath:       "./secrets-age",
						SopsSecret: &capiv1_protos.SopsSecret{
							ApiVersion: "v1",
							Kind:       "Secret",
							Metadata: &capiv1_protos.SopsSecretMetadata{
								Name:      "my-secret",
								Namespace: "my-namepsace",
								Labels: map[string]string{
									"label-1": "value-1",
									"label-2": "value-2",
								},
							},
							Type:      "Opaque",
							Immutable: true,
							Data: map[string]string{
								"username": "admin",
								"password": "password",
							},
							StringData: map[string]string{
								"username": "admin",
								"password": "password",
							},
						},
					},
				},
			},
			err: errors.New("expected only one of data or stringData fields, but found both"),
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
			})

			// request
			createPullRequestResponse, err := s.CreateAutomationsPullRequest(context.Background(), tt.req)

			// Check the response looks good
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to create a pull request:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					fmt.Println(tt.err.Error(), err.Error())
					t.Fatalf("got the wrong error:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, createPullRequestResponse.WebUrl, protocmp.Transform()); diff != "" {
					fmt.Println("==============", tt.name)
					t.Fatalf("pull request url didn't match expected:\n%s", diff)
				}
				fakeGitProvider := (tt.provider).(*gitfakes.FakeGitProvider)
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
		name               string
		clusterState       []runtime.Object
		pruneEnvVar        string
		req                *capiv1_protos.RenderAutomationRequest
		expected           string
		kustomizationFiles []*capiv1_protos.CommitFile
		helmreleaseFiles   []*capiv1_protos.CommitFile
		err                error
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
			kustomizationFiles: []*capiv1_protos.CommitFile{
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
			},
			helmreleaseFiles: []*capiv1_protos.CommitFile{
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
  targetNamespace: flux-system
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
				if diff := cmp.Diff(tt.kustomizationFiles, renderAutomationResponse.KustomizationFiles, protocmp.Transform()); diff != "" {
					t.Fatalf("kustomization files do not match expected committed files:\n%s", diff)
				}
				if diff := cmp.Diff(tt.helmreleaseFiles, renderAutomationResponse.HelmReleaseFiles, protocmp.Transform()); diff != "" {
					t.Fatalf("helmrelease files do not match expected committed files:\n%s", diff)
				}
			}
		})
	}
}
