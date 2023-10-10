package steps

import (
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

var CheckUIDomainStep = BootstrapStep{
	Name: "Preparing dashboard domain",
	Step: checkUIDomain,
}

// checkUIDomain display the message to be for external dns or localhost.
func checkUIDomain(input []StepInput, c *Config) ([]StepOutput, error) {
	if err := utils.ReconcileHelmRelease(WGEHelmReleaseName); err != nil {
		return []StepOutput{}, err
	}
	if !strings.Contains(c.UserDomain, domainTypeLocalhost) {
		return []StepOutput{
			{
				Name:  "domain msg",
				Type:  successMsg,
				Value: fmt.Sprintf(installSuccessMsg, c.WGEVersion, c.UserDomain),
			},
		}, nil
	}

	return []StepOutput{
		{
			Name:  "localhost msg",
			Type:  successMsg,
			Value: fmt.Sprintf(localInstallSuccessMsg, c.WGEVersion),
		},
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
