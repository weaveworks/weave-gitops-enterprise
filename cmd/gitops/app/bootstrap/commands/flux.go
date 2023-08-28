package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

const (
	FluxBootstrapMsg         = "Do you want to bootstrap flux with the generic way on your cluster"
	GitRepoUrlMsg            = "Please enter your git repository url (example: ssh://git@github.com/my-org-name/my-repo-name)"
	GitBranchMsg             = "Please enter your git repository branch (default: main)"
	GitRepoPathMsg           = "Please enter your path for your cluster (default: clusters/my-cluster)"
	FluxInstallInfoMsg       = "Installing flux ..."
	FluxBootstrapInfoMsg     = "Bootstrapping flux ..."
	FluxInstallCheckMsg      = "Checking flux installation ..."
	FluxInstallValidationMsg = "Checking flux installation is valid ..."
	FluxInstallConfirmMsg    = "flux is installed"
	FluxReconcileConfirmMsg  = "flux is installed and can reconcile successfully"
	FluxNotInstalledMsg      = "%v\n\n✖️  Flux is not installed on your cluster. Continue in the next step to bootstrap flux with the generic method. \nIf you wish for more information or advanced scenarios please refer to flux docs https://fluxcd.io/flux/installation/.\n\n"
	FluxInstallationErrorMsg = "✖️  An error occurred. Please refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster.\n%v\n"
	FluxDocsReferenceMsg     = "Please refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster.\n%v"
)

const (
	DefaultBranch      = "main"
	DefaultGitRepoPath = "clusters/my-cluster"
)

// BootstrapFlux get flux values from user and bootstraps it using generic way
func BootstrapFlux() error {
	bootstrapFlux := utils.GetConfirmInput(FluxBootstrapMsg)

	if bootstrapFlux != "y" {
		os.Exit(1)
	}

	gitURL, err := utils.GetStringInput(GitRepoUrlMsg, "")
	if err != nil {
		return err
	}

	gitBranch, err := utils.GetStringInput(GitBranchMsg, DefaultBranch)
	if err != nil {
		return err
	}

	gitPath, err := utils.GetStringInput(GitRepoPathMsg, DefaultGitRepoPath)
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	defaultPrivateKeyPath := filepath.Join(home, ".ssh", "id_rsa")
	privateKeyPathMsg := fmt.Sprintf("Please enter your private key path (default: %s)", defaultPrivateKeyPath)
	privateKeyPath, err := utils.GetStringInput(privateKeyPathMsg, defaultPrivateKeyPath)
	if err != nil {
		return err
	}

	utils.Warning(FluxInstallInfoMsg)
	// TODO: create default repo structure

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "bootstrap", "git", "--url", gitURL, "--branch", gitBranch, "--path", gitPath, "--private-key-file", privateKeyPath, "-s")
	if err != nil {
		errMsg := fmt.Sprintf(FluxDocsReferenceMsg, string(out))
		return fmt.Errorf("%s%s", err.Error(), errMsg)
	}

	utils.Info(FluxBootstrapInfoMsg)
	return nil
}

// CheckFluxIsInstalled checks for valid flux installation
func CheckFluxIsInstalled() error {
	utils.Warning(FluxInstallCheckMsg)

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "check")
	if err != nil {
		utils.Warning(FluxNotInstalledMsg, string(out))
		return BootstrapFlux()
	}

	utils.Info(FluxInstallConfirmMsg)

	return nil
}

// CheckFluxIsInstalled checks if flux installation is valid and can reconcile
func CheckFluxReconcile() error {
	utils.Warning(FluxInstallValidationMsg)

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		errMsg := fmt.Sprintf(FluxInstallationErrorMsg, string(out))
		return fmt.Errorf("%s%s", err.Error(), errMsg)
	}

	utils.Info(FluxReconcileConfirmMsg)

	return nil
}
