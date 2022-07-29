package get

import (
	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/cli/app/get/clusters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/cli/app/get/credentials"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/cli/app/get/profiles"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/cli/app/get/templates"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/cli/app/get/templates/terraform"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/cli/pkg/adapters"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get/bcrypt"
)

func GetCommand(opts *config.Options, client *adapters.HTTPClient) *cobra.Command {
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

	templateCommand := templates.TemplateCommand(opts, client)
	terraformCommand := terraform.TerraformCommand(opts, client)
	templateCommand.AddCommand(terraformCommand)

	cmd.AddCommand(templateCommand)
	cmd.AddCommand(credentials.CredentialCommand(opts, client))
	cmd.AddCommand(clusters.ClusterCommand(opts, client))
	cmd.AddCommand(profiles.ProfilesCommand(opts, client))
	cmd.AddCommand(bcrypt.HashCommand(opts))

	return cmd
}
