package commands

import (
	"os"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"golang.org/x/crypto/bcrypt"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	adminUsernameMsg           = "Please enter WeaveGitOps dashboard admin username (default: wego-admin)"
	adminPasswordMsg           = "Please enter admin password (Minimum characters: 6)"
	secretConfirmationMsg      = "Admin login credentials has been created successfully!"
	adminSecretExistsMsgFormat = "Admin login credentials already exist on the cluster. To reset admin credentials please remove secret '%s' in namespace '%s', then try again."
	existingCredsMsg           = "Do you want to continue using existing credentials"
	existingCredsExitMsg       = "If you want to reset admin credentials please remove secret '%s' in namespace '%s', then try again.\nExiting gitops bootstrap..."
)

const (
	defaultAdminUsername = "wego-admin"
	defaultAdminPassword = "password"
	adminSecretName      = "cluster-user-auth"
	confirmYes           = "y"
)

// AskAdminCredsSecrets asks user about admin username and password.
// admin username and password are you used for accessing WGE Dashboard
// for emergency access. OIDC can be used instead.
// there an option to revert these creds in case OIDC setup is successful
// if the creds already exist. user will be asked to continue with the current creds
// Or existing and deleting the creds then re-run the bootstrap process
func AskAdminCredsSecret(client k8s_client.Client) error {
	// search for existing admin credentials in secret cluster-user-auth
	secret, err := utils.GetSecret(client, adminSecretName, WGEDefaultNamespace)
	if secret != nil && err == nil {
		existingCreds := utils.GetConfirmInput(existingCredsMsg)
		if existingCreds == confirmYes {
			return nil
		} else {
			utils.Warning(existingCredsExitMsg, adminSecretName, WGEDefaultNamespace)
			os.Exit(0)
		}
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

	if err := utils.CreateSecret(client, adminSecretName, WGEDefaultNamespace, data); err != nil {
		return err
	}

	utils.Info(secretConfirmationMsg)

	return nil
}
