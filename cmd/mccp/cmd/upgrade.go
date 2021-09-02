package cmd

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/mccp/pkg/upgrade"
)

func upgradeCmd(client *resty.Client) *cobra.Command {
	var cmd = &cobra.Command{
		Use:           "upgrade",
		Short:         "Upgrade to WGE",
		Example:       `mccp upgrade`,
		RunE:          upgradeCmdRun(client),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.RepositoryURL, "pr-repo", "", "The repository to open a pull request against")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.Remote, "pr-remote", "origin", "The remote to push the branch to")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.BaseBranch, "pr-base", "main", "The base branch to open the pull request against")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.HeadBranch, "pr-branch", "", "The branch to create the pull request from")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.Title, "pr-title", "Upgrade to WGE", "The title of the pull request")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.Description, "pr-description", "PR to upgrade to WGE", "The description of the pull request")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.CommitMessage, "pr-commit-message", "Upgrade to WGE", "The commit message to use when adding the CAPI template")

	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.Name, "name", "", "The commit message to use when adding the CAPI template")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.Namespace, "namespace", "default", "The namespace to use for generating resources")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.ProfileBranch, "profile-branch", "main", "The branch to use on the repository in which the profile is")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.ConfigMap, "configmap", "", "The name of the ConfigMap which contains values for this profile")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.Out, "out", ".", "Optional location to create the profile installation folder in. This should be relative to the current working directory")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.ProfileRepoURL, "profile-repo-url", "", "The URL of the repository that contains the profile to be added")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.ProfilePath, "profile-path", "", "The path to a profile when url is provided")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.GitRepository, "git-repository", "", "The namespace and name of the GitRepository object governing the flux repo")

	return cmd
}

type upgradeFlags struct {
	RepositoryURL  string
	Remote         string
	BaseBranch     string
	HeadBranch     string
	Title          string
	Description    string
	CommitMessage  string
	Name           string
	Namespace      string
	ProfileBranch  string
	ConfigMap      string
	Out            string
	ProfileRepoURL string
	ProfilePath    string
	GitRepository  string
}

var upgradeCmdFlags upgradeFlags

func upgradeCmdRun(client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return upgrade.Upgrade(upgrade.UpgradeParams{
			RepositoryURL:  upgradeCmdFlags.RepositoryURL,
			Remote:         upgradeCmdFlags.Remote,
			HeadBranch:     upgradeCmdFlags.HeadBranch,
			BaseBranch:     upgradeCmdFlags.BaseBranch,
			Title:          upgradeCmdFlags.Title,
			Description:    upgradeCmdFlags.Description,
			CommitMessage:  upgradeCmdFlags.CommitMessage,
			Name:           upgradeCmdFlags.Name,
			Namespace:      upgradeCmdFlags.Namespace,
			ProfileBranch:  upgradeCmdFlags.ProfileBranch,
			ConfigMap:      upgradeCmdFlags.ConfigMap,
			Out:            upgradeCmdFlags.Out,
			ProfileRepoURL: upgradeCmdFlags.ProfileRepoURL,
			ProfilePath:    upgradeCmdFlags.ProfilePath,
			GitRepository:  upgradeCmdFlags.GitRepository,
		})
	}
}
