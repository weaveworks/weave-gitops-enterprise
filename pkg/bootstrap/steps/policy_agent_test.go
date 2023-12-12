package steps

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestInstallPolicyAgent(t *testing.T) {
	tests := []struct {
		name   string
		input  []StepInput
		output []StepOutput
		config Config
		err    bool
	}{
		{
			name:  "install policy agent controller",
			input: []StepInput{},
			output: []StepOutput{
				{
					Name: agentHelmReleaseFileName,
					Type: typeFile,
					Value: fileContent{
						Name:      agentHelmReleaseFileName,
						Content:   getControllerHelmReleaseTestFile(agentControllerURL),
						CommitMsg: agentHelmReleaseCommitMsg,
					},
				},
			},
			config: Config{},
			err:    false,
		},
		{
			name:   "do not install policy agent controller if it's already installed",
			input:  []StepInput{},
			output: []StepOutput{},
			config: Config{
				ComponentsExtra: ComponentsExtraConfig{
					Requested: []string{policyAgentController},
					Existing:  []string{policyAgentController},
				},
			},
			err: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := MakeTestConfig(t, tt.config)
			step := NewInstallPolicyAgentStep(config)
			out, err := step.Execute(&config)
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
		name   string
		config Config
		want   BootstrapStep
	}{
		{
			name: "return bootstrap step",
			want: BootstrapStep{
				Name:  "install Policy Agent",
				Input: []StepInput{},
			},
			config: Config{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := MakeTestConfig(t, tt.config)
			step := NewInstallPolicyAgentStep(config)

			assert.Equal(t, tt.want.Name, step.Name)
			if diff := cmp.Diff(tt.want.Input, step.Input); diff != "" {
				t.Fatalf("different step expected:\n%s", diff)
			}
		})
	}
}
