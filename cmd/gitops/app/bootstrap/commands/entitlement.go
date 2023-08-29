package commands

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

const (
	EntitlementCheckMsg   = "entitlement file is checked and valid!"
	EntitlementSecretName = "weave-gitops-enterprise-credentials"
)

// CheckEntitlementFile checks for valid entitlement secret
func CheckEntitlementFile() error {
	utils.Warning("Checking entitlement file ...")

	secret, err := utils.GetSecret(EntitlementSecretName, WGEDefaultNamespace)
	if err != nil || secret.Data["entitlement"] == nil {
		errorMsg := fmt.Sprintf("\n✖️  Invalid entitlement file, Please check secret: '%s' under namespace: '%s'  on your cluster\nTo purchase an entitlement to Weave GitOps Enterprise, please contact sales@weave.works.\n", EntitlementSecretName, WGEDefaultNamespace)
		return fmt.Errorf("%s%s", err.Error(), errorMsg)
	}

	// TODO: verify valid entitlement file

	utils.Info(EntitlementCheckMsg)

	return nil
}
