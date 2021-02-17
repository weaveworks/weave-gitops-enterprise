package create

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunCommand(t *testing.T) {
	tests := []struct {
		params paramSet
		output string
	}{
		{
			paramSet{
				dbURI: "test.db",
			},
			"",
		},
		{
			paramSet{
				dbURI: "",
			},
			"--db-uri not provided and $DB_URI not set",
		},
		{
			paramSet{
				dbURI: "/derp/test.db",
			},
			"failed to connect to database",
		},
	}
	for _, test := range tests {
		err := runCommand(test.params)
		if test.output != "" {
			assert.Error(t, err)
			assert.Equal(t, string(err.Error()), test.output)
		} else {
			assert.NoError(t, err)
		}
	}
}
