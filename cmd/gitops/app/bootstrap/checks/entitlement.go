package checks

import (
	"fmt"
	"os"

	"golang.org/x/exp/slices"
)

const ENTITLEMENT_SECRET_NAME string = "weave-gitops-enterprise-credentials"
const ENTITLEMENT_SECRET_NAMESPACE string = "flux-system"

func CheckEntitlementFile() {
	entitlementCheckPromptContent := promptContent{
		"Please provide an answer with (y/n).",
		"Do you have a valid entitlment file on your cluster (y/n)?",
	}
	entitlementExists := promptGetBoolInput(entitlementCheckPromptContent)
	if !slices.Contains([]string{"Y", "y"}, entitlementExists) {
		fmt.Println("\nPlease apply the entitlement file\nTo purchase an entitlement to Weave GitOps Enterprise, please contact sales@weave.works.")
		os.Exit(1)
	}

	//get secret from getSecret()
	secret, err := getSecret(ENTITLEMENT_SECRET_NAMESPACE, ENTITLEMENT_SECRET_NAME)
	if err != nil || secret.Data["entitlement"] == nil {
		fmt.Printf("\nInvalid entitlement file, Please check secret: '%s' under namespace: '%s'  on your cluster\n", ENTITLEMENT_SECRET_NAME, ENTITLEMENT_SECRET_NAMESPACE)
		os.Exit(1)
	}
	fmt.Println("✔ entitlement file is checked and valid!")
}
