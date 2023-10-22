package steps

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

// user messages
const (
	fluxBoostrapCheckMsg     = "checking flux"
	fluxExistingInstallMsg   = "flux is installed"
	fluxExistingBootstrapMsg = "flux is bootstrapped"
	fluxRecoverMsg           = "please bootstrap Flux: more info https://fluxcd.io/flux/installation"
)

// VerifyFluxInstallation checks that Flux is present in the cluster. It fails in case not and returns next steps to install it.
var VerifyFluxInstallation = BootstrapStep{
	Name: fluxBoostrapCheckMsg,
	Step: verifyFluxInstallation,
}

// VerifyFluxInstallation checks for valid flux installation.
func verifyFluxInstallation(input []StepInput, c *Config) ([]StepOutput, error) {
	var runner runner.CLIRunner

	c.Logger.Actionf("verifying flux installation in namespace: %s", c.Namespace)
	out, err := runner.Run("flux", "check", "-n", c.Namespace)
	if err != nil {
		return []StepOutput{}, fmt.Errorf("flux installed error: %v. %s", string(out), fluxRecoverMsg)
	}
	c.Logger.Successf(fluxExistingInstallMsg)

	c.Logger.Actionf("verifying flux reconcillation in namespace: %s", c.Namespace)
	err = utils.ReconcileFlux(c.Namespace)
	if err != nil {
		return []StepOutput{}, fmt.Errorf("flux bootstrapped error: %v. %s", err, fluxRecoverMsg)
	}

	c.Logger.Successf(fluxExistingBootstrapMsg)

	return []StepOutput{}, nil
}
