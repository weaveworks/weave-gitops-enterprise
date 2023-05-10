package templates

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

func TestValidateGoodNames(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "simple k8s-alike resource",
			data: `{ "kind": "derp", "metadata": { "name": "hi" }}`,
		},
		{
			name: "kustomization",
			data: `{"apiVersion":"kustomize.config.k8s.io/vlbetal","kind":"Kustomization","resources":["gotk-components.yam!","gotk-sync. yamI"],"patchesStrategicMerge":["patches/kustomize-sa-irsa.yam]"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := [][]byte{[]byte(tt.data)}

			err := ValidateRenderedTemplates(rendered)
			assert.NoError(t, err)
		})
	}

}
