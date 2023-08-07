package checks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/exp/slices"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
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

	path := filepath.Join(homedir.HomeDir(), ".kube", "config")
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		panic(err.Error())
	}
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	secret, err := clientset.CoreV1().Secrets(ENTITLEMENT_SECRET_NAMESPACE).Get(context.TODO(), ENTITLEMENT_SECRET_NAME, v1.GetOptions{})
	if err != nil || secret.Data["entitlement"] == nil {
		fmt.Println("invalid entitlement file!")
		os.Exit(1)
	}
	fmt.Println("✅ entitlement file is checked and valid!")
}
