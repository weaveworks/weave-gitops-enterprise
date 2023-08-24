package profiles

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/commands"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

var CapiCommand = &cobra.Command{
	Use:   "capi",
	Short: "Bootstraps capi controller",
	Example: `
# Bootstrap Weave Policy Agent
gitops bootstrap controllers capi`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return InstallCapi()
	},
}

const (
	TEMPLATES_NAMESPACE_MSG = "Please input the default namespace for templates"
	CLUSTERS_NAMESPACE_MSG  = "Please input the default namespace for clusters"
)

// InstallCapi start installing policy agent helm chart
func InstallCapi() error {
	utils.Warning("For more information about the configurations please refer to the docs https://docs.gitops.weave.works/docs/enterprise/getting-started/install-enterprise/#valuesconfigcapirepositoryurl")

	templatesNamespace, err := utils.GetStringInput(TEMPLATES_NAMESPACE_MSG, "default")
	if err != nil {
		return err
	}

	clustersNamespace, err := utils.GetStringInput(CLUSTERS_NAMESPACE_MSG, "default")
	if err != nil {
		return err
	}

	values, err := constructCAPIValues(templatesNamespace, clustersNamespace)
	if err != nil {
		return err
	}

	utils.Warning("Installing CAPI Controller ...")
	err = commands.InstallController(domain.CAPI_VALUES_NAME, values)
	if err != nil {
		return err
	}

	utils.Info("CAPI Controller is installed successfully")
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
