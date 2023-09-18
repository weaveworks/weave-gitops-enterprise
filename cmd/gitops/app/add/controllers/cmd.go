package controllers

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/add/controllers/profiles"
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
	cmd.AddCommand(profiles.PolicyAgentCommand(opts))
	cmd.AddCommand(profiles.CapiCommand(opts))
	cmd.AddCommand(profiles.TerraformCommand(opts))
	return cmd
}
