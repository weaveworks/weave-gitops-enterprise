package commands

import (
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"golang.org/x/crypto/bcrypt"
)

const (
	ADMIN_USERNAME_MSG = "Please enter your admin username (default: wego-admin)"
	ADMIN_PASSWORD_MSG = "Please enter your admin password"
)

const (
	DEFAULT_ADMIN_USERNAME = "wego-admin"
	ADMIN_SECRET_NAME      = "cluster-user-auth"
)

// getAdminPasswordSecrets asks user about admin username and password
func getAdminPasswordSecrets() (string, []byte, error) {

	if _, err := utils.GetSecret(WGE_DEFAULT_NAMESPACE, ADMIN_SECRET_NAME); err == nil {
		utils.Info("admin secret already existed on the cluster, to reset please remove secret '%s' in namespace '%s'", ADMIN_SECRET_NAME, WGE_DEFAULT_NAMESPACE)
		return "", nil, nil
	} else if err != nil && !strings.Contains(err.Error(), "not found") {
		return "", nil, err
	}

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
	if adminUsername == "" || adminPassword == nil {
		return nil
	}

	data := map[string][]byte{
		"username": []byte(adminUsername),
		"password": adminPassword,
	}

	if err := utils.CreateSecret(ADMIN_SECRET_NAME, WGE_DEFAULT_NAMESPACE, data); err != nil {
		return err
	}

	utils.Info("admin secret is created")

	return nil
}
