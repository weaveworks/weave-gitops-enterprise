package steps

import (
	"errors"
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/runner"
)

// ExposeWge step that exposes weave gitops enterprise ui to the user doing bootstrap.
// It could be either by providing its Url if public or using port-forward if private.
var ExposeWge = BootstrapStep{
	Name: "expose wge ui",
	Step: exposeWge,
}

// exposeWge display the message to be for external dns or localhost.
func exposeWge(input []StepInput, c *Config) ([]StepOutput, error) {
	//TODO we should not reconcile or do a write operation in a read-only operation
	//if err := utils.ReconcileHelmRelease(WgeHelmReleaseName); err != nil {
	//	return []StepOutput{}, err
	//}
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
	_, err := runner.Run("kubectl", "-n", "flux-system", "port-forward", "svc/clusters-service", "8000:8000")
	if err != nil {
		// adding an error message, err is meaningless
		return errors.New("failed to make portforward 8000")
	}
	return nil
}
