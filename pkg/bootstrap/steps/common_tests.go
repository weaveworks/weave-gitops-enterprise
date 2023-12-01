package steps

import (
	"testing"

	"github.com/weaveworks/weave-gitops-enterprise/test/utils"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type fakeGitClient struct {
}

func (f fakeGitClient) CloneRepo(kubeClient client.Client, repoName string, namespace string, authType string, privateKeyPath string, privateKeyPassword string, username string, token string) (string, error) {
	return "/", nil
}

func (f fakeGitClient) CreateFileToRepo(filename, filecontent, path, commitmsg, authType, privateKeyPath, privateKeyPassword, username, token string) error {
	return nil
}

type fakeFluxClient struct {
}

func (f fakeFluxClient) ReconcileFlux() error {
	return nil
}

func (f fakeFluxClient) ReconcileHelmRelease(hrName string) error {
	return nil
}

func makeTestConfig(t *testing.T, config Config, objects ...runtime.Object) Config {
	fakeClient := utils.CreateFakeClient(t, objects...)

	cliLogger := utils.CreateLogger()

	return Config{
		KubernetesClient:        fakeClient,
		GitClient:               fakeGitClient{},
		FluxClient:              fakeFluxClient{},
		Logger:                  cliLogger,
		WGEVersion:              config.WGEVersion,
		ClusterUserAuth:         config.ClusterUserAuth,
		GitScheme:               config.GitScheme,
		GitRepository:           config.GitRepository,
		FluxInstallated:         config.FluxInstallated,
		PrivateKeyPath:          config.PrivateKeyPath,
		PrivateKeyPassword:      config.PrivateKeyPassword,
		GitUsername:             config.GitUsername,
		GitToken:                config.GitToken,
		RepoURL:                 config.RepoURL,
		Branch:                  config.Branch,
		RepoPath:                config.RepoPath,
		AuthType:                config.AuthType,
		InstallOIDC:             config.InstallOIDC,
		DiscoveryURL:            config.DiscoveryURL,
		IssuerURL:               config.IssuerURL,
		ClientID:                config.ClientID,
		ClientSecret:            config.ClientSecret,
		RedirectURL:             config.RedirectURL,
		PromptedForDiscoveryURL: config.PromptedForDiscoveryURL,
	}
}
