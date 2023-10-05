package steps

import (
	"errors"

	"github.com/weaveworks/weave-gitops/pkg/runner"
)

// user messages
const (
	fluxBoostrapCheckMsg           = "Checking Flux is bootstrapped"
	fluxExistingBootstrapMsg       = "Flux is already bootstrapped"
	fluxInstallationErrorMsgFormat = "an error occurred. Please refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster"
)

// VerifyFluxInstallation checks that Flux is present in the cluster. It fails in case not and returns next steps to install it.
var VerifyFluxInstallation = BootstrapStep{
	Name: fluxBoostrapCheckMsg,
	Step: verifyFluxInstallation,
}

// VerifyFluxInstallation checks for valid flux installation.
func verifyFluxInstallation(input []StepInput, c *Config) ([]StepOutput, error) {

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
			Name:  "Flux success msg",
			Type:  successMsg,
			Value: fluxExistingBootstrapMsg,
		},
	}, nil
}
