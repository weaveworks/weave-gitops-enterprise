package profiles

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/services/profiles"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"k8s.io/cli-runtime/pkg/printers"
)

type profileCommandFlags struct {
	RepoName         string
	RepoNamespace    string
	RepoKind         string
	ClusterName      string
	ClusterNamespace string
	Kind             string
}

var flags profileCommandFlags

func GetCommand(opts *config.Options, client *adapters.HTTPClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "profile",
		Aliases:       []string{"profiles"},
		Short:         "Show information about available profiles",
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: `
	# Get all profiles
	gitops get profiles
	`,
		PreRunE: getProfilesCmdPreRunE(&opts.Endpoint),
		RunE:    getProfilesCmdRunE(opts, client),
	}

	cmd.Flags().StringVar(&flags.RepoName, "repo-name", "weaveworks-charts", "Name of the repository")
	cmd.Flags().StringVar(&flags.RepoNamespace, "repo-namespace", "flux-system", "Namespace of the repository")
	cmd.Flags().StringVar(&flags.RepoKind, "repo-kind", "", "Kind of the repository")
	cmd.Flags().StringVar(&flags.ClusterName, "cluster-name", "", "Name of the cluster")
	cmd.Flags().StringVar(&flags.ClusterNamespace, "cluster-namespace", "", "Namespace of the cluster")
	cmd.Flags().StringVar(&flags.Kind, "kind", "", "Kind of the profile")

	return cmd
}

func getProfilesCmdPreRunE(endpoint *string) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		if *endpoint == "" {
			return cmderrors.ErrNoWGEEndpoint
		}

		return nil
	}
}

func getProfilesCmdRunE(opts *config.Options, client *adapters.HTTPClient) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		err := client.ConfigureClientWithOptions(opts, os.Stdout)
		if err != nil {
			return err
		}

		w := printers.GetNewTabWriter(os.Stdout)

		defer w.Flush()

		opts := profiles.GetOptions{
			Kind: flags.Kind,
			Repository: profiles.RepositoryRef{
				Name:      flags.RepoName,
				Namespace: flags.RepoNamespace,
				Kind:      flags.RepoKind,
				Cluster: profiles.ClusterRef{
					Name:      flags.ClusterName,
					Namespace: flags.ClusterNamespace,
				},
			},
		}

		return profiles.NewService(logger.NewCLILogger(os.Stdout)).Get(context.Background(), client, w, opts)
	}
}
