package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

const OIDC_SECRET_NAME string = "oidc-auth"
const OIDC_SECRET_NAMESPACE string = "flux-system"

func GetOIDCSecrets() (string, string, string, string) {

	oidcIssuerURLPrompt := utils.PromptContent{
		ErrorMsg:     "",
		Label:        "Please enter OIDC issuer URL",
		DefaultValue: "",
	}
	oidcIssuerURL := utils.GetPromptStringInput(oidcIssuerURLPrompt)

	oidcClientIDPrompt := utils.PromptContent{
		ErrorMsg:     "",
		Label:        "Please enter OIDC client ID",
		DefaultValue: "",
	}
	oidcClientID := utils.GetPromptStringInput(oidcClientIDPrompt)

	oidcClientSecretPrompt := utils.PromptContent{
		ErrorMsg:     "",
		Label:        "Please enter OIDC client-Secret",
		DefaultValue: "",
	}
	oidcClientSecret := utils.GetPromptStringInput(oidcClientSecretPrompt)

	oidcRedirectURLPrompt := utils.PromptContent{
		ErrorMsg:     "",
		Label:        "Please enter OIDC redirect URL",
		DefaultValue: "",
	}
	oidcRedirectURL := utils.GetPromptStringInput(oidcRedirectURLPrompt)

	return oidcIssuerURL, oidcClientID, oidcClientSecret, oidcRedirectURL
}

func CreateOIDCConfig(version string) {

	oidcConfigPrompt := utils.PromptContent{
		ErrorMsg:     "",
		Label:        "Do you want to add OIDC config?(y/n)",
		DefaultValue: "",
	}
	controllerName := utils.GetPromptStringInput(oidcConfigPrompt)

	if strings.Compare(controllerName, "y") == 0 {

		oidcIssuerURL, oidcClientID, oidcClientSecret, oidcRedirectURL := GetOIDCSecrets()

		oidcSecretData := map[string][]byte{
			"issuerURL":    []byte(oidcIssuerURL),
			"clientID":     []byte(oidcClientID),
			"clientSecret": []byte(oidcClientSecret),
			"redirectURL":  []byte(oidcRedirectURL),
		}

		utils.CreateSecret(OIDC_SECRET_NAME, OIDC_SECRET_NAMESPACE, oidcSecretData)

		valuesFile, err := os.OpenFile(VALUES_FILES_LOCATION, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}
		defer valuesFile.Close()
		defer os.Remove(VALUES_FILES_LOCATION)

		var values string
		values = fmt.Sprintf(`config:
  oidc:
    enabled: true
    issuerURL: %s
    redirectURL: %s
    clientCredentialsSecret: %s
`, oidcIssuerURL, oidcRedirectURL, OIDC_SECRET_NAME)

		if _, err = valuesFile.WriteString(values); err != nil {
			panic(err)
		}

		var runner runner.CLIRunner
		fmt.Println("Installing OIDC ...")
		out, err := runner.Run("flux", "create", "hr", HELMRELEASE_NAME,
			"--source", fmt.Sprintf("HelmRepository/%s", HELMREPOSITORY_NAME),
			"--chart", "mccp",
			"--chart-version", version,
			"--interval", "65m",
			"--crds", "CreateReplace",
			"--values", VALUES_FILES_LOCATION,
		)
		if err != nil {
			fmt.Printf("An error occurred updating helmrelease\n%v\n", string(out))
			os.Exit(1)
		}
		fmt.Println("✔ OIDC config created successfully")
	}
}
