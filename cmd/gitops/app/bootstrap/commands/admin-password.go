package commands

import (
	"os"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"golang.org/x/crypto/bcrypt"
	"k8s.io/client-go/kubernetes"
)

const (
	adminUsernameMsg           = "Please enter your admin username (default: wego-admin)"
	adminPasswordMsg           = "Please enter your admin password"
	secretConfirmationMsg      = "admin secret is created"
	adminSecretExistsMsgFormat = "admin secret already existed on the cluster, to reset please remove secret '%s' in namespace '%s'"
	existingCredsMsg           = "Do you want to continue using existing credentials"
	existingCredsExitMsg       = "If you want to reset admin credentials please remove secret '%s' in namespace '%s', then try again.\nExiting gitops bootstrap..."
)

const (
	defaultAdminUsername = "wego-admin"
	adminSecretName      = "cluster-user-auth"
)

// isAdminCredsAvailable if exists return not found error otherwise return nil
func isAdminCredsAvailable(kubernetesClient kubernetes.Interface) (bool, error) {
	if _, err := utils.GetSecret(adminSecretName, wgeDefaultNamespace, kubernetesClient); err == nil {
		return true, nil
	} else if err != nil && strings.Contains(err.Error(), "not found") {
		return false, nil
	} else {
		return false, err
	}
}

// AskAdminCredsSecrets asks user about admin username and password if it doesn't exist.
func AskAdminCredsSecret(opts config.Options) error {
	kubernetesClient, err := utils.GetKubernetesClient(opts.Kubeconfig)
	if err != nil {
		return err
	}

	available, err := isAdminCredsAvailable(kubernetesClient)
	if err != nil {
		return err
	}

	if available {
		utils.Info(adminSecretExistsMsgFormat, adminSecretName, wgeDefaultNamespace)
		existingCreds := utils.GetConfirmInput(existingCredsMsg)
		if existingCreds == "y" {
			return nil
		} else {
			utils.Warning(existingCredsExitMsg, adminSecretName, wgeDefaultNamespace)
			os.Exit(0)
		}
		return nil
	}

	adminUsername, err := utils.GetStringInput(adminUsernameMsg, defaultAdminUsername)
	if err != nil {
		return err
	}

	adminPassword, err := utils.GetPasswordInput(adminPasswordMsg)
	if err != nil {
		return err
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	data := map[string][]byte{
		"username": []byte(adminUsername),
		"password": encryptedPassword,
	}

	if err := utils.CreateSecret(adminSecretName, wgeDefaultNamespace, data, kubernetesClient); err != nil {
		return err
	}

	utils.Info(secretConfirmationMsg)

	return nil
}
