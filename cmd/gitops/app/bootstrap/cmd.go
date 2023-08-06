package create

import (
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstrap Weave-gitops-enterprise",
		Example: `
# Bootstrap weave gitops enterprise
gitops bootstrap`,
	}

	return cmd
}
