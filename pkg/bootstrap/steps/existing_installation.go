package steps

import (
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

func NewContinueWithExistingWGEInstallationStep(config Config) BootstrapStep {
	return BootstrapStep{
		Name: versionStepName,
		Input: []StepInput{
			continueUsingCurrentVersionInput,
		},
		Step: continueWithExistingInstallation,
	}
}

// selectWgeVersion step ask user to select wge version from the latest 3 versions.
func continueWithExistingInstallation(input []StepInput, c *Config) ([]StepOutput, error) {
	for _, param := range input {
		if param.Name == ExistingInstallation {
			continueExistingInstallationFlag, ok := param.Value.(string)
			if ok {
				if continueExistingInstallationFlag != "y" {
					c.Logger.Println(abortMsg)
					os.Exit(0)
				}
			}
		}
	}
	return []StepOutput{}, nil
}

func askContinueWithExistingVersion(input []StepInput, c *Config) (interface{}, error) {
	c.Logger.Actionf(checkMsg, WGEDefaultNamespace)
	installedVersion, err := utils.GetHelmReleaseVersion(c.KubernetesClient, WgeHelmReleaseName, WGEDefaultNamespace)
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
