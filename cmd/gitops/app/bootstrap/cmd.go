package bootstrap

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/commands"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/domain"
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

## gitops bootstrap auth --type=oidc

This sub-command adds OIDC configuration to your cluster. You can specify the type of authentication using the '--type' flag. Currently, only OIDC is supported.

`
	redError = "\x1b[31;1m%w\x1b[0m"
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

	// Add the auth sub-command to bootstrap command to add OIDC authentication to the cluster
	authCmd, err := createAuthCommand(opts)
	if err != nil {
		fmt.Printf("failed to create auth command: %v\n", err)
		return cmd
	}
	cmd.AddCommand(authCmd)

	return cmd
}

func getBootstrapCmdRunE(opts *config.Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := bootstrap(opts, bootstrapArgs); err != nil {
			return fmt.Errorf(redError, err)
		}
		return nil
	}
}

// Bootstrap initiated by the command runs the WGE bootstrap steps
func bootstrap(opts *config.Options, bootstrapArgs bootstrapFlags) error {
	// creating kubernetes client to use it in the commands
	kubernetesClient, err := utils.GetKubernetesClient(opts.Kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to get kubernetes client. error: %s", err)
	}

	installedVersion, err := utils.GetHelmRelease(kubernetesClient, commands.WgeHelmReleaseName, commands.WGEDefaultNamespace)
	if err == nil {
		utils.Info("WGE version: %s is already installed on your cluster!", installedVersion)
		os.Exit(0)
	}

	if err := commands.CheckEntitlementSecret(kubernetesClient); err != nil {
		return fmt.Errorf("failed to check entitlement secret. error: %s", err)
	}

	if err := commands.VerifyFluxInstallation(kubernetesClient); err != nil {
		return fmt.Errorf("failed to get verify flux installation. error: %s", err)
	}

	wgeVersion, err := commands.SelectWgeVersion(kubernetesClient, bootstrapArgs.silent)
	if err != nil {
		return fmt.Errorf("failed to select WGE version. error: %s", err)
	}

	if err := commands.AskAdminCredsSecret(kubernetesClient, bootstrapArgs.silent); err != nil {
		return fmt.Errorf("failed to create admin secret. error: %s", err)
	}

	userDomain, err := commands.InstallWge(kubernetesClient, wgeVersion, bootstrapArgs.silent)
	if err != nil {
		return fmt.Errorf("failed to install WGE. error: %s", err)
	}

	// initialize config Struct with the userDomain and wgeVersion
	var params domain.OIDCConfigParams = domain.OIDCConfigParams{
		UserDomain: userDomain,
		WGEVersion: wgeVersion,
	}
	if !bootstrapArgs.silent {
		if err = commands.CreateOIDCPrompt(kubernetesClient, params); err != nil {
			return err
		}

		// Ask the user if he wants to revert the admin user
		if err := commands.CheckAdminPasswordRevert(kubernetesClient); err != nil {
			return err
		}
	}

	if err = commands.CheckUIDomain(kubernetesClient, userDomain, wgeVersion); err != nil {
		return fmt.Errorf("failed to get WGE dashboard domain. error: %s", err)
	}

	return nil
}
