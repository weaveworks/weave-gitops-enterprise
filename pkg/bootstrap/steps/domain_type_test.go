package steps

import (
	"os"
	"testing"

	"github.com/weaveworks/weave-gitops/pkg/logger"
)

func TestSelectDomainType(t *testing.T) {
	tests := []struct {
		name  string
		input []StepInput
		err   bool
	}{
		{
			name: "domain type exist",
			input: []StepInput{
				{
					Name:  domainType,
					Value: "localhost",
				},
			},
			err: false,
		},
		{
			name: "domain type doesn't exist",
			input: []StepInput{
				{
					Name:  "anothervalue",
					Value: "localhost",
				},
			},
			err: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cliLogger := logger.NewCLILogger(os.Stdout)
			config := Config{
				Logger: cliLogger,
			}
			_, err := selectDomainType(tt.input, &config)
			if err != nil {
				if tt.err {
					return
				}
				t.Fatalf("error getting domain type: %v", err)
			}

		})
	}
}
