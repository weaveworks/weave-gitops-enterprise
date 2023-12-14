package steps

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/test/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	expectedHelmRepository = `apiVersion: source.toolkit.fluxcd.io/v1beta2
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
	expectedHelmRelease = `apiVersion: helm.toolkit.fluxcd.io/v2beta1
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
      annotations: {}
      className: ""
      enabled: false
      hosts:
      - host: ""
        paths:
        - path: /
          pathType: ImplementationSpecific
      service:
        name: clusters-service
        port: 8000
      tls: []
    service:
      annotations: {}
      clusterIP: ""
      externalIPs: []
      externalTrafficPolicy: ""
      healthCheckNodePort: 0
      loadBalancerIP: ""
      loadBalancerSourceRanges: []
      nodePorts:
        http: ""
        https: ""
        tcp: {}
        udp: {}
      port:
        https: 8000
      targetPort:
        https: 8000
      type: ClusterIP
    tls:
      enabled: false
status: {}
`
)

func TestInstallWge_Execute(t *testing.T) {

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "testuser" || password != "testpassword" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		_, _ = fmt.Fprintln(w, `entries:
  mccp:
  - version: 1.0.0
    name: mccp
  - version: 1.1.0
    name: mccp
  - version: 1.2.0
    name: mccp`)
	}))
	defer mockServer.Close()

	secretName := "weave-gitops-enterprise-credentials"
	secretNamespace := "flux-system"
	fakeClient := utils.CreateFakeClient(t, &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: secretNamespace},
		Type:       "Opaque",
		Data: map[string][]byte{
			"username": []byte("testuser"),
			"password": []byte("testpassword"),
		},
	})

	tests := []struct {
		name       string
		config     Config
		wantOutput []StepOutput
		wantErr    string
	}{
		{
			name: "should install weave gitops enterprise",
			config: MakeTestConfig(t, Config{
				WGEVersion:  "1.0.0",
				GitUsername: "test",
				GitToken:    "abc",
				GitRepository: GitRepositoryConfig{
					Url:    "https://test.com.git",
					Branch: "main",
					Path:   "/",
					Scheme: "https",
				},
				IsExistingWgeInstallation: false,
				ChartURL:                  mockServer.URL,
			}, fluxSystemGitRepository(), fluxSystemKustomization()),
			wantOutput: []StepOutput{
				{
					Name: wgeHelmrepoFileName,
					Type: typeFile,
					Value: fileContent{
						Name:      wgeHelmrepoFileName,
						Content:   expectedHelmRepository,
						CommitMsg: wgeHelmRepoCommitMsg,
					},
				},
				{
					Name: wgeHelmReleaseFileName,
					Type: typeFile,
					Value: fileContent{
						Name:      wgeHelmReleaseFileName,
						Content:   expectedHelmRelease,
						CommitMsg: wgeHelmReleaseCommitMsg,
					},
				},
			},
		},
		{
			name: "should not install weave gitops enterprise if it already exists",
			config: MakeTestConfig(t, Config{
				WGEVersion:  "1.0.0",
				GitUsername: "test",
				GitToken:    "abc",
				GitRepository: GitRepositoryConfig{
					Url:    "https://test.com.git",
					Branch: "main",
					Path:   "/",
					Scheme: "https",
				},
				IsExistingWgeInstallation: true,
				ChartURL:                  mockServer.URL,
			}, fluxSystemGitRepository(), fluxSystemKustomization()),
			wantOutput: []StepOutput{},
		},
		// a case when WGEversion is not specified
		{
			name: "should install weave gitops enterprise by select the version interactively",
			config: MakeTestConfig(t, Config{
				GitUsername: "test",
				GitToken:    "abc",
				GitRepository: GitRepositoryConfig{
					Url:    "https://test.com.git",
					Branch: "main",
					Path:   "/",
					Scheme: "https",
				},
				IsExistingWgeInstallation: false,
				ChartURL:                  mockServer.URL,
			}, fluxSystemGitRepository(), fluxSystemKustomization()),
			wantOutput: []StepOutput{
				{
					Name: wgeHelmrepoFileName,
					Type: typeFile,
					Value: fileContent{
						Name:      wgeHelmrepoFileName,
						Content:   expectedHelmRepository,
						CommitMsg: wgeHelmRepoCommitMsg,
					},
				},
				{
					Name: wgeHelmReleaseFileName,
					Type: typeFile,
					Value: fileContent{
						Name:      wgeHelmReleaseFileName,
						Content:   expectedHelmRelease,
						CommitMsg: wgeHelmReleaseCommitMsg,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//use fake client
			tt.config.KubernetesClient = fakeClient
			step := NewInstallWGEStep(tt.config)
			gotOutputs, err := step.Execute(&tt.config)
			if tt.wantErr != "" {
				if msg := err.Error(); msg != tt.wantErr {
					t.Fatalf("got error %q, want %q", msg, tt.wantErr)
				}
				return
			}

			assert.NoError(t, err)
			if diff := cmp.Diff(tt.wantOutput, gotOutputs); diff != "" {
				t.Fatalf("unexpected wge outputs:\n%s", diff)
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
