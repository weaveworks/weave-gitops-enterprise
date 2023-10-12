package disconnect

import (
	"github.com/spf13/cobra"
	disconnect "github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/disconnect/clusters"

	// "github.com/weaveworks/weave-gitops/cmd/gitops/app"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/app"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

// Command returns the command for disconnect
func Command(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disconnect",
		Short: "Disconnect clusters",
		Example: `
# Disconnect a cluster
gitops disconnect cluster`,
		PreRunE: app.DisinheritAPIFlags,
	}

	cmd.AddCommand(disconnect.DisconnectCommand(opts))

	return cmd
}
