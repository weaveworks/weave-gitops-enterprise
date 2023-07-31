package connector

import (
	"context"
	"testing"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// ReconcileServiceAccount accepts a client and the name for a service account.
//
// It creates the ServiceAccount, and if the service account exists, this is not
// an error.
func ReconcileServiceAccount(ctx context.Context, client corev1.CoreV1Interface, serviceAccountName string) error {
}

// Test with non-existing SA
// Test with existing SA
// Look for the fake client in client-go
func TestReconcileServiceAccount(t *testing.T) {
	// Call ReconcileServiceAccount with a fake client and service name
	// If it doesn't fail, load the ServiceAccount with the name, it should exist
}
