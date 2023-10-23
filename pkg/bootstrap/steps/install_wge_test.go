package steps

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
    ingress:
      annotations:
        external-dns.alpha.kubernetes.io/hostname: localhost
      className: public-nginx
      enabled: true
      hosts:
      - host: localhost
        paths:
        - path: /
          pathType: ImplementationSpecific
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
			input: []StepInput{
				{
					Name:  UserDomain,
					Value: "example.com",
				},
			},
			err: true,
		},
		{
			name:       "install with domaintype localhost",
			domainType: domainTypeLocalhost,
			input: []StepInput{
				{
					Name:  UserDomain,
					Value: "localhost",
				},
			},
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
			name:       "install with domaintype external dns",
			domainType: domainTypeExternalDNS,
			input: []StepInput{
				{
					Name:  UserDomain,
					Value: "example.com",
				},
			},
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
				DomainType: tt.domainType,
			}

			config, err := makeTestConfig(t, testConfig)
			if err != nil {
				t.Fatalf("error creating config: %v", err)
			}

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
