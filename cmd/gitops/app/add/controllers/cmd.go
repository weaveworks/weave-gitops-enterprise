package controllers

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/controllers/profiles"
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
	cmd.AddCommand(profiles.PolicyAgentCommand)
	cmd.AddCommand(profiles.CapiCommand)
	cmd.AddCommand(profiles.TerraformCommand)
	return cmd
}
