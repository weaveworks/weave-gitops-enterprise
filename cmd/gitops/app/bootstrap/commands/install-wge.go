package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

const HELMREPOSITORY_NAME string = "weave-gitops-enterprise-charts"
const HELMRELEASE_NAME string = "weave-gitops-enterprise"
const VALUES_FILES_LOCATION string = "/tmp/mccp-values.yaml"
const DOMAIN_TYPE_LOCALHOST string = "localhost (Using Portforward)"
const DOMAIN_TYPE_EXTERNALDNS string = "external DNS"
const UI_URL_LOCALHOST string = "localhost:8000"

func InstallWge(version string) {

	domainTypes := []string{
		DOMAIN_TYPE_LOCALHOST,
		DOMAIN_TYPE_EXTERNALDNS,
	}

	domainSelectorPrompt := utils.PromptContent{
		ErrorMsg:     "",
		Label:        "Please select the domain to be used",
		DefaultValue: "",
	}
	domainType := utils.GetPromptSelect(domainSelectorPrompt, domainTypes)

	var userDomain string
	if strings.Compare(domainType, DOMAIN_TYPE_EXTERNALDNS) == 0 {
		fmt.Printf("\n\nPlease make sure to have the external DNS service is installed in your cluster, or you have a domain points to your cluster\nFor more information about external DNS please refer to https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-configuring.html\n\n")
		userDomainPrompt := utils.PromptContent{
			ErrorMsg:     "Domain can't be empty",
			Label:        "Please enter your cluster domain",
			DefaultValue: "",
		}
		userDomain = utils.GetPromptStringInput(userDomainPrompt)
	}

	fmt.Printf("✔ All set installing WGE v%s, This may take few minutes...\n", version)
	var runner runner.CLIRunner

	out, err := runner.Run("flux", "create", "source", "helm", HELMREPOSITORY_NAME, "--url", CHART_URL, "--secret-ref", ENTITLEMENT_SECRET_NAME)
	if err != nil {
		fmt.Printf("An error occurred creating helmrepository\n%v\n", string(out))
		os.Exit(1)
	}

	if strings.Compare(domainType, DOMAIN_TYPE_EXTERNALDNS) == 0 {
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
		utils.CheckIfError(err)

		defer valuesFile.Close()
		_, err = valuesFile.WriteString(values)
		utils.CheckIfError(err)

		err = valuesFile.Sync()
		utils.CheckIfError(err)

		fmt.Println("Installing WGE ...")
		out, err := runner.Run("flux", "create", "hr", HELMRELEASE_NAME,
			"--source", fmt.Sprintf("HelmRepository/%s", HELMREPOSITORY_NAME),
			"--chart", "mccp",
			"--chart-version", version,
			"--interval", "65m",
			"--crds", "CreateReplace",
			"--values", VALUES_FILES_LOCATION,
		)
		if err != nil {
			fmt.Printf("An error occurred creating helmrelease\n%v\n", string(out))
			os.Exit(1)
		}
	} else {
		fmt.Println("Installing WGE ...")
		out, err := runner.Run("flux", "create", "hr", HELMRELEASE_NAME,
			"--source", fmt.Sprintf("HelmRepository/%s", HELMREPOSITORY_NAME),
			"--chart", "mccp",
			"--chart-version", version,
			"--interval", "65m",
			"--crds", "CreateReplace")
		if err != nil {
			fmt.Printf("An error occurred creating helmrelease\n%v\n", string(out))
			os.Exit(1)
		}
	}

	if strings.Compare(domainType, DOMAIN_TYPE_EXTERNALDNS) == 0 {
		fmt.Printf("✔ WGE v%s is installed successfully\n\n✅ You can visit the UI at https://%s/\n", version, userDomain)
	} else {
		fmt.Printf("✔ WGE v%s is installed successfully\n\n✅ You can visit the UI at https://%s/\n", version, UI_URL_LOCALHOST)
		out, err := runner.Run("kubectl", "-n", "flux-system", "port-forward", "svc/clusters-service", "8000:8000")
		if err != nil {
			fmt.Printf("An error occurred port-forwarding\n%v\n", string(out))
			os.Exit(1)
		}
	}
}
