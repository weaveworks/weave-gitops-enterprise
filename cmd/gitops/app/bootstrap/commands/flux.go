package commands

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

const (
	fluxBoostrapCheckMsg     = "Checking flux is bootstrapped ..."
	fluxExistingBootstrapMsg = "Flux is already bootstrapped!"

	fluxInstallationErrorMsgFormat = "✖️  An error occurred. Please refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster.\n%v\n"
)

// VerifyFluxInstallation checks for valid flux installation.
func VerifyFluxInstallation(opts config.Options) error {
	utils.Warning(fluxBoostrapCheckMsg)

	var runner runner.CLIRunner
	_, err := runner.Run("flux", "check")
	if err != nil {
		return err
	}

	out, err := runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		errMsg := fmt.Sprintf(fluxInstallationErrorMsgFormat, string(out))
		return fmt.Errorf("%s: %w", errMsg, err)
	}

	utils.Info(fluxExistingBootstrapMsg)

	return nil
}
