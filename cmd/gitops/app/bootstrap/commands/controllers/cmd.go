package controllers

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/commands/controllers/controllers"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "controllers",
		Short: "Bootstraps Weave gitops controllers",
		Example: `
# Bootstrap controllers
gitops bootstrap controllers <controller-name>`,
	}
	cmd.AddCommand(controllers.PolicyAgentCommand)
	return cmd
}
