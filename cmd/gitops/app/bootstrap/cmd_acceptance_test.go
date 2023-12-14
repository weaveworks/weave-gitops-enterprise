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
	"github.com/weaveworks/weave-gitops-enterprise/test/utils"
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

	// ssh git repo configuration
	privateKeyFile := os.Getenv("GIT_PRIVATEKEY_PATH")
	g.Expect(privateKeyFile).NotTo(BeEmpty())

	privateKeyPassword := os.Getenv("GIT_PRIVATEKEY_PASSWORD")
	g.Expect(privateKeyPassword).NotTo(BeEmpty())

	repoURLSSH := os.Getenv("GIT_REPO_URL_SSH")
	g.Expect(repoURLSSH).NotTo(BeEmpty())

	repoURLSSHNoScheme := os.Getenv("GIT_REPO_URL_SSH_NO_SCHEME")
	g.Expect(repoURLSSHNoScheme).NotTo(BeEmpty())

	privateKeyFlag := fmt.Sprintf("--private-key=%s", privateKeyFile)
	privateKeyPasswordFlag := fmt.Sprintf("--private-key-password=%s", privateKeyPassword)
	gitRepoUrlSshNoSchemeFlag := fmt.Sprintf("--repo-url=%s", repoURLSSHNoScheme)

	// https git repo configuration
	repoURLHTTPS := os.Getenv("GIT_REPO_URL_HTTPS")
	g.Expect(repoURLHTTPS).NotTo(BeEmpty())

	gitUsername := os.Getenv("GIT_USERNAME")
	g.Expect(gitUsername).NotTo(BeEmpty())

	gitPassword := os.Getenv("GIT_PASSWORD")
	g.Expect(gitPassword).NotTo(BeEmpty())

	// git repo configuration
	gitBranch := os.Getenv("GIT_BRANCH")
	g.Expect(gitBranch).NotTo(BeEmpty())

	gitRepoPath := os.Getenv("GIT_REPO_PATH")
	g.Expect(gitRepoPath).NotTo(BeEmpty())

	gitBranchFlag := fmt.Sprintf("--branch=%s", gitBranch)
	gitRepoPathFlag := fmt.Sprintf("--repo-path=%s", gitRepoPath)

	// wge version
	wgeVersion := os.Getenv("WGE_VERSION")
	g.Expect(wgeVersion).NotTo(BeEmpty())

	// oidc configuration
	oidcClientSecret := os.Getenv("OIDC_CLIENT_SECRET")
	g.Expect(oidcClientSecret).NotTo(BeEmpty())

	oidcClientSecretFlag := fmt.Sprintf("--client-secret=%s", oidcClientSecret)

	kubeconfigFlag := fmt.Sprintf("--kubeconfig=%s", kubeconfigPath)

	_ = k8sClient.Create(context.Background(), &fluxSystemNamespace)

	tests := []struct {
		name             string
		flags            []string
		userInputs       []string
		expectedErrorStr string
		setup            func(t *testing.T)
		reset            func(t *testing.T)
	}{
		{
			name: "journey flux exists: should bootstrap with valid arguments",
			flags: []string{kubeconfigFlag,
				"--version=0.37.0",
				privateKeyFlag, privateKeyPasswordFlag,
				"--password=admin123",
				"--discovery-url=https://dex-01.wge.dev.weave.works/.well-known/openid-configuration",
				"--client-id=weave-gitops-enterprise",
				oidcClientSecretFlag,
				"-s",
			},
			setup: func(t *testing.T) {
				bootstrapFluxSsh(g, kubeconfigFlag)
				createEntitlements(t, testLog)
			},
			reset: func(t *testing.T) {
				deleteEntitlements(t, testLog)
				deleteClusterUser(t, testLog)
				uninstallFlux(g, kubeconfigFlag)
			},
			expectedErrorStr: "",
		},
		{
			name: "journey flux does not exist: should bootstrap with valid arguments",
			flags: []string{kubeconfigFlag,
				"--version=0.37.0",
				"--password=admin123",
				"--discovery-url=https://dex-01.wge.dev.weave.works/.well-known/openid-configuration",
				"--client-id=weave-gitops-enterprise",
				gitRepoUrlSshNoSchemeFlag, gitBranchFlag, gitRepoPathFlag,
				privateKeyFlag, privateKeyPasswordFlag,
				oidcClientSecretFlag, "-s",
				"--components-extra=policy-agent,tf-controller",
				"--bootstrap-flux",
			},
			setup: func(t *testing.T) {
				createEntitlements(t, testLog)
			},
			reset: func(t *testing.T) {
				deleteEntitlements(t, testLog)
				deleteClusterUser(t, testLog)
				uninstallFlux(g, kubeconfigFlag)
			},
			expectedErrorStr: "",
		},
		{
			name: "journey flux does not exist interactive",
			flags: []string{kubeconfigFlag,
				"--version=0.37.0",
				"--password=admin123",
				"--components-extra=none",
			},
			userInputs: []string{
				"y\n",                             // do you want to bootstrap flux using the generic way?
				fmt.Sprintf("%s\n", repoURLHTTPS), // please enter your flux git https or ssh repository url
				fmt.Sprintf("%s\n", gitBranch),    // please enter your flux git repository branch
				fmt.Sprintf("%s\n", gitRepoPath),  // please enter your flux path for your cluster
				fmt.Sprintf("%s\n", gitUsername),  // please enter your git username
				fmt.Sprintf("%s\n", gitPassword),  // please enter your git password
				"N\n",                             // Do you want to setup OIDC to access Weave GitOps Dashboards?
			},
			setup: func(t *testing.T) {
				createEntitlements(t, testLog)
			},
			reset: func(t *testing.T) {
				deleteEntitlements(t, testLog)
				deleteClusterUser(t, testLog)
				uninstallFlux(g, kubeconfigFlag)
			},
			expectedErrorStr: "",
		},
		{
			name: "journey WGE version is not provided: should take the wge from user input",
			flags: []string{kubeconfigFlag,
				"--password=admin123",
				"--components-extra=none",
			},
			userInputs: []string{
				"y\n",                             // do you want to bootstrap flux using the generic way?
				fmt.Sprintf("%s\n", repoURLHTTPS), // please enter your flux git https or ssh repository url
				fmt.Sprintf("%s\n", gitBranch),    // please enter your flux git repository branch
				fmt.Sprintf("%s\n", gitRepoPath),  // please enter your flux path for your cluster
				fmt.Sprintf("%s\n", gitUsername),  // please enter your git username
				fmt.Sprintf("%s\n", gitPassword),  // please enter your git password
				fmt.Sprintf("%s\n", wgeVersion),   // please select your wge version
				"N\n",                             // Do you want to setup OIDC to access Weave GitOps Dashboards?
			},
			setup: func(t *testing.T) {
				createEntitlements(t, testLog)
			},
			reset: func(t *testing.T) {
				deleteEntitlements(t, testLog)
				deleteClusterUser(t, testLog)
				uninstallFlux(g, kubeconfigFlag)
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

			if len(tt.userInputs) > 0 {
				cmd.SetIn(&utils.MockReader{Inputs: tt.userInputs})
			}

			fmt.Println("bootstrap args: ", bootstrapCmdArgs)

			err := cmd.Execute()
			if tt.expectedErrorStr != "" {
				g.Expect(err.Error()).To(ContainSubstring(tt.expectedErrorStr))
				return
			}
			g.Expect(err).To(BeNil())

		})
	}
}

func bootstrapFluxSsh(g *WithT, kubeconfigFlag string) {
	var cliRunner runner.CLIRunner

	repoUrl := os.Getenv("GIT_REPO_URL_SSH")
	g.Expect(repoUrl).NotTo(BeEmpty())
	fmt.Println(repoUrl)

	privateKeyFile := os.Getenv("GIT_PRIVATEKEY_PATH")
	g.Expect(privateKeyFile).NotTo(BeEmpty())
	fmt.Println(privateKeyFile)

	privateKeyPassword := os.Getenv("GIT_PRIVATEKEY_PASSWORD")
	g.Expect(privateKeyFile).NotTo(BeEmpty())

	args := []string{"bootstrap", "git", kubeconfigFlag, "-s", fmt.Sprintf("--url=%s", repoUrl), fmt.Sprintf("--private-key-file=%s", privateKeyFile), fmt.Sprintf("--password=%s", privateKeyPassword), "--path=clusters/management"}
	fmt.Println(args)

	s, err := cliRunner.Run("flux", args...)
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

func uninstallFlux(g *WithT, kubeconfigFlag string) {
	var runner runner.CLIRunner
	args := []string{"uninstall", kubeconfigFlag, "-s", "--keep-namespace"}
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
