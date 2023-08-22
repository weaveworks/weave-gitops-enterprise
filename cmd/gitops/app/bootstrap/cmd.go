package bootstrap

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/commands"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/controllers"
)

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
	commands.InstallWge(wgeVersion)
	commands.CheckExtraControllers(wgeVersion)
	return nil
}
