package steps

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestNewInstallExtraControllers(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   BootstrapStep
	}{
		{
			name:   "return bootstrap step with inputs in case provided by user",
			config: Config{},
			want: BootstrapStep{
				Name: "install extra controllers",
				Input: []StepInput{
					{
						Name: inExtraControllers,
						Type: multiSelectionChoice,
						Msg:  extraControllersMsg,
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
				Step: installExtraControllers,
			},
		},
		{
			name: "return bootstrap step with empty inputs in case not provided by user",
			config: Config{
				ExtraControllers: []string{
					policyAgentController,
					tfController,
					capiController,
				},
			},
			want: BootstrapStep{
				Name:  "install extra controllers",
				Input: []StepInput{},
				Step:  installExtraControllers,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := makeTestConfig(t, tt.config)
			step := NewInstallExtraControllers(config)

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
					Name:  inExtraControllers,
					Type:  multiSelectionChoice,
					Msg:   extraControllersMsg,
					Value: defaultController,
				},
			},
			err: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := makeTestConfig(t, Config{})
			_, err := installExtraControllers(tt.stepInput, &config)
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
