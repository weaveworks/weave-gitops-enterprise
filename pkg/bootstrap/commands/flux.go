package commands

import (
	"errors"

	"github.com/weaveworks/weave-gitops/pkg/runner"
)

const (
	fluxBoostrapCheckMsg     = "Checking flux is bootstrapped"
	fluxExistingBootstrapMsg = "Flux is already bootstrapped"

	fluxInstallationErrorMsgFormat = "An error occurred. Please refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster"
)

// VerifyFluxInstallation checks for valid flux installation.
func (c *Bootstrapper) VerifyFluxInstallation() error {
	c.Logger.Waitingf(fluxBoostrapCheckMsg)

	var runner runner.CLIRunner
	_, err := runner.Run("flux", "check")
	if err != nil {
		return errors.New(fluxInstallationErrorMsgFormat)
	}

	_, err = runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		return errors.New(fluxInstallationErrorMsgFormat)
	}
	c.Logger.Successf(fluxExistingBootstrapMsg)
	return nil
}
