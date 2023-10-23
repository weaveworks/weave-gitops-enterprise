package steps

import (
	"testing"
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
			config, err := makeTestConfig(t, Config{})
			if err != nil {
				t.Fatalf("error creating config: %v", err)
			}

			_, err = selectDomainType(tt.input, &config)
			if err != nil {
				if tt.err {
					return
				}
				t.Fatalf("error getting domain type: %v", err)
			}

		})
	}
}
