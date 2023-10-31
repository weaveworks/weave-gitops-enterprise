package steps

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestCanAskForHttpsGitConfig(t *testing.T) {

	tests := []struct {
		name   string
		config *Config
		output bool
	}{
		{
			name:   "test empty config",
			config: &Config{},
			output: false,
		},
		{
			name: "test valid config",
			config: &Config{
				FluxInstallated: false,
				GitAuthType:     httpsAuthType,
			},
			output: true,
		},
		{
			name: "test with flux installed and no type defined",
			config: &Config{
				FluxInstallated: true,
			},
			output: false,
		},
		{
			name: "test with type and no flux installed",
			config: &Config{
				GitAuthType: httpsAuthType,
			},
			output: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := canAskForHttpsGitConfig(nil, tt.config)
			assert.Equal(t, out, tt.output, "invalid output")
		})
	}
}
