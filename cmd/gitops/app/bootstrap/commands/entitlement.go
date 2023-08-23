package commands

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

const (
	ENTITLEMENT_SECRET_NAME = "weave-gitops-enterprise-credentials"
)

// CheckEntitlementFile checks for valid entitlement secret
func CheckEntitlementFile() error {
	utils.Warning("Checking entitlement file ...")

	secret, err := utils.GetSecret(WGE_DEFAULT_NAMESPACE, ENTITLEMENT_SECRET_NAME)
	if err != nil || secret.Data["entitlement"] == nil {
		errorMsg := fmt.Sprintf("✖️  Invalid entitlement file, Please check secret: '%s' under namespace: '%s'  on your cluster\nTo purchase an entitlement to Weave GitOps Enterprise, please contact sales@weave.works.\n", ENTITLEMENT_SECRET_NAME, WGE_DEFAULT_NAMESPACE)
		return utils.CheckIfError(err, errorMsg)
	}

	utils.Info("✔  entitlement file is checked and valid!")

	return nil
}
