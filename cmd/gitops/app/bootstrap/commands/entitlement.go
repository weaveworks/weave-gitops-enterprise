package commands

import (
	"fmt"
	"os"
)

const ENTITLEMENT_SECRET_NAME string = "weave-gitops-enterprise-credentials"
const ENTITLEMENT_SECRET_NAMESPACE string = "flux-system"

func CheckEntitlementFile() {
	fmt.Println("Checking entitlement file ...")

	secret, err := getSecret(ENTITLEMENT_SECRET_NAMESPACE, ENTITLEMENT_SECRET_NAME)
	if err != nil || secret.Data["entitlement"] == nil {
		fmt.Printf("✖️  Invalid entitlement file, Please check secret: '%s' under namespace: '%s'  on your cluster\nTo purchase an entitlement to Weave GitOps Enterprise, please contact sales@weave.works.\n", ENTITLEMENT_SECRET_NAME, ENTITLEMENT_SECRET_NAMESPACE)
		os.Exit(1)
	}
	fmt.Println("✔  entitlement file is checked and valid!")
}
