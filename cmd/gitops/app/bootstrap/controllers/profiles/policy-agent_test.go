package profiles

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestConstructPolicyAgentValues(t *testing.T) {
	tests := []struct {
		name            string
		enableAdmission bool
		enableMutate    bool
		enableAudit     bool
		failurePolicy   string
		expected        map[string]interface{}
	}{
		{
			name:            "all enabled",
			enableAdmission: true,
			enableMutate:    true,
			enableAudit:     true,
			failurePolicy:   "Fail",
			expected: map[string]interface{}{
				"enabled": true,
				"config": map[string]interface{}{
					"admission": map[string]interface{}{
						"enabled": true,
						"sinks": map[string]interface{}{
							"k8sEventsSink": map[string]interface{}{
								"enabled": true,
							},
						},
						"mutate": true,
					},
					"audit": map[string]interface{}{
						"enabled": true,
						"sinks": map[string]interface{}{
							"k8sEventsSink": map[string]interface{}{
								"enabled": true,
							},
						},
					},
				},
				"excludeNamespaces": []string{
					"kube-system",
				},
				"failurePolicy":  "Fail",
				"useCertManager": true,
			},
		},
		{
			name:            "all disabled",
			enableAdmission: false,
			enableMutate:    false,
			enableAudit:     false,
			failurePolicy:   "Ignore",
			expected: map[string]interface{}{
				"enabled": true,
				"config": map[string]interface{}{
					"admission": map[string]interface{}{
						"enabled": false,
						"sinks": map[string]interface{}{
							"k8sEventsSink": map[string]interface{}{
								"enabled": true,
							},
						},
						"mutate": false,
					},
					"audit": map[string]interface{}{
						"enabled": false,
						"sinks": map[string]interface{}{
							"k8sEventsSink": map[string]interface{}{
								"enabled": true,
							},
						},
					},
				},
				"excludeNamespaces": []string{
					"kube-system",
				},
				"failurePolicy":  "Ignore",
				"useCertManager": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := constructPolicyAgentValues(tt.enableAdmission, tt.enableMutate, tt.enableAudit, tt.failurePolicy)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
