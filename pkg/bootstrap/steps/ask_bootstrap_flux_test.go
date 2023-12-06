package steps

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestAskBootstrapFlux(t *testing.T) {
	tests := []struct {
		name   string
		input  []StepInput
		config *Config
		err    bool
		canAsk bool
	}{
		{
			name:  "check with flux installed",
			input: []StepInput{},
			config: &Config{
				FluxInstallated: true,
			},
			err:    false,
			canAsk: false,
		},
		{
			name: "check with flux not installed and user selected no",
			input: []StepInput{
				{
					Name:  inBootstrapFlux,
					Value: confirmNo,
				},
			},
			config: &Config{
				FluxInstallated: false,
			},
			err:    true,
			canAsk: true,
		},
		{
			name: "check with flux not installed and user selected yes",
			input: []StepInput{
				{
					Name:  inBootstrapFlux,
					Value: "y",
				},
			},
			config: &Config{
				FluxInstallated: false,
			},
			err:    false,
			canAsk: true,
		},
		{
			name:  "check with silent mode and bootstrap flux flag available",
			input: []StepInput{},
			config: &Config{
				FluxInstallated: false,
				BootstrapFlux:   true,
				Silent:          true,
			},
			err:    false,
			canAsk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := makeTestConfig(t, *tt.config)

			_, err := askBootstrapFlux(tt.input, &config)
			if err != nil {
				if tt.err {
					return
				}
				t.Fatalf("unexpected error occurred: %v", err)
			}
			ask := canAskForFluxBootstrap(tt.input, tt.config)
			assert.Equal(t, tt.canAsk, ask, "mismatch result")

		})
	}
}
