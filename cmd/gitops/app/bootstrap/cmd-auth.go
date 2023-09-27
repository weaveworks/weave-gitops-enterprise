package bootstrap

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/commands"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

const (
	cmdAuthName             = "auth"
	cmdAuthShortDescription = "Add authentication to your cluster"
	authOIDC                = "oidc"

	helmDomainProperty = "domain"
)

func createAuthCommand(opts *config.Options) (*cobra.Command, error) {

	var params domain.OIDCConfigParams

	// get kubernetes client
	kubernetesClient, err := utils.GetKubernetesClient(opts.Kubeconfig)
	if err != nil {
		return nil, err
	}

	// get userDomain from helm release
	params.UserDomain, err = utils.GetHelmReleaseProperty(kubernetesClient, commands.WGEHelmReleaseName, commands.WGEDefaultNamespace, helmDomainProperty)
	if err != nil {
		return nil, err
	}

	// get current version of WGE
	params.WGEVersion, err = utils.GetHelmReleaseProperty(kubernetesClient, commands.WGEHelmReleaseName, commands.WGEDefaultNamespace, commands.HelmVersionProperty)
	if err != nil {
		return nil, err
	}

	var authType string

	// Add the auth sub-command to the bootstrap command
	authCmd := &cobra.Command{
		Use:   cmdAuthName,
		Short: cmdAuthShortDescription,
		RunE: func(cmd *cobra.Command, args []string) error {

			switch authType {
			case authOIDC:
				return commands.CreateOIDCConfig(kubernetesClient, params)
			default:
				return fmt.Errorf("Unsupported authentication type: %s", authType)
			}
		},
	}

	authCmd.Flags().StringVar(&authType, "type", "oidc", "Type of authentication")
	authCmd.Flags().StringVar(&params.DiscoveryURL, "discovery-url", "", "OIDC Discovery URL (optional)")
	authCmd.Flags().StringVar(&params.ClientID, "client-id", "", "OIDC Client ID (optional)")
	authCmd.Flags().StringVar(&params.ClientSecret, "client-secret", "", "OIDC Client Secret (optional)")

	return authCmd, nil
}
