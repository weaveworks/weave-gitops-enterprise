package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

const (
	//TODO: make sure we skip the following message if we are coming from the oidc cmd directly.
	oidcInstallMsg = "Do you want to setup OIDC to access Weave GitOps Dashboards?"
	//TODO: review the URL after updating the docs.
	oidcConfigInfoMsg = "Setting up OIDC require configurations provided by your OIDC provider. To learn more about these OIDC configurations, checkout https://docs.gitops.weave.works/docs/next/configuration/oidc-access/#configuration"

	oidcDiscoverUrlMsg    = "Please enter OIDC Discovery URL (example: https://example-idp.com/.well-known/openid-configuration)"
	discoveryUrlVerifyMsg = "Verifying OIDC discovery URL ..."
	//TODO: mmegahid - clarify that this is a failed
	discoveryUrlErrorMsgFormat = "error: OIDC discovery URL returned status %d"
	discoveryUrlNoIssuerMsg    = "error: OIDC discovery URL returned no issuer"
	//TODO: prompt the user to enter the URL again, using the oidcDiscoverUrlMsg.

	oidcClientIDMsg     = "Please enter OIDC clientID"
	oidcClientSecretMsg = "Please enter OIDC clientSecret"

	oidcInstallInfoMsg  = "Configuring OIDC ..."
	oidcConfirmationMsg = "OIDC has been configured successfully!"

	//TODO: replace (cmd) with the command to run again
	oidcConfigExistWarningMsgFormat = "OIDC is already configured on the cluster. To reset configurations please remove secret '%s' in namespace '%s' and run 'cmd' command again."

	adminUserRevertMsg     = "Do you want to revoke admin user login, and only use OIDC for dashboard access?"
	adminUsernameRevertMsg = "Admin user login has been revoked!"
)

const (
	oidcSecretName = "oidc-auth"
)

// getOIDCSecrets ask the user for the OIDC configuraions.
func getOIDCSecrets(userDomain string) (domain.OIDCConfig, error) {
	configs := domain.OIDCConfig{}

	oidcDiscoveryURL, err := utils.GetStringInput(oidcDiscoverUrlMsg, "")
	if err != nil {
		return configs, err
	}

	utils.Info(discoveryUrlVerifyMsg)
	oidcIssuerURL, err := getIssuer(oidcDiscoveryURL)
	if err != nil {
		return configs, err
	}

	oidcClientID, err := utils.GetStringInput(oidcClientIDMsg, "")
	if err != nil {
		return configs, err
	}

	oidcClientSecret, err := utils.GetStringInput(oidcClientSecretMsg, "")
	if err != nil {
		return configs, err
	}

	oidcConfig := domain.OIDCConfig{
		IssuerURL:    oidcIssuerURL,
		ClientID:     oidcClientID,
		ClientSecret: oidcClientSecret,
	}

	if strings.Contains(userDomain, domainTypelocalhost) {
		oidcConfig.RedirectURL = "http://localhost:8000/oauth2/callback"
	} else {
		oidcConfig.RedirectURL = fmt.Sprintf("https://%s/oauth2/callback", userDomain)
	}

	return oidcConfig, nil
}

// CreateOIDCConfig creates OIDC config for the cluster to be used for authentication
func CreateOIDCConfig(opts config.Options, userDomain string, version string) error {
	oidcConfigPrompt := utils.GetConfirmInput(oidcInstallMsg)

	if oidcConfigPrompt != "y" {
		return nil
	}

	utils.Info(oidcConfigInfoMsg)
	kubernetesClient, err := utils.GetKubernetesClient(opts.Kubeconfig)
	if err != nil {
		return err
	}
	if _, err := utils.GetSecret(oidcSecretName, wgeDefaultNamespace, kubernetesClient); err == nil {
		utils.Info(oidcConfigExistWarningMsgFormat, oidcSecretName, wgeDefaultNamespace)
		return nil
	} else if err != nil && !strings.Contains(err.Error(), "not found") {
		return err
	}

	oidcConfig, err := getOIDCSecrets(userDomain)
	if err != nil {
		return err
	}

	oidcSecretData := map[string][]byte{
		"issuerURL":    []byte(oidcConfig.IssuerURL),
		"clientID":     []byte(oidcConfig.ClientID),
		"clientSecret": []byte(oidcConfig.ClientSecret),
		"redirectURL":  []byte(oidcConfig.RedirectURL),
	}

	if err = utils.CreateSecret(oidcSecretName, wgeDefaultNamespace, oidcSecretData, kubernetesClient); err != nil {
		return err
	}

	values := constructOIDCValues(oidcConfig)

	utils.Warning(oidcInstallInfoMsg)

	if err := UpdateHelmReleaseValues(domain.OIDCValuesName, values); err != nil {
		return err
	}

	utils.Info(oidcConfirmationMsg)

	// Ask the user if he wants to revert the admin user
	if err := checkAdminPasswordRevert(opts); err != nil {
		return err
	}

	return nil
}

func checkAdminPasswordRevert(opts config.Options) error {
	adminUserRevert := utils.GetConfirmInput(adminUserRevertMsg)

	if adminUserRevert != "y" {
		return nil
	}
	kubernetesClient, err := utils.GetKubernetesClient(opts.Kubeconfig)
	if err != nil {
		return err
	}
	if err := utils.DeleteSecret(adminSecretName, wgeDefaultNamespace, kubernetesClient); err != nil {
		return err
	}

	utils.Info(adminUsernameRevertMsg)
	return nil
}

// constructOIDCValues construct the OIDC values
func constructOIDCValues(oidcConfig domain.OIDCConfig) map[string]interface{} {
	values := map[string]interface{}{
		"enabled":                 true,
		"issuerURL":               oidcConfig.IssuerURL,
		"redirectURL":             oidcConfig.RedirectURL,
		"clientCredentialsSecret": oidcSecretName,
	}

	return values
}

func getIssuer(oidcDiscoveryURL string) (string, error) {
	resp, err := http.Get(oidcDiscoveryURL)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(discoveryUrlErrorMsgFormat, resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	issuer, ok := result["issuer"].(string)
	if !ok || issuer == "" {
		return "", fmt.Errorf(discoveryUrlNoIssuerMsg)
	}

	return issuer, nil
}
