package profiles

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/commands"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

const (
	tfGettingSarted     = "Terraform Controller is installed successfully, please follow the getting started guide to continue: https://docs.gitops.weave.works/enterprise/getting-started/terraform/"
	tfInstallInfoMsg    = "Installing Terraform Controller ..."
	tfInstallConfirmMsg = "Terraform Controller is installed successfully"
)

const (
	tfCommitMsg             = "Add terraform controller"
	tfControllerUrl         = "https://raw.githubusercontent.com/weaveworks/tf-controller/main/docs/release.yaml"
	tfFileName              = "tf-controller.yaml"
	tfReleaseErrorMsgFormat = "error getting terraform release %d"
)

func TerraformCommand(opts *config.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "terraform",
		Short: "Add Terraform Controller",
		Example: `
# Add Terraform Controller
gitops add controllers terraform`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return InstallTerraform(opts)
		},
	}
}

// InstallTerraform start installing policy agent helm chart
func InstallTerraform(opts *config.Options) error {
	utils.Warning(tfGettingSarted)
	utils.Warning(tfInstallInfoMsg)

	resp, err := http.Get(tfControllerUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var bodyBytes []byte
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(tfReleaseErrorMsgFormat, resp.StatusCode)
	}

	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	pathInRepo, err := utils.CloneRepo()
	if err != nil {
		return err
	}

	defer func() {
		err = utils.CleanupRepo()
		if err != nil {
			utils.Warning(utils.RepoCleanupMsg)
		}
	}()

	err = utils.CreateFileToRepo(tfFileName, string(bodyBytes), pathInRepo, tfCommitMsg)
	if err != nil {
		return err
	}

	values := map[string]interface{}{
		domain.TerraformValuesName: true,
	}
	err = commands.UpdateHelmReleaseValues(domain.TerraformValuesName, values)
	if err != nil {
		return err
	}
	utils.Info(tfInstallConfirmMsg)
	return nil
}
