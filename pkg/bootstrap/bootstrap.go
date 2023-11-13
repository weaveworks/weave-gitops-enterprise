package bootstrap

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/steps"
)

// Bootstrap initiated by the command runs the WGE bootstrap workflow
func Bootstrap(config steps.Config) error {

	adminCredentials, err := steps.NewAskAdminCredsSecretStep(config.ClusterUserAuth)
	if err != nil {
		return fmt.Errorf("cannot create ask admin creds step: %v", err)
	}

	// TODO have a single workflow source of truth and documented in https://docs.gitops.weave.works/docs/0.33.0/enterprise/getting-started/install-enterprise/
	var steps = []steps.BootstrapStep{
		//steps.VerifyFluxInstallation,
		//steps.NewAskBootstrapFluxStep(config),
		//steps.NewGitRepositoryConfig(config),
		//steps.NewBootstrapFlux(config),
		//steps.CheckEntitlementSecret,
		//steps.NewSelectWgeVersionStep(config),
		adminCredentials,
		//steps.NewSelectDomainType(config),
		//steps.NewInstallWGEStep(config),
		//steps.NewInstallOIDCStep(config),
		//steps.NewOIDCConfigStep(config),
		//steps.CheckUIDomainStep,
	}

	for _, step := range steps {
		config.Logger.Waitingf(step.Name)
		err := step.Execute(&config)
		if err != nil {
			return err
		}
	}
	return nil
}
