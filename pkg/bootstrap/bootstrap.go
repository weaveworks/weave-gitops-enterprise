package bootstrap

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/steps"
)

// Bootstrap initiated by the command runs the WGE bootstrap workflow

func Bootstrap(config steps.Config) error {
	// TODO have a single workflow source of truth and documented in https://docs.gitops.weave.works/docs/0.33.0/enterprise/getting-started/install-enterprise/
	fmt.Println(config)

	var steps = []steps.BootstrapStep{
		steps.VerifyFluxInstallation,
		steps.CheckEntitlementSecret,
		steps.NewAskPrivateKeyStep(config),
		steps.NewSelectWgeVersionStep(config),
		steps.NewAskAdminCredsSecretStep(config),
		steps.NewSelectDomainType(config),
		steps.NewInstallWGEStep(config),
		steps.CheckUIDomainStep,
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
