package profiles

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/commands"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

var TerraformCommand = &cobra.Command{
	Use:   "terraform",
	Short: "Bootstraps Terraform Controller",
	Example: `
# Bootstrap Terraform Controller
gitops bootstrap controllers terraform`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return InstallTerraform()
	},
}

const (
	TF_CONTROLLER_URL = "https://raw.githubusercontent.com/weaveworks/tf-controller/main/docs/release.yaml"
	TF_FILE_NAME      = "tf-controller.yaml"
	TF_COMMIT_MSG     = "Add terraform controller"
)

// InstallTerraform start installing policy agent helm chart
func InstallTerraform() error {
	utils.Warning("For more information about the configurations please refer to the docs https://weaveworks.github.io/tf-controller/getting_started/")
	utils.Warning("Installing Terraform Controller ...")

	resp, err := http.Get(TF_CONTROLLER_URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var bodyBytes []byte
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error getting terraform release %d", resp.StatusCode)
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
			utils.Warning("cleanup failed!")
		}
	}()

	err = utils.CreateFileToRepo(TF_FILE_NAME, string(bodyBytes), pathInRepo, TF_COMMIT_MSG)
	if err != nil {
		return err
	}

	values := map[string]interface{}{
		domain.TERRAFORM_VALUES_NAME: true,
	}
	err = commands.InstallController(domain.TERRAFORM_VALUES_NAME, values)
	if err != nil {
		return err
	}
	utils.Info("✔ Terraform Controller is installed successfully")
	return nil
}
