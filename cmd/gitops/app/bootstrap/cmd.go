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
# Run Weave GitOps Enterprise bootstrapping in interactive session creating resources to the cluster and the Git repo.
gitops bootstrap

# Run Weave GitOps Enterprise bootstrapping in non-interactive session. It uses values from flags and default values. It fails in case cannot complete the journey without asking the user. 
gitops bootstrap --silent

# Run Weave GitOps Enterprise bootstrapping in interactive session writing resources to stdout 
gitops bootstrap --export  > bootstrap-weave-gitops-enterprise.yaml

# Run Weave GitOps Enterprise bootstrapping from a specific Kubeconfig
gitops bootstrap --kubeconfig <your-kubeconfig-location>

# Run Weave GitOps Enterprise bootstrapping with OIDC and Flux bootstrap with https
gitops bootstrap --silent --version=<version> --password=<admin-password> --discovery-url=<oidc-discovery-url> --client-id=<oidc-client-id> --client-secret=<oidc-secret> --git-username=<git-username-https> -gitPassword=<gitPassword>--branch=<git-branch> --repo-path=<path-in-repo-for-management-cluster> --repo-url=https://<repo-url> 

# Run Weave GitOps Enterprise bootstrapping with OIDC and flux bootstrap with ssh
gitops bootstrap --silent --version=<version> --password=<admin-password> --discovery-url=<oidc-discovery-url> --client-id=<oidc-client-id> --client-secret=<oidc-secret> --private-key-path=<private-key-path> --private-key-password=<private-key-password> --branch=<git-branch> --repo-path=<path-in-repo-for-management-cluster> --repo-url=ssh://<repo-url>

# Run Weave GitOps Enterprise bootstrapping with extra components 
gitops bootstrap --components-extra="policy-agent,tf-controller"
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
	export bool

	// flux flag
	bootstrapFlux bool

	// extra controllers
	componentsExtra []string
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
	cmd.Flags().StringSliceVar(&flags.componentsExtra, "components-extra", nil, "extra components to be installed. Supported components: none, policy-agent, tf-controller")
	cmd.PersistentFlags().BoolVarP(&flags.silent, "silent", "s", false, "non-interactive session: it will not ask questions but rather to use default values to complete the introduced flags")
	cmd.PersistentFlags().BoolVarP(&flags.bootstrapFlux, "bootstrap-flux", "", false, "flags that you want to bootstrap Flux in case is not detected")
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
	cmd.PersistentFlags().BoolVar(&flags.export, "export", false, "write to stdout the bootstrapping manifests without writing in the cluster or Git. It requires Flux to be bootstrapped.")
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
			WithBootstrapFluxFlag(flags.bootstrapFlux).
			WithComponentsExtra(flags.componentsExtra).
			WithSilent(flags.silent).
			WithExport(flags.export).
			WithInReader(cmd.InOrStdin()).
			WithOutWriter(cmd.OutOrStdout()).
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
