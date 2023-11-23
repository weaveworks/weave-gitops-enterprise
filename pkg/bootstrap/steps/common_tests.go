package steps

import (
	"testing"

	"github.com/weaveworks/weave-gitops-enterprise/test/utils"
	"k8s.io/apimachinery/pkg/runtime"
)

func makeTestConfig(t *testing.T, config Config, objects ...runtime.Object) Config {
	fakeClient := utils.CreateFakeClient(t, objects...)
	cliLogger := utils.CreateLogger()
	return Config{
		KubernetesClient:        fakeClient,
		Logger:                  cliLogger,
		WGEVersion:              config.WGEVersion,
		ClusterUserAuth:         config.ClusterUserAuth,
		DomainType:              config.DomainType,
		UserDomain:              config.UserDomain,
		GitScheme:               config.GitScheme,
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
