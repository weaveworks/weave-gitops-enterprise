package commands

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"golang.org/x/crypto/bcrypt"
)

func GetAdminPasswordSecrets() (string, []byte, error) {
	AdminUsernamePromptContent := utils.PromptContent{
		ErrorMsg:     "Admin username can't be empty",
		Label:        "Please enter your admin username (default: wego-admin)",
		DefaultValue: DEFAULT_ADMIN_USERNAME,
	}
	adminUsername, err := utils.GetPromptStringInput(AdminUsernamePromptContent)
	if err != nil {
		return "", nil, utils.CheckIfError(err)
	}

	AdminPasswordPromptContent := utils.PromptContent{
		ErrorMsg:     "Admin password can't be empty",
		Label:        "Please enter your admin password",
		DefaultValue: "",
	}
	adminPassword, err := utils.GetPromptPasswordInput(AdminPasswordPromptContent)
	if err != nil {
		return "", nil, utils.CheckIfError(err)
	}
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, utils.CheckIfError(err)
	}
	return adminUsername, encryptedPassword, nil
}

func CreateAdminPasswordSecret() error {
	adminUsername, adminPassword, err := GetAdminPasswordSecrets()
	if err != nil {
		return utils.CheckIfError(err)
	}
	data := map[string][]byte{
		"username": []byte(adminUsername),
		"password": adminPassword,
	}
	err = utils.CreateSecret(ADMIN_SECRET_NAME, WGE_DEFAULT_NAMESPACE, data)
	if err != nil {
		return utils.CheckIfError(err)
	}
	fmt.Println("✔ admin secret is created")
	return nil
}
