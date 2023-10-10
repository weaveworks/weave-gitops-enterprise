package steps

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
	existingOIDCMsg   = "Do you want to continue using existing OIDC configurations"
	oidcConfigInfoMsg = "Setting up OIDC require configurations provided by your OIDC provider. To learn more about these OIDC configurations, checkout https://docs.gitops.weave.works/docs/next/configuration/oidc-access/#configuration"

	oidcDiscoverUrlMsg    = "Please enter OIDC Discovery URL (example: https://example-idp.com/.well-known/openid-configuration)"
	discoveryUrlVerifyMsg = "Verifying OIDC discovery URL"

	discoveryUrlErrorMsgFormat = "error: OIDC discovery URL returned status %d"
	discoveryUrlNoIssuerMsg    = "error: OIDC discovery URL returned no issuer"

	oidcClientIDMsg     = "Please enter OIDC clientID"
	oidcClientSecretMsg = "Please enter OIDC clientSecret"

	oidcInstallInfoMsg  = "Configuring OIDC"
	oidcConfirmationMsg = "OIDC has been configured successfully!"

	oidcConfigExistWarningMsg = "OIDC is already configured on the cluster. To reset configurations please remove secret '%s' in namespace '%s' and run 'bootstrap auth --type=oidc' command again."

	adminUserRevertMsg     = "Do you want to revoke admin user login, and only use OIDC for dashboard access"
	adminUsernameRevertMsg = "Admin user login has been revoked!"
)

const (
	oidcSecretName = "oidc-auth"
)

func OIDCConfigStep(config Config) BootstrapStep {

	inputs := []StepInput{
		{
			Name:         "oidcInstall",
			Type:         confirmInput,
			Msg:          oidcInstallMsg,
			DefaultValue: "",
			Valuesfn:     canAskOIDCPrompot,
		},
		{
			Name:            "existingOIDC",
			Type:            confirmInput,
			Msg:             existingOIDCMsg,
			DefaultValue:    "",
			Valuesfn:        checkExistingOIDCConfig,
			StepInformation: fmt.Sprintf(oidcConfigExistWarningMsg, oidcSecretName, WGEDefaultNamespace),
		},
	}

	steps := []StepInput{
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
	}

	inputs = append(inputs, steps...)

	return BootstrapStep{
		Name:  "OIDC Configuration",
		Input: inputs,
		Step:  createOIDCConfig,
	}

}

// createOIDCConfig creates OIDC secrets on the cluster and updates the OIDC values in the helm release.
// If the OIDC configs already exist, we will ask the user to delete the secret and run the command again.
func createOIDCConfig(input []StepInput, c *Config) ([]StepOutput, error) {

	oidcConfig := OIDCConfig{}

	continueWithExistingConfigs := confirmYes

	for _, param := range input {

		if param.Name == DiscoveryURL && param.Value != nil {
			issuerUrl, err := getIssuerFromDiscoveryUrl(param, c)
			if err != nil {
				return []StepOutput{}, err
			}
			oidcConfig.IssuerURL = issuerUrl[0].Value.(string)
		}

		if param.Name == ClientID && param.Value != nil {
			if val, ok := param.Value.(string); ok {
				oidcConfig.ClientID = val
			} else {
				return []StepOutput{}, errors.New("ClientID not found")
			}
		}
		if param.Name == ClientSecret && param.Value != nil {
			if val, ok := param.Value.(string); ok {
				oidcConfig.ClientSecret = val
			} else {
				return []StepOutput{}, errors.New("ClientSecret not found")
			}
		}
		if param.Name == "existingOIDC" && param.Value != nil {
			existing, ok := param.Value.(string)
			if ok {
				continueWithExistingConfigs = existing
			}
		}
	}

	if existing, _ := checkExistingOIDCConfig(input, c); existing.(bool) {
		if continueWithExistingConfigs != confirmYes {
			c.Logger.Warningf(oidcConfigExistWarningMsg, oidcSecretName, WGEDefaultNamespace)
			return []StepOutput{}, nil
		} else {
			//get oidc-auth secret and construct oidcConfig
			secret, err := utils.GetSecret(c.KubernetesClient, oidcSecretName, WGEDefaultNamespace)
			if err != nil {
				return []StepOutput{}, err
			}
			oidcConfig.IssuerURL = string(secret.Data["issuerURL"])
			oidcConfig.ClientID = string(secret.Data["clientID"])
			oidcConfig.ClientSecret = string(secret.Data["clientSecret"])
		}
	}

	if strings.Contains(c.UserDomain, domainTypeLocalhost) {
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

	//if secret doesn't exist, create it
	if existing, _ := checkExistingOIDCConfig(input, c); !existing.(bool) {
		if err := utils.CreateSecret(c.KubernetesClient, oidcSecretName, WGEDefaultNamespace, oidcSecretData); err != nil {
			return []StepOutput{}, err
		}
	}

	values := constructOIDCValues(oidcConfig)

	c.Logger.Waitingf(oidcInstallInfoMsg)

	if err := updateHelmReleaseValues(c, oidcValuesName, values); err != nil {
		return []StepOutput{}, err
	}
	c.Logger.Successf(oidcConfirmationMsg)

	return []StepOutput{}, nil
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
	if ask, _ := checkExistingOIDCConfig(input, c); ask.(bool) {
		return false, nil
	}
	return true, nil
}

func canAskOIDCPrompot(input []StepInput, c *Config) (interface{}, error) {
	//check for config.PromptedForDiscoveryURL
	if c.PromptedForDiscoveryURL {
		return true, nil
	}
	return false, nil
}

// func to get issuer url from discovery url
func getIssuerFromDiscoveryUrl(param StepInput, c *Config) ([]StepOutput, error) {

	// check if discovery url is valid, try for 3 times if not valid
	issuerURLErrCount := 0
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
			return []StepOutput{{Name: "issuerURL", Value: issuerURL}}, nil
		} else {
			return []StepOutput{}, errors.New("DiscoveryURL not found")
		}
	}
}
