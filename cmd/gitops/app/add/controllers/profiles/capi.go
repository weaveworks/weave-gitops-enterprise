package profiles

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/commands"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

const (
	templatesNamespaceMsg = "Please input the default namespace for templates"
	clusterNamespaceMsg   = "Please input the default namespace for clusters"
	capiGettingSartedMsg  = "CAPI Controller is installed successfully, please follow the getting started guide to continue: https://docs.gitops.weave.works/enterprise/getting-started/capi/"
	capiInstallInfoMsg    = "Installing CAPI Controller ..."
	capiInstallConfirmMsg = "CAPI Controller is installed successfully"
)

func CapiCommand(opts *config.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "capi",
		Short: "Add capi controller",
		Example: `
# Add Weave Policy Agent
gitops add controllers capi`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return InstallCapi(opts)
		},
	}
}

// InstallCapi start installing CAPI controller
func InstallCapi(opts *config.Options) error {
	utils.Warning(capiGettingSartedMsg)

	templatesNamespace, err := utils.GetStringInput(templatesNamespaceMsg, "default")
	if err != nil {
		return err
	}

	clustersNamespace, err := utils.GetStringInput(clusterNamespaceMsg, "default")
	if err != nil {
		return err
	}

	values, err := constructCAPIValues(templatesNamespace, clustersNamespace)
	if err != nil {
		return err
	}

	utils.Warning(capiInstallInfoMsg)
	err = commands.UpdateHelmReleaseValues(domain.CAPIValuesName, values)
	if err != nil {
		return err
	}

	utils.Info(capiInstallConfirmMsg)
	return nil
}

func constructCAPIValues(templatesNamespace string, clustersNamespace string) (map[string]interface{}, error) {
	branch, err := utils.GetRepoBranch()
	if err != nil {
		return map[string]interface{}{}, nil
	}

	url, err := utils.GetRepoUrl()
	if err != nil {
		return map[string]interface{}{}, nil
	}
	url = strings.Replace(url, ":", "/", 1)
	url = strings.Replace(url, "git@", "https://", 1)

	path, err := utils.GetRepoPath()
	if err != nil {
		return map[string]interface{}{}, nil
	}

	values := map[string]interface{}{
		"repositoryURL":          url,
		"repositoryPath":         fmt.Sprintf("%s/clusters", path),
		"repositoryClustersPath": path,
		"baseBranch":             branch,
		"templates": map[string]interface{}{
			"namespace": templatesNamespace,
		},
		"clusters": map[string]interface{}{
			"namespace": clustersNamespace,
		},
	}

	return values, nil
}
