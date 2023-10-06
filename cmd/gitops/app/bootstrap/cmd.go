package bootstrap

import (
	"os"

	"github.com/spf13/cobra"
	. "github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/steps"
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

# Start WGE installation with given 'username' and 'password'
gitops bootstrap --username wego-admin --password=hell0!
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

		cliLogger := logger.NewCLILogger(os.Stdout)

		// create config from flags
		c, err := steps.NewConfigBuilder().
			WithLogWriter(cliLogger).
			WithKubeconfig(opts.Kubeconfig).
			WithUsername(flags.username).
			WithPassword(flags.password).
			WithVersion(flags.version).
			Build()

		if err != nil {
			cliLogger.Failuref("cannot config bootstrap: %w", err)
		}

		err = Bootstrap(c)
		if err != nil {
			cliLogger.Failuref("cannot execute bootstrap: %w", err)
		}

	}
}
