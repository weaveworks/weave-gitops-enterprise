package bootstrap

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/steps"
)

// Bootstrap initiated by the command runs the WGE bootstrap workflow
func Bootstrap(config steps.Config) error {

	adminCredentials, err := steps.NewAskAdminCredsSecretStep(config.ClusterUserAuth, config.Silent)
	if err != nil {
		return fmt.Errorf("cannot create ask admin creds step: %v", err)
	}

	repositoryConfig := steps.NewGitRepositoryConfigStep(config.GitRepository)

	// TODO have a single workflow source of truth and documented in https://docs.gitops.weave.works/docs/0.33.0/enterprise/getting-started/install-enterprise/
	var steps = []steps.BootstrapStep{
		steps.VerifyFluxInstallation,
		steps.NewAskBootstrapFluxStep(config),
		repositoryConfig,
		steps.NewBootstrapFlux(config),
		steps.CheckEntitlementSecret,
		steps.NewSelectWgeVersionStep(config),
		adminCredentials,
		steps.NewInstallWGEStep(),
		steps.NewInstallOIDCStep(config),
		steps.NewOIDCConfigStep(config),
		steps.CheckUIDomainStep,
	}

	for _, step := range steps {
		config.Logger.Waitingf(step.Name)
		_, err := step.Execute(&config)
		if err != nil {
			return err
		}
	}
	return nil
}
