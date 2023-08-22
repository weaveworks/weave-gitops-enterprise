package commands

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

const (
	OIDC_VALUES_FILES_LOCATION = "/tmp/agent-values.yaml"
	OIDC_SECRET_NAME           = "oidc-auth"
	OIDC_SECRET_NAMESPACE      = "flux-system"
)

func GetOIDCSecrets() (string, string, string, string) {

	oidcIssuerURLPrompt := utils.PromptContent{
		ErrorMsg:     "OIDC issuerURL can't be empty",
		Label:        "Please enter OIDC issuerURL",
		DefaultValue: "",
	}
	oidcIssuerURL := utils.GetPromptStringInput(oidcIssuerURLPrompt)

	oidcClientIDPrompt := utils.PromptContent{
		ErrorMsg:     "OIDC clientID can't be empty",
		Label:        "Please enter OIDC clientID",
		DefaultValue: "",
	}
	oidcClientID := utils.GetPromptStringInput(oidcClientIDPrompt)

	oidcClientSecretPrompt := utils.PromptContent{
		ErrorMsg:     "OIDC clientSecret can't be empty",
		Label:        "Please enter OIDC clientSecret",
		DefaultValue: "",
	}
	oidcClientSecret := utils.GetPromptStringInput(oidcClientSecretPrompt)

	oidcRedirectURLPrompt := utils.PromptContent{
		ErrorMsg:     "OIDC redirectURL can't be empty",
		Label:        "Please enter OIDC redirectURL",
		DefaultValue: "",
	}
	oidcRedirectURL := utils.GetPromptStringInput(oidcRedirectURLPrompt)

	return oidcIssuerURL, oidcClientID, oidcClientSecret, oidcRedirectURL
}

func CreateOIDCConfig(version string) {

	oidcConfigPrompt := promptui.Prompt{
		Label:     "Do you want to add OIDC config to your cluster",
		IsConfirm: true,
	}

	result, _ := oidcConfigPrompt.Run()

	if result != "y" {
		return
	}

	oidcIssuerURL, oidcClientID, oidcClientSecret, oidcRedirectURL := GetOIDCSecrets()

	oidcSecretData := map[string][]byte{
		"issuerURL":    []byte(oidcIssuerURL),
		"clientID":     []byte(oidcClientID),
		"clientSecret": []byte(oidcClientSecret),
		"redirectURL":  []byte(oidcRedirectURL),
	}

	utils.CreateSecret(OIDC_SECRET_NAME, OIDC_SECRET_NAMESPACE, oidcSecretData)

	valuesFile, err := os.OpenFile(OIDC_VALUES_FILES_LOCATION, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer valuesFile.Close()
	defer os.Remove(OIDC_VALUES_FILES_LOCATION)

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
		"--values", OIDC_VALUES_FILES_LOCATION,
	)
	utils.CheckIfError(err, string(out))

	fmt.Println("✔ OIDC config created successfully")

	//make a function to ask the user if he want to revet the admin user or keep it
	checkAdminPassword()
}

func checkAdminPassword() {
	adminUserPrompot := promptui.Prompt{
		Label:     "Do you want to revert the admin user",
		IsConfirm: true,
	}

	result, _ := adminUserPrompot.Run()

	if result != "y" {
		return
	}

	utils.DeleteSecret(ADMIN_SECRET_NAME, ADMIN_SECRET_NAMESPACE)
	fmt.Println("✔ Admin user reverted successfully")
}
