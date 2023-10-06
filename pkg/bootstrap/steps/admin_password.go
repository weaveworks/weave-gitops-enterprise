package steps

import (
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	adminUsernameMsg           = "dashboard admin username (default: wego-admin)"
	adminPasswordMsg           = "dashboard admin password (Minimum characters: 6)"
	secretConfirmationMsg      = "Admin login credentials has been created successfully!"
	adminSecretExistsMsgFormat = "Admin login credentials already exist on the cluster. To reset admin credentials please remove secret '%s' in namespace '%s', then try again."
	existingCredsMsg           = "Do you want to continue using existing credentials"
	existingCredsExitMsg       = "If you want to reset admin credentials please remove secret '%s' in namespace '%s', then try again.\nExiting gitops bootstrap..."
)

const (
	adminSecretName = "cluster-user-auth"
	confirmYes      = "y"
)

var getUsernameInput = StepInput{
	Name:         UserName,
	Type:         stringInput,
	Msg:          adminUsernameMsg,
	DefaultValue: defaultAdminUsername,
	Valuesfn:     canAskForCreds,
}

var getPasswordInput = StepInput{
	Name:         Password,
	Type:         passwordInput,
	Msg:          adminPasswordMsg,
	DefaultValue: defaultAdminPassword,
	Valuesfn:     canAskForCreds,
}

// NewAskAdminCredsSecretStep asks user about admin username and password.
// admin username and password are you used for accessing WGE Dashboard
// for emergency access. OIDC can be used instead.
// there an option to revert these creds in case OIDC setup is successful
// if the creds already exist. user will be asked to continue with the current creds
// Or existing and deleting the creds then re-run the bootstrap process
func NewAskAdminCredsSecretStep(config Config) BootstrapStep {
	inputs := []StepInput{
		{
			Name:            "existingCreds",
			Type:            confirmInput,
			Msg:             existingCredsMsg,
			DefaultValue:    "",
			Valuesfn:        checkExistingAdminSecret,
			StepInformation: fmt.Sprintf(adminSecretExistsMsgFormat, adminSecretName, WGEDefaultNamespace),
		},
	}

	if config.Username != "" {
		inputs = append(inputs, getUsernameInput)
	}

	if config.Password != "" {
		inputs = append(inputs, getPasswordInput)
	}

	return BootstrapStep{
		Name:  "User Authentication",
		Input: inputs,
		Step:  createCredentials,
	}
}

func createCredentials(input []StepInput, c *Config) ([]StepOutput, error) {
	// search for existing admin credentials in secret cluster-user-auth
	continueWithExistingCreds := confirmYes
	for _, param := range input {
		if param.Name == UserName {
			username, ok := param.Value.(string)
			if ok {
				c.Username = username
			}
		}
		if param.Name == Password {
			password, ok := param.Value.(string)
			if ok {
				c.Password = password
			}
		}
		if param.Name == "existingCreds" {
			existing, ok := param.Value.(string)
			if ok {
				continueWithExistingCreds = existing
			}
		}
	}

	if existing, _ := checkExistingAdminSecret(input, c); existing.(bool) {
		if continueWithExistingCreds != confirmYes {
			c.Logger.Warningf(existingCredsExitMsg, adminSecretName, WGEDefaultNamespace)
			os.Exit(0)
		} else {
			return []StepOutput{}, nil
		}
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	data := map[string][]byte{
		"username": []byte(c.Username),
		"password": encryptedPassword,
	}

	secret := corev1.Secret{
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
			Type:  typeSecret,
			Value: secret,
		},
	}, nil

}

// checkExistingAdminSecret checks for admin secret on management cluster
// returns true if admin secret is already on the cluster
// returns false if no admin secret on the cluster
func checkExistingAdminSecret(input []StepInput, c *Config) (interface{}, error) {
	_, err := utils.GetSecret(c.KubernetesClient, adminSecretName, WGEDefaultNamespace)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func canAskForCreds(input []StepInput, c *Config) (interface{}, error) {
	if ask, _ := checkExistingAdminSecret(input, c); ask.(bool) {
		return false, nil
	}
	return true, nil
}
