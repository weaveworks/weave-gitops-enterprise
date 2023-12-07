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
				FluxInstalled: true,
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
				FluxInstalled: false,
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
				FluxInstalled: false,
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
			canAsk: true,
		},
		{
			name: "should error if not installed and export mode as not supported",
			input: []StepInput{
				{
					Name:  inBootstrapFlux,
					Value: "y",
				},
			},
			config: &Config{
				FluxInstalled: false,
				ModesConfig: ModesConfig{
					Export: true,
				},
			},
			err:    true,
			canAsk: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := MakeTestConfig(t, *tt.config)
			_, err := askBootstrapFlux(tt.input, &config)
			if tt.err {
				assert.Error(t, err, "error expected")
				return
			}
			assert.NoError(t, err, "error not expected")
			ask := canAskForFluxBootstrap(tt.input, tt.config)
			assert.Equal(t, tt.canAsk, ask, "mismatch result")

		})
	}
}
