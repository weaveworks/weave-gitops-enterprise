package bootstrap

import (
	"fmt"

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
			if err := Bootstrap(); err != nil {
				return fmt.Errorf("\x1b[31;1m%s\x1b[0m", err.Error())
			}
			return nil
		},
	}

	cmd.AddCommand(controllers.Command())

	return cmd
}

// Bootstrap initiated by the command runs the WGE bootstrap steps
func Bootstrap() error {

	if err := commands.CheckEntitlementFile(); err != nil {
		return err
	}

	if err := commands.CheckFluxIsInstalled(); err != nil {
		return err
	}

	if err := commands.CheckFluxReconcile(); err != nil {
		return err
	}

	wgeVersion, err := commands.SelectWgeVersion()
	if err != nil {
		return err
	}

	if err := commands.CreateAdminPasswordSecret(); err != nil {
		return err
	}

	userDomain, err := commands.InstallWge(wgeVersion)
	if err != nil {
		return err
	}

	if err = commands.CreateOIDCConfig(userDomain, wgeVersion); err != nil {
		return err
	}

	// check if the UI is running on localhost or external domain
	if err = commands.CheckUIDomain(userDomain, wgeVersion); err != nil {
		return err
	}

	return nil
}
