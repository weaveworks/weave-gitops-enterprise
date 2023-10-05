package bootstrap

import (
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/steps"
)

// Bootstrap initiated by the command runs the WGE bootstrap steps
func Bootstrap(config steps.Config) error {
	var steps = []steps.BootstrapStep{
		steps.NewCheckEntitlementSecret(),
		steps.VerifyFluxInstallationStep,
		steps.SelectWgeVersionStep,
		steps.AskAdminCredsSecretStep,
		steps.SelectDomainType,
		steps.InstallWGEStep,
		steps.CheckUIDomainStep,
	}

	for _, step := range steps {
		config.Logger.Waitingf(step.Name)
		err := step.Execute(&config, nil)
		if err != nil {
			return err
		}
	}
	return nil
}
