package commands

import (
	"errors"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/runner"
)

var CheckUIDomainStep = BootstrapStep{
	Name: "check ui domain",
	Step: checkUIDomain,
}

// checkUIDomain display the message to be for external dns or localhost.
func checkUIDomain(input []StepInput, c *Config) ([]StepOutput, error) {
	if !strings.Contains(c.UserDomain, domainTypelocalhost) {
		c.Logger.Successf(installSuccessMsg, c.WGEVersion, c.UserDomain)
		return []StepOutput{}, nil
	}

	c.Logger.Successf(localInstallSuccessMsg, c.WGEVersion)

	var runner runner.CLIRunner
	_, err := runner.Run("kubectl", "-n", "flux-system", "port-forward", "svc/clusters-service", "8000:8000")
	if err != nil {
		// adding an error message, err is meaningless
		return []StepOutput{}, errors.New("failed to make portforward 8000")
	}

	return []StepOutput{}, nil
}
