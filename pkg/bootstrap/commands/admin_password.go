package commands

import (
	"os"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"golang.org/x/crypto/bcrypt"
	v1 "k8s.io/api/core/v1"
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
	DefaultAdminUsername = "wego-admin"
	DefaultAdminPassword = "password"
	adminSecretName      = "cluster-user-auth"
	confirmYes           = "y"
)

var AskAdminCredsSecretStep = BootstrapStep{
	Name: "ask admin",
	Input: []StepInput{
		{
			Name:         "username",
			Type:         "string",
			Msg:          adminUsernameMsg,
			DefaultValue: DefaultAdminUsername,
		},
		{
			Name:         "password",
			Type:         "string",
			Msg:          adminPasswordMsg,
			DefaultValue: DefaultAdminPassword,
		},
	},
	Step: encryptCredentials,
	Output: []StepOutput{
		{
			Name: adminSecretName,
			Type: "secret",
		},
	},
}

func encryptCredentials(input []StepInput, c *Config) ([]StepOutput, error) {
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	data := map[string][]byte{
		"username": []byte(c.Username),
		"password": encryptedPassword,
	}

	if err := utils.CreateSecret(c.KubernetesClient, adminSecretName, WGEDefaultNamespace, data); err != nil {
		return nil, err
	}

	return []StepOutput{
		{
			Name:  "adminSecret",
			Type:  "secret",
			Value: v1.Secret{},
		},
	}, nil

}

// AskAdminCredsSecrets asks user about admin username and password.
// admin username and password are you used for accessing WGE Dashboard
// for emergency access. OIDC can be used instead.
// there an option to revert these creds in case OIDC setup is successful
// if the creds already exist. user will be asked to continue with the current creds
// Or existing and deleting the creds then re-run the bootstrap process
func (c *Config) AskAdminCredsSecret() error {
	// search for existing admin credentials in secret cluster-user-auth
	secret, err := utils.GetSecret(c.KubernetesClient, adminSecretName, WGEDefaultNamespace)
	if secret != nil && err == nil {
		existingCreds := utils.GetConfirmInput(existingCredsMsg)
		if existingCreds == confirmYes {
			return nil
		} else {
			c.Logger.Warningf(existingCredsExitMsg, adminSecretName, WGEDefaultNamespace)
			os.Exit(0)
		}
	}

	if c.Username == "" {
		c.Username, err = utils.GetStringInput(adminUsernameMsg, DefaultAdminUsername)
		if err != nil {
			return err
		}
	}

	if c.Password == "" {
		c.Password, err = utils.GetPasswordInput(adminPasswordMsg)
		if err != nil {
			return err
		}
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	data := map[string][]byte{
		"username": []byte(c.Username),
		"password": encryptedPassword,
	}

	if err := utils.CreateSecret(c.KubernetesClient, adminSecretName, WGEDefaultNamespace, data); err != nil {
		return err
	}

	c.Logger.Successf(secretConfirmationMsg)

	return nil
}
