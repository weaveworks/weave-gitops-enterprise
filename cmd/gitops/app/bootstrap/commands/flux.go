package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

func BootstrapFlux() error {

	prompt := promptui.Prompt{
		Label:     "Do you want to bootstrap flux with the generic way on your cluster",
		IsConfirm: true,
	}

	result, err := prompt.Run()
	if err != nil {
		return utils.CheckIfError(err)
	}

	if result != "y" {
		os.Exit(1)
	}

	gitURLPrompt := utils.PromptContent{
		ErrorMsg:     "Host can't be empty",
		Label:        "Please enter your git repository url (example: ssh://git@github.com/my-org-name/my-repo-name)",
		DefaultValue: "",
	}
	gitURL, err := utils.GetPromptStringInput(gitURLPrompt)
	if err != nil {
		return utils.CheckIfError(err)
	}

	gitBranchPrompt := utils.PromptContent{
		ErrorMsg:     "Branch can't be empty",
		Label:        "Please enter your git repository branch (default: main)",
		DefaultValue: "main",
	}
	gitBranch, err := utils.GetPromptStringInput(gitBranchPrompt)
	if err != nil {
		return utils.CheckIfError(err)
	}

	gitPathPrompt := utils.PromptContent{
		ErrorMsg:     "Path can't be empty",
		Label:        "Please enter your path for your cluster (default: clusters/my-cluster)",
		DefaultValue: "clusters/my-cluster",
	}
	gitPath, err := utils.GetPromptStringInput(gitPathPrompt)
	if err != nil {
		return utils.CheckIfError(err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return utils.CheckIfError(err)
	}

	defaultPrivateKeyPath := filepath.Join(home, ".ssh", "id_rsa")
	privateKeyPathPrompt := utils.PromptContent{
		ErrorMsg:     "Private key path can't be empty",
		Label:        fmt.Sprintf("Please enter your private key path (default: %s)", defaultPrivateKeyPath),
		DefaultValue: defaultPrivateKeyPath,
	}
	privateKeyPath, err := utils.GetPromptStringInput(privateKeyPathPrompt)
	if err != nil {
		return utils.CheckIfError(err)
	}
	fmt.Println("Installing flux ...")
	var runner runner.CLIRunner
	out, err := runner.Run("flux", "bootstrap", "git", "--url", gitURL, "--branch", gitBranch, "--path", gitPath, "--private-key-file", privateKeyPath, "-s")
	if err != nil {
		errMsg := fmt.Sprintf("Please refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster.\n%v", string(out))
		return utils.CheckIfError(err, errMsg)
	}
	fmt.Println("✔  flux is bootstrapped successfully")
	return nil
}

func CheckFluxIsInstalled() error {
	fmt.Println("Checking flux is installed ...")

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "check")
	if err != nil {
		fmt.Printf("%v\n\n✖️  Flux is not installed on your cluster. Continue in the next step to bootstrap flux with the generic method. \nIf you wish for more information or advanced scenarios please refer to flux docs https://fluxcd.io/flux/installation/.\n\n", string(out))
		return BootstrapFlux()
	}
	fmt.Println("✔  flux is installed")
	return nil

}

func CheckFluxReconcile() error {
	fmt.Println("Checking flux installation is valid ...")

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		errMsg := fmt.Sprintf("✖️  An error occurred. Please refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster.\n%v\n", string(out))
		return utils.CheckIfError(err, errMsg)

	}
	fmt.Println("✔  flux is installed properly and can reconcile successfully")
	return nil
}
