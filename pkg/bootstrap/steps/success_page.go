package steps

import (
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
)

const (
	installSuccessMsg = "WGE v%s is installed successfully\nYou can visit the UI at https://%s/ . admin username: `wego-admin`"
	portforwardMsg    = "WGE v%s is installed successfully. To access the dashboard, run the following command to create portforward to the dasboard local domain http://localhost:8000 . admin username: `wego-admin`"
	portforwardCmdMsg = "kubectl -n %s port-forward svc/clusters-service 8000:8000"
)

var (
	CheckUIDomainStep = BootstrapStep{
		Name: "preparing dashboard domain",
		Step: checkUIDomain,
	}
)

// checkUIDomain display the message to be for external dns or localhost.
func checkUIDomain(input []StepInput, c *Config) ([]StepOutput, error) {
	if err := utils.ReconcileHelmRelease(WgeHelmReleaseName); err != nil {
		return []StepOutput{}, err
	}
	if !strings.Contains(c.UserDomain, domainTypeLocalhost) {
		c.Logger.Successf(installSuccessMsg, c.WGEVersion, c.UserDomain)
		return []StepOutput{}, nil
	}

	c.Logger.Successf(portforwardMsg, c.WGEVersion)
	c.Logger.Println(portforwardCmdMsg, WGEDefaultNamespace)
	return []StepOutput{}, nil
}
