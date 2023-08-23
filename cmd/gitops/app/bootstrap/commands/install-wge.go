package commands

import (
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

const (
	DOMAIN_MSG         = "Please select the domain to be used"
	CLUSTER_DOMAIN_MSG = "Please enter your cluster domain"
)

// InstallWge installs weave gitops enterprise chart
func InstallWge(version string) error {
	domainTypes := []string{
		DOMAIN_TYPE_LOCALHOST,
		DOMAIN_TYPE_EXTERNALDNS,
	}

	domainType, err := utils.GetSelectInput(DOMAIN_MSG, domainTypes)
	if err != nil {
		return utils.CheckIfError(err)
	}

	userDomain := "localhost"

	if strings.Compare(domainType, DOMAIN_TYPE_EXTERNALDNS) == 0 {

		utils.Warning("\n\nPlease make sure to have the external DNS service is installed in your cluster, or you have a domain points to your cluster\nFor more information about external DNS please refer to https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-configuring.html\n\n")

		userDomain, err = utils.GetStringInput(CLUSTER_DOMAIN_MSG, "")
		if err != nil {
			return utils.CheckIfError(err)
		}

	}

	utils.Info("✔ All set installing WGE v%s, This may take few minutes...\n", version)

	pathInRepo, err := utils.CloneRepo()
	if err != nil {
		return utils.CheckIfError(err)
	}

	defer func() {
		err = utils.CleanupRepo()
		if err != nil {
			fmt.Println("cleanup failed!")
		}
	}()

	wgehelmRepo, err := constructWgeHelmRepository()
	if err != nil {
		return utils.CheckIfError(err)
	}

	err = utils.CreateFileToRepo(WGE_HELMREPO_FILENAME, wgehelmRepo, pathInRepo, WGE_HELMREPO_COMMITMSG)
	if err != nil {
		return utils.CheckIfError(err)
	}

	wgeHelmRelease, err := constructWGEhelmRelease(userDomain, version)
	if err != nil {
		return utils.CheckIfError(err)
	}

	err = utils.CreateFileToRepo(WGE_HELMRELEASE_FILENAME, wgeHelmRelease, pathInRepo, WGE_HELMRELEASE_COMMITMSG)
	if err != nil {
		return utils.CheckIfError(err)
	}

	err = utils.ReconcileFlux(WGE_HELMRELEASE_NAME)
	if err != nil {
		return utils.CheckIfError(err)
	}

	if strings.Compare(domainType, DOMAIN_TYPE_EXTERNALDNS) == 0 {
		utils.Info("✔ WGE v%s is installed successfully\n\n✅ You can visit the UI at https://%s/\n", version, userDomain)
	} else {
		utils.Info("✔ WGE v%s is installed successfully\n\n✅ You can visit the UI at http://%s:8000/\n", version, userDomain)
		var runner runner.CLIRunner
		out, err := runner.Run("kubectl", "-n", "flux-system", "port-forward", "svc/clusters-service", "8000:8000")
		if err != nil {
			return utils.CheckIfError(err, string(out))
		}
	}
	return nil
}
