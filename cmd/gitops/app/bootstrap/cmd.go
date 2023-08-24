package bootstrap

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/commands"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/controllers"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

var extraControllers []string = []string{
	"None",
	"policy-agent",
	"pipeline-controller",
	"gitopssets-controller",
}

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstraps Weave gitops enterprise",
		Example: `
# Bootstrap Weave-gitops-enterprise
gitops bootstrap`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Bootstrap()
		},
	}

	cmd.AddCommand(controllers.Command())

	return cmd
}

// Bootstrap initiated by the command runs the WGE bootstrap steps
func Bootstrap() error {
	err := commands.CheckEntitlementFile()
	if err != nil {
		return err
	}

	err = commands.CheckFluxIsInstalled()
	if err != nil {
		return err
	}

	err = commands.CheckFluxReconcile()
	if err != nil {
		return err
	}

	wgeVersion, err := commands.SelectWgeVersion()
	if err != nil {
		return err
	}

	err = commands.CreateAdminPasswordSecret()
	if err != nil {
		return err
	}

	err, isExternalDomain, userDomain := commands.InstallWge(wgeVersion)
	if err != nil {
		return err
	}

	err = commands.CreateOIDCConfig(isExternalDomain, userDomain, wgeVersion)
	if err != nil {
		return err
	}

	// check if the UI is running on localhost or external domain
	CheckUIDomain(isExternalDomain, userDomain, wgeVersion)

	return nil
}

func CheckUIDomain(isExternalDomain bool, userDomain string, wgeVersion string) {
	if isExternalDomain {
		fmt.Printf("✔ WGE v%s is installed successfully\n\n✅ You can visit the UI at https://%s/\n", wgeVersion, userDomain)
	} else {
		utils.Info("✔ WGE v%s is installed successfully\n\n✅ You can visit the UI at http://localhost:8000/\n", wgeVersion)

		var runner runner.CLIRunner
		out, err := runner.Run("kubectl", "-n", "flux-system", "port-forward", "svc/clusters-service", "8000:8000")
		utils.CheckIfError(err, string(out))
	}
}
