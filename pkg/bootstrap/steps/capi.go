package steps

import (
	"encoding/json"
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
)

const (
	capiInstallInfoMsg    = "installing Capi Controller ..."
	capiInstallConfirmMsg = "Capi Controller is installed successfully"
)

const (
	capiCommitMsg    = "Add Capi Controller"
	defaultNamespace = "default"
)

// NewInstallCapiControllerStep ask for continue installing OIDC
func NewInstallCapiControllerStep(config Config) BootstrapStep {
	return BootstrapStep{
		Name: "install Capi Controller",
		Step: installCapi,
	}
}

// installCapi start installing CAPI controller
func installCapi(input []StepInput, c *Config) ([]StepOutput, error) {
	c.Logger.Actionf(capiInstallInfoMsg)

	capiValues := map[string]interface{}{
		"repositoryURL":          c.RepoURL,
		"repositoryPath":         fmt.Sprintf("%s/clusters", c.RepoPath),
		"repositoryClustersPath": c.RepoPath,
		"baseBranch":             c.Branch,
		"templates": map[string]interface{}{
			"namespace": defaultNamespace,
		},
		"clusters": map[string]interface{}{
			"namespace": defaultNamespace,
		},
	}

	// enable capi in wge chart
	valuesBytes, err := utils.GetHelmReleaseValues(c.KubernetesClient, WgeHelmReleaseName, WGEDefaultNamespace)
	if err != nil {
		return []StepOutput{}, err
	}
	var wgeValues valuesFile

	err = json.Unmarshal(valuesBytes, &wgeValues)
	if err != nil {
		return []StepOutput{}, err
	}

	wgeValues.Global.CapiEnabled = true
	wgeValues.Config.CAPI = capiValues

	wgeHelmRelease, err := constructWGEhelmRelease(wgeValues, c.WGEVersion)
	if err != nil {
		return []StepOutput{}, err
	}
	c.Logger.Actionf("rendered WGE HelmRelease file")

	c.Logger.Actionf("updating HelmRelease file")
	helmreleaseFile := fileContent{
		Name:      wgeHelmReleaseFileName,
		Content:   wgeHelmRelease,
		CommitMsg: capiCommitMsg,
	}

	c.Logger.Successf(capiInstallConfirmMsg)
	return []StepOutput{
		{
			Name:  wgeHelmReleaseFileName,
			Type:  typeFile,
			Value: helmreleaseFile,
		},
	}, nil
}
