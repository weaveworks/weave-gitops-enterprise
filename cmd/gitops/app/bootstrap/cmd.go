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
- Weave GitOps: check or install a supported Weave GitOps version with default configuration.
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

# Start WGE installation with OIDC and flux bootstrap with https
gitops bootstrap --version=<version> --password=<admin-password> --discovery-url=<oidc-discovery-url> --client-id=<oidc-client-id> --git-username=<git-username-https> -gitPassword=<gitPassword>--branch=<git-branch> --repo-path=<path-in-repo-for-management-cluster> --repo-url=https://<repo-url> --client-secret=<oidc-secret> -s

# Start WGE installation with OIDC and flux bootstrap with ssh
gitops bootstrap --version=<version> --password=<admin-password> --discovery-url=<oidc-discovery-url> --client-id=<oidc-client-id> --private-key-path=<private-key-path> --private-key-password=<private-key-password> --branch=<git-branch> --repo-path=<path-in-repo-for-management-cluster> --repo-url=ssh://<repo-url> --client-secret=<oidc-secret> -s

# Start WGE installation with more than one extra controller 
gitops bootstrap --components-extra="policy-agent,capi,tf-controller"
`
)

type bootstrapFlags struct {
	// wge version flags
	version string

	// ssh git auth flags
	privateKeyPath     string
	privateKeyPassword string

	// https git auth flags
	gitUsername string
	gitPassword string

	// git repo flags
	repoURL  string
	branch   string
	repoPath string

	// oidc flags
	discoveryURL string
	clientID     string
	clientSecret string

	// modes flags
	silent bool

	// extra controllers
	extraComponents []string
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

	cmd.Flags().StringVarP(&flags.version, "version", "v", "", "version of Weave GitOps Enterprise (should be from the latest 3 versions)")
	cmd.Flags().StringSliceVar(&flags.extraComponents, "components-extra", nil, "extra components to be installed from (policy-agent, tf-controller, capi)")
	cmd.PersistentFlags().BoolVarP(&flags.silent, "silent", "s", false, "chose the defaults with current provided information without asking any questions")
	cmd.PersistentFlags().StringVarP(&flags.gitUsername, "git-username", "", "", "git username used in https authentication type")
	cmd.PersistentFlags().StringVarP(&flags.gitPassword, "git-password", "", "", "git password/token used in https authentication type")
	cmd.PersistentFlags().StringVarP(&flags.branch, "branch", "b", "", "git branch for your flux repository (example: main)")
	cmd.PersistentFlags().StringVarP(&flags.repoPath, "repo-path", "r", "", "git path for your flux repository (example: clusters/my-cluster)")
	cmd.PersistentFlags().StringVarP(&flags.repoURL, "repo-url", "", "", "git repo url for your flux repository (example: ssh://git@github.com/my-org-name/my-repo-name or https://github.com/my-org-name/my-repo-name)")
	cmd.PersistentFlags().StringVarP(&flags.privateKeyPath, "private-key", "k", "", "private key path. This key will be used to push the Weave GitOps Enterprise's resources to the default cluster repository")
	cmd.PersistentFlags().StringVarP(&flags.privateKeyPassword, "private-key-password", "c", "", "private key password. If the private key is encrypted using password")
	cmd.PersistentFlags().StringVarP(&flags.discoveryURL, "discovery-url", "", "", "OIDC discovery URL")
	cmd.PersistentFlags().StringVarP(&flags.clientID, "client-id", "i", "", "OIDC client ID")
	cmd.PersistentFlags().StringVarP(&flags.clientSecret, "client-secret", "", "", "OIDC client secret")
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
			WithGitRepository(flags.repoURL,
				flags.branch,
				flags.repoPath,
			).
			WithGitAuthentication(flags.privateKeyPath,
				flags.privateKeyPassword,
				flags.gitUsername,
				flags.gitPassword,
			).
			WithOIDCConfig(flags.discoveryURL, flags.clientID, flags.clientSecret, true).
			WithSilentFlag(flags.silent).
			WithExtraComponents(flags.extraComponents).
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
