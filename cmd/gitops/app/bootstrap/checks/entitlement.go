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
	entitlementExists := promptGetInput(entitlementCheckPromptContent)
	if !slices.Contains([]string{"Y", "y"}, entitlementExists) {
		fmt.Println("\nPlease apply the entitlement file")
		os.Exit(1)
	}

	//get secret from getSecret()
	secret, err := getSecret(ENTITLEMENT_SECRET_NAMESPACE, ENTITLEMENT_SECRET_NAME)
	if err != nil || secret.Data["entitlement"] == nil {
		fmt.Println("invalid entitlement file!")
		os.Exit(1)
	}
	fmt.Println("✅ entitlement file is checked and valid!")
}
