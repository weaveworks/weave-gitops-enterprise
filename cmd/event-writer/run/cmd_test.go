package run

import (
	"testing"

	"github.com/tj/assert"
)

func TestRunCommand(t *testing.T) {
	tests := []struct {
		params paramSet
		output string
	}{
		{
			paramSet{
				natsURL:     "",
				natsSubject: "test.subject",
				dbURI:       "test.db",
			},
			"please specify the NATS server URL the event-writer should connect to",
		},
		{
			paramSet{
				natsURL:     "localhost:4222",
				natsSubject: "",
				dbURI:       "test.db",
			},
			"please specify the NATS subject the event-writer should subscribe to",
		},
		{
			paramSet{
				natsURL:     "localhost:4222",
				natsSubject: "test.subject",
				dbURI:       "",
			},
			"--db-uri not provided and $DB_URI not set",
		},
	}
	for _, test := range tests {
		err := runCommand(test.params)
		assert.Error(t, err)
		assert.Equal(t, string(err.Error()), test.output)
	}
}
