package steps

import (
	"encoding/json"
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"golang.org/x/exp/slices"
)

const (
	tfInstallInfoMsg    = "installing Terraform Controller ..."
	tfInstallConfirmMsg = "Terraform Controller is installed successfully"
)

const (
	tfCommitMsg     = "Add Terraform Controller HelmRelease"
	tfControllerUrl = "https://raw.githubusercontent.com/weaveworks/tf-controller/main/docs/release.yaml"
	tfFileName      = "tf-controller.yaml"
)

// NewInstallTFControllerStep creates the terraform installation step
func NewInstallTFControllerStep(config Config) BootstrapStep {
	if slices.Contains(config.ComponentsExtra.Existing, tfController) {
		config.Logger.Warningf(" not installing %s: found in the cluster", tfController)
		return BootstrapStep{
			Name:  "existing tf controller installation",
			Input: []StepInput{},
			Step:  doNothingStep,
		}
	}

	return BootstrapStep{
		Name: "install Terraform Controller",
		Step: installTerraform,
	}
}

// installTerraform start installing terraform controller helm release
func installTerraform(input []StepInput, c *Config) ([]StepOutput, error) {
	c.Logger.Actionf(tfInstallInfoMsg)

	bodyBytes, err := doBasicAuthGetRequest(tfControllerUrl, "", "")
	if err != nil {
		return []StepOutput{}, fmt.Errorf("error getting Terraform Controller HelmRelease: %v", err)
	}

	// enable tf ui
	valuesBytes, err := utils.GetHelmReleaseValues(c.KubernetesClient, WgeHelmReleaseName, WGEDefaultNamespace)
	if err != nil {
		return []StepOutput{}, err
	}
	var wgeValues valuesFile

	err = json.Unmarshal(valuesBytes, &wgeValues)
	if err != nil {
		return []StepOutput{}, err
	}
	wgeValues.EnableTerraformUI = true

	wgeHelmRelease, err := constructWGEhelmRelease(wgeValues, c.WGEVersion)
	if err != nil {
		return []StepOutput{}, err
	}

	c.Logger.Actionf("rendered WGE HelmRelease file")

	c.Logger.Actionf("updating HelmRelease file")
	helmreleaseFile := fileContent{
		Name:      wgeHelmReleaseFileName,
		Content:   wgeHelmRelease,
		CommitMsg: tfCommitMsg,
	}

	tfHelmFile := fileContent{
		Name:      tfFileName,
		Content:   string(bodyBytes),
		CommitMsg: tfCommitMsg,
	}

	c.Logger.Successf(tfInstallConfirmMsg)
	return []StepOutput{
		{
			Name:  tfFileName,
			Type:  typeFile,
			Value: tfHelmFile,
		},
		{
			Name:  wgeHelmReleaseFileName,
			Type:  typeFile,
			Value: helmreleaseFile,
		},
	}, nil
}
