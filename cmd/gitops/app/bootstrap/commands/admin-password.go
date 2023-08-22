package commands

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"golang.org/x/crypto/bcrypt"
)

const ADMIN_SECRET_NAME string = "cluster-user-auth"
const ADMIN_SECRET_NAMESPACE string = "flux-system"

func GetAdminPasswordSecrets() (string, []byte) {
	AdminUsernamePromptContent := utils.PromptContent{
		ErrorMsg:     "Admin username can't be empty",
		Label:        "Please enter your admin username (default: wego-admin)",
		DefaultValue: "wego-admin",
	}
	adminUsername := utils.GetPromptStringInput(AdminUsernamePromptContent)

	AdminPasswordPromptContent := utils.PromptContent{
		ErrorMsg:     "Admin password can't be empty",
		Label:        "Please enter your admin password",
		DefaultValue: "",
	}
	adminPassword := utils.GetPromptPasswordInput(AdminPasswordPromptContent)
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	utils.CheckIfError(err)
	return adminUsername, encryptedPassword
}

func CreateAdminPasswordSecret() {
	adminUsername, adminPassword := GetAdminPasswordSecrets()
	data := map[string][]byte{
		"username": []byte(adminUsername),
		"password": adminPassword,
	}
	utils.CreateSecret(ADMIN_SECRET_NAME, ADMIN_SECRET_NAMESPACE, data)
	fmt.Println("✔ admin secret is created")
}
