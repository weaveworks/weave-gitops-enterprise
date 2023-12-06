package steps

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestNewInstallExtraComponents(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   BootstrapStep
	}{
		{
			name:   "return bootstrap step with inputs in case provided by user",
			config: Config{},
			want: BootstrapStep{
				Name: "install extra components",
				Input: []StepInput{
					{
						Name:         inComponentsExtra,
						Type:         multiSelectionChoice,
						Msg:          componentsExtraMsg,
						Values:       ComponentsExtra,
						DefaultValue: "",
					},
				},
				Step: installExtraComponents,
			},
		},
		{
			name: "return bootstrap step with empty inputs in case not provided by user",
			config: Config{
				ComponentsExtra: ComponentsExtraConfig{
					Requested: []string{
						policyAgentController,
						tfController,
					},
				},
			},
			want: BootstrapStep{
				Name:  "install extra components",
				Input: []StepInput{},
				Step:  installExtraComponents,
			},
		},
		{
			name: "return empty step in case of silent mode",
			config: Config{
				Silent: true,
			},
			want: BootstrapStep{
				Name:  "install extra components",
				Input: []StepInput{},
				Step:  installExtraComponents,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := makeTestConfig(t, tt.config)
			stepConfig, err := NewInstallExtraComponentsConfig(tt.config.ComponentsExtra.Requested, config.KubernetesClient)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			step := NewInstallExtraComponentsStep(stepConfig, config.Silent)

			assert.Equal(t, tt.want.Name, step.Name)
			if diff := cmp.Diff(tt.want.Input, step.Input); diff != "" {
				t.Fatalf("different step expected:\n%s", diff)
			}
		})
	}
}

func TestInstallExtraComponents(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		stepInput []StepInput
	}{
		{
			name: "test skip installing nothing in silent mode",
			config: Config{
				Silent: true,
			},
		},
		{
			name: "test skip installing controllers with defaults (\"\")",
			stepInput: []StepInput{
				{
					Name:  inComponentsExtra,
					Type:  multiSelectionChoice,
					Msg:   componentsExtraMsg,
					Value: "",
				},
			},
		},
		{
			name: "test skip installing controllers as default from silent mode",
			config: Config{
				Silent: true,
			},
		},
		{
			name:   "test install controllers with policy agent ",
			config: Config{},
			stepInput: []StepInput{
				{
					Name:  inComponentsExtra,
					Value: "policy-agent",
				},
			},
		},
		{
			name:   "test install controllers with policy agent and terraform",
			config: Config{},
			stepInput: []StepInput{
				{
					Name:  inComponentsExtra,
					Value: policyAgentController,
				},
				{
					Name:  inComponentsExtra,
					Value: tfController,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wgeObject, err := createWGEHelmReleaseFakeObject("1.0.0")
			if err != nil {
				t.Fatalf("error create wge object: %v", err)
			}
			config := makeTestConfig(t, tt.config, &wgeObject)
			_, err = installExtraComponents(tt.stepInput, &config)
			assert.NoError(t, err, "unexpected error")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}

}
