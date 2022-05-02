package tfcontroller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateBadName(t *testing.T) {
	data := `
kind: foo
metadata:
  name: h i
`
	rendered := [][]byte{[]byte(data)}
	err := ValidateRenderedTemplates(rendered)
	assert.Contains(t, err.Error(), "invalid value")
}

func TestValidateGoodName(t *testing.T) {
	data := `{ "kind": "derp", "metadata": { "name": "hi" }}`
	rendered := [][]byte{[]byte(data)}
	err := ValidateRenderedTemplates(rendered)
	assert.NoError(t, err)
}
