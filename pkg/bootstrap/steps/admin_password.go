package steps

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	adminPasswordMsg           = "dashboard admin password (minimum characters: 6)"
	secretConfirmationMsg      = "admin login credentials has been created successfully!"
	adminSecretExistsMsgFormat = "admin login credentials already exist on the cluster. To reset admin credentials please remove secret '%s' in namespace '%s', then try again"
	existingCredsMsg           = "do you want to continue using existing credentials"
	existingCredsExitMsg       = "if you want to reset admin credentials please remove secret '%s' in namespace '%s', then try again.\nExiting gitops bootstrap"
)

const (
	adminSecretName      = "cluster-user-auth"
	confirmYes           = "y"
	defaultAdminUsername = "wego-admin"
)

var getPasswordInputConfig = StepInputConfig{
	Name:         inPassword,
	Type:         passwordInput,
	Msg:          adminPasswordMsg,
	DefaultValue: defaultAdminPassword,
	Enabled:      canAskForCreds,
	Required:     true,
}

// NewAskAdminCredsSecretStep asks user about admin  password.
// admin password are you used for accessing WGE Dashboard
// for emergency access. OIDC can be used instead.
// there an option to revert these creds in case OIDC setup is successful
// if the creds already exist. user will be asked to continue with the current creds
// Or existing and deleting the creds then re-run the bootstrap process
func NewAskAdminCredsSecretStep(config Config) (BootstrapStep, error) {
	inputs := []StepInput{
		{
			Name:            inExistingCreds,
			Type:            confirmInput,
			Msg:             existingCredsMsg,
			StepInformation: fmt.Sprintf(adminSecretExistsMsgFormat, adminSecretName, WGEDefaultNamespace),
			DefaultValue:    "",
			Enabled:         isExistingAdminSecret,
		},
	}

	if config.Password == "" {
		getPasswordInput, err := NewStepInput(&getPasswordInputConfig)
		if err != nil {
			return BootstrapStep{}
		}
		inputs = append(inputs, getPasswordInput)
	}

	return BootstrapStep{
		Name:  "user authentication",
		Input: inputs,
		Step:  createCredentials,
	}
}

func createCredentials(input []StepInput, c *Config) ([]StepOutput, error) {
	// search for existing admin credentials in secret cluster-user-auth
	continueWithExistingCreds := confirmYes
	for _, param := range input {
		if param.Name == inPassword {
			password, ok := param.Value.(string)
			if ok {
				c.Password = password
			}
		}
		if param.Name == inExistingCreds {
			existing, ok := param.Value.(string)
			if ok {
				continueWithExistingCreds = existing
			}
		}
	}

	if existing := isExistingAdminSecret(input, c); existing {
		if continueWithExistingCreds != confirmYes {
			return []StepOutput{}, fmt.Errorf(existingCredsExitMsg, adminSecretName, WGEDefaultNamespace)
		} else {
			return []StepOutput{}, nil
		}
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	data := map[string][]byte{
		"username": []byte(defaultAdminUsername),
		"password": encryptedPassword,
	}
	c.Logger.Actionf("dashboard admin username: %s is configured", defaultAdminUsername)

	secret := corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      adminSecretName,
			Namespace: WGEDefaultNamespace,
		},
		Data: data,
	}
	c.Logger.Successf(secretConfirmationMsg)

	return []StepOutput{
		{
			Name:  adminSecretName,
			Type:  typeSecret,
			Value: secret,
		},
	}, nil

}

// isExistingAdminSecret checks for admin secret on management cluster
// returns true if admin secret is already on the cluster
// returns false if no admin secret on the cluster
func isExistingAdminSecret(input []StepInput, c *Config) bool {
	_, err := utils.GetSecret(c.KubernetesClient, adminSecretName, WGEDefaultNamespace)
	return err == nil
}

func canAskForCreds(input []StepInput, c *Config) bool {
	return !isExistingAdminSecret(input, c)
}
