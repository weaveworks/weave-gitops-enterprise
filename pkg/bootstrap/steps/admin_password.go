package steps

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"
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

var getPasswordInput = StepInput{
	Name:         inPassword,
	Type:         passwordInput,
	Msg:          adminPasswordMsg,
	DefaultValue: defaultAdminPassword,
	Enabled:      canAskForCreds,
	Required:     true,
}

var getPasswordInput2 = StepInput{
	Name:         inPassword,
	Type:         passwordInput,
	Msg:          adminPasswordMsg,
	DefaultValue: defaultAdminPassword,
}

type ClusterUserAuthConfig struct {
	Username         string
	Password         string
	ExistCredentials bool
}

// NewClusterUserAuthConfig creates new configuration out of the user input and discovered state
func NewClusterUserAuthConfig(password string, client k8s_client.Client) (ClusterUserAuthConfig, error) {
	if password != "" && len(password) < 6 {
		return ClusterUserAuthConfig{}, fmt.Errorf("password minimum characters should be >= 6")
	}
	return ClusterUserAuthConfig{
		Username:         defaultAdminUsername,
		Password:         password,
		ExistCredentials: isExistingAdminSecret(client),
	}, nil
}

// NewAskAdminCredsSecretStep asks user about admin password.
// Admin password are you used for accessing WGE Dashboard for emergency access.
// Users will be asked to continue with the current creds or overriding existing credentials during bootstrapping.
func NewAskAdminCredsSecretStep(config ClusterUserAuthConfig, silent bool) (BootstrapStep, error) {
	inputs := []StepInput{}

	if !silent {
		if !config.ExistCredentials {
			if config.Password == "" {
				inputs = append(inputs, getPasswordInput2)
			}
		}
	}

	return BootstrapStep{
		Name:  "user authentication",
		Input: inputs,
		Step:  createCredentials,
	}, nil
}

func createCredentials(input []StepInput, c *Config) ([]StepOutput, error) {
	// search for existing admin credentials in secret cluster-user-auth
	continueWithExistingCreds := confirmYes
	for _, param := range input {
		if param.Name == inPassword {
			password, ok := param.Value.(string)
			if ok {
				c.ClusterUserAuth.Password = password
			}
		}
		if param.Name == inExistingCreds {
			existing, ok := param.Value.(string)
			if ok {
				continueWithExistingCreds = existing
			}
		}
	}

	if existing := isExistingAdminSecret(c.KubernetesClient); existing {
		if continueWithExistingCreds != confirmYes {
			return []StepOutput{}, fmt.Errorf(existingCredsExitMsg, adminSecretName, WGEDefaultNamespace)
		} else {
			return []StepOutput{}, nil
		}
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(c.ClusterUserAuth.Password), bcrypt.DefaultCost)
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
func isExistingAdminSecret(client k8s_client.Client) bool {
	_, err := utils.GetSecret(client, adminSecretName, WGEDefaultNamespace)
	return err == nil
}

func canAskForCreds(input []StepInput, c *Config) bool {
	return !isExistingAdminSecret(c.KubernetesClient)
}
