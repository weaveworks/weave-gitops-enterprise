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
	fluxRecoverMsg           = "flux is not bootstrapped in 'flux-system' namespace: more info https://fluxcd.io/flux/installation"
	fluxFatalErrorMsg        = "flux is not bootstrapped, please bootstrap Flux in 'flux-system' namespace: more info https://fluxcd.io/flux/installation"
)

// VerifyFluxInstallation checks that Flux is present in the cluster. It fails in case not and returns next steps to install it.
var VerifyFluxInstallation = BootstrapStep{
	Name: fluxBoostrapCheckMsg,
	Step: verifyFluxInstallation,
}

// VerifyFluxInstallation checks for valid flux installation.
func verifyFluxInstallation(input []StepInput, c *Config) ([]StepOutput, error) {
	var runner runner.CLIRunner

	c.Logger.Actionf("verifying flux installation")
	out, err := runner.Run("flux", "check")
	if err != nil {
		c.Logger.Failuref("flux installed error: %v. %s", string(out), fluxRecoverMsg)
		return []StepOutput{}, nil
	}
	c.Logger.Successf(fluxExistingInstallMsg)

	c.Logger.Actionf("verifying flux reconciliation")
	out, err = runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		return []StepOutput{}, fmt.Errorf("flux bootstrapped error: %v. %s", string(out), fluxFatalErrorMsg)
	}
	c.Logger.Successf(fluxExistingBootstrapMsg)

	repo, err := utils.GetGitRepositoryObject(c.KubernetesClient, WGEDefaultRepoName, WGEDefaultNamespace)
	if err != nil {
		return []StepOutput{}, fmt.Errorf("failed to get flux repository: %v", err)
	}
	_, scheme, err := normaliseUrl(repo.Spec.URL)
	if err != nil {
		return []StepOutput{}, fmt.Errorf("failed to parse flux repository: %v", err)
	}
	c.GitRepository.Scheme = scheme
	c.Logger.Successf("detected git scheme: %s", c.GitRepository.Scheme)

	c.FluxInstalled = true

	return []StepOutput{}, nil
}
