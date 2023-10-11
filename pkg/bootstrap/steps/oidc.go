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
)

const (
	oidcSecretName = "oidc-auth"
)

var discoveryUrlStep = StepInput{
	Name:            DiscoveryURL,
	Type:            stringInput,
	Msg:             oidcDiscoverUrlMsg,
	DefaultValue:    "",
	Valuesfn:        canAskForConfig,
	StepInformation: oidcConfigInfoMsg,
}

var clientIDStep = StepInput{
	Name:         ClientID,
	Type:         stringInput,
	Msg:          oidcClientIDMsg,
	DefaultValue: "",
	Valuesfn:     canAskForConfig,
}

var clientSecretStep = StepInput{
	Name:         ClientSecret,
	Type:         passwordInput,
	Msg:          oidcClientSecretMsg,
	DefaultValue: "",
	Valuesfn:     canAskForConfig,
}

func OIDCConfigStep(config Config) BootstrapStep {

	inputs := []StepInput{
		{
			Name:         oidcInstalled,
			Type:         confirmInput,
			Msg:          oidcInstallMsg,
			DefaultValue: "",
			Valuesfn:     canAskOIDCPrompot,
		},
		{
			Name:            existingOIDC,
			Type:            confirmInput,
			Msg:             existingOIDCMsg,
			DefaultValue:    "",
			Valuesfn:        checkExistingOIDCConfig,
			StepInformation: fmt.Sprintf(oidcConfigExistWarningMsg, oidcSecretName, WGEDefaultNamespace),
		},
	}

	if config.DiscoveryURL == "" {
		inputs = append(inputs, discoveryUrlStep)
	}
	if config.ClientID == "" {
		inputs = append(inputs, clientIDStep)
	}
	if config.ClientSecret == "" {
		inputs = append(inputs, clientSecretStep)
	}

	return BootstrapStep{
		Name:  "OIDC Configuration",
		Input: inputs,
		Step:  createOIDCConfig,
	}

}

// createOIDCConfig creates OIDC secrets on the cluster and updates the OIDC values in the helm release.
// If the OIDC configs already exist, we will ask the user to delete the secret and run the command again.
func createOIDCConfig(input []StepInput, c *Config) ([]StepOutput, error) {

	//oidcConfig := OIDCConfig{}

	continueWithExistingConfigs := confirmYes

	for _, param := range input {

		if param.Name == DiscoveryURL && param.Value != nil {
			if val, ok := param.Value.(string); ok {
				c.DiscoveryURL = val
			} else {
				return []StepOutput{}, errors.New("DiscoveryURL not found")
			}
		}
		if param.Name == ClientID && param.Value != nil {
			if val, ok := param.Value.(string); ok {
				c.ClientID = val
			} else {
				return []StepOutput{}, errors.New("ClientID not found")
			}
		}
		if param.Name == ClientSecret && param.Value != nil {
			if val, ok := param.Value.(string); ok {
				c.ClientSecret = val
			} else {
				return []StepOutput{}, errors.New("ClientSecret not found")
			}
		}
		if param.Name == existingOIDC && param.Value != nil {
			existing, ok := param.Value.(string)
			if ok {
				continueWithExistingConfigs = existing
			}
		}
	}
	//get issuer url from discovery url
	issuerUrl, err := getIssuerFromDiscoveryUrl(c)
	if err != nil {
		return []StepOutput{}, err
	}
	c.IssuerURL = issuerUrl[0].Value.(string)

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
			c.IssuerURL = string(secret.Data["issuerURL"])
			c.ClientID = string(secret.Data["clientID"])
			c.ClientSecret = string(secret.Data["clientSecret"])
		}
	}

	if strings.Contains(c.UserDomain, domainTypeLocalhost) {
		c.RedirectURL = "http://localhost:8000/oauth2/callback"
	} else {
		c.RedirectURL = fmt.Sprintf("https://%s/oauth2/callback", c.UserDomain)
	}

	oidcSecretData := map[string][]byte{
		"issuerURL":    []byte(c.IssuerURL),
		"clientID":     []byte(c.ClientID),
		"clientSecret": []byte(c.ClientSecret),
		"redirectURL":  []byte(c.RedirectURL),
	}

	//if secret doesn't exist, create it
	if existing, _ := checkExistingOIDCConfig(input, c); !existing.(bool) {
		if err := utils.CreateSecret(c.KubernetesClient, oidcSecretName, WGEDefaultNamespace, oidcSecretData); err != nil {
			return []StepOutput{}, err
		}
	}

	values := constructOIDCValues(c)

	c.Logger.Waitingf(oidcInstallInfoMsg)

	if err := updateHelmReleaseValues(c, oidcValuesName, values); err != nil {
		return []StepOutput{}, err
	}
	c.Logger.Successf(oidcConfirmationMsg)

	return []StepOutput{}, nil
}

// constructOIDCValues construct the OIDC values
func constructOIDCValues(c *Config) map[string]interface{} {
	values := map[string]interface{}{
		"enabled":                 true,
		"issuerURL":               c.IssuerURL,
		"redirectURL":             c.RedirectURL,
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
func getIssuerFromDiscoveryUrl(c *Config) ([]StepOutput, error) {

	// check if discovery url is valid, try for 3 times if not valid
	issuerURLErrCount := 0
	for {
		issuerURL, err := getIssuer(c.DiscoveryURL)
		if err != nil {
			issuerURLErrCount++
			// if we fail to get issuer url after 3 attempts, we will return an error
			if issuerURLErrCount > 3 {
				return []StepOutput{}, errors.New("Failed to retrieve IssuerURL after multiple attempts. Please verify the DiscoveryURL and try again.")
			}
			c.Logger.Warningf("Failed to retrieve IssuerURL. Please verify the DiscoveryURL and try again.")
			// ask for discovery url again
			c.DiscoveryURL, err = utils.GetStringInput(oidcDiscoverUrlMsg, "")
			if err != nil {
				return []StepOutput{}, err
			}
			continue
		}
		return []StepOutput{{Name: "issuerURL", Value: issuerURL}}, nil
	}
}
