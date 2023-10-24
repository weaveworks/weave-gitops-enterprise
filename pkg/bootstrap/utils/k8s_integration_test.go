//go:build integration

package utils

import (
	"fmt"
	"os"
	"testing"

	"github.com/alecthomas/assert"
	testutils "github.com/weaveworks/weave-gitops-enterprise/test/utils"
)

// TestGetKubernetesClient test TestGetKubernetesClient
func TestGetKubernetesClientIt(t *testing.T) {

	tests := []struct {
		name        string
		setup       func() (string, error)
		reset       func()
		shouldError bool
	}{
		{
			name: "should create kubernetes http without kubeconfig",
			setup: func() (string, error) {
				kp, err := testutils.CreateKubeconfigFileForRestConfig(*cfg)
				if err != nil {
					return "", fmt.Errorf("cannot create kubeconfig: %w", err)
				}
				os.Setenv("KUBECONFIG", kp)
				return "", nil
			},
			reset: func() {
				os.Unsetenv("KUBECONFIG")
			},
			shouldError: false,
		},
		{
			name: "should create kubernetes http kubeconfig",
			setup: func() (string, error) {
				kp, err := testutils.CreateKubeconfigFileForRestConfig(*cfg)
				if err != nil {
					return "", fmt.Errorf("cannot create kubeconfig: %w", err)
				}
				return kp, nil
			},
			reset:       func() {},
			shouldError: false,
		},
		{
			name: "should not create kubernetes http with invalid kubeconfig ",
			setup: func() (string, error) {
				return "idontexist.yaml", nil
			},
			reset:       func() {},
			shouldError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeconfigPath, err := tt.setup()
			assert.NoError(t, err, "error on setup")
			defer tt.reset()

			kubehttp, err := GetKubernetesHttp(kubeconfigPath)
			if tt.shouldError {
				assert.Error(t, err, "error getting Kubernetes client")
				return
			}
			assert.NoError(t, err, "should have Kubernetes client")
			assert.NotNil(t, kubehttp, "should have Kubernetes client")
		})
	}

}
