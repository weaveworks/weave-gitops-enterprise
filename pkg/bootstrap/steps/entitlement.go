package steps

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/weaveworks/weave-gitops-enterprise-credentials/pkg/entitlement"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"
)

// user messages
const (
	entitlementCheckConfirmMsg      = "entitlement file exists and is valid"
	nonExistingEntitlementSecretMsg = "entitlement file is not found, To get Weave GitOps Entitelment secret, please contact *sales@weave.works* and add it to your cluster"
	invalidEntitlementSecretMsg     = "entitlement file is invalid, please verify the secret content. If you still facing issues, please contact *sales@weave.works*"
	entitlementCheckMsg             = "Verifying Weave GitOps Entitlement File"
)

// wge consts
const (
	entitlementSecretName = "weave-gitops-enterprise-credentials"
	HelmVersionProperty   = "version"
)

var (
	//go:embed public.pem
	publicKey string
)

var CheckEntitlementSecret = BootstrapStep{
	Name: entitlementCheckMsg,
	Step: checkEntitlementSecret,
}

func checkEntitlementSecret(input []StepInput, c *Config) ([]StepOutput, error) {

	err := verifyEntitlementSecret(c.KubernetesClient)
	if err != nil {
		return []StepOutput{}, err
	}

	return []StepOutput{
		{
			Name:  "entitlement success msg",
			Type:  successMsg,
			Value: entitlementCheckConfirmMsg,
		},
	}, nil
}

// verifyEntitlementSecret ensures the entitlement is valid and not expired also verifying username & password
// verifing entitlement by the public key (private key is used for encrypting and public is for verification)
// and making sure it's not expired
// verifying username and password by making http request for downloading charts and ensuring it's authenticated
func verifyEntitlementSecret(client k8s_client.Client) error {
	secret, err := utils.GetSecret(client, entitlementSecretName, WGEDefaultNamespace)
	if err != nil || secret.Data["entitlement"] == nil || secret.Data["username"] == nil || secret.Data["password"] == nil {
		return errors.New(nonExistingEntitlementSecretMsg)
	}

	ent, err := entitlement.VerifyEntitlement(strings.NewReader(string(publicKey)), string(secret.Data["entitlement"]))
	if err != nil || time.Now().Compare(ent.IssuedAt) <= 0 {
		return fmt.Errorf("%s: %v", invalidEntitlementSecretMsg, err)
	}

	body, err := doBasicAuthGetRequest(wgeChartUrl, string(secret.Data["username"]), string(secret.Data["password"]))
	if err != nil || body == nil {
		return fmt.Errorf("%s: %v", invalidEntitlementSecretMsg, err)
	}

	return nil
}
