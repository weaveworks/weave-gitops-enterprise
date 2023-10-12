package connect

import (
	"github.com/spf13/cobra"
	connect "github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/connect/clusters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/app"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

func Command(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect clusters",
		Example: `
# Connect a cluster
gitops connect cluster`,
		PreRunE: app.DisinheritAPIFlags,
	}

	cmd.AddCommand(connect.ConnectCommand(opts))

	return cmd
}
