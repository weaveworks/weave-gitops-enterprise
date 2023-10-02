package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
)

const (
	oidcInstallMsg    = "Do you want to setup OIDC to access Weave GitOps Dashboards"
	oidcConfigInfoMsg = "Setting up OIDC require configurations provided by your OIDC provider. To learn more about these OIDC configurations, checkout https://docs.gitops.weave.works/docs/next/configuration/oidc-access/#configuration"

	oidcDiscoverUrlMsg    = "Please enter OIDC Discovery URL (example: https://example-idp.com/.well-known/openid-configuration)"
	discoveryUrlVerifyMsg = "Verifying OIDC discovery URL"

	discoveryUrlErrorMsgFormat = "error: OIDC discovery URL returned status %d"
	discoveryUrlNoIssuerMsg    = "error: OIDC discovery URL returned no issuer"

	oidcClientIDMsg     = "Please enter OIDC clientID"
	oidcClientSecretMsg = "Please enter OIDC clientSecret"

	oidcInstallInfoMsg  = "Configuring OIDC"
	oidcConfirmationMsg = "OIDC has been configured successfully!"

	oidcConfigExistWarningMsgFormat = "OIDC is already configured on the cluster. To reset configurations please remove secret '%s' in namespace '%s' and run 'bootstrap auth --type=oidc' command again."

	adminUserRevertMsg     = "Do you want to revoke admin user login, and only use OIDC for dashboard access"
	adminUsernameRevertMsg = "Admin user login has been revoked!"
)

const (
	oidcSecretName = "oidc-auth"
)

func (c *Config) CreateOIDCPrompt(params AuthConfigParams) error {
	oidcConfigPrompt := utils.GetConfirmInput(oidcInstallMsg)
	if oidcConfigPrompt != confirmYes {
		return nil
	}
	return c.CreateOIDCConfig(params)

}

// CreateOIDCConfig creates OIDC secrets on the cluster and updates the OIDC values in the helm release.
// If the OIDC configs already exist, we will ask the user to delete the secret and run the command again.
func (c *Config) CreateOIDCConfig(params AuthConfigParams) error {
	c.Logger.Actionf(oidcConfigInfoMsg)
	if secret, err := utils.GetSecret(c.KubernetesClient, oidcSecretName, WGEDefaultNamespace); secret != nil && err == nil {
		c.Logger.Warningf(oidcConfigExistWarningMsgFormat, oidcSecretName, WGEDefaultNamespace)
		return nil
	} else if err != nil && !strings.Contains(err.Error(), "not found") {
		return err
	}

	oidcConfig, err := c.getOIDCSecrets(params)
	if err != nil {
		return err
	}

	oidcSecretData := map[string][]byte{
		"issuerURL":    []byte(oidcConfig.IssuerURL),
		"clientID":     []byte(oidcConfig.ClientID),
		"clientSecret": []byte(oidcConfig.ClientSecret),
		"redirectURL":  []byte(oidcConfig.RedirectURL),
	}

	if err = utils.CreateSecret(c.KubernetesClient, oidcSecretName, WGEDefaultNamespace, oidcSecretData); err != nil {
		return err
	}

	values := constructOIDCValues(oidcConfig)

	c.Logger.Waitingf(oidcInstallInfoMsg)

	if err := updateHelmReleaseValues(c.KubernetesClient, oidcValuesName, values); err != nil {
		return err
	}
	c.Logger.Successf(oidcConfirmationMsg)

	return nil
}

func (c *Config) CheckAdminPasswordRevert() error {
	adminUserRevert := utils.GetConfirmInput(adminUserRevertMsg)

	if adminUserRevert != "y" {
		return nil
	}

	if err := utils.DeleteSecret(c.KubernetesClient, adminSecretName, WGEDefaultNamespace); err != nil {
		return err
	}

	c.Logger.Successf(adminUsernameRevertMsg)
	return nil
}

// constructOIDCValues construct the OIDC values
func constructOIDCValues(oidcConfig OIDCConfig) map[string]interface{} {
	values := map[string]interface{}{
		"enabled":                 true,
		"issuerURL":               oidcConfig.IssuerURL,
		"redirectURL":             oidcConfig.RedirectURL,
		"clientCredentialsSecret": oidcSecretName,
	}

	return values
}

// getOIDCSecrets gets the OIDC config from the user if not provided
func (c *Config) getOIDCSecrets(inputs AuthConfigParams) (OIDCConfig, error) {
	configs := OIDCConfig{}

	var oidcDiscoveryURL string
	var oidcIssuerURL string
	var err error

	// If the user didn't provide a discovery URL, ask for it
	if inputs.DiscoveryURL == "" {
		// Keep asking for the discovery URL until we get a valid one
		for {
			// Ask for discovery URL from the user
			oidcDiscoveryURL, err = utils.GetStringInput(oidcDiscoverUrlMsg, "")
			if err != nil {
				return configs, err
			}

			c.Logger.Waitingf(discoveryUrlVerifyMsg)

			// Try to get the issuer
			oidcIssuerURL, err = getIssuer(oidcDiscoveryURL)
			if err != nil {
				c.Logger.Failuref("An error occurred: %s. Please enter the discovery URL again.", err.Error())
				continue // Go to the next iteration to re-ask for the URL
			}

			// If we reach this point, it means that the URL is valid. Break out of the loop.
			break
		}
	} else {
		oidcDiscoveryURL = inputs.DiscoveryURL
		oidcIssuerURL, err = getIssuer(oidcDiscoveryURL)
		// If the discovery URL is invalid, return the error
		if err != nil {
			return configs, err
		}
	}

	var oidcClientID string
	var oidcClientSecret string

	// If the user didn't provide a client ID or secret, ask for them
	if inputs.ClientID == "" {
		oidcClientID, err = utils.GetStringInput(oidcClientIDMsg, "")
		if err != nil {
			return configs, err
		}
	} else {
		oidcClientID = inputs.ClientID
	}

	// If the user didn't provide a client ID or secret, ask for them
	if inputs.ClientSecret == "" {
		oidcClientSecret, err = utils.GetPasswordInput(oidcClientSecretMsg)
		if err != nil {
			return configs, err
		}
	} else {
		oidcClientSecret = inputs.ClientSecret
	}

	oidcConfig := OIDCConfig{
		IssuerURL:    oidcIssuerURL,
		ClientID:     oidcClientID,
		ClientSecret: oidcClientSecret,
	}

	if strings.Contains(inputs.UserDomain, domainTypelocalhost) {
		oidcConfig.RedirectURL = "http://localhost:8000/oauth2/callback"
	} else {
		oidcConfig.RedirectURL = fmt.Sprintf("https://%s/oauth2/callback", inputs.UserDomain)
	}

	return oidcConfig, nil
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
		return "", errors.New(discoveryUrlNoIssuerMsg)
	}

	return issuer, nil
}
