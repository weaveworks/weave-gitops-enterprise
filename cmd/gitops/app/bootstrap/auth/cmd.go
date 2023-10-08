package auth

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/commands"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
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
		Short: "Generate authentication configuration for Weave GitOps. You can specify the type of authentication using the '--type' flag. Currently, only OIDC is supported.",
		Example: `
## Add OIDC configuration to your cluster. 
gitops bootstrap auth --type=oidc
	`,
		Run: getAuthCmdRun(opts),
	}
	cmd.Flags().StringVarP(&params.Type, "type", "t", "", "Type of authentication")
	cmd.Flags().StringVarP(&params.DiscoveryURL, "discovery-url", "du", "", "OIDC Discovery URL (optional)")
	cmd.Flags().StringVarP(&params.ClientID, "client-id", "ci", "", "OIDC Client ID (optional)")
	cmd.Flags().StringVarP(&params.ClientSecret, "client-secret", "cs", "", "OIDC Client Secret (optional)")
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
	kubernetesClient, err := utils.GetKubernetesClient(opts.Kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to get kubernetes client. error: %s", err)
	}

	installedVersion, err := utils.GetHelmReleaseProperty(kubernetesClient, commands.WGEHelmReleaseName, commands.WGEDefaultNamespace, commands.HelmVersionProperty)
	if err != nil {
		return err
	}

	oidcDomain, err := utils.GetHelmReleaseProperty(kubernetesClient, commands.WGEHelmReleaseName, commands.WGEDefaultNamespace, helmDomainProperty)
	if err != nil {
		return err
	}

	config := commands.Config{}
	config.KubernetesClient = kubernetesClient
	config.Logger = logger

	authParams := commands.AuthConfigParams{
		UserDomain:   oidcDomain,
		WGEVersion:   installedVersion,
		DiscoveryURL: params.DiscoveryURL,
		ClientID:     params.ClientID,
		ClientSecret: params.ClientSecret,
	}

	switch params.Type {
	case authOIDC:
		return config.CreateOIDCConfig(authParams)
	default:
		return fmt.Errorf("unsupported authentication type: %s", params.Type)
	}

}
