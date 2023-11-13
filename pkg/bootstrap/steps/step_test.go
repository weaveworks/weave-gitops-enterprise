package steps

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_isUpdate(t *testing.T) {
	tests := []struct {
		name  string
		input StepInput
		in    string
		want  bool
	}{
		{
			name: "test input with no existing value",
			input: StepInput{
				Name:         "idontexist",
				AlreadyExist: false,
			},
			want: false,
		},
		{
			name: "test input with existing value and want to update",
			input: StepInput{
				Name:         "iexist",
				AlreadyExist: true,
			},
			in:   "y",
			want: true,
		},
		{
			name: "test input with existing value and dont want to update",
			input: StepInput{
				Name:         "iexist",
				AlreadyExist: true,
			},
			in:   "n",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to store the input string.
			var buf bytes.Buffer

			// Create a new io.ReaderCloser using the buffer.
			reader := io.NopCloser(&buf)

			// Write the input string to the buffer.
			buf.WriteString(fmt.Sprintf("%s\n", tt.in))

			update := isUpdate(tt.input, reader)
			assert.Equal(t, tt.want, update)
		})
	}
}
