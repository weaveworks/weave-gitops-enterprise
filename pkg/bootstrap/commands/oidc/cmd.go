package commands

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/commands"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

func Command(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "controllers",
		Short: "Add Weave gitops controllers",
		Example: `
# Add controllers
gitops add controllers <controller-name>`,
	}
	cmd.AddCommand(commands.CreateOIDCConfig()
	return cmd
}
