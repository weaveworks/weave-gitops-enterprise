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

# Start WGE installation with given admin 'password'
gitops bootstrap --password=hell0!

# Start WGE installation using OIDC
gitops bootstrap --client-id <client-id> --client-secret <client-secret> --discovery-url <discovery-url>
`
)

type bootstrapFlags struct {
	// wge version flags
	version string

	// domain flags
	domainType string
	domain     string

	// private key (ssh-auth) flags
	privateKeyPath     string
	privateKeyPassword string

	// https git auth flags
	gitUsername string
	gitToken    string

	repoURL  string
	branch   string
	repoPath string

	// oidc flags
	discoveryURL string
	clientID     string
	clientSecret string
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

	cmd.Flags().StringVarP(&flags.domainType, "domain-type", "t", "", "dashboard domain type: could be 'localhost' or 'externaldns'")
	cmd.Flags().StringVarP(&flags.domain, "domain", "d", "", "the domain to access the dashboard in case of using externaldns")
	cmd.Flags().StringVarP(&flags.version, "version", "v", "", "version of Weave GitOps Enterprise (should be from the latest 3 versions)")
	cmd.PersistentFlags().StringVarP(&flags.gitUsername, "git-username", "", "", "git username used in https authentication type")
	cmd.PersistentFlags().StringVarP(&flags.gitToken, "git-token", "", "", "git token used in https authentication type")
	cmd.PersistentFlags().StringVarP(&flags.branch, "branch", "b", "", "git branch for your flux repository (example: main)")
	cmd.PersistentFlags().StringVarP(&flags.repoPath, "repo-path", "r", "", "git path for your flux repository (example: clusters/my-cluster)")
	cmd.PersistentFlags().StringVarP(&flags.repoURL, "repo-url", "", "", "git repo url for you flux repository (example: ssh://git@github.com/my-org-name/my-repo-name or https://github.com/my-org-name/my-repo-name)")
	cmd.PersistentFlags().StringVarP(&flags.privateKeyPath, "private-key", "k", "", "private key path. This key will be used to push the Weave GitOps Enterprise's resources to the default cluster repository")
	cmd.PersistentFlags().StringVarP(&flags.privateKeyPassword, "private-key-password", "c", "", "private key password. If the private key is encrypted using password")
	cmd.PersistentFlags().StringVarP(&flags.discoveryURL, "discovery-url", "", "", "OIDC discovery URL")
	cmd.PersistentFlags().StringVarP(&flags.clientID, "client-id", "i", "", "OIDC client ID")
	cmd.PersistentFlags().StringVarP(&flags.clientSecret, "client-secret", "s", "", "OIDC client secret")

	cmd.AddCommand(AuthCommand(opts))

	return cmd
}

func getBootstrapCmdRun(opts *config.Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {

		cliLogger := logger.NewCLILogger(os.Stdout)

		// create config from flags
		c, err := steps.NewConfigBuilder().
			WithLogWriter(cliLogger).
			WithKubeconfig(opts.Kubeconfig).
			WithPassword(opts.Password).
			WithVersion(flags.version).
			WithDomainType(flags.domainType).
			WithDomain(flags.domain).
			WithFluxGitRepository(flags.repoURL,
				flags.branch,
				flags.repoPath,
			).
			WithGitAuthentication(flags.privateKeyPath,
				flags.privateKeyPassword,
				flags.gitUsername,
				flags.gitToken,
			).
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
