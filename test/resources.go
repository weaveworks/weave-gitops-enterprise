package test

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewNamespace creates a new namespace with a random name, prefixed with `kube-test`
func NewNamespace() *corev1.Namespace {
	ns := &corev1.Namespace{}
	ns.Name = "kube-test-" + rand.String(5)
	return ns
}

// Create uses a controller-runtime client to create a set of Kubernetes objects
func Create(ctx context.Context, t *testing.T, cfg *rest.Config, state ...client.Object) {
	t.Helper()
	k, err := client.New(cfg, client.Options{})
	if err != nil {
		t.Errorf("failed to create client: %s", err)
	}

	for _, o := range state {
		err := k.Create(ctx, o)
		if err != nil {
			t.Errorf("failed to create object: %s", err)
		}
	}
	t.Cleanup(func() {
		for _, o := range state {
			err := k.Delete(ctx, o)
			if err != nil {
				t.Logf("failed to cleanup object: %s", err)
			}
		}
	})
}
