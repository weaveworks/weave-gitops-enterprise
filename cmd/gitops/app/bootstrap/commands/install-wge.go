package commands

import (
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

const HELMREPOSITORY_NAME string = "weave-gitops-enterprise-charts"
const HELMRELEASE_NAME string = "weave-gitops-enterprise"
const DOMAIN_TYPE_LOCALHOST string = "localhost (Using Portforward)"
const DOMAIN_TYPE_EXTERNALDNS string = "external DNS"

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

	userDomain := "localhost"
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

	pathInRepo, err := utils.CloneRepo()
	utils.CheckIfError(err)

	wgeHelmRepo := fmt.Sprintf(`apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: %s
  namespace: flux-system
spec:
  interval: 1m0s
  secretRef:
    name: %s
  url: %s
`, HELMREPOSITORY_NAME, ENTITLEMENT_SECRET_NAME, CHART_URL)

	err = utils.CreateFileToRepo("wge-hrepo.yaml", wgeHelmRepo, pathInRepo, "create wge helmrepository")
	utils.CheckIfError(err)

	wgeHelmRelease := fmt.Sprintf(`apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: %s
  namespace: flux-system
spec:
  chart:
    spec:
      chart: mccp
      reconcileStrategy: ChartVersion
      sourceRef:
        kind: HelmRepository
        name: %s
      version: %s
  install:
    crds: CreateReplace
  interval: 1h5m0s
  upgrade:
    crds: CreateReplace
  values:
    ingress:
      annotations:
        external-dns.alpha.kubernetes.io/hostname: %s
        service.beta.kubernetes.io/aws-load-balancer-backend-protocol: http
        service.beta.kubernetes.io/aws-load-balancer-type: nlb
      className: public-nginx
      enabled: true
      hosts:
        - host: %s
          paths:
          - path: /
            pathType: ImplementationSpecific
    tls:
        enabled: false
`, HELMRELEASE_NAME, HELMREPOSITORY_NAME, version, userDomain, userDomain)

	err = utils.CreateFileToRepo("wge-hrelease.yaml", wgeHelmRelease, pathInRepo, "create wge helmrelease")
	utils.CheckIfError(err)

	var runner runner.CLIRunner
	out, err := runner.Run("flux", "reconcile", "source", "git", "flux-system")
	utils.CheckIfError(err, string(out))
	out, err = runner.Run("flux", "reconcile", "kustomization", "flux-system")
	utils.CheckIfError(err, string(out))
	out, err = runner.Run("flux", "reconcile", "helmrelease", HELMRELEASE_NAME)
	utils.CheckIfError(err, string(out))

	if strings.Compare(domainType, DOMAIN_TYPE_EXTERNALDNS) == 0 {
		fmt.Printf("✔ WGE v%s is installed successfully\n\n✅ You can visit the UI at https://%s/\n", version, userDomain)
	} else {
		out, err := runner.Run("kubectl", "-n", "flux-system", "port-forward", "svc/clusters-service", "8000:8000")
		utils.CheckIfError(err, string(out))
	}
}
