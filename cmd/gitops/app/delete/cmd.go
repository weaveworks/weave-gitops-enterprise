package delete

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/delete/clusters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

func DeleteCommand(opts *config.Options, client *adapters.HTTPClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete one or many Weave GitOps resources",
		Example: `
# Delete a CAPI cluster given its name
gitops delete cluster <cluster-name>`,
	}

	cmd.AddCommand(clusters.ClusterCommand(opts, client))

	return cmd
}
