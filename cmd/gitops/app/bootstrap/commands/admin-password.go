package commands

import (
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"golang.org/x/crypto/bcrypt"
)

const (
	adminUsernameMsg           = "Please enter your admin username (default: wego-admin)"
	adminPasswordMsg           = "Please enter your admin password"
	secretConfirmationMsg      = "admin secret is created"
	adminSecretExistsMsgFormat = "admin secret already existed on the cluster, to reset please remove secret '%s' in namespace '%s'"
)

const (
	defaultAdminUsername = "wego-admin"
	adminSecretName      = "cluster-user-auth"
)

// getAdminPasswordSecrets asks user about admin username and password.
func getAdminPasswordSecrets(opts config.Options) (string, []byte, error) {
	kubernetesClient, err := utils.GetKubernetesClient(opts.Kubeconfig)
	if err != nil {
		return "", nil, err
	}
	if _, err := utils.GetSecret(adminSecretName, wgeDefaultNamespace, kubernetesClient); err == nil {
		utils.Info(adminSecretExistsMsgFormat, adminSecretName, wgeDefaultNamespace)
		return "", nil, nil
	} else if err != nil && !strings.Contains(err.Error(), "not found") {
		return "", nil, err
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
func CreateAdminPasswordSecret(opts config.Options) error {
	adminUsername, adminPassword, err := getAdminPasswordSecrets(opts)
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
	kubernetesClient, err := utils.GetKubernetesClient(opts.Kubeconfig)
	if err != nil {
		return err
	}
	if err := utils.CreateSecret(adminSecretName, wgeDefaultNamespace, data, kubernetesClient); err != nil {
		return err
	}

	utils.Info(secretConfirmationMsg)

	return nil
}
