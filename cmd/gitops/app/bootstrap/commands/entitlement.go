package commands

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

const (
	entitlementCheckConfirmMsg  = "entitlement file is checked and valid!"
	invalidEntitlementMsgFormat = "\n✖️  Invalid entitlement file, Please check secret: '%s' under namespace: '%s' on your cluster\nTo purchase an entitlement to Weave GitOps Enterprise, please contact sales@weave.works.\n"
	entitlementCheckMsg         = "Checking entitlement file ..."
)
const EntitlementSecretName = "weave-gitops-enterprise-credentials"

// CheckEntitlementFile checks for valid entitlement secret
func CheckEntitlementFile() error {
	utils.Warning(entitlementCheckMsg)

	secret, err := utils.GetSecret(EntitlementSecretName, WGEDefaultNamespace)
	if err != nil || secret.Data["entitlement"] == nil {
		errorMsg := fmt.Sprintf(invalidEntitlementMsgFormat, EntitlementSecretName, WGEDefaultNamespace)
		return fmt.Errorf("%s%s", err.Error(), errorMsg)
	}

	// TODO: verify valid entitlement file

	utils.Info(entitlementCheckConfirmMsg)

	return nil
}
