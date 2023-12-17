package steps

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestNewCheckUIDomainStep(t *testing.T) {
	tests := []struct {
		name            string
		config          ModesConfig
		wantStep        BootstrapStep
		wantErrorString string
	}{
		{
			name:     "should be not required in export mode",
			config:   ModesConfig{Export: true},
			wantStep: checkUIDomainStepNotRequired,
		},
		{
			name:     "should be required by default",
			config:   ModesConfig{},
			wantStep: checkUIDomainStep,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStep, err := NewCheckUIDomainStep(tt.config)
			if tt.wantErrorString != "" {
				assert.EqualError(t, err, tt.wantErrorString)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStep.Name, gotStep.Name)
		})
	}
}

func TestCheckUIDomainStep_Execute(t *testing.T) {
	tests := []struct {
		name            string
		setup           func() (BootstrapStep, Config)
		config          Config
		wantOutput      []StepOutput
		wantErrorString string
	}{
		{
			name: "can execute in export mode",
			setup: func() (BootstrapStep, Config) {
				config := MakeTestConfig(t, Config{
					ModesConfig: ModesConfig{
						Export: true,
					},
				})
				step, err := NewCheckUIDomainStep(config.ModesConfig)
				assert.NoError(t, err)
				return step, config
			},
			wantOutput: []StepOutput{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step, config := tt.setup()
			gotOutputs, err := step.Execute(&config)
			if tt.wantErrorString != "" {
				assert.EqualError(t, err, tt.wantErrorString)
				return
			}
			assert.NoError(t, err)
			if diff := cmp.Diff(tt.wantOutput, gotOutputs, cmpopts.IgnoreFields(v1.Secret{}, "Data")); diff != "" {
				t.Fatalf("expected output:\n%s", diff)
			}
		})
	}
}
