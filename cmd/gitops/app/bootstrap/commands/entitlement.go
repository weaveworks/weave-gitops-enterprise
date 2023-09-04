package commands

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

const (
	entitlementCheckConfirmMsg  = "Entitlement File exists and is valid!"
	//TODO: change the following messages to be separated into: 
	// 1. Non-existing entitlement secret file : "Entitlement File is not found. To get Weave GitOps Entitelment secret, please contact *sales@weave.works* and add it to your cluster."
	// 2. Invalid entitlement secret: "Entitlement file is invalid, please verify the secret content. If you still facing issues, please contact *sales@weave.works*."
	invalidEntitlementMsgFormat = "\n✖️  Invalid entitlement file, Please check secret: '%s' under namespace: '%s' on your cluster\nTo purchase an entitlement to Weave GitOps Enterprise, please contact sales@weave.works.\n"
	entitlementCheckMsg         = "Verifying Weave GitOps Entitlement File ..."
)

const (
	entitlementSecretName = "weave-gitops-enterprise-credentials"
)

// CheckEntitlementFile checks for valid entitlement secret.
func CheckEntitlementFile() error {
	utils.Warning(entitlementCheckMsg)

	secret, err := utils.GetSecret(entitlementSecretName, wgeDefaultNamespace)
	if err != nil || secret.Data["entitlement"] == nil {
		errorMsg := fmt.Sprintf(invalidEntitlementMsgFormat, entitlementSecretName, wgeDefaultNamespace)
		return fmt.Errorf("%s%s", err.Error(), errorMsg)
	}

	// TODO: verify valid entitlement file

	utils.Info(entitlementCheckConfirmMsg)

	return nil
}
