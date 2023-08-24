package commands

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

const (
	OIDC_INSTALL_MSG           = "Do you want to add OIDC config to your cluster"
	OIDC_DISCOVERY_URL_MSG     = "Please enter OIDC Discovery URL (example: https://example-idp.com/.well-known/openid-configuration)"
	OIDC_CLIENT_ID_MSG         = "Please enter OIDC clientID"
	CLIENT_SECRET_MSG          = "Please enter OIDC clientSecret"
	OIDC_VALUES_FILES_LOCATION = "/tmp/agent-values.yaml"
	OIDC_SECRET_NAME           = "oidc-auth"
	OIDC_SECRET_NAMESPACE      = "flux-system"
	ADMIN_USER_REVERT_MSG      = "Do you want to revert the admin user, this will delete the admin user and OIDC will be the only way to login? (y/n)"
)

// getOIDCSecrets ask the user for the OIDC configuraions
func getOIDCSecrets(isExternalDomain bool, userDomain string) (domain.OIDCConfig, error) {

	configs := domain.OIDCConfig{}

	oidcDiscoveryURL, err := utils.GetStringInput(OIDC_DISCOVERY_URL_MSG, "")
	if err != nil {
		return configs, utils.CheckIfError(err)
	}

	utils.Info("Verifying OIDC Discovery URL ...")
	oidcIssuerURL, err := getIssuer(oidcDiscoveryURL)
	if err != nil {
		return configs, utils.CheckIfError(err)
	}

	oidcClientID, err := utils.GetStringInput(OIDC_CLIENT_ID_MSG, "")
	if err != nil {
		return configs, utils.CheckIfError(err)
	}

	oidcClientSecret, err := utils.GetStringInput(CLIENT_SECRET_MSG, "")
	if err != nil {
		return configs, utils.CheckIfError(err)
	}

	oidcConfig := domain.OIDCConfig{
		IssuerURL:    oidcIssuerURL,
		ClientID:     oidcClientID,
		ClientSecret: oidcClientSecret,
	}

	if isExternalDomain {
		oidcConfig.RedirectURL = fmt.Sprintf("https://%s/oauth2/callback", userDomain)
	} else {
		oidcConfig.RedirectURL = fmt.Sprintf("http://localhost:8000/oauth2/callback")
	}

	return oidcConfig, nil
}

func CreateOIDCConfig(isExternalDomain bool, userDomain string, version string) error {

	oidcConfigPrompt, err := utils.GetConfirmInput(OIDC_INSTALL_MSG)
	if err != nil {
		return utils.CheckIfError(err)
	}

	if oidcConfigPrompt != "y" {
		return nil
	}

	utils.Info("For more information about the OIDC config please refer to https://docs.gitops.weave.works/docs/next/configuration/oidc-access/#configuration")

	oidcConfig, err := getOIDCSecrets(isExternalDomain, userDomain)

	oidcSecretData := map[string][]byte{
		"issuerURL":    []byte(oidcConfig.IssuerURL),
		"clientID":     []byte(oidcConfig.ClientID),
		"clientSecret": []byte(oidcConfig.ClientSecret),
		"redirectURL":  []byte(oidcConfig.RedirectURL),
	}

	err = utils.CreateSecret(OIDC_SECRET_NAME, OIDC_SECRET_NAMESPACE, oidcSecretData)
	if err != nil {
		return utils.CheckIfError(err)
	}

	values := constructOIDCValues(oidcConfig)

	utils.Warning("Installing OIDC config ...")

	err = InstallController(domain.OIDC_VALUES_NAME, values)
	if err != nil {
		return utils.CheckIfError(err)
	}

	utils.Info("✔ OIDC config created successfully")

	// Ask the user if he wants to revert the admin user
	err = checkAdminPasswordRevet()
	if err != nil {
		return utils.CheckIfError(err)
	}

	return nil
}

func checkAdminPasswordRevet() error {

	adminUserRevet, err := utils.GetStringInput(ADMIN_USER_REVERT_MSG, "")
	if err != nil {
		return utils.CheckIfError(err)
	}

	if adminUserRevet != "y" {
		return nil
	}

	err = utils.DeleteSecret(ADMIN_SECRET_NAME, WGE_DEFAULT_NAMESPACE)
	if err != nil {
		return utils.CheckIfError(err)
	}

	fmt.Println("✔ Admin user reverted successfully")
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
		return "", fmt.Errorf("failed to fetch OIDC configuration: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode OIDC configuration: %v", err)
	}

	issuer, ok := result["issuer"].(string)
	if !ok || issuer == "" {
		return "", fmt.Errorf("issuer not found or not a string")
	}

	return issuer, nil
}
