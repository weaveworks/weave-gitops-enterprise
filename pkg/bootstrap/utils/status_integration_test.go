//go:build integration

package utils

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	testutils "github.com/weaveworks/weave-gitops-enterprise/test/utils"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/object"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"
	k8s_config "sigs.k8s.io/controller-runtime/pkg/client/config"
)

func TestStatusCheckerIt(t *testing.T) {
	// Setup function to create kubeconfig and set the KUBECONFIG environment variable
	setup := func() error {
		kp, err := testutils.CreateKubeconfigFileForRestConfig(*cfg)
		if err != nil {
			return fmt.Errorf("cannot create kubeconfig: %w", err)
		}
		os.Setenv("KUBECONFIG", kp)
		return nil
	}

	// Reset function to clean up after the test
	reset := func() {
		os.Unsetenv("KUBECONFIG")
	}

	// Run the setup function and handle errors
	assert.NoError(t, setup(), "Error setting up kubeconfig")
	defer reset()

	// Create the Kubernetes client using the kubeconfig
	config, err := k8s_config.GetConfig()
	assert.NoError(t, err, "Error getting Kubernetes config")
	client, err := k8s_client.New(config, k8s_client.Options{})
	assert.NoError(t, err, "Error creating Kubernetes client")

	// Initialize logger
	logInstance := logger.NewCLILogger(os.Stdout)

	// Initialize the StatusChecker
	statusChecker, err := NewStatusChecker(client, 5*time.Second, 1*time.Minute, logInstance)
	assert.NoError(t, err, "Error creating StatusChecker")

	// Define test cases
	testCases := []struct {
		name        string
		identifiers []object.ObjMetadata
	}{
		{
			name: "Check specific Kubernetes resource",
			identifiers: []object.ObjMetadata{
				{
					Name:      "cluster-controller-manager",
					Namespace: "flux-system",
					GroupKind: schema.GroupKind{Group: "apps", Kind: "Deployment"},
				},
				// Add more resources as needed
			},
		},
		// ... other test cases ...
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the Assess function to check the status of resources
			err := statusChecker.Assess(tc.identifiers...)
			assert.NoError(t, err, "Error assessing resource status")
		})
	}
}
