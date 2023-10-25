//go:build acceptance

package bootstrap_test

import (
	"context"
	"fmt"
	"os"
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

func createEntitlementSecretFromEnv(t *testing.T, namespace string) corev1.Secret {

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
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"username":    []byte(username),
			"password":    []byte(password),
			"entitlement": []byte(entitlement),
		},
	}
}

type testConfig struct {
	KubeconfigFlag string
	NamespaceFlag  string
	Namespace      string
	Log            logr.Logger
}

// TestBootstrapCmd is an integration test for bootstrapping command.
// It uses envtest to simulate a cluster.
func TestBootstrapCmd(t *testing.T) {
	g := NewGomegaWithT(t)
	g.SetDefaultEventuallyTimeout(defaultTimeout)
	g.SetDefaultEventuallyPollingInterval(defaultInterval)
	testLog := testr.New(t)

	privateKeyFile := os.Getenv("GIT_PRIVATEKEY_PATH")
	g.Expect(privateKeyFile).NotTo(BeEmpty())
	privateKeyFlag := fmt.Sprintf("--private-key=%s", privateKeyFile)
	kubeconfigFlag := fmt.Sprintf("--kubeconfig=%s", kubeconfigPath)
	namespace := "acceptance"
	namespaceFlag := fmt.Sprintf("--namespace=%s", namespace)

	tc := testConfig{
		KubeconfigFlag: kubeconfigFlag,
		NamespaceFlag:  namespaceFlag,
		Namespace:      namespace,
		Log:            testLog,
	}

	var bootstrappingNamespace = corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	_ = k8sClient.Create(context.Background(), &bootstrappingNamespace)

	tests := []struct {
		name             string
		flags            []string
		expectedErrorStr string
		setup            func(t *testing.T)
		reset            func(t *testing.T)
	}{
		{
			name: "should install with ssh repo",
			flags: []string{kubeconfigFlag,
				namespaceFlag,
				"--version=0.33.0",
				privateKeyFlag, "--private-key-password=\"\"",
				"--username=admin",
				"--password=admin123",
				"--domain-type=localhost",
			},
			setup: func(t *testing.T) {
				bootstrapFluxSsh(g, tc)
				createEntitlements(t, tc)
			},
			reset: func(t *testing.T) {
				deleteEntitlements(t, tc)
				deleteClusterUser(t, tc)
				uninstallFlux(g, tc)
			},
			expectedErrorStr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setup != nil {
				tt.setup(t)
			}

			if tt.reset != nil {
				defer tt.reset(t)
			}

			cmd := root.Command(adapters.NewHTTPClient())
			bootstrapCmdArgs := []string{"bootstrap"}
			bootstrapCmdArgs = append(bootstrapCmdArgs, tt.flags...)
			cmd.SetArgs(bootstrapCmdArgs)
			fmt.Println(bootstrapCmdArgs)

			err := cmd.Execute()
			if tt.expectedErrorStr != "" {
				g.Expect(err.Error()).To(ContainSubstring(tt.expectedErrorStr))
				return
			}
			g.Expect(err).To(BeNil())

		})
	}
}

func bootstrapFluxSsh(g *WithT, tc testConfig) {
	var runner runner.CLIRunner
	kubeconfigFlag := tc.KubeconfigFlag
	namespaceFlag := tc.NamespaceFlag

	repoUrl := os.Getenv("GIT_URL_SSH")
	g.Expect(repoUrl).NotTo(BeEmpty())
	fmt.Println(repoUrl)

	privateKeyFile := os.Getenv("GIT_PRIVATEKEY_PATH")
	g.Expect(privateKeyFile).NotTo(BeEmpty())
	fmt.Println(privateKeyFile)

	args := []string{"bootstrap", "git", kubeconfigFlag, namespaceFlag, "-s", fmt.Sprintf("--url=%s", repoUrl), fmt.Sprintf("--private-key-file=%s", privateKeyFile), "--path=clusters/management"}
	fmt.Println(args)

	s, err := runner.Run("flux", args...)
	fmt.Println(string(s))
	g.Expect(err).To(BeNil())

}

func createKindCluster(name string) ([]byte, error) {
	var runner runner.CLIRunner
	args := []string{"create", "cluster", "-n", name}
	return runner.Run("kind", args...)
}

func deleteKindCluster(name string) ([]byte, error) {
	var runner runner.CLIRunner
	args := []string{"delete", "cluster", "-n", name}
	return runner.Run("kind", args...)
}

func uninstallFlux(g *WithT, tc testConfig) {
	kubeconfigFlag := tc.KubeconfigFlag
	namespaceFlag := tc.NamespaceFlag
	var runner runner.CLIRunner
	args := []string{"uninstall", kubeconfigFlag, namespaceFlag, "-s", "--keep-namespace"}
	_, err := runner.Run("flux", args...)
	g.Expect(err).To(BeNil())
}

func createEntitlements(t *testing.T, tc testConfig) {
	testLog := tc.Log
	secret := createEntitlementSecretFromEnv(t, tc.Namespace)
	objects := []client.Object{
		&secret,
	}
	createResources(testLog, t, k8sClient, objects...)
}

func deleteEntitlements(t *testing.T, tc testConfig) {
	testLog := tc.Log
	secret := createEntitlementSecretFromEnv(t, tc.Namespace)
	objects := []client.Object{
		&secret,
	}
	deleteResources(testLog, t, k8sClient, objects...)
}

func deleteClusterUser(t *testing.T, tc testConfig) {
	testLog := tc.Log
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
