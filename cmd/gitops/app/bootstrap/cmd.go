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

	userDomain, err := commands.InstallWge(wgeVersion)
	if err != nil {
		return err
	}

	err = commands.CreateOIDCConfig(userDomain, wgeVersion)
	if err != nil {
		return err
	}

	// check if the UI is running on localhost or external domain
	err = commands.CheckUIDomain(userDomain, wgeVersion)
	if err != nil {
		return err
	}

	return nil
}
