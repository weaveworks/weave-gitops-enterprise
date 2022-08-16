package get

import (
	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/get/clusters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/get/credentials"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/get/profiles"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/get/templates"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/get/templates/terraform"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get/bcrypt"
)

func Command(opts *config.Options, client *adapters.HTTPClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Display one or many Weave GitOps resources",
		Example: `
# Get all CAPI templates
gitops get templates

# Get all CAPI credentials
gitops get credentials

# Get all CAPI clusters
gitops get clusters`,
	}

	templateCommand := templates.GetCommand(opts, client)
	terraformCommand := terraform.GetCommand(opts, client)
	templateCommand.AddCommand(terraformCommand)

	cmd.AddCommand(templateCommand)
	cmd.AddCommand(credentials.GetCommand(opts, client))
	cmd.AddCommand(clusters.GetCommand(opts, client))
	cmd.AddCommand(profiles.GetCommand(opts, client))
	cmd.AddCommand(bcrypt.HashCommand(opts))

	return cmd
}
