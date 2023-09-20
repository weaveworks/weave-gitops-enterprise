package profiles

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/commands"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	values, err := constructCAPIValues(opts, templatesNamespace, clustersNamespace)
	if err != nil {
		return err
	}

	utils.Warning(capiInstallInfoMsg)

	config, err := clientcmd.BuildConfigFromFlags("", opts.Kubeconfig)
	if err != nil {
		return err
	}
	cl, err := client.New(config, client.Options{})
	if err != nil {
		return err
	}

	err = commands.UpdateHelmReleaseValues(cl, domain.CAPIValuesName, values)
	if err != nil {
		return err
	}

	utils.Info(capiInstallConfirmMsg)
	return nil
}

func constructCAPIValues(opts *config.Options, templatesNamespace string, clustersNamespace string) (map[string]interface{}, error) {

	config, err := clientcmd.BuildConfigFromFlags("", opts.Kubeconfig)
	if err != nil {
		return nil, err
	}
	cl, err := client.New(config, client.Options{})
	if err != nil {
		return nil, err
	}

	branch, err := utils.GetRepoBranch(cl, commands.WGEDefaultRepoName, commands.WGEDefaultNamespace)
	if err != nil {
		return map[string]interface{}{}, nil
	}

	url, err := utils.GetRepoUrl(cl, commands.WGEDefaultRepoName, commands.WGEDefaultNamespace)
	if err != nil {
		return map[string]interface{}{}, nil
	}
	url = strings.Replace(url, ":", "/", 1)
	url = strings.Replace(url, "git@", "https://", 1)

	path, err := utils.GetRepoPath(cl, commands.WGEDefaultRepoName, commands.WGEDefaultNamespace)
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
