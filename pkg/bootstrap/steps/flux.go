package steps

import (
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

// user messages
const (
	fluxBoostrapCheckMsg     = "checking flux"
	fluxExistingInstallMsg   = "flux is installed"
	fluxExistingBootstrapMsg = "flux is bootstrapped"
	fluxRecoverMsg           = "please bootstrap Flux in 'flux-system' namespace: more info https://fluxcd.io/flux/installation"
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

	c.Logger.Actionf("verifying flux reconcillation")
	out, err = runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		c.Logger.Failuref("flux bootstrapped error: %v. %s", string(out), fluxRecoverMsg)
		return []StepOutput{}, nil
	}
	c.Logger.Successf(fluxExistingBootstrapMsg)

	repo, err := utils.GetGitRepositoryObject(c.KubernetesClient, WGEDefaultRepoName, WGEDefaultNamespace)
	if err != nil {
		return []StepOutput{}, fmt.Errorf("failed to get flux repository: %v", err)
	}

	if strings.Contains(repo.Spec.URL, sshAuthType) {
		c.GitAuthType = sshAuthType
	} else {
		c.GitAuthType = httpsAuthType
	}
	c.Logger.Successf("set git authentication method to: %s", c.GitAuthType)
	c.FluxInstallated = true

	return []StepOutput{}, nil
}
