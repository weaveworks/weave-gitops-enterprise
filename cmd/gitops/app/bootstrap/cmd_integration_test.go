//go:build integration
// +build integration

package bootstrap_test

import (
	"context"
	"testing"
	"time"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/root"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/adapters"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/gomega"
)

const (
	defaultTimeout  = time.Second * 5
	defaultInterval = time.Second
)

var validEntitlement = `eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJsaWNlbmNlZFVudGlsIjoxNzg5MzgxMDE1LCJpYXQiOjE2MzE2MTQ2MTUsImlzcyI6InNhbGVzQHdlYXZlLndvcmtzIiwibmJmIjoxNjMxNjE0NjE1LCJzdWIiOiJ0ZWFtLXBlc3RvQHdlYXZlLndvcmtzIn0.klRpQQgbCtshC3PuuD4DdI3i-7Z0uSGQot23YpsETphFq4i3KK4NmgfnDg_WA3Pik-C2cJgG8WWYkWnemWQJAw`

var fluxSystemNamespace = corev1.Namespace{
	TypeMeta: metav1.TypeMeta{
		Kind:       "Namespace",
		APIVersion: "v1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: "flux-system",
	},
}

var entitlementSecret = corev1.Secret{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "weave-gitops-credentials",
		Namespace: "flux-system",
	},
	Type: corev1.SecretTypeOpaque,
	Data: map[string][]byte{"entitlement": []byte(validEntitlement)},
}

// TestBootstrapCmd is an integration test for bootstrapping command.
// It uses envtest to simulate a cluster.
func TestBootstrapCmd(t *testing.T) {
	g := NewGomegaWithT(t)
	g.SetDefaultEventuallyTimeout(defaultTimeout)
	g.SetDefaultEventuallyPollingInterval(defaultInterval)

	tests := []struct {
		name             string
		flags            []string
		expectedErrorStr string
		objects          []client.Object
	}{
		{
			name:             "should fail without entitlements",
			flags:            []string{},
			expectedErrorStr: "entitlement file is not found",
		},
		{
			name:  "should not fail with entitlements",
			flags: []string{},
			objects: []client.Object{
				&fluxSystemNamespace,
				&entitlementSecret,
			},
			expectedErrorStr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			createResources(t, k8sClient, tt.objects...)

			client := adapters.NewHTTPClient()
			cmd := root.Command(client)
			bootstrapCmdArgs := []string{"bootstrap"}
			bootstrapCmdArgs = append(bootstrapCmdArgs, tt.flags...)
			cmd.SetArgs(bootstrapCmdArgs)

			err := cmd.Execute()
			if tt.expectedErrorStr != "" {
				g.Expect(err.Error()).To(ContainSubstring(tt.expectedErrorStr))
				return
			}
			g.Expect(err).To(BeNil())

		})
	}
}

func createResources(t *testing.T, k client.Client, state ...client.Object) {
	ctx := context.Background()
	t.Helper()
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
