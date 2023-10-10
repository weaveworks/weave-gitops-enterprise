package bootstrap

import (
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/steps"
)

// Bootstrap initiated by the command runs the WGE bootstrap steps
func Bootstrap(config steps.Config) error {
	var steps = []steps.BootstrapStep{
		steps.CheckEntitlementSecret,
		steps.VerifyFluxInstallation,
		steps.NewAskPrivateKeyStep(config),
		steps.NewSelectWgeVersionStep(config),
		steps.NewAskAdminCredsSecretStep(config),
		steps.NewSelectDomainType(config),
		steps.NewInstallWGEStep(config),
		steps.OIDCConfigStep(config),
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
