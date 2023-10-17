package steps

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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

	oidcConfigExistWarningMsg = "OIDC is already configured on the cluster. To reset configurations please remove secret '%s' in namespace '%s' and run 'bootstrap auth --type=oidc' command again"
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

func NewOIDCConfigStep(config Config) BootstrapStep {
	inputs := []StepInput{
		{
			Name:            existingOIDC,
			Type:            confirmInput,
			Msg:             existingOIDCMsg,
			DefaultValue:    "",
			Valuesfn:        isExistingOIDCConfig,
			StepInformation: fmt.Sprintf(oidcConfigExistWarningMsg, oidcSecretName, WGEDefaultNamespace),
		},
		{
			Name:         oidcInstalled,
			Type:         confirmInput,
			Msg:          oidcInstallMsg,
			DefaultValue: "",
			Valuesfn:     canAskOIDCPrompot,
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
	continueWithExistingConfigs := confirmYes

	// process params
	for _, param := range input {
		switch param.Name {
		case DiscoveryURL:
			discoveryUrl, ok := param.Value.(string)
			if ok {
				c.DiscoveryURL = discoveryUrl
			}
		case ClientID:
			clientId, ok := param.Value.(string)
			if ok {
				c.ClientID = clientId
			}
		case ClientSecret:
			clientSecret, ok := param.Value.(string)
			if ok {
				c.ClientSecret = clientSecret
			}
		case existingOIDC:
			existing, ok := param.Value.(string)
			if ok {
				continueWithExistingConfigs = existing
			}
		}
	}

	// check existing oidc configuration
	if existing, _ := isExistingOIDCConfig(input, c); existing.(bool) {
		if continueWithExistingConfigs != confirmYes {
			return []StepOutput{}, fmt.Errorf(oidcConfigExistWarningMsg, oidcSecretName, WGEDefaultNamespace)
		} else {
			return []StepOutput{}, nil
		}
	}

	// process user domain if not passed
	if c.UserDomain == "" {
		domain, err := utils.GetHelmReleaseProperty(c.KubernetesClient, WgeHelmReleaseName, WGEDefaultNamespace, utils.HelmDomainProperty)
		if err != nil {
			return []StepOutput{}, fmt.Errorf("error getting helm release domain: %v", err)
		}
		c.Logger.Actionf("setting user domain: %s", domain)
		c.UserDomain = domain

	}

	issuerUrl, err := getIssuerFromDiscoveryUrl(c)
	if err != nil {
		return []StepOutput{}, err
	}
	c.Logger.Actionf("retrieved issure url: %s", issuerUrl)
	c.IssuerURL = issuerUrl

	if c.DomainType == domainTypeLocalhost {
		c.RedirectURL = "http://localhost:8000/oauth2/callback"
	} else {
		c.RedirectURL = fmt.Sprintf("https://%s/oauth2/callback", c.UserDomain)
	}
	c.Logger.Actionf("setting redirect url: %s", c.RedirectURL)

	oidcSecretData := map[string][]byte{
		"issuerURL":    []byte(c.IssuerURL),
		"clientID":     []byte(c.ClientID),
		"clientSecret": []byte(c.ClientSecret),
		"redirectURL":  []byte(c.RedirectURL),
	}

	c.Logger.Actionf("creating secret %s:%s", WGEDefaultNamespace, oidcSecretName)
	if err := utils.CreateSecret(c.KubernetesClient, oidcSecretName, WGEDefaultNamespace, oidcSecretData); err != nil {
		return []StepOutput{}, err
	}
	c.Logger.Actionf("created secret %s:%s", WGEDefaultNamespace, oidcSecretName)

	c.Logger.Waitingf(oidcInstallInfoMsg)
	values, err := utils.GetHelmReleaseValues(c.KubernetesClient, WgeHelmReleaseName, WGEDefaultNamespace)
	if err != nil {
		return []StepOutput{}, err
	}

	wgeValues, ok := values.(valuesFile)
	if !ok {
		return []StepOutput{}, fmt.Errorf("failed to parse Weave GitOps Enterprise HelmRelease's values")
	}

	c.Logger.Actionf("configuring oidc values")
	wgeValues.Config.OIDC = map[string]interface{}{
		"enabled":                 true,
		"issuerURL":               c.IssuerURL,
		"redirectURL":             c.RedirectURL,
		"clientCredentialsSecret": oidcSecretName,
	}

	wgeHelmRelease, err := constructWGEhelmRelease(wgeValues, c.WGEVersion)
	if err != nil {
		return []StepOutput{}, err
	}
	c.Logger.Actionf("rendered HelmRelease file")
	c.Logger.Successf(oidcConfirmationMsg)

	c.Logger.Actionf("updating HelmRelease file")
	helmreleaseFile := fileContent{
		Name:      wgeHelmReleaseFileName,
		Content:   wgeHelmRelease,
		CommitMsg: wgeHelmReleaseCommitMsg,
	}

	return []StepOutput{{
		Name:  wgeHelmReleaseFileName,
		Type:  typeFile,
		Value: helmreleaseFile,
	}}, nil
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

// isExistingOIDCConfig checks for OIDC secret on management cluster
// returns false if OIDC is already on the cluster
// returns true if no OIDC on the cluster
func isExistingOIDCConfig(input []StepInput, c *Config) (interface{}, error) {
	_, err := utils.GetSecret(c.KubernetesClient, oidcSecretName, WGEDefaultNamespace)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func canAskForConfig(input []StepInput, c *Config) (interface{}, error) {
	if ask, _ := isExistingOIDCConfig(input, c); ask.(bool) {
		return false, nil
	}
	return true, nil
}

func canAskOIDCPrompot(input []StepInput, c *Config) (interface{}, error) {
	return c.PromptedForDiscoveryURL, nil
}

// func to get issuer url from discovery url
func getIssuerFromDiscoveryUrl(c *Config) (string, error) {
	// check if discovery url is valid, try for 3 times if not valid
	issuerURLErrCount := 0
	for {
		issuerURL, err := getIssuer(c.DiscoveryURL)
		if err != nil {
			issuerURLErrCount++
			// if we fail to get issuer url after 3 attempts, we will return an error
			if issuerURLErrCount > 3 {
				return "", fmt.Errorf("failed to retrieve IssuerURL after multiple attempts. Please verify the DiscoveryURL and try again")
			}
			c.Logger.Warningf("Failed to retrieve IssuerURL. Please verify the DiscoveryURL and try again")
			// ask for discovery url again
			c.DiscoveryURL, err = utils.GetStringInput(oidcDiscoverUrlMsg, "")
			if err != nil {
				return "", err
			}
			continue
		}
		return issuerURL, nil
	}
}
