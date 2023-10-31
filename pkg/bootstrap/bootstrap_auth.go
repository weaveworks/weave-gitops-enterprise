package bootstrap

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/steps"
)

// BootstrapAuth initiated by the command runs the WGE bootstrap auth steps
func BootstrapAuth(config steps.Config) error {
	// use bootstrapAuth function to bootstrap the authentication
	switch config.AuthType {
	case steps.AuthOIDC:
		err := bootstrapOIDC(config)
		if err != nil {
			return fmt.Errorf("cannot bootstrap auth: %v", err)
		}
	default:
		return fmt.Errorf("authentication type %s is not supported", config.AuthType)

	}
	return nil
}

func bootstrapOIDC(config steps.Config) error {
	var steps = []steps.BootstrapStep{
		steps.VerifyFluxInstallation,
		steps.CheckEntitlementSecret,
		steps.NewBootstrapFluxUsingSSH(config),
		steps.NewBootstrapFluxUsingHTTPS(config),
		steps.NewInstallOIDCStep(config),
		steps.NewOIDCConfigStep(config),
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
