package bootstrap

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/commands"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

const (
	cmdName             = "bootstrap"
	cmdShortDescription = `gitops bootstrap installs Weave GitOps Enterprise in simple steps:
- Entitlements: check that you have valid entitlements.
- Flux: check or bootstrap Flux. 
- Weave Gitops: check or install a supported Weave GitOps version with default configuration.
- Authentication: check or setup cluster user authentication to access the dashboard.
`
	cmdExamples = `
# Start WGE installation from the current kubeconfig
gitops bootstrap

# Start WGE installation from a specific kubeconfig
gitops bootstrap --kubeconfig <your-kubeconfig-location>
`
)

type bootstrapFlags struct {
	username string
	password string
	version  string
}

var flags bootstrapFlags

func Command(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     cmdName,
		Short:   cmdShortDescription,
		Example: cmdExamples,
		Run:     getBootstrapCmdRun(opts),
	}

	cmd.Flags().StringVarP(&flags.username, "username", "u", "", "Dashboard admin username")
	cmd.Flags().StringVarP(&flags.password, "password", "p", "", "Dashboard admin password")
	cmd.Flags().StringVarP(&flags.version, "version", "v", "", "Weave GitOps Enterprise version to install.")
	return cmd
}

func getBootstrapCmdRun(opts *config.Options) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		logger := logger.NewCLILogger(os.Stdout)

		if err := bootstrap(opts, logger); err != nil {
			logger.Failuref(err.Error())
		}
	}
}

// Bootstrap initiated by the command runs the WGE bootstrap steps
func bootstrap(opts *config.Options, logger logger.Logger) error {
	kubernetesClient, err := utils.GetKubernetesClient(opts.Kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to get kubernetes client. error: %s", err)
	}

	installedVersion, err := utils.GetHelmRelease(kubernetesClient, commands.WgeHelmReleaseName, commands.WGEDefaultNamespace)
	if err == nil {
		logger.Successf("WGE version: %s is already installed on your cluster!", installedVersion)
		return nil
	}

	config := commands.Config{}
	config.KubernetesClient = kubernetesClient
	config.Logger = logger

	flagsMap := map[string]string{
		commands.UserName:   flags.username,
		commands.Password:   flags.password,
		commands.WGEVersion: flags.version,
	}

	var steps = []commands.BootstrapStep{
		commands.CheckEntitlementSecretStep,
		commands.VerifyFluxInstallationStep,
		commands.SelectWgeVersionStep,
		commands.AskAdminCredsSecretStep,
		commands.SelectDomainType,
		commands.InstallWGEStep,
		commands.CheckUIDomainStep,
	}

	for _, step := range steps {
		err := step.Execute(&config, flagsMap)
		if err != nil {
			return err
		}
	}

	return nil
}
