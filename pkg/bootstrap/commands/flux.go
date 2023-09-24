package commands

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	fluxBoostrapCheckMsg     = "Checking flux is bootstrapped ..."
	fluxExistingBootstrapMsg = "Flux is already bootstrapped!"

	fluxInstallationErrorMsgFormat = "✖️  An error occurred. Please refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster.\n%v\n"
)

// VerifyFluxInstallation checks for valid flux installation.
func VerifyFluxInstallation(client k8s_client.Client) error {
	utils.Warning(fluxBoostrapCheckMsg)

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "check")
	if err != nil {
		errMsg := fmt.Sprintf(fluxInstallationErrorMsgFormat, string(out))
		return fmt.Errorf("%s: %w", errMsg, err)
	}

	out, err = runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		errMsg := fmt.Sprintf(fluxInstallationErrorMsgFormat, string(out))
		return fmt.Errorf("%s: %w", errMsg, err)
	}

	utils.Info(fluxExistingBootstrapMsg)

	return nil
}
