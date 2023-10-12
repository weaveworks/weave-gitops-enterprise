//go:build integration
// +build integration

package bootstrap_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/root"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/runner"
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

	lock := sync.Mutex{}

	privateKeyFile := os.Getenv("GIT_PRIVATEKEY_PATH")
	g.Expect(privateKeyFile).NotTo(BeEmpty())
	privateKeyArg := fmt.Sprintf("--private-key=%s", privateKeyFile)

	// ensure flux-system ns exists
	_ = k8sClient.Create(context.Background(), &fluxSystemNamespace)

	tests := []struct {
		name             string
		flags            []string
		expectedErrorStr string
		setup            func(t *testing.T)
		reset            func(t *testing.T)
	}{
		{
			name:             "should fail without flux bootstrapped",
			flags:            []string{},
			expectedErrorStr: "please bootstrap Flux in 'flux-system' namespace: more info https://fluxcd.io/flux/installation",
		},
		{
			name:  "should fail without entitlements",
			flags: []string{},
			setup: func(t *testing.T) {
				bootstrapFluxSsh(g)
			},
			reset: func(t *testing.T) {
				uninstallFlux(g)
			},

			expectedErrorStr: "entitlement file is not found",
		},
		{
			name:  "should fail without private key",
			flags: []string{},
			setup: func(t *testing.T) {
				bootstrapFluxSsh(g)
				createEntitlements(t, testLog)
			},
			reset: func(t *testing.T) {
				deleteEntitlements(t, testLog)
				uninstallFlux(g)
			},
			expectedErrorStr: "cannot process input 'private key path and password",
		},
		{
			name: "should fail without selected wge version",
			flags: []string{
				privateKeyArg,
				"--private-key-password=\"\"",
			},
			setup: func(t *testing.T) {
				bootstrapFluxSsh(g)
				createEntitlements(t, testLog)
			},
			reset: func(t *testing.T) {
				deleteEntitlements(t, testLog)
				uninstallFlux(g)
			},
			expectedErrorStr: "cannot process input 'select WGE version'",
		},
		{
			name: "should fail without user authentication",
			flags: []string{"--version=0.33.0",
				privateKeyArg,
				"--private-key-password=\"\"",
			},
			setup: func(t *testing.T) {
				bootstrapFluxSsh(g)
				createEntitlements(t, testLog)
			},
			reset: func(t *testing.T) {
				deleteEntitlements(t, testLog)
				uninstallFlux(g)
			},
			expectedErrorStr: "cannot process input 'user authentication'",
		},
		{
			name: "should fail without dashboard access",
			flags: []string{"--version=0.33.0",
				privateKeyArg,
				"--private-key-password=\"\"",
				"--username=admin",
				"--password=admin123"},
			setup: func(t *testing.T) {
				bootstrapFluxSsh(g)
				createEntitlements(t, testLog)
			},
			reset: func(t *testing.T) {
				deleteClusterUser(t, testLog)
				deleteEntitlements(t, testLog)
				uninstallFlux(g)
			},
			expectedErrorStr: "cannot process input 'dashboard access'",
		},
		{
			name: "should install with ssh repo",
			flags: []string{"--version=0.33.0",
				privateKeyArg,
				"--private-key-password=\"\"",
				"--username=admin",
				"--password=admin123",
				"--domain-type=localhost",
			},
			setup: func(t *testing.T) {
				bootstrapFluxSsh(g)
				createEntitlements(t, testLog)
			},
			reset: func(t *testing.T) {
				deleteEntitlements(t, testLog)
				deleteClusterUser(t, testLog)
				uninstallFlux(g)
			},
			expectedErrorStr: "",
		},
	}
	for _, tt := range tests {
		lock.Lock()
		t.Run(tt.name, func(t *testing.T) {

			defer lock.Unlock()

			if tt.setup != nil {
				tt.setup(t)
			}

			if tt.reset != nil {
				defer tt.reset(t)
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

func bootstrapFluxSsh(g *WithT) {
	var runner runner.CLIRunner

	repoUrl := os.Getenv("GIT_URL_SSH")
	g.Expect(repoUrl).NotTo(BeEmpty())
	fmt.Println(repoUrl)

	privateKeyFile := os.Getenv("GIT_PRIVATEKEY_PATH")
	g.Expect(privateKeyFile).NotTo(BeEmpty())
	fmt.Println(privateKeyFile)

	args := []string{"bootstrap", "git", "-s", fmt.Sprintf("--url=%s", repoUrl), fmt.Sprintf("--private-key-file=%s", privateKeyFile), "--path=clusters/management"}
	fmt.Println(args)

	s, err := runner.Run("flux", args...)
	fmt.Println(string(s))
	g.Expect(err).To(BeNil())

}

func uninstallFlux(g *WithT) {
	var runner runner.CLIRunner
	args := []string{"uninstall", "-s", "--keep-namespace"}
	_, err := runner.Run("flux", args...)
	g.Expect(err).To(BeNil())
}

func createEntitlements(t *testing.T, testLog logr.Logger) {
	secret := createEntitlementSecretFromEnv(t)
	objects := []client.Object{
		&secret,
	}
	createResources(testLog, t, k8sClient, objects...)
}

func deleteEntitlements(t *testing.T, testLog logr.Logger) {
	secret := createEntitlementSecretFromEnv(t)
	objects := []client.Object{
		&secret,
	}
	deleteResources(testLog, t, k8sClient, objects...)
}

func deleteClusterUser(t *testing.T, testLog logr.Logger) {
	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		Type: corev1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-user-auth",
			Namespace: "flux-system",
		},
		Data: map[string][]byte{},
	}

	objects := []client.Object{
		&secret,
	}
	deleteResources(testLog, t, k8sClient, objects...)
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
}

func deleteResources(log logr.Logger, t *testing.T, k client.Client, objects ...client.Object) {
	ctx := context.Background()
	t.Helper()
	for _, o := range objects {
		err := k.Delete(ctx, o)
		if err != nil {
			t.Logf("failed to cleanup object: %s", err)
		}
		log.Info("deleted object", "name", o.GetName(), "ns", o.GetNamespace(), "kind", o.GetObjectKind().GroupVersionKind().Kind)

	}
}
