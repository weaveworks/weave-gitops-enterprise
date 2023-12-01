package steps

import (
	"testing"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	helmrepositoryTestFile = `apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: weave-gitops-enterprise-charts
  namespace: flux-system
spec:
  interval: 1m0s
  secretRef:
    name: weave-gitops-enterprise-credentials
  url: https://charts.dev.wkp.weave.works/releases/charts-v3
status: {}
`
	hrFileContentLocalhost = `apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: weave-gitops-enterprise
  namespace: flux-system
spec:
  chart:
    spec:
      chart: mccp
      reconcileStrategy: ChartVersion
      sourceRef:
        kind: HelmRepository
        name: weave-gitops-enterprise-charts
        namespace: flux-system
      version: 1.0.0
  install:
    crds: CreateReplace
  interval: 1h0m0s
  upgrade:
    crds: CreateReplace
  values:
    cluster-controller:
      controllerManager:
        manager:
          image:
            repository: docker.io/weaveworks/cluster-controller
            tag: v1.5.2
      enabled: true
      fullnameOverride: cluster
    config: {}
    enablePipelines: true
    gitopssets-controller:
      controllerManager:
        manager:
          args:
          - --health-probe-bind-address=:8081
          - --metrics-bind-address=127.0.0.1:8080
          - --leader-elect
          - --enabled-generators=GitRepository,Cluster,PullRequests,List,APIClient,Matrix,Config
      enabled: true
    global: {}
    tls:
      enabled: false
status: {}
`
	hrFileContentExternalDns = `apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: weave-gitops-enterprise
  namespace: flux-system
spec:
  chart:
    spec:
      chart: mccp
      reconcileStrategy: ChartVersion
      sourceRef:
        kind: HelmRepository
        name: weave-gitops-enterprise-charts
        namespace: flux-system
      version: 1.0.0
  install:
    crds: CreateReplace
  interval: 1h0m0s
  upgrade:
    crds: CreateReplace
  values:
    cluster-controller:
      controllerManager:
        manager:
          image:
            repository: docker.io/weaveworks/cluster-controller
            tag: v1.5.2
      enabled: true
      fullnameOverride: cluster
    config: {}
    enablePipelines: true
    gitopssets-controller:
      controllerManager:
        manager:
          args:
          - --health-probe-bind-address=:8081
          - --metrics-bind-address=127.0.0.1:8080
          - --leader-elect
          - --enabled-generators=GitRepository,Cluster,PullRequests,List,APIClient,Matrix,Config
      enabled: true
    global: {}
    ingress:
      annotations:
        external-dns.alpha.kubernetes.io/hostname: example.com
      className: public-nginx
      enabled: true
      hosts:
      - host: example.com
        paths:
        - path: /
          pathType: ImplementationSpecific
    tls:
      enabled: false
status: {}
`
)

func TestInstallWge_Execute(t *testing.T) {

	tests := []struct {
		name       string
		config     Config
		wantOutput []StepOutput
		wantErr    string
	}{
		{
			name: "should install weave gitops enterprise",
			config: makeTestConfig(t, Config{
				WGEVersion:  "1.0.0",
				GitUsername: "test",
				GitToken:    "abc",
				GitRepository: GitRepositoryConfig{
					Url:    "https://test.com.git",
					Branch: "main",
					Path:   "/",
					Scheme: "https",
				},
			}, fluxSystemGitRepository(), fluxSystemKustomization()),
			wantOutput: []StepOutput{
				{
					Name: wgeHelmrepoFileName,
					Type: typeFile,
					Value: fileContent{
						Name:      wgeHelmrepoFileName,
						Content:   helmrepositoryTestFile,
						CommitMsg: wgeHelmRepoCommitMsg,
					},
				},
				{
					Name: wgeHelmReleaseFileName,
					Type: typeFile,
					Value: fileContent{
						Name:      wgeHelmReleaseFileName,
						Content:   hrFileContentLocalhost,
						CommitMsg: wgeHelmReleaseCommitMsg,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := NewInstallWGEStep()
			gotOutputs, err := step.Execute(&tt.config)
			if tt.wantErr != "" {
				if msg := err.Error(); msg != tt.wantErr {
					t.Fatalf("got error %q, want %q", msg, tt.wantErr)
				}
				return
			}

			assert.NoError(t, err)
			if diff := cmp.Diff(tt.wantOutput, gotOutputs, cmpopts.IgnoreFields(v1.Secret{}, "Data")); diff != "" {
				t.Fatalf("expected output:\n%s", diff)
			}
		})
	}
}

func TestInstallWge(t *testing.T) {
	tests := []struct {
		name       string
		domainType string
		input      []StepInput
		output     []StepOutput
		err        bool
	}{
		{
			name:       "unsupported domain type",
			domainType: "wrongType",
			input:      []StepInput{},
			err:        true,
		},
		{
			name:  "install with domaintype localhost",
			input: []StepInput{},
			output: []StepOutput{
				{
					Name: wgeHelmrepoFileName,
					Type: typeFile,
					Value: fileContent{
						Name:      wgeHelmrepoFileName,
						Content:   helmrepositoryTestFile,
						CommitMsg: wgeHelmRepoCommitMsg,
					},
				},
				{
					Name: wgeHelmReleaseFileName,
					Type: typeFile,
					Value: fileContent{
						Name:      wgeHelmReleaseFileName,
						Content:   hrFileContentLocalhost,
						CommitMsg: wgeHelmReleaseCommitMsg,
					},
				},
			},
			err: false,
		},
		{
			name:  "install with domaintype external dns",
			input: []StepInput{},
			output: []StepOutput{
				{
					Name: wgeHelmrepoFileName,
					Type: typeFile,
					Value: fileContent{
						Name:      wgeHelmrepoFileName,
						Content:   helmrepositoryTestFile,
						CommitMsg: wgeHelmRepoCommitMsg,
					},
				},
				{
					Name: wgeHelmReleaseFileName,
					Type: typeFile,
					Value: fileContent{
						Name:      wgeHelmReleaseFileName,
						Content:   hrFileContentExternalDns,
						CommitMsg: wgeHelmReleaseCommitMsg,
					},
				},
			},
			err: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testConfig := Config{
				WGEVersion: "1.0.0",
			}

			config := makeTestConfig(t, testConfig)

			out, err := installWge(tt.input, &config)
			if err != nil {
				if tt.err {
					return
				}
				t.Fatalf("error install wge: %v", err)
			}

			for i, item := range out {
				assert.Equal(t, item.Name, tt.output[i].Name, "wrong name")
				assert.Equal(t, item.Type, tt.output[i].Type, "wrong type")
				inFileContent, ok := tt.output[i].Value.(fileContent)
				if !ok {
					t.Fatalf("error install wge: %v", err)
				}
				outFileContent, ok := item.Value.(fileContent)
				if !ok {
					t.Fatalf("error install wge: %v", err)
				}
				assert.Equal(t, outFileContent.CommitMsg, inFileContent.CommitMsg, "wrong commit msg")
				assert.Equal(t, outFileContent.Name, inFileContent.Name, "wrong filename")
				assert.Equal(t, outFileContent.Content, inFileContent.Content, "wrong content")
			}
		})
	}
}

func fluxSystemGitRepository() *sourcev1.GitRepository {
	return &sourcev1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flux-system",
			Namespace: "flux-system",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.GitRepositoryKind,
			APIVersion: sourcev1.GroupVersion.String(),
		},
		Spec: sourcev1.GitRepositorySpec{
			URL: "https://example.com/owner/repo",
			Reference: &sourcev1.GitRepositoryRef{
				Branch: "main",
			},
		},
	}
}

func fluxSystemKustomization() *kustomizev1.Kustomization {
	return &kustomizev1.Kustomization{
		TypeMeta: metav1.TypeMeta{
			Kind:       kustomizev1.KustomizationKind,
			APIVersion: kustomizev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flux-system",
			Namespace: "flux-system",
		},
		Spec: kustomizev1.KustomizationSpec{
			Path: "/foo",
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind:      sourcev1.GitRepositoryKind,
				Name:      "flux-system",
				Namespace: "flux-system",
			},
		},
	}
}
