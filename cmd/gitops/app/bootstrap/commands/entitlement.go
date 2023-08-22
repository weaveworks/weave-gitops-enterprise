package commands

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

const ENTITLEMENT_SECRET_NAME string = "weave-gitops-enterprise-credentials"
const ENTITLEMENT_SECRET_NAMESPACE string = "flux-system"

func CheckEntitlementFile() error {
	fmt.Println("Checking entitlement file ...")

	secret, err := utils.GetSecret(ENTITLEMENT_SECRET_NAMESPACE, ENTITLEMENT_SECRET_NAME)
	if err != nil || secret.Data["entitlement"] == nil {
		errorMsg := fmt.Sprintf("✖️  Invalid entitlement file, Please check secret: '%s' under namespace: '%s'  on your cluster\nTo purchase an entitlement to Weave GitOps Enterprise, please contact sales@weave.works.\n", ENTITLEMENT_SECRET_NAME, ENTITLEMENT_SECRET_NAMESPACE)
		return utils.CheckIfError(err, errorMsg)

	}
	fmt.Println("✔  entitlement file is checked and valid!")
	return nil
}
