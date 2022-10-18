package add

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/add/clusters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/add/profiles"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/add/terraform"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

func Command(opts *config.Options, client *adapters.HTTPClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new Weave GitOps resource",
		Example: `
# Add a new cluster using a CAPI template
gitops add cluster`,
	}

	cmd.AddCommand(clusters.AddCommand(opts, client))
	cmd.AddCommand(profiles.AddCommand(opts, client))
	cmd.AddCommand(terraform.AddCommand(opts, client))

	return cmd
}
