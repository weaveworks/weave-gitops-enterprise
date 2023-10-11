package steps

import (
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

var CheckUIDomainStep = BootstrapStep{
	Name: "preparing dashboard domain",
	Step: checkUIDomain,
}

// checkUIDomain display the message to be for external dns or localhost.
func checkUIDomain(input []StepInput, c *Config) ([]StepOutput, error) {
	if err := utils.ReconcileHelmRelease(WGEHelmReleaseName); err != nil {
		return []StepOutput{}, err
	}
	if !strings.Contains(c.UserDomain, domainTypeLocalhost) {
		c.Logger.Successf(installSuccessMsg, c.WGEVersion, c.UserDomain)
		return []StepOutput{}, nil
	}

	c.Logger.Successf(localInstallSuccessMsg, c.WGEVersion)
	return []StepOutput{
		{
			Name:  "portforward",
			Type:  typePortforward,
			Value: createPortforward,
		},
	}, nil
}

func createPortforward() error {
	var runner runner.CLIRunner
	out, err := runner.Run("kubectl", "-n", "flux-system", "port-forward", "svc/clusters-service", "8000:8000")
	if err != nil {
		return fmt.Errorf("failed to create portforward 8000: %s", string(out))
	}
	return nil
}
