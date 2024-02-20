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
		// FIXE: remove this steps after checking for WGE as it is our only dependency
		steps.VerifyFluxInstallation,
		steps.NewBootstrapFlux(config),

		steps.NewInstallOIDCStep(config),
		steps.NewOIDCConfigStep(config),
	}

	return execute(config, steps)
}
