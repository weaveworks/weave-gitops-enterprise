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
				URI:  "test.db",
				Type: "",
			},
			"--db-type not provided and $DB_TYPE not set",
		},
		{
			paramSet{
				URI:  "test.db",
				Type: "sqlite",
			},
			"",
		},
		{
			paramSet{
				URI:  "",
				Type: "sqlite",
			},
			"--db-uri not provided and $DB_URI not set",
		},
		{
			paramSet{
				URI:  "/derp/test.db",
				Type: "sqlite",
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
