package commands

import (
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"golang.org/x/crypto/bcrypt"
)

const (
	ADMIN_USERNAME_MSG     = "Please enter your admin username (default: wego-admin)"
	ADMIN_PASSWORD_MSG     = "Please enter your admin password"
	DEFAULT_ADMIN_USERNAME = "wego-admin"
	ADMIN_SECRET_NAME      = "cluster-user-auth"
)

// getAdminPasswordSecrets asks user about admin username and password
func getAdminPasswordSecrets() (string, []byte, error) {
	adminUsername, err := utils.GetStringInput(ADMIN_USERNAME_MSG, DEFAULT_ADMIN_USERNAME)
	if err != nil {
		return "", nil, err
	}

	adminPassword, err := utils.GetPasswordInput(ADMIN_PASSWORD_MSG)
	if err != nil {
		return "", nil, err
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, err
	}

	return adminUsername, encryptedPassword, nil
}

// CreateAdminPasswordSecret creates the secret for admin access with username and password
func CreateAdminPasswordSecret() error {
	adminUsername, adminPassword, err := getAdminPasswordSecrets()
	if err != nil {
		return err
	}

	data := map[string][]byte{
		"username": []byte(adminUsername),
		"password": adminPassword,
	}

	err = utils.CreateSecret(ADMIN_SECRET_NAME, WGE_DEFAULT_NAMESPACE, data)
	if err != nil {
		return err
	}

	utils.Info("✔ admin secret is created")

	return nil
}
