package steps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testTFControllerFile = `---
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: tf-controller
  namespace: flux-system
spec:
  interval: 1h0s
  type: oci
  url: oci://ghcr.io/weaveworks/charts
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: tf-controller
  namespace: flux-system
spec:
  chart:
    spec:
      chart: tf-controller
      sourceRef:
        kind: HelmRepository
        name: tf-controller
      version: '>=0.16.0-rc.3'
  interval: 1h0s
  releaseName: tf-controller
  targetNamespace: flux-system
  install:
    crds: Create
    remediation:
      retries: -1
  upgrade:
    crds: CreateReplace
    remediation:
      retries: -1
  values:
    replicaCount: 3
    concurrency: 24
    resources:
      limits:
        cpu: 1000m
        memory: 2Gi
      requests:
        cpu: 400m
        memory: 64Mi
    caCertValidityDuration: 24h
    certRotationCheckFrequency: 30m
    image:
      tag: v0.16.0-rc.3
    runner:
      image:
        tag: v0.16.0-rc.3
      grpc:
        maxMessageSize: 30
`

	wgeHRFakeFileTFController = `apiVersion: helm.toolkit.fluxcd.io/v2beta1
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
    enableTerraformUI: true
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
)

func TestInstallTerraform(t *testing.T) {
	tests := []struct {
		name   string
		input  []StepInput
		output []StepOutput
		err    bool
	}{
		{
			name: "install tf controller",
			output: []StepOutput{
				{
					Name: tfFileName,
					Type: typeFile,
					Value: fileContent{
						Name:      tfFileName,
						Content:   testTFControllerFile,
						CommitMsg: tfCommitMsg,
					},
				},
				{
					Name: wgeHelmReleaseFileName,
					Type: typeFile,
					Value: fileContent{
						Name:      wgeHelmReleaseFileName,
						Content:   wgeHRFakeFileTFController,
						CommitMsg: tfCommitMsg,
					},
				},
			},
			err: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testConfig := Config{
				Silent:     true,
				WGEVersion: "1.0.0",
			}
			wgeObject, err := createWGEHelmReleaseFakeObject("1.0.0")
			if err != nil {
				t.Fatalf("error create wge object: %v", err)
			}

			config := makeTestConfig(t, testConfig, &wgeObject)

			out, err := installTerraform(tt.input, &config)
			if err != nil {
				if tt.err {
					return
				}
				t.Fatalf("error install tf controller: %v", err)
			}

			for i, item := range out {
				assert.Equal(t, item.Name, tt.output[i].Name, "wrong name")
				assert.Equal(t, item.Type, tt.output[i].Type, "wrong type")
				inFileContent, ok := tt.output[i].Value.(fileContent)
				if !ok {
					t.Fatalf("error install tf controller: %v", err)
				}
				outFileContent, ok := item.Value.(fileContent)
				if !ok {
					t.Fatalf("error install tf controller: %v", err)
				}
				assert.Equal(t, outFileContent.CommitMsg, inFileContent.CommitMsg, "wrong commit msg")
				assert.Equal(t, outFileContent.Name, inFileContent.Name, "wrong filename")
				assert.Equal(t, outFileContent.Content, inFileContent.Content, "wrong content")
			}
		})
	}

}
