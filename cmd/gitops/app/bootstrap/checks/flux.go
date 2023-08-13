package checks

import (
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/runner"
)

func CheckFluxIsInstalled() {
	fmt.Println("Checking flux is installed ...")

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "check")
	if err != nil {
		fmt.Printf("✖️  An error occurred. Please refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster\nerror: %v\n", string(out))
		os.Exit(1)
	}

	fmt.Println("✔  flux is installed")
}

func CheckFluxReconcile() {
	fmt.Println("Checking flux installation is valid ...")

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		fmt.Printf("✖️  An error occurred. Please refer to flux docs https://fluxcd.io/flux/installation/ to install and bootstrap flux on your cluster\nerror: %v\n", string(out))
		os.Exit(1)
	}

	fmt.Println("✔  flux is installed properly and can reconcile successfully")
}
