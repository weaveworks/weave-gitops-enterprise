package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

const (
	FLUX_BOOTSTRAP_MSG    = "Do you want to bootstrap flux with the generic way on your cluster"
	GIT_REPO_URL_MSG      = "Please enter your git repository url (example: ssh://git@github.com/my-org-name/my-repo-name)"
	GIT_BRANCH_MSG        = "Please enter your git repository branch (default: main)"
	DEFAULT_BRANCH        = "main"
	GIT_REPO_PATH_MSG     = "Please enter your path for your cluster (default: clusters/my-cluster)"
	DEFAULT_GIT_REPO_PATH = "clusters/my-cluster"
)

// BootstrapFlux get flux values from user and bootstraps it using generic way
func BootstrapFlux() error {
	bootstrapFlux, err := utils.GetConfirmInput(FLUX_BOOTSTRAP_MSG)
	if err != nil {
		return err
	}

	if bootstrapFlux != "y" {
		os.Exit(1)
	}

	gitURL, err := utils.GetStringInput(GIT_REPO_URL_MSG, "")
	if err != nil {
		return err
	}

	gitBranch, err := utils.GetStringInput(GIT_BRANCH_MSG, DEFAULT_BRANCH)
	if err != nil {
		return err
	}

	gitPath, err := utils.GetStringInput(GIT_REPO_PATH_MSG, DEFAULT_GIT_REPO_PATH)
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

	utils.Warning("Installing flux ...")

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "bootstrap", "git", "--url", gitURL, "--branch", gitBranch, "--path", gitPath, "--private-key-file", privateKeyPath, "-s")
	if err != nil {
		errMsg := fmt.Sprintf("Please refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster.\n%v", string(out))
		return fmt.Errorf("%s%s", err.Error(), errMsg)
	}

	utils.Info("flux is bootstrapped successfully")
	return nil
}

// CheckFluxIsInstalled checks for valid flux installation
func CheckFluxIsInstalled() error {
	utils.Warning("Checking flux is installed ...")

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "check")
	if err != nil {
		utils.Warning("%v\n\n✖️  Flux is not installed on your cluster. Continue in the next step to bootstrap flux with the generic method. \nIf you wish for more information or advanced scenarios please refer to flux docs https://fluxcd.io/flux/installation/.\n\n", string(out))
		return BootstrapFlux()
	}

	utils.Info("flux is installed")

	return nil
}

// CheckFluxIsInstalled checks if flux installation is valid and can reconcile
func CheckFluxReconcile() error {
	utils.Warning("Checking flux installation is valid ...")

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		errMsg := fmt.Sprintf("✖️  An error occurred. Please refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster.\n%v\n", string(out))
		return fmt.Errorf("%s%s", err.Error(), errMsg)
	}

	utils.Info("flux is installed properly and can reconcile successfully")

	return nil
}
