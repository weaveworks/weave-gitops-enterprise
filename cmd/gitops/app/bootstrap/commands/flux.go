package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

const (
	fluxBootstrapMsg               = "Do you want to bootstrap flux with the generic way on your cluster"
	gitRepoUrlMsg                  = "Please enter your git repository url (example: ssh://git@github.com/my-org-name/my-repo-name)"
	gitBranchMsg                   = "Please enter your git repository branch (default: main)"
	gitRepoPathMsg                 = "Please enter your path for your cluster (default: clusters/my-cluster)"
	fluxInstallInfoMsg             = "Installing flux ..."
	fluxBootstrapInfoMsg           = "Bootstrapping flux ..."
	fluxInstallCheckMsg            = "Checking flux installation ..."
	fluxInstallValidationMsg       = "Checking flux installation is valid ..."
	fluxInstallConfirmMsg          = "flux is installed"
	fluxReconcileConfirmMsg        = "flux is installed and can reconcile successfully"
	fluxNotInstalledMsgFormat      = "%v\n\n✖️  Flux is not installed on your cluster. Continue in the next step to bootstrap flux with the generic method. \nIf you wish for more information or advanced scenarios please refer to flux docs https://fluxcd.io/flux/installation/.\n\n"
	fluxInstallationErrorMsgFormat = "✖️  An error occurred. Please refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster.\n%v\n"
	fluxDocsReferenceMsgFormat     = "Please refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster.\n%v"
	privateKeyPathPromptMsgFormat  = "Please enter your private key path (default: %s)"
)

const (
	defaultBranch      = "main"
	defaultGitRepoPath = "clusters/my-cluster"
)

// BootstrapFlux get flux values from user and bootstraps it using generic way
func BootstrapFlux() error {
	bootstrapFlux := utils.GetConfirmInput(fluxBootstrapMsg)

	if bootstrapFlux != "y" {
		os.Exit(1)
	}

	gitURL, err := utils.GetStringInput(gitRepoUrlMsg, "")
	if err != nil {
		return err
	}

	gitBranch, err := utils.GetStringInput(gitBranchMsg, defaultBranch)
	if err != nil {
		return err
	}

	gitPath, err := utils.GetStringInput(gitRepoPathMsg, defaultGitRepoPath)
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	defaultPrivateKeyPath := filepath.Join(home, ".ssh", "id_rsa")
	privateKeyPathMsg := fmt.Sprintf(privateKeyPathPromptMsgFormat, defaultPrivateKeyPath)
	privateKeyPath, err := utils.GetStringInput(privateKeyPathMsg, defaultPrivateKeyPath)
	if err != nil {
		return err
	}

	utils.Warning(fluxInstallInfoMsg)
	// TODO: create default repo structure

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "bootstrap", "git", "--url", gitURL, "--branch", gitBranch, "--path", gitPath, "--private-key-file", privateKeyPath, "-s")
	if err != nil {
		errMsg := fmt.Sprintf(fluxDocsReferenceMsgFormat, string(out))
		return fmt.Errorf("%s%s", err.Error(), errMsg)
	}

	utils.Info(fluxBootstrapInfoMsg)
	return nil
}

// CheckFluxIsInstalled checks for valid flux installation
func CheckFluxIsInstalled() error {
	utils.Warning(fluxInstallCheckMsg)

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "check")
	if err != nil {
		utils.Warning(fluxNotInstalledMsgFormat, string(out))
		return BootstrapFlux()
	}

	utils.Info(fluxInstallConfirmMsg)

	return nil
}

// CheckFluxIsInstalled checks if flux installation is valid and can reconcile
func CheckFluxReconcile() error {
	utils.Warning(fluxInstallValidationMsg)

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		errMsg := fmt.Sprintf(fluxInstallationErrorMsgFormat, string(out))
		return fmt.Errorf("%s%s", err.Error(), errMsg)
	}

	utils.Info(fluxReconcileConfirmMsg)

	return nil
}
