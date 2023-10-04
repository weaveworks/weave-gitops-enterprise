package commands

import (
	"errors"

	"github.com/weaveworks/weave-gitops/pkg/runner"
)

// user messages
const (
	fluxBoostrapCheckMsg           = "Checking flux is bootstrapped"
	fluxExistingBootstrapMsg       = "Flux is already bootstrapped"
	fluxInstallationErrorMsgFormat = "an error occurred. Please refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster"
)

// CheckEntitlementSecretStep checks for valid entitlement secret.
var VerifyFluxInstallationStep = BootstrapStep{
	Name: "verify flux setup",
	Step: verifyFluxInstallation,
}

// VerifyFluxInstallation checks for valid flux installation.
func verifyFluxInstallation(input []StepInput, c *Config) ([]StepOutput, error) {
	c.Logger.Waitingf(fluxBoostrapCheckMsg)

	var runner runner.CLIRunner
	_, err := runner.Run("flux", "check")
	if err != nil {
		return []StepOutput{}, errors.New(fluxInstallationErrorMsgFormat)
	}

	_, err = runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		return []StepOutput{}, errors.New(fluxInstallationErrorMsgFormat)
	}

	return []StepOutput{
		{
			Name:  "flux success msg",
			Type:  successMsg,
			Value: fluxExistingBootstrapMsg,
		},
	}, nil
}
