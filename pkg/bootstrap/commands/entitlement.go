package commands

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/weaveworks/weave-gitops-enterprise-credentials/pkg/entitlement"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	entitlementCheckConfirmMsg      = "Entitlement File exists and is valid!"
	nonExistingEntitlementSecretMsg = "\n✖️ Entitlement file is not found, To get Weave GitOps Entitelment secret, please contact *sales@weave.works* and add it to your cluster.\n"
	invalidEntitlementSecretMsg     = "\n✖️ Entitlement file is invalid, please verify the secret content. If you still facing issues, please contact *sales@weave.works*."
	entitlementCheckMsg             = "Verifying Weave GitOps Entitlement File ..."
)

const (
	entitlementSecretName = "weave-gitops-enterprise-credentials"
)

var (
	//go:embed public.pem
	publicKey string
)

// CheckEntitlementSecret checks for valid entitlement secret.
func CheckEntitlementSecret(client k8s_client.Client) error {
	utils.Warning(entitlementCheckMsg)

	err := verifyEntitlementSecret(client)
	if err != nil {
		return err
	}

	utils.Info(entitlementCheckConfirmMsg)
	return nil
}

func verifyEntitlementSecret(client k8s_client.Client) error {
	secret, err := utils.GetSecret(entitlementSecretName, WGEDefaultNamespace, client)
	if err != nil || secret.Data["entitlement"] == nil {
		return fmt.Errorf("%s: %w", nonExistingEntitlementSecretMsg, err)
	}

	ent, err := entitlement.VerifyEntitlement(strings.NewReader(string(publicKey)), string(secret.Data["entitlement"]))
	if err != nil || time.Now().Compare(ent.IssuedAt) <= 0 {
		return fmt.Errorf("%s: %w", invalidEntitlementSecretMsg, err)
	}

	// TODO: verify username and password

	return nil
}
