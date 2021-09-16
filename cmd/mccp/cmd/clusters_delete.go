package cmd

import (
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/mccp/pkg/adapters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/mccp/pkg/clusters"
)

func clustersDeleteCmd(client *resty.Client) *cobra.Command {
	var cmd = &cobra.Command{
		Use:           "delete",
		Short:         "Delete CAPI cluster",
		Example:       "mccp clusters delete <cluster-name>",
		RunE:          getClustersDeleteCmdRun(client),
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&clustersDeleteCmdFlags.RepositoryURL, "pr-repo", "", "The repository to open a pull request against")
	cmd.PersistentFlags().StringVar(&clustersDeleteCmdFlags.BaseBranch, "pr-base", "", "The base branch to open the pull request against")
	cmd.PersistentFlags().StringVar(&clustersDeleteCmdFlags.HeadBranch, "pr-branch", "", "The branch to create the pull request from")
	cmd.PersistentFlags().StringVar(&clustersDeleteCmdFlags.Title, "pr-title", "", "The title of the pull request")
	cmd.PersistentFlags().StringVar(&clustersDeleteCmdFlags.Description, "pr-description", "", "The description of the pull request")
	cmd.PersistentFlags().StringVar(&clustersDeleteCmdFlags.CommitMessage, "pr-commit-message", "", "The commit message to use when deleting the clusters")

	return cmd
}

type clustersDeleteFlags struct {
	RepositoryURL string
	BaseBranch    string
	HeadBranch    string
	Title         string
	Description   string
	ClustersNames string
	CommitMessage string
}

var clustersDeleteCmdFlags clustersDeleteFlags

func getClustersDeleteCmdRun(client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		r, err := adapters.NewHttpClient(endpoint, client, os.Stdout)
		if err != nil {
			return err
		}

		return clusters.DeleteClusters(clusters.DeleteClustersParams{
			RepositoryURL: clustersDeleteCmdFlags.RepositoryURL,
			HeadBranch:    clustersDeleteCmdFlags.HeadBranch,
			BaseBranch:    clustersDeleteCmdFlags.BaseBranch,
			Title:         clustersDeleteCmdFlags.Title,
			Description:   clustersDeleteCmdFlags.Description,
			ClustersNames: args,
			CommitMessage: clustersDeleteCmdFlags.CommitMessage,
		}, r, os.Stdout)
	}
}
