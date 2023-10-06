//go:build integration
// +build integration

package bootstrap_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/require"
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

var fluxSystemNamespace = corev1.Namespace{
	TypeMeta: metav1.TypeMeta{
		Kind:       "Namespace",
		APIVersion: "v1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: "flux-system",
	},
}

func createEntitlementSecretFromEnv(t *testing.T) corev1.Secret {

	username := os.Getenv("WGE_ENTITLEMENT_USERNAME")
	require.NotEmpty(t, username)
	password := os.Getenv("WGE_ENTITLEMENT_PASSWORD")
	require.NotEmpty(t, password)
	entitlement := os.Getenv("WGE_ENTITLEMENT_ENTITLEMENT")
	require.NotEmpty(t, entitlement)

	return corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weave-gitops-enterprise-credentials",
			Namespace: "flux-system",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"username":    []byte(username),
			"password":    []byte(password),
			"entitlement": []byte(entitlement),
		},
	}
}

// TestBootstrapCmd is an integration test for bootstrapping command.
// It uses envtest to simulate a cluster.
func TestBootstrapCmd(t *testing.T) {
	g := NewGomegaWithT(t)
	g.SetDefaultEventuallyTimeout(defaultTimeout)
	g.SetDefaultEventuallyPollingInterval(defaultInterval)
	testLog := testr.New(t)

	tests := []struct {
		name             string
		flags            []string
		expectedErrorStr string
		setup            func(t *testing.T)
	}{
		{
			name:             "should fail without entitlements",
			flags:            []string{},
			expectedErrorStr: "entitlement file is not found",
		},
		{
			name:  "should fail without flux bootstrapped",
			flags: []string{},
			setup: func(t *testing.T) {
				secret := createEntitlementSecretFromEnv(t)
				objects := []client.Object{
					&fluxSystemNamespace,
					&secret,
				}
				createResources(testLog, t, k8sClient, objects...)
			},
			expectedErrorStr: "Please bootstrap Flux into your cluster",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setup != nil {
				tt.setup(t)
			}

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

func createResources(log logr.Logger, t *testing.T, k client.Client, objects ...client.Object) {
	ctx := context.Background()
	t.Helper()
	for _, o := range objects {
		err := k.Create(ctx, o)
		if err != nil {
			t.Errorf("failed to create object: %s", err)
		}
		log.Info("created object", "name", o.GetName(), "ns", o.GetNamespace(), "kind", o.GetObjectKind().GroupVersionKind().Kind)
	}
	t.Cleanup(func() {
		for _, o := range objects {
			err := k.Delete(ctx, o)
			if err != nil {
				t.Logf("failed to cleanup object: %s", err)
			}
			log.Info("deleted object", "name", o.GetName(), "ns", o.GetNamespace(), "kind", o.GetObjectKind().GroupVersionKind().Kind)

		}
	})
}
