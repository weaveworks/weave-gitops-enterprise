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
	cmdShortDescription = `gitops bootstrap will help getting started with Weave GitOps Enterprise through simple steps in bootstrap by performing the following tasks:
- Verify the entitlement file exist on the cluster and valid.
- Verify Flux installation is valid.
- Allow option to bootstrap Flux in the generic git server way if not installed.
- Allow selecting the version of WGE to be installed from the latest 3 versions.
- Set the admin password for WGE Dashboard.
- Easy steps to make OIDC flow
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

	// all the flags are going to be here
	cmd.Flags().StringVarP(&flags.username, "username", "u", "", "Dashboard admin username")
	cmd.Flags().StringVarP(&flags.password, "password", "p", "", "Dashboard admin password")
	cmd.Flags().StringVarP(&flags.version, "version", "v", "", "Weave GitOps Enterprise version")
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
	// parse the flags

	// create configuration
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
	config.Username = flags.username
	config.Password = flags.password
	config.WGEVersion = flags.version
	config.KubernetesClient = kubernetesClient
	config.Logger = logger

	// configure workflow
	var steps = []commands.BootstrapStep{
		commands.CheckEntitlementSecretStep,
		commands.VerifyFluxInstallationStep,
		commands.SelectWgeVersionStep,
		commands.AskAdminCredsSecretStep,
		commands.InstallWGEStep,
		commands.CheckUIDomainStep,
	}

	// here at this point we should have two
	// for a given step we have at this stage configuration
	// 1 - user flags -> --username=wego
	// 3 - user input

	// if there is flag -> we take that flag so we dont need to
	// to add that as an input
	flags.username

	input = commands.AskAdminCredsSecretStep.WithConfig(config)

	// 2 - default configuration  -> password=password

	//execute workflow
	for _, step := range steps {
		err := step.Execute(input, &config)
		if err != nil {
			return err
		}
	}

	return nil
}
