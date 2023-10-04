package commands

import (
	"os"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	adminSecretName = "cluster-user-auth"
	confirmYes      = "y"
)

// AskAdminCredsSecretStep asks user about admin username and password.
// admin username and password are you used for accessing WGE Dashboard
// for emergency access. OIDC can be used instead.
// there an option to revert these creds in case OIDC setup is successful
// if the creds already exist. user will be asked to continue with the current creds
// Or existing and deleting the creds then re-run the bootstrap process
var AskAdminCredsSecretStep = BootstrapStep{
	Name: "ask admin creds",
	Input: []StepInput{
		{
			Name:         "username",
			Type:         stringInput,
			Msg:          adminUsernameMsg,
			DefaultValue: defaultAdminUsername,
		},
		{
			Name:         "password",
			Type:         passwordInput,
			Msg:          adminPasswordMsg,
			DefaultValue: defaultAdminPassword,
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
	// search for existing admin credentials in secret cluster-user-auth
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	data := map[string][]byte{
		"username": []byte(c.Username),
		"password": encryptedPassword,
	}

	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      adminSecretName,
			Namespace: WGEDefaultNamespace,
		},
		Data: data,
	}

	return []StepOutput{
		{
			Name:  "secret is created",
			Type:  successMsg,
			Value: secretConfirmationMsg,
		},
		{
			Name:  "adminSecret",
			Type:  "secret",
			Value: secret,
		},
	}, nil

}

func checkExistingAdminSecret(input []StepInput, c *Config) ([]StepOutput, error) {
	secret, err := utils.GetSecret(c.KubernetesClient, adminSecretName, WGEDefaultNamespace)
	if secret != nil && err == nil {
		existingCreds := utils.GetConfirmInput(existingCredsMsg)
		if existingCreds == confirmYes {
			return []StepOutput{}, nil
		} else {
			c.Logger.Warningf(existingCredsExitMsg, adminSecretName, WGEDefaultNamespace)
			os.Exit(0)
		}
	}
	return []StepOutput{}, nil
}
