package commands

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

const (
	OIDC_INSTALL_MSG           = "Do you want to add OIDC config to your cluster"
	ISSUER_URL_MSG             = "Please enter OIDC issuerURL"
	OIDC_CLIENT_ID_MSG         = "Please enter OIDC clientID"
	CLIENT_SECRET_MSG          = "Please enter OIDC clientSecret"
	REDIRECT_URL_MSG           = "Please enter OIDC redirectURL"
	OIDC_VALUES_FILES_LOCATION = "/tmp/agent-values.yaml"
	OIDC_SECRET_NAME           = "oidc-auth"
	OIDC_SECRET_NAMESPACE      = "flux-system"
	OIDC_VALUES_NAME           = "oidc"
	ADMIN_USER_REVERT_MSG      = "Do you want to revert the admin user, this will delete the admin user and OIDC will be the only way to login"
)

// getOIDCSecrets ask the user for the OIDC configuraions
func getOIDCSecrets() (domain.OIDCConfig, error) {

	configs := domain.OIDCConfig{}

	oidcIssuerURL, err := utils.GetStringInput(ISSUER_URL_MSG, "")
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

	oidcRedirectURL, err := utils.GetStringInput(REDIRECT_URL_MSG, "")
	if err != nil {
		return configs, utils.CheckIfError(err)
	}

	return domain.OIDCConfig{
		IssuerURL:    oidcIssuerURL,
		ClientID:     oidcClientID,
		ClientSecret: oidcClientSecret,
		RedirectURL:  oidcRedirectURL,
	}, nil
}

func CreateOIDCConfig(version string) error {

	oidcConfigPrompt, err := utils.GetConfirmInput(OIDC_INSTALL_MSG)
	if err != nil {
		return utils.CheckIfError(err)
	}

	if oidcConfigPrompt != "y" {
		return nil
	}

	utils.Info("For more information about the OIDC config please refer to https://docs.gitops.weave.works/docs/next/configuration/oidc-access/#configuration")

	oidcConfig, err := getOIDCSecrets()

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

	utils.Warning("Installing policy agent ...")
	err = InstallController(OIDC_VALUES_NAME, values)
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

// constructOIDCValues
func constructOIDCValues(oidcConfig domain.OIDCConfig) map[string]interface{} {
	values := map[string]interface{}{
		"config": map[string]interface{}{
			"oidc": map[string]interface{}{
				"enabled":                 true,
				"issuerURL":               oidcConfig.IssuerURL,
				"redirectURL":             oidcConfig.RedirectURL,
				"clientCredentialsSecret": OIDC_SECRET_NAME,
			},
		},
	}

	return values
}
