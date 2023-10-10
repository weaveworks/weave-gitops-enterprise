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
	cmdName             = "auth"
	cmdShortDescription = "Generate authentication configuration for Weave GitOps. You can specify the type of authentication using the '--type' flag. Currently, only OIDC is supported."
	cmdExamples         = `
	## Add OIDC configuration to your cluster. 
	gitops bootstrap auth --type=oidc
	`
)
const (
	authOIDC           = "oidc"
	helmDomainProperty = "domain"
)

type authConfigParams struct {
	Type         string
	UserDomain   string
	WGEVersion   string
	DiscoveryURL string
	ClientID     string
	ClientSecret string
}

var params authConfigParams

func Command(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     cmdName,
		Short:   cmdShortDescription,
		Example: cmdExamples,
		Run:     getAuthCmdRun(opts),
	}
	cmd.Flags().StringVarP(&params.Type, "type", "t", "", "Type of authentication")
	cmd.Flags().StringVarP(&params.DiscoveryURL, "discovery-url", "u", "", "OIDC Discovery URL (optional)")
	cmd.Flags().StringVarP(&params.ClientID, "client-id", "c", "", "OIDC Client ID (optional)")
	cmd.Flags().StringVarP(&params.ClientSecret, "client-secret", "s", "", "OIDC Client Secret (optional)")
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

	params.WGEVersion, err = utils.GetHelmReleaseProperty(kubernetesClient, commands.WGEHelmReleaseName, commands.WGEDefaultNamespace, commands.HelmVersionProperty)
	if err != nil {
		return err
	}

	params.UserDomain, err = utils.GetHelmReleaseProperty(kubernetesClient, commands.WGEHelmReleaseName, commands.WGEDefaultNamespace, helmDomainProperty)
	if err != nil {
		return err
	}

	config := commands.Config{}
	config.KubernetesClient = kubernetesClient
	config.Logger = logger

	flagsMap := map[string]string{
		commands.WGEVersion:   params.WGEVersion,
		commands.UserDomain:   params.UserDomain,
		commands.DiscoveryURL: params.DiscoveryURL,
		commands.ClientID:     params.ClientID,
		commands.ClientSecret: params.ClientSecret,
	}
	var steps = []commands.BootstrapStep{
		commands.OIDCConfigStep,
		commands.CheckUIDomainStep,
	}

	for _, step := range steps {
		if err := step.Execute(&config, flagsMap); err != nil {
			return err
		}
	}

	return nil
}
