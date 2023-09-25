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

## gitops bootstrap auth --type=oidc

This sub-command adds OIDC configuration to your cluster. You can specify the type of authentication using the '--type' flag. Currently, only OIDC is supported.

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

	//get kubernetes client
	kubernetesClient, err := utils.GetKubernetesClient(opts.Kubeconfig)
	if err != nil {
		fmt.Println(err)
	}

	//get userDomain from helm release
	userDomain, err := utils.GetCurrentDominForHelmRelease(commands.WGEHelmReleaseName, commands.WGEDefaultNamespace)
	if err != nil {
		fmt.Println(err)
	}

	//get current version of WGE
	wgeVersion, err := utils.GetCurrentVersionForHelmRelease(commands.WGEHelmReleaseName, commands.WGEDefaultNamespace)
	if err != nil {
		fmt.Println(err)
	}

	var authType string

	// Add the auth sub-command to the bootstrap command
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Add OIDC configuration to your cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if authType == "oidc" {
				return commands.CreateOIDCConfig(kubernetesClient, userDomain, wgeVersion, true)
			} else {
				// Handle other types of authentication or return an error
				return fmt.Errorf("Unsupported authentication type: %s", authType)
			}
		},
	}

	// Flags for the auth sub-command
	authCmd.Flags().StringVar(&authType, "type", "oidc", "Type of authentication")

	cmd.AddCommand(authCmd)

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

	if err = commands.CreateOIDCConfig(kubernetesClient, userDomain, wgeVersion, false); err != nil {
		return err
	}

	if err = commands.CheckUIDomain(kubernetesClient, userDomain, wgeVersion); err != nil {
		return err
	}

	return nil
}
