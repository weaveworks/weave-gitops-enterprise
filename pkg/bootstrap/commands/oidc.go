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

var OIDCConfigStep = BootstrapStep{
	Name: "OIDC config",
	Input: []StepInput{
		{
			Name:            "oidcConfig",
			Type:            confirmInput,
			Msg:             oidcInstallMsg,
			DefaultValue:    "",
			Valuesfn:        checkExistingOIDCConfig,
			StepInformation: fmt.Sprintf(oidcConfigExistWarningMsgFormat, oidcSecretName, WGEDefaultNamespace),
		},
		{
			Name:            DiscoveryURL,
			Type:            stringInput,
			Msg:             oidcDiscoverUrlMsg,
			DefaultValue:    "",
			Valuesfn:        canAskForConfig,
			StepInformation: oidcConfigInfoMsg,
		},
		{
			Name:         ClientID,
			Type:         stringInput,
			Msg:          oidcClientIDMsg,
			DefaultValue: "",
			Valuesfn:     canAskForConfig,
		},
		{
			Name:         ClientSecret,
			Type:         passwordInput,
			Msg:          oidcClientSecretMsg,
			DefaultValue: "",
			Valuesfn:     canAskForConfig,
		},
	},
	Step: createOIDCConfig,
}

// createOIDCConfig creates OIDC secrets on the cluster and updates the OIDC values in the helm release.
// If the OIDC configs already exist, we will ask the user to delete the secret and run the command again.
func createOIDCConfig(input []StepInput, c *Config) ([]StepOutput, error) {

	if existing, _ := checkExistingOIDCConfig(input, c); existing.(bool) {
		return []StepOutput{}, nil
	}

	oidcConfig := OIDCConfig{}
	issuerURLErrCount := 0

	for _, param := range input {
		if param.Name == DiscoveryURL {
			// keep asking until a valid issuer URL is obtained
			for {
				if val, ok := param.Value.(string); ok {
					issuerURL, err := getIssuer(val)
					if err != nil {
						issuerURLErrCount++
						// if we fail to get issuer url after 3 attempts, we will return an error
						if issuerURLErrCount > 3 {
							return []StepOutput{}, errors.New("Failed to retrieve IssuerURL after multiple attempts. Please verify the DiscoveryURL and try again.")
						}
						c.Logger.Warningf("Failed to retrieve IssuerURL. Please verify the DiscoveryURL and try again.")
						// ask for discovery url again
						val, err = utils.GetStringInput(oidcDiscoverUrlMsg, "")
						if err != nil {
							return []StepOutput{}, err
						}
						param.Value = val
						continue
					}
					oidcConfig.IssuerURL = issuerURL
					break
				} else {
					return []StepOutput{}, errors.New("DiscoveryURL not found")
				}
			}
		}

		if param.Name == ClientID {
			if val, ok := param.Value.(string); ok {
				oidcConfig.ClientID = val
			} else {
				return []StepOutput{}, errors.New("ClientID not found")
			}
		}
		if param.Name == ClientSecret {
			if val, ok := param.Value.(string); ok {
				oidcConfig.ClientSecret = val
			} else {
				return []StepOutput{}, errors.New("ClientSecret not found")
			}
		}
	}

	if strings.Contains(c.UserDomain, domainTypelocalhost) {
		oidcConfig.RedirectURL = "http://localhost:8000/oauth2/callback"
	} else {
		oidcConfig.RedirectURL = fmt.Sprintf("https://%s/oauth2/callback", c.UserDomain)
	}

	oidcSecretData := map[string][]byte{
		"issuerURL":    []byte(oidcConfig.IssuerURL),
		"clientID":     []byte(oidcConfig.ClientID),
		"clientSecret": []byte(oidcConfig.ClientSecret),
		"redirectURL":  []byte(oidcConfig.RedirectURL),
	}

	if err := utils.CreateSecret(c.KubernetesClient, oidcSecretName, WGEDefaultNamespace, oidcSecretData); err != nil {
		return []StepOutput{}, err
	}

	values := constructOIDCValues(oidcConfig)

	c.Logger.Waitingf(oidcInstallInfoMsg)

	if err := updateHelmReleaseValues(c.KubernetesClient, oidcValuesName, values); err != nil {
		return []StepOutput{}, err
	}
	c.Logger.Successf(oidcConfirmationMsg)

	return []StepOutput{}, nil
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

// checkExistingOIDCConfig checks for OIDC secret on management cluster
// returns false if OIDC is already on the cluster
// returns true if no OIDC on the cluster
func checkExistingOIDCConfig(input []StepInput, c *Config) (interface{}, error) {
	_, err := utils.GetSecret(c.KubernetesClient, oidcSecretName, WGEDefaultNamespace)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func canAskForConfig(input []StepInput, c *Config) (interface{}, error) {
	if c.PromptedForDiscoveryURL {
		return false, nil
	}
	if ask, _ := checkExistingOIDCConfig(input, c); ask.(bool) {
		return false, nil
	}
	return true, nil
}
