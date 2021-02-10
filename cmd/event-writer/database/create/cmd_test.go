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
	}
	for _, test := range tests {
		err := runCommand(test.params)
		if test.params.dbURI == "" {
			assert.Error(t, err)
			assert.Equal(t, string(err.Error()), test.output)
		} else {
			assert.NoError(t, err)
		}
	}
}
