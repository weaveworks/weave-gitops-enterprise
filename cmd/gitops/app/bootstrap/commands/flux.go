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

	fluxSetupValidationMsg  = "Verifying flux setup is valid ..."
	fluxReconcileConfirmMsg = "Flux is bootstrapped and can reconcile successfully!"

	fluxInstallationErrorMsgFormat = "✖️  An error occurred. Please refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster.\n%v\n"
)

// CheckFluxIsInstalled checks for valid flux installation.
func CheckFluxIsInstalled(opts config.Options) error {
	utils.Warning(fluxBoostrapCheckMsg)

	var runner runner.CLIRunner
	_, err := runner.Run("flux", "check")
	if err != nil {
		return err
	}

	utils.Info(fluxExistingBootstrapMsg)

	return nil
}

// CheckFluxIsInstalled checks if flux installation is valid and can reconcile.
func CheckFluxReconcile(opts config.Options) error {
	utils.Warning(fluxSetupValidationMsg)

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		errMsg := fmt.Sprintf(fluxInstallationErrorMsgFormat, string(out))
		return fmt.Errorf("%s%s", err.Error(), errMsg)
	}

	utils.Info(fluxReconcileConfirmMsg)

	return nil
}
