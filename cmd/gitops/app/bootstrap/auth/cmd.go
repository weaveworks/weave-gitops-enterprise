package auth

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/commands"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

const (
	authOIDC           = "oidc"
	helmDomainProperty = "domain"
)

var params commands.AuthConfigParams

func Command(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Add authentication to your cluster. You can specify the type of authentication using the '--type' flag. Currently, only OIDC is supported.",
		Example: `
## Add OIDC configuration to your cluster. 
gitops bootstrap auth --type=oidc
	`,
		Run: getAuthCmdRun(opts),
	}
	cmd.Flags().StringVar(&params.Type, "type", "oidc", "Type of authentication")
	cmd.Flags().StringVar(&params.DiscoveryURL, "discovery-url", "", "OIDC Discovery URL (optional)")
	cmd.Flags().StringVar(&params.ClientID, "client-id", "", "OIDC Client ID (optional)")
	cmd.Flags().StringVar(&params.ClientSecret, "client-secret", "", "OIDC Client Secret (optional)")
	return cmd
}

func getAuthCmdRun(opts *config.Options) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		logger := logger.NewCLILogger(os.Stdout)

		if err := createAuthCommand(opts, logger); err != nil {
			logger.Failuref(err.Error())
		}
	}
}

func createAuthCommand(opts *config.Options, logger logger.Logger) error {
	// // get userDomain from helm release
	// oidcDomain, err := utils.GetHelmReleaseProperty(kubernetesClient, commands.WGEHelmReleaseName, commands.WGEDefaultNamespace, helmDomainProperty)
	// if err != nil {
	// 	return nil, err
	// }

	// // get current version of WGE
	// wgeVersion, err := utils.GetHelmReleaseProperty(kubernetesClient, commands.WGEHelmReleaseName, commands.WGEDefaultNamespace, commands.HelmVersionProperty)
	// if err != nil {
	// 	return nil, err
	// }

	// var authType string

	// // Add the auth sub-command to the bootstrap command
	// authCmd := &cobra.Command{
	// 	Use:   cmdAuthName,
	// 	Short: cmdAuthShortDescription,
	// 	RunE: func(cmd *cobra.Command, args []string) error {

	// 		switch authType {
	// 		case authOIDC:
	// 			return commands.CreateOIDCConfig()
	// 		default:
	// 			return fmt.Errorf("Unsupported authentication type: %s", authType)
	// 		}
	// 	},
	// }
	return nil
}
