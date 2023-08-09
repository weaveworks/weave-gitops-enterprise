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

	fluxExists := promptGetBoolInput(fluxPromptContent)
	if !slices.Contains([]string{"Y", "y"}, fluxExists) {
		fmt.Println("\nPlease refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster")
		os.Exit(1)
	}

	var runner runner.CLIRunner

	_, err := runner.Run("flux", "check")
	if err != nil {
		fmt.Printf("An error occurred %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ flux is installed")
}

func CheckFluxReconcile() {
	var runner runner.CLIRunner

	_, err := runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		fmt.Printf("An error occurred %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ flux is installed properly and can reconcile successfully")
}
