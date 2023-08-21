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

func Bootstrap() error {
	commands.CheckEntitlementFile()
	commands.CheckFluxIsInstalled()
	commands.CheckFluxReconcile()
	wgeVersion := commands.SelectWgeVersion()
	commands.CreateAdminPasswordSecret()
	isExternalDomain, uiDomain := commands.InstallWge(wgeVersion)
	commands.CreateOIDCConfig(wgeVersion)
	commands.CheckExtraControllers(wgeVersion)
	// check if the UI is running on localhost or external domain
	CheckUIDomain(isExternalDomain, uiDomain, wgeVersion)

	return nil
}

func CheckUIDomain(isExternalDomain bool, uiDomain string, wgeVersion string) {
	var runner runner.CLIRunner
	if isExternalDomain {
		fmt.Printf("✔ WGE v%s is installed successfully\n\n✅ You can visit the UI at https://%s/\n", wgeVersion, uiDomain)
	} else {
		out, err := runner.Run("kubectl", "-n", "flux-system", "port-forward", "svc/clusters-service", "8000:8000")
		utils.CheckIfError(err, string(out))
	}
}
