package steps

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

const (
	testAgentHelmRepoFile = `apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: policy-agent
  namespace: flux-system
spec:
  interval: 1m0s
  url: https://weaveworks.github.io/policy-agent/
status: {}
`
	testAgentHelmReleaseFileAdmissionDisabled = `apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: policy-agent
  namespace: flux-system
spec:
  chart:
    spec:
      chart: policy-agent
      reconcileStrategy: ChartVersion
      sourceRef:
        kind: HelmRepository
        name: policy-agent
        namespace: flux-system
      version: 2.5.0
  install:
    crds: CreateReplace
    createNamespace: true
  interval: 10m0s
  values:
    config:
      admission:
        enabled: false
        sinks:
          k8sEventsSink:
            enabled: true
      audit:
        enabled: true
        sinks:
          k8sEventsSink:
            enabled: true
    excludeNamespaces:
    - kube-system
    - flux-system
    failurePolicy: Fail
    useCertManager: true
status: {}
`
	testAgentHelmReleaseFileAdmissionEnabled = `apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: policy-agent
  namespace: flux-system
spec:
  chart:
    spec:
      chart: policy-agent
      reconcileStrategy: ChartVersion
      sourceRef:
        kind: HelmRepository
        name: policy-agent
        namespace: flux-system
      version: 2.5.0
  install:
    crds: CreateReplace
    createNamespace: true
  interval: 10m0s
  values:
    config:
      admission:
        enabled: true
        sinks:
          k8sEventsSink:
            enabled: true
      audit:
        enabled: true
        sinks:
          k8sEventsSink:
            enabled: true
    excludeNamespaces:
    - kube-system
    - flux-system
    failurePolicy: Fail
    useCertManager: true
status: {}
`
)

func TestInstallPolicyAgent(t *testing.T) {
	tests := []struct {
		name   string
		input  []StepInput
		output []StepOutput
		err    bool
	}{
		{
			name: "install policy agent controller with admission disabled",
			input: []StepInput{
				{
					Name:  inEnableAdmission,
					Type:  confirmInput,
					Msg:   admissionMsg,
					Value: confirmNo,
				},
			},
			output: []StepOutput{
				{
					Name: agentHelmRepoFileName,
					Type: typeFile,
					Value: fileContent{
						Name:      agentHelmRepoFileName,
						Content:   testAgentHelmRepoFile,
						CommitMsg: agentHelmRepoCommitMsg,
					},
				},
				{
					Name: agentHelmReleaseFileName,
					Type: typeFile,
					Value: fileContent{
						Name:      agentHelmReleaseFileName,
						Content:   testAgentHelmReleaseFileAdmissionDisabled,
						CommitMsg: agentHelmReleaseCommitMsg,
					},
				},
			},
			err: false,
		},
		{
			name: "install policy agent controller with admission enabled",
			input: []StepInput{
				{
					Name:  inEnableAdmission,
					Type:  confirmInput,
					Msg:   admissionMsg,
					Value: confirmYes,
				},
			},
			output: []StepOutput{
				{
					Name: agentHelmRepoFileName,
					Type: typeFile,
					Value: fileContent{
						Name:      agentHelmRepoFileName,
						Content:   testAgentHelmRepoFile,
						CommitMsg: agentHelmRepoCommitMsg,
					},
				},
				{
					Name: agentHelmReleaseFileName,
					Type: typeFile,
					Value: fileContent{
						Name:      agentHelmReleaseFileName,
						Content:   testAgentHelmReleaseFileAdmissionEnabled,
						CommitMsg: agentHelmReleaseCommitMsg,
					},
				},
			},
			err: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testConfig := Config{}

			config := makeTestConfig(t, testConfig)
			out, err := installPolicyAgent(tt.input, &config)
			if err != nil {
				if tt.err {
					return
				}
				t.Fatalf("error install policy-agent: %v", err)
			}

			for i, item := range out {
				assert.Equal(t, item.Name, tt.output[i].Name, "wrong name")
				assert.Equal(t, item.Type, tt.output[i].Type, "wrong type")
				inFileContent, ok := tt.output[i].Value.(fileContent)
				if !ok {
					t.Fatalf("error install policy-agent: %v", err)
				}
				outFileContent, ok := item.Value.(fileContent)
				if !ok {
					t.Fatalf("error install policy-agent: %v", err)
				}
				assert.Equal(t, outFileContent.CommitMsg, inFileContent.CommitMsg, "wrong commit msg")
				assert.Equal(t, outFileContent.Name, inFileContent.Name, "wrong filename")
				assert.Equal(t, outFileContent.Content, inFileContent.Content, "wrong content")
			}
		})
	}

}

func TestNewInstallPolicyAgentStep(t *testing.T) {
	tests := []struct {
		name string
		want BootstrapStep
	}{

		{
			name: "return bootstrap step",
			want: BootstrapStep{
				Name: "install Policy Agent",
				Input: []StepInput{
					enableAdmission,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := makeTestConfig(t, Config{})
			step := NewInstallPolicyAgentStep(config)

			assert.Equal(t, tt.want.Name, step.Name)
			if diff := cmp.Diff(tt.want.Input, step.Input); diff != "" {
				t.Fatalf("different step expected:\n%s", diff)
			}
		})
	}
}