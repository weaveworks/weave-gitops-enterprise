package controllers

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/controllers/profiles"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "controllers",
		Short: "Bootstraps Weave gitops controllers",
		Example: `
# Bootstrap controllers
gitops bootstrap controllers <controller-name>`,
	}
	cmd.AddCommand(profiles.PolicyAgentCommand)
	cmd.AddCommand(profiles.CapiCommand)
	cmd.AddCommand(profiles.TerraformCommand)
	return cmd
}
