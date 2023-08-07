package checks

import (
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/runner"
	"golang.org/x/exp/slices"
)

func CheckFluxIsInstalled() {

	fluxPromptContent := promptContent{
		"Please provide an answer with (y/n).",
		"Do you have a valid flux installation on your cluster (y/n)?",
	}

	fluxExists := promptGetInput(fluxPromptContent)
	if !slices.Contains([]string{"Y", "y"}, fluxExists) {
		fmt.Println("\nPlease install flux")
		os.Exit(1)
	}

	var runner runner.CLIRunner

	_, err := runner.Run("flux", "check")
	if err != nil {
		fmt.Printf("An error occured %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ flux is installed")
}

func CheckFluxReconcile() {

	var runner runner.CLIRunner

	_, err := runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		fmt.Printf("An error occured %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ flux is installed properly and can reconcile successfully")
}
