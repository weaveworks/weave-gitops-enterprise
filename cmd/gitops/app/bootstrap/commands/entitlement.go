package commands

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/weaveworks/weave-gitops-enterprise-credentials/pkg/entitlement"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"k8s.io/client-go/kubernetes"
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

// CheckEntitlementFile checks for valid entitlement secret.
func CheckEntitlementFile(opts config.Options) error {
	utils.Warning(entitlementCheckMsg)

	kubernetesClient, err := utils.GetKubernetesClient(opts.Kubeconfig)
	if err != nil {
		return err
	}

	err = verifyEntitlementFile(kubernetesClient)
	if err != nil {
		return err
	}

	utils.Info(entitlementCheckConfirmMsg)
	return nil
}

func verifyEntitlementFile(kubernetesClient kubernetes.Interface) error {
	secret, err := utils.GetSecret(entitlementSecretName, WGEDefaultNamespace, kubernetesClient)
	if err != nil || secret.Data["entitlement"] == nil {
		return fmt.Errorf("%s: %w", nonExistingEntitlementSecretMsg, err)
	}

	ent, err := entitlement.VerifyEntitlement(strings.NewReader(string(publicKey)), string(secret.Data["entitlement"]))
	if err != nil || time.Now().Compare(ent.IssuedAt) <= 0 {
		return fmt.Errorf("%s: %w", invalidEntitlementSecretMsg, err)
	}

	return nil
}
