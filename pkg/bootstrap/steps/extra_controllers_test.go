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
						Name: inExtraComponents,
						Type: multiSelectionChoice,
						Msg:  extraComponentsMsg,
						Values: []string{
							defaultController,
							policyAgentController,
							tfController,
							capiController,
							allControllers,
						},
						DefaultValue: defaultController,
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
	// note: can't test controllers as it requires pushing to git
	// will test the functionality with the default controller only here
	tests := []struct {
		name      string
		stepInput []StepInput
		err       bool
	}{
		{
			name: "test skip installing controllers with defaults (none)",
			stepInput: []StepInput{
				{
					Name:  inExtraComponents,
					Type:  multiSelectionChoice,
					Msg:   extraComponentsMsg,
					Value: defaultController,
				},
			},
			err: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := makeTestConfig(t, Config{})
			_, err := installExtraComponents(tt.stepInput, &config)
			assert.NoError(t, err, "unexpected error")
			if err != nil {
				if tt.err {
					return
				}
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}

}
