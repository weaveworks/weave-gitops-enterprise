package bootstrap

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	. "github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/steps"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

const (
	autCmdName             = "auth"
	autCmdShortDescription = "Generate authentication configuration for Weave GitOps. You can specify the type of authentication using the '--type' flag. Currently, only OIDC is supported."
	authCmdExamples        = `
# Add OIDC configuration to your cluster. 
gitops bootstrap auth --type=oidc

# Add OIDC configuration from a specific kubeconfig
gitops bootstrap auth --type=oidc --kubeconfig <your-kubeconfig-location>

# Add OIDC configuration with given oidc configurations 'discoveryURL' 'client-id' 'client-secret'
gitops bootstrap auth --type=oidc --client-id <client-id> --client-secret <client-secret> --discovery-url <discovery-url>
`
)

type authConfigFlags struct {
	authType string
}

var authFlags authConfigFlags

func AuthCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     autCmdName,
		Short:   autCmdShortDescription,
		Example: authCmdExamples,
		Run: func(cmd *cobra.Command, args []string) {
			err := getAuthCmdRun(opts)(cmd, args)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&authFlags.authType, "type", "t", "", "type of authentication to be configured")

	return cmd
}

func getAuthCmdRun(opts *config.Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cliLogger := logger.NewCLILogger(os.Stdout)

		c, err := steps.NewConfigBuilder().
			WithLogWriter(cliLogger).
			WithKubeconfig(opts.Kubeconfig).
			WithGitRepository(flags.repoURL,
				flags.branch,
				flags.repoPath,
			).
			WithGitAuthentication(flags.privateKeyPath,
				flags.privateKeyPassword,
				flags.gitUsername,
				flags.gitPassword,
			).
			WithOIDCConfig(flags.discoveryURL, flags.clientID, flags.clientSecret, false).
			WithSilent(flags.silent).
			WithExport(flags.export).
			WithInReader(cmd.InOrStdin()).
			WithOutWriter(cmd.OutOrStdout()).
			Build()

		if err != nil {
			return fmt.Errorf("cannot config bootstrap auth: %v", err)
		}

		err = BootstrapAuth(c)
		if err != nil {
			return fmt.Errorf("cannot bootstrap auth: %v", err)
		}

		return nil

	}
}
