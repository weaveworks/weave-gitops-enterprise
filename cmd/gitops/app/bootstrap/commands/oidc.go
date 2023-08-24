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
	OIDC_INSTALL_MSG       = "Do you want to add OIDC config to your cluster"
	OIDC_DISCOVERY_URL_MSG = "Please enter OIDC Discovery URL (example: https://example-idp.com/.well-known/openid-configuration)"
	OIDC_CLIENT_ID_MSG     = "Please enter OIDC clientID"
	CLIENT_SECRET_MSG      = "Please enter OIDC clientSecret"
	OIDC_SECRET_NAME       = "oidc-auth"
	ADMIN_USER_REVERT_MSG  = "Do you want to revert the admin user, this will delete the admin user and OIDC will be the only way to login"
)

// getOIDCSecrets ask the user for the OIDC configuraions
func getOIDCSecrets(userDomain string) (domain.OIDCConfig, error) {

	configs := domain.OIDCConfig{}

	oidcDiscoveryURL, err := utils.GetStringInput(OIDC_DISCOVERY_URL_MSG, "")
	if err != nil {
		return configs, err
	}

	utils.Info("Verifying OIDC Discovery URL ...")
	oidcIssuerURL, err := getIssuer(oidcDiscoveryURL)
	if err != nil {
		return configs, err
	}

	oidcClientID, err := utils.GetStringInput(OIDC_CLIENT_ID_MSG, "")
	if err != nil {
		return configs, err
	}

	oidcClientSecret, err := utils.GetStringInput(CLIENT_SECRET_MSG, "")
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

	oidcConfigPrompt, err := utils.GetConfirmInput(OIDC_INSTALL_MSG)
	if err != nil {
		return err
	}

	if oidcConfigPrompt != "y" {
		return nil
	}

	utils.Info("For more information about the OIDC config please refer to https://docs.gitops.weave.works/docs/next/configuration/oidc-access/#configuration")

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

	err = utils.CreateSecret(OIDC_SECRET_NAME, WGE_DEFAULT_NAMESPACE, oidcSecretData)
	if err != nil {
		return err
	}

	values := constructOIDCValues(oidcConfig)

	utils.Warning("Installing OIDC config ...")

	err = UpdateHelmReleaseValues(domain.OIDC_VALUES_NAME, values)
	if err != nil {
		return err
	}

	utils.Info("OIDC config created successfully")

	// Ask the user if he wants to revert the admin user
	err = checkAdminPasswordRevert()
	if err != nil {
		return err
	}

	return nil
}

func checkAdminPasswordRevert() error {

	adminUserRevert, err := utils.GetConfirmInput(ADMIN_USER_REVERT_MSG)
	if err != nil {
		return err
	}

	if adminUserRevert != "y" {
		return nil
	}

	err = utils.DeleteSecret(ADMIN_SECRET_NAME, WGE_DEFAULT_NAMESPACE)
	if err != nil {
		return err
	}

	utils.Info("Admin user reverted successfully")
	return nil
}

// constructOIDCValues construct the OIDC values
func constructOIDCValues(oidcConfig domain.OIDCConfig) map[string]interface{} {
	values := map[string]interface{}{
		"enabled":                 true,
		"issuerURL":               oidcConfig.IssuerURL,
		"redirectURL":             oidcConfig.RedirectURL,
		"clientCredentialsSecret": OIDC_SECRET_NAME,
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
		return "", fmt.Errorf("OIDC discovery URL returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	issuer, ok := result["issuer"].(string)
	if !ok || issuer == "" {
		return "", fmt.Errorf("OIDC discovery URL did not return an issuer")
	}

	return issuer, nil
}
