package steps

import (
	"errors"
	"os"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// user messages
const (
	continueExistingInstallation = "do you want to continue with updating the current installation"
	checkMsg                     = "checking for existing installation in namespace: %s"
	abortMsg                     = "installation aborted"
	noInstallationExistMsg       = "no installation found in namespace: %s"
	installationExistMsg         = "found WGE v%s already installed on your cluster"
)

var continueUsingCurrentVersionInput = StepInput{
	Name:     ExistingInstallation,
	Type:     confirmInput,
	Msg:      continueExistingInstallation,
	Valuesfn: askContinueWithExistingVersion,
}

// NewContinueWithExistingWGEInstallationStep step to search for existing installation and ask user to continue or no
func NewContinueWithExistingWGEInstallationStep(config Config) BootstrapStep {
	return BootstrapStep{
		Name: versionStepName,
		Input: []StepInput{
			continueUsingCurrentVersionInput,
		},
		Step: continueWithExistingInstallation,
	}
}

// continueWithExistingInstallation step ask user to if he wish to continue with currently installed WGE
func continueWithExistingInstallation(input []StepInput, c *Config) ([]StepOutput, error) {
	for _, param := range input {
		if param.Name == ExistingInstallation {
			continueExistingInstallationFlag, ok := param.Value.(string)
			if ok {
				if continueExistingInstallationFlag != "y" {
					return []StepOutput{}, errors.New(abortMsg)
				}
			}
		}
	}
	return []StepOutput{}, nil
}

func askContinueWithExistingVersion(input []StepInput, c *Config) (interface{}, error) {
	c.Logger.Actionf(checkMsg, WGEDefaultNamespace)
	installedVersion, err := utils.GetHelmReleaseProperty(c.KubernetesClient, WgeHelmReleaseName, WGEDefaultNamespace, utils.HelmVersionProperty)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			c.Logger.Failuref("unexpected error finding existing WGE helm release: %v", err)
			os.Exit(1)
		}
		c.Logger.Successf(noInstallationExistMsg, WGEDefaultNamespace)
		return false, nil
	}
	c.Logger.Warningf(installationExistMsg, installedVersion)
	return true, nil
}
