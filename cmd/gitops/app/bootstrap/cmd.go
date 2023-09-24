package bootstrap

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/commands"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

const (
	cmdName             = "bootstrap"
	cmdShortDescription = "Bootstraps Weave GitOps Enterprise"
	cmdLongDescription  = `
# Bootstrap Weave GitOps Enterprise

gitops bootstrap

This will help getting started with Weave GitOps Enterprise through simple steps in bootstrap by performing the following tasks:
- Verify the entitlement file exist on the cluster and valid.
- Verify Flux installation is valid.
- Allow option to bootstrap Flux in the generic git server way if not installed.
- Allow selecting the version of WGE to be installed from the latest 3 versions.
- Set the admin password for WGE Dashboard.
- Easy steps to make OIDC flow
`
	redColor = "\x1b[31;1m%w\x1b[0m"
)

type bootstrapFlags struct {
	silent bool
}

var bootstrapArgs bootstrapFlags

func Command(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     cmdName,
		Short:   cmdShortDescription,
		Example: cmdLongDescription,
		RunE:    getBootstrapCmdRunE(opts),
	}

	cmd.Flags().BoolVarP(&bootstrapArgs.silent, "silent", "s", false, "install with the default values without user confirmation")

	return cmd
}

func getBootstrapCmdRunE(opts *config.Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := bootstrap(opts, bootstrapArgs); err != nil {
			return fmt.Errorf(redColor, err)
		}
		return nil
	}
}

// Bootstrap initiated by the command runs the WGE bootstrap steps
func bootstrap(opts *config.Options, bootstrapArgs bootstrapFlags) error {
	// creating kubernetes client to use it in the commands

	kubernetesClient, err := utils.GetKubernetesClient(opts.Kubeconfig)
	if err != nil {
		return err
	}

	if err := commands.CheckEntitlementSecret(kubernetesClient); err != nil {
		return err
	}

	if err := commands.VerifyFluxInstallation(kubernetesClient); err != nil {
		return err
	}

	wgeVersion, err := commands.SelectWgeVersion(kubernetesClient, bootstrapArgs.silent)
	if err != nil {
		return err
	}

	if err := commands.AskAdminCredsSecret(kubernetesClient, bootstrapArgs.silent); err != nil {
		return err
	}

	userDomain, err := commands.InstallWge(kubernetesClient, wgeVersion, bootstrapArgs.silent)
	if err != nil {
		return err
	}

	if err = commands.CheckUIDomain(kubernetesClient, userDomain, wgeVersion); err != nil {
		return err
	}

	return nil
}
