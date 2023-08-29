package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

const (
	oidcInstallMsg                  = "Do you want to add OIDC config to your cluster"
	oidcDiscoverUrlMsg              = "Please enter OIDC Discovery URL (example: https://example-idp.com/.well-known/openid-configuration)"
	discoveryUrlVerifyMsg           = "Verifying OIDC discovery URL ..."
	oidcClientIDMsg                 = "Please enter OIDC clientID"
	oidcClientSecretMsg             = "Please enter OIDC clientSecret"
	adminUserRevertMsg              = "Do you want to revert the admin user, this will delete the admin user and OIDC will be the only way to login"
	oidcConfigInfoMsg               = "For more information about the OIDC config please refer to https://docs.gitops.weave.works/docs/next/configuration/oidc-access/#configuration"
	oidcInstallInfoMsg              = "Installing OIDC config ..."
	oidcConfirmationMsg             = "OIDC config created successfully"
	adminUsernameRevertMsg          = "Admin user reverted successfully"
	oidcSecretName                  = "oidc-auth"
	oidcConfigExistWarningMsgFormat = "OIDC already configured on the cluster, to reset please remove secret '%s' in namespace '%s'"
	discoveryUrlErrorMsgFormat      = "OIDC discovery URL returned status %d"
	discoveryUrlNoIssuerMsg         = "OIDC discovery URL returned no issuer"
)

// getOIDCSecrets ask the user for the OIDC configuraions
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

	if strings.Contains(userDomain, "localhost") {
		oidcConfig.RedirectURL = "http://localhost:8000/oauth2/callback"
	} else {
		oidcConfig.RedirectURL = fmt.Sprintf("https://%s/oauth2/callback", userDomain)
	}

	return oidcConfig, nil
}

// CreateOIDCConfig creates OIDC config for the cluster to be used for authentication
func CreateOIDCConfig(userDomain string, version string) error {

	oidcConfigPrompt := utils.GetConfirmInput(oidcInstallMsg)

	if oidcConfigPrompt != "y" {
		return nil
	}

	utils.Info(oidcConfigInfoMsg)

	if _, err := utils.GetSecret(oidcSecretName, WGEDefaultNamespace); err == nil {
		utils.Info(oidcConfigExistWarningMsgFormat, oidcSecretName, WGEDefaultNamespace)
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

	err = utils.CreateSecret(oidcSecretName, WGEDefaultNamespace, oidcSecretData)
	if err != nil {
		return err
	}

	values := constructOIDCValues(oidcConfig)

	utils.Warning(oidcInstallInfoMsg)

	err = UpdateHelmReleaseValues(domain.OIDCValuesName, values)
	if err != nil {
		return err
	}

	utils.Info(oidcConfirmationMsg)

	// Ask the user if he wants to revert the admin user
	if err := checkAdminPasswordRevert(); err != nil {
		return err
	}

	return nil
}

func checkAdminPasswordRevert() error {

	adminUserRevert := utils.GetConfirmInput(adminUserRevertMsg)

	if adminUserRevert != "y" {
		return nil
	}

	err := utils.DeleteSecret(AdminSecretName, WGEDefaultNamespace)
	if err != nil {
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
