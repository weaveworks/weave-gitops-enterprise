package commands

import (
	"os"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"golang.org/x/crypto/bcrypt"
)

const (
	adminUsernameMsg           = "Please enter WeaveGitOps dashboard admin username (default: wego-admin)"
	adminPasswordMsg           = "Please enter admin password"
	secretConfirmationMsg      = "Admin login credentials has been created successfully!"
	adminSecretExistsMsgFormat = "Admin login credentials already exist on the cluster. To reset admin credentials please remove secret '%s' in namespace '%s', then try again."
	existingCredsMsg           = "Do you want to continue using existing credentials"
	existingCredsExitMsg       = "If you want to reset admin credentials please remove secret '%s' in namespace '%s', then try again.\nExiting gitops bootstrap..."
)
const (
	defaultAdminUsername = "wego-admin"
	adminSecretName      = "cluster-user-auth"
)

// getAdminPasswordSecrets asks user about admin username and password.
func getAdminPasswordSecrets() (string, []byte, error) {
	if _, err := utils.GetSecret(adminSecretName, wgeDefaultNamespace); err != nil && !strings.Contains(err.Error(), "not found") {
		return "", nil, err
	} else if err == nil {
		utils.Warning(adminSecretExistsMsgFormat, adminSecretName, wgeDefaultNamespace)
		existingCreds := utils.GetConfirmInput(existingCredsMsg)
		if existingCreds == "y" {
			return "", nil, nil
		} else {
			utils.Warning(existingCredsExitMsg, adminSecretName, wgeDefaultNamespace)
			os.Exit(0)
		}
	}
	adminUsername, err := utils.GetStringInput(adminUsernameMsg, defaultAdminUsername)
	if err != nil {
		return "", nil, err
	}

	adminPassword, err := utils.GetPasswordInput(adminPasswordMsg)
	if err != nil {
		return "", nil, err
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, err
	}

	return adminUsername, encryptedPassword, nil
}

// CreateAdminPasswordSecret creates the secret for admin access with username and password.
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

	if err := utils.CreateSecret(adminSecretName, wgeDefaultNamespace, data); err != nil {
		return err
	}

	utils.Info(secretConfirmationMsg)

	return nil
}
