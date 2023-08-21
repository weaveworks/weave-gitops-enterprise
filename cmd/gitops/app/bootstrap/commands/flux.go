package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

func BootstrapFlux() {

	prompt := promptui.Prompt{
		Label:     "Do you want to bootstrap flux with the generic way on your cluster",
		IsConfirm: true,
	}

	result, _ := prompt.Run()

	if result == "y" {

		gitURLPrompt := promptContent{
			"Host can't be empty",
			"Please enter your git repository url (example: ssh://git@github.com/my-org-name/my-repo-name)",
			"",
		}
		gitURL := promptGetStringInput(gitURLPrompt)

		gitBranchPrompt := promptContent{
			"Branch can't be empty",
			"Please enter your git repository branch (default: main)",
			"main",
		}
		gitBranch := promptGetStringInput(gitBranchPrompt)

		gitPathPrompt := promptContent{
			"Path can't be empty",
			"Please enter your path for your cluster (default: clusters/my-cluster)",
			"clusters/my-cluster",
		}
		gitPath := promptGetStringInput(gitPathPrompt)

		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		defaultPrivateKeyPath := filepath.Join(home, ".ssh", "id_rsa")
		privateKeyPathPrompt := promptContent{
			"Private key path can't be empty",
			fmt.Sprintf("Please enter your private key path (default: %s)", defaultPrivateKeyPath),
			defaultPrivateKeyPath,
		}
		privateKeyPath := promptGetStringInput(privateKeyPathPrompt)
		fmt.Println("Installing flux ...")
		var runner runner.CLIRunner
		out, err := runner.Run("flux", "bootstrap", "git", "--url", gitURL, "--branch", gitBranch, "--path", gitPath, "--private-key-file", privateKeyPath, "-s")
		if err != nil {
			fmt.Printf("✖️  An error occurred. Please refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster.\n%v\n", string(out))
			os.Exit(1)
		}

		fmt.Println("✔  flux is bootstrapped successfully")
	} else {
		os.Exit(1)
	}

}

func CheckFluxIsInstalled() {
	fmt.Println("Checking flux is installed ...")

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "check")
	if err != nil {
		fmt.Printf("%v\n\n✖️  Flux is not installed on your cluster. Continue in the next step to bootstrap flux with the generic method. \nIf you wish for more information or advanced scenarios please refer to flux docs https://fluxcd.io/flux/installation/.\n\n", string(out))
		BootstrapFlux()
	} else {
		fmt.Println("✔  flux is installed")
	}

}

func CheckFluxReconcile() {
	fmt.Println("Checking flux installation is valid ...")

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		fmt.Printf("✖️  An error occurred. Please refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster.\n%v\n", string(out))
		os.Exit(1)
	}

	fmt.Println("✔  flux is installed properly and can reconcile successfully")
}
