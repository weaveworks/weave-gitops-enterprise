package commands

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const ADMIN_SECRET_NAME string = "cluster-user-auth"
const ADMIN_SECRET_NAMESPACE string = "flux-system"

func GetAdminPasswordSecrets() (string, []byte) {
	AdminUsernamePromptContent := promptContent{
		"Admin username can't be empty",
		"Please enter your admin username (default: wego-admin)",
		"wego-admin",
	}
	adminUsername := promptGetStringInput(AdminUsernamePromptContent)

	AdminPasswordPromptContent := promptContent{
		"Admin password can't be empty",
		"Please enter your admin password",
		"",
	}
	adminPassword := promptGetPasswordInput(AdminPasswordPromptContent)
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		panic(err.Error())
	}
	return adminUsername, encryptedPassword
}

func CreateAdminPasswordSecret() {
	adminUsername, adminPassword := GetAdminPasswordSecrets()
	data := map[string][]byte{
		"username": []byte(adminUsername),
		"password": adminPassword,
	}
	createSecret(ADMIN_SECRET_NAME, ADMIN_SECRET_NAMESPACE, data)
	fmt.Println("✔ admin secret is created")
}
