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
						Name:         inExtraComponents,
						Type:         multiSelectionChoice,
						Msg:          extraComponentsMsg,
						Values:       ExtraComponents,
						DefaultValue: "",
					},
				},
				Step: installExtraComponents,
			},
		},
		{
			name: "return bootstrap step with empty inputs in case not provided by user",
			config: Config{
				ExtraComponents: []string{
					policyAgentController,
					tfController,
					capiController,
				},
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
			step := NewInstallExtraComponents(config)

			assert.Equal(t, tt.want.Name, step.Name)
			if diff := cmp.Diff(tt.want.Input, step.Input); diff != "" {
				t.Fatalf("different step expected:\n%s", diff)
			}
		})
	}
}

func TestInstallExtraControllers(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		stepInput []StepInput
	}{
		{
			name: "test skip installing controllers with defaults (\"\")",
			stepInput: []StepInput{
				{
					Name:  inExtraComponents,
					Type:  multiSelectionChoice,
					Msg:   extraComponentsMsg,
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
					Name:  inExtraComponents,
					Value: "policy-agent",
				},
			},
		},
		{
			name:   "test install controllers with policy agent and capi",
			config: Config{},
			stepInput: []StepInput{
				{
					Name:  inExtraComponents,
					Value: policyAgentController,
				},
				{
					Name:  inExtraComponents,
					Value: capiController,
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
