package steps

import (
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/runner"
)

// user messages
const (
	fluxBoostrapCheckMsg     = "Checking Flux is bootstrapped"
	fluxExistingInstallMsg   = "Flux is installed"
	fluxExistingBootstrapMsg = "Flux is bootstrapped"
	fluxRecoverMsg           = "Please bootstrap Flux into your cluster. Refer to https://fluxcd.io/flux/installation/ for more info."
)

// VerifyFluxInstallation checks that Flux is present in the cluster. It fails in case not and returns next steps to install it.
var VerifyFluxInstallation = BootstrapStep{
	Name: fluxBoostrapCheckMsg,
	Step: verifyFluxInstallation,
}

// VerifyFluxInstallation checks for valid flux installation.
func verifyFluxInstallation(input []StepInput, c *Config) ([]StepOutput, error) {
	var runner runner.CLIRunner

	out, err := runner.Run("flux", "check")
	if err != nil {
		return []StepOutput{}, fmt.Errorf("flux installed error: %v. %s", string(out), fluxRecoverMsg)
	}
	c.Logger.Successf(fluxExistingInstallMsg)

	out, err = runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		return []StepOutput{}, fmt.Errorf("flux bootstrapped error: %v. %s", string(out), fluxRecoverMsg)
	}
	c.Logger.Successf(fluxExistingBootstrapMsg)

	return []StepOutput{}, nil
}
