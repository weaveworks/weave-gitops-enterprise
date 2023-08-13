package checks

import (
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/runner"
)

const HELMREPOSITORY_NAME string = "weave-gitops-enterprise-charts"
const HELMRELEASE_NAME string = "weave-gitops-enterprise"
const VALUES_FILES_LOCATION string = "/tmp/mccp-values.yaml"

// const UI_URL string = "https://localhost:8000"

func InstallWge(version string) {

	userDomainPrompt := promptContent{
		"Domain can't be empty",
		"Please enter your cluster domain: ",
	}
	userDomain := promptGetStringInput(userDomainPrompt)

	fmt.Printf("✔ All set installing WGE v%s, This may take few minutes...\n", version)
	var runner runner.CLIRunner

	_, err := runner.Run("flux", "create", "source", "helm", HELMREPOSITORY_NAME, "--url", CHART_URL, "--secret-ref", ENTITLEMENT_SECRET_NAME)
	if err != nil {
		fmt.Printf("An error occurred creating helmrepository %v\n", err)
		os.Exit(1)
	}

	values := fmt.Sprintf(`ingress:
  enabled: true
  className: "public-nginx"
  annotations:
    external-dns.alpha.kubernetes.io/hostname: %s
    service.beta.kubernetes.io/aws-load-balancer-type: nlb
    service.beta.kubernetes.io/aws-load-balancer-backend-protocol: http
  hosts:
    - host: %s
      paths:
        - path: /
          pathType: ImplementationSpecific
tls:
  enabled: false
`, userDomain, userDomain)

	valuesFile, err := os.Create(VALUES_FILES_LOCATION)
	if err != nil {
		fmt.Printf("An error occurred creating values file %v\n", err)
		os.Exit(1)
	}

	defer valuesFile.Close()
	_, err = valuesFile.WriteString(values)
	if err != nil {
		fmt.Printf("An error occurred writing values file %v\n", err)
		os.Exit(1)
	}

	err = valuesFile.Sync()
	if err != nil {
		fmt.Printf("An error occurred finializing writing values file %v\n", err)
		os.Exit(1)
	}

	_, err = runner.Run("flux", "create", "hr", HELMRELEASE_NAME,
		"--source", fmt.Sprintf("HelmRepository/%s", HELMREPOSITORY_NAME),
		"--chart", "mccp",
		"--chart-version", version,
		"--interval", "65m",
		"--crds", "CreateReplace",
		"--values", VALUES_FILES_LOCATION,
	)

	if err != nil {
		fmt.Printf("An error occurred creating helmrelease %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✔ WGE v%s is installed successfully\n\n✅ You can visit the UI at https://%s/\n", version, userDomain)

	// _, err = runner.Run("kubectl", "-n", "flux-system", "port-forward", "svc/clusters-service", "8000:8000")
	// if err != nil {
	// 	fmt.Printf("An error occurred port-forwarding %v\n", err)
	// 	os.Exit(1)
	// }

}
