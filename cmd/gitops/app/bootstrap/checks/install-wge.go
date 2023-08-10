package checks

import (
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/runner"
)

const HELMREPOSITORY_NAME string = "weave-gitops-enterprise-charts"
const HELMRELEASE_NAME string = "weave-gitops-enterprise"
const UI_URL string = "https://localhost:8000"

func InstallWge(version string) {

	fmt.Printf("✔ All set installing WGE v%s, This may take few minutes...\n", version)
	var runner runner.CLIRunner

	_, err := runner.Run("flux", "create", "source", "helm", HELMREPOSITORY_NAME, "--url", CHART_URL, "--secret-ref", ENTITLEMENT_SECRET_NAME)
	if err != nil {
		fmt.Printf("An error occurred creating helmrepository %v\n", err)
		os.Exit(1)
	}

	_, err = runner.Run("flux", "create", "hr", HELMRELEASE_NAME,
		"--source", fmt.Sprintf("HelmRepository/%s", HELMREPOSITORY_NAME),
		"--chart", "mccp",
		"--chart-version", version,
		"--interval", "65m",
		"--crds", "CreateReplace",
	)

	if err != nil {
		fmt.Printf("An error occurred creating helmrelease %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✔ WGE v%s is installed successfully\n\n✅ You can visit the UI using portforward at %s\n", version, UI_URL)

	_, err = runner.Run("kubectl", "-n", "flux-system", "port-forward", "svc/clusters-service", "8000:8000")
	if err != nil {
		fmt.Printf("An error occurred port-forwarding %v\n", err)
		os.Exit(1)
	}

}
