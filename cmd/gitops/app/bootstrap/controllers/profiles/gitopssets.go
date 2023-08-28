package profiles

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/commands"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

var GitopsSetsCommand = &cobra.Command{
	Use:   "gitopssets",
	Short: "Bootstraps GitOpsSets Controller",
	Example: `
# Bootstrap GitOpsSets Controller
gitops bootstrap controllers gitopssets`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return InstallGitopsSetsController()
	},
}

// InstallPolicyAgent start installing policy agent helm chart
func InstallGitopsSetsController() error {
	utils.Warning("For more information about the Gitopssets controller please refer to the docs https://docs.gitops.weave.works/docs/gitopssets/gitopssets-intro/")

	values := constructGitopsSetsValues()

	utils.Warning("Installing Gitopssets controller ...")
	err := commands.UpdateHelmReleaseValues(domain.GITOPSSETS_VALUES_NAME, values)
	if err != nil {
		return err
	}

	utils.Info("Gitopssets controller is installed successfully")
	return nil
}

func constructGitopsSetsValues() map[string]interface{} {
	values := map[string]interface{}{
		"enabled": true,
		"controllerManager": map[string]interface{}{
			"manager": map[string]interface{}{
				"args": []string{
					"--health-probe-bind-address=:8081",
					"--metrics-bind-address=127.0.0.1:8080",
					"--leader-elect",
					"--enabled-generators=GitRepository,Cluster,PullRequests,List,APIClient,Matrix,Config",
				},
			},
		},
	}

	return values
}
