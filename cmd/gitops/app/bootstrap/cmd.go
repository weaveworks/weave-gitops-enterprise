package bootstrap

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/checks"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstraps Weave gitops enterprise",
		Example: `
# Bootstrap Weave-gitops-enterprise
gitops bootstrap`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Bootstrap()
		},
	}

	return cmd
}

func Bootstrap() error {
	checks.CheckEntitlementFile()
	checks.CheckFluxIsInstalled()
	checks.CheckFluxReconcile()
	checks.CheckWgeVersion()
	return nil
}
