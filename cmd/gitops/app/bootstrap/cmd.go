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
	cmdName             = "bootstrap"
	cmdShortDescription = "Installs Weave GitOps Enterprise in simple steps"
	cmdLongDescription  = `Installs Weave GitOps Enterprise in simple steps:
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
	username           string
	password           string
	version            string
	domainType         string
	domain             string
	privateKeyPath     string
	privateKeyPassword string
	discoveryURL       string
	clientID           string
	clientSecret       string
}

var flags bootstrapFlags

func Command(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     cmdName,
		Short:   cmdShortDescription,
		Long:    cmdLongDescription,
		Example: cmdExamples,
		RunE:    getBootstrapCmdRun(opts),
	}

	cmd.AddCommand(AuthCommand(opts))

	cmd.Flags().StringVarP(&flags.username, "username", "u", "", "dashboard admin username")
	cmd.Flags().StringVarP(&flags.password, "password", "p", "", "dashboard admin password")
	cmd.Flags().StringVarP(&flags.version, "version", "v", "", "version of Weave GitOps Enterprise (should be from the latest 3 versions)")
	cmd.Flags().StringVarP(&flags.domainType, "domain-type", "t", "", "dashboard domain type: could be 'localhost' or 'externaldns'")
	cmd.Flags().StringVarP(&flags.domain, "domain", "d", "", "indicate the domain to use in case of using `externaldns`")
	cmd.Flags().StringVarP(&flags.privateKeyPath, "private-key", "k", "", "private key path. This key will be used to push the Weave GitOps Enterprise's resources to the default cluster repository")
	cmd.Flags().StringVarP(&flags.privateKeyPassword, "private-key-password", "c", "", "private key password. If the private key is encrypted using password")
	cmd.Flags().StringVarP(&flags.discoveryURL, "discovery-url", "", "", "OIDC discovery URL")
	cmd.Flags().StringVarP(&flags.clientID, "client-id", "i", "", "OIDC client ID")
	cmd.Flags().StringVarP(&flags.clientSecret, "client-secret", "s", "", "OIDC client secret")

	return cmd
}

func getBootstrapCmdRun(opts *config.Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {

		cliLogger := logger.NewCLILogger(os.Stdout)

		// create config from flags
		c, err := steps.NewConfigBuilder().
			WithLogWriter(cliLogger).
			WithKubeconfig(opts.Kubeconfig).
			WithUsername(flags.username).
			WithPassword(flags.password).
			WithVersion(flags.version).
			WithDomainType(flags.domainType).
			WithDomain(flags.domain).
			WithPrivateKey(flags.privateKeyPath, flags.privateKeyPassword).
			WithOIDCConfig(flags.discoveryURL, flags.clientID, flags.clientSecret, true).
			Build()

		if err != nil {
			return fmt.Errorf("cannot config bootstrap: %v", err)
		}

		err = Bootstrap(c)
		if err != nil {
			return fmt.Errorf("cannot execute bootstrap: %v", err)
		}
		return nil
	}
}
