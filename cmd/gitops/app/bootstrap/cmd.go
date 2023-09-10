package bootstrap

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/commands"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

const (
	cmdName             = "bootstrap"
	cmdShortDescription = "Bootstraps Weave gitops enterprise"
	cmdLongDescription  = `
# Bootstrap Weave-gitops-enterprise

gitops bootstrap

This will help getting started with Weave GitOps Enterprise through simple steps in bootstrap by performing the following tasks:
- Verify the entitlement file exist on the cluster and valid.
- Verify Flux installation is valid.
- Allow option to bootstrap Flux in the generic git server way if not installed.
- Allow selecting the version of WGE to be installed from the latest 3 versions.
- Set the admin password for WGE Dashboard.
- Easy steps to make OIDC flow
`
)

func Command(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     cmdName,
		Short:   cmdShortDescription,
		Example: cmdLongDescription,
		RunE:    getBootstrapCmdRunE(opts),
	}
	return cmd
}

func getBootstrapCmdRunE(opts *config.Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := bootstrap(opts); err != nil {
			return fmt.Errorf("\x1b[31;1m%s\x1b[0m", err.Error())
		}
		return nil
	}
}

// Bootstrap initiated by the command runs the WGE bootstrap steps
func bootstrap(opts *config.Options) error {
	if err := commands.CheckEntitlementFile(*opts); err != nil {
		return err
	}

	if err := commands.CheckFluxIsInstalled(*opts); err != nil {
		return err
	}

	if err := commands.CheckFluxReconcile(*opts); err != nil {
		return err
	}

	wgeVersion, err := commands.SelectWgeVersion(*opts)
	if err != nil {
		return err
	}

	if err := commands.CreateAdminPasswordSecret(*opts); err != nil {
		return err
	}

	userDomain, err := commands.InstallWge(*opts, wgeVersion)
	if err != nil {
		return err
	}

	if err = commands.CreateOIDCConfig(*opts, userDomain, wgeVersion); err != nil {
		return err
	}

	// check if the UI is running on localhost or external domain
	if err = commands.CheckUIDomain(*opts, userDomain, wgeVersion); err != nil {
		return err
	}

	return nil
}
