package checks

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type promptContent struct {
	errorMsg string
	label    string
}

func promptGetStringInput(pc promptContent) string {
	validate := func(input string) error {
		if input == "" {
			return errors.New(pc.errorMsg)
		}
		return nil
	}

	return getPrompt(pc, validate)
}

func promptGetPasswordInput(pc promptContent) string {
	validate := func(input string) error {
		if len(input) < 6 {
			return errors.New("password must have more than 6 characters")
		}
		return nil
	}
	prompt := promptui.Prompt{
		Label:    pc.label,
		Validate: validate,
		Mask:     '*',
	}

	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return result
}

func getPrompt(pc promptContent, validate promptui.ValidateFunc) string {
	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     pc.label,
		Templates: templates,
		Validate:  validate,
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return result
}

func promptGetSelect(pc promptContent, items []string) string {
	index := -1
	var result string
	var err error

	for index < 0 {
		prompt := promptui.Select{
			Label: pc.label,
			Items: items,
		}

		index, result, err = prompt.Run()

		if index == -1 {
			items = append(items, result)
		}
	}

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Selected: %s\n", result)

	return result
}

func getKubernetesClient() (*kubernetes.Clientset, error) {
	// Path to the kubeconfig file. This is typically located at "~/.kube/config".
	// Obtain the user's home directory.
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Construct the full path to the kubeconfig file.
	kubeconfig := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	// Create a new Kubernetes client using the config.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func getSecret(secretNamespace, secretName string) (*corev1.Secret, error) {
	// Create a new Kubernetes client using the config.
	clientset, err := getKubernetesClient()
	if err != nil {
		panic(err.Error())
	}

	// Fetch the secret from the Kubernetes cluster.
	secret, err := clientset.CoreV1().Secrets(secretNamespace).Get(context.TODO(), secretName, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func createSecret(secretName string, secretNamespace string, secretData map[string][]byte) {
	// Create a new Kubernetes client using the config.
	clientset, err := getKubernetesClient()
	if err != nil {
		panic(err.Error())
	}

	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNamespace,
		},
		Data: secretData,
	}

	_, err = clientset.CoreV1().Secrets(secretNamespace).Create(context.TODO(), secret, v1.CreateOptions{
		TypeMeta: secret.TypeMeta,
	})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		panic(err.Error())
	}
}

// func createConfigMap(Name string, Namespace string, Data map[string]string) {
// 	// Create a new Kubernetes client using the config.
// 	clientset, err := getKubernetesClient()
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	configMap := &corev1.ConfigMap{
// 		ObjectMeta: v1.ObjectMeta{
// 			Name:      Name,
// 			Namespace: Namespace,
// 		},
// 		Data: Data,
// 	}

// 	_, err = clientset.CoreV1().ConfigMaps(Namespace).Create(context.TODO(), configMap, v1.CreateOptions{
// 		TypeMeta: configMap.TypeMeta,
// 	})
// 	if err != nil && !strings.Contains(err.Error(), "already exists") {
// 		panic(err.Error())
// 	}
// }

func isValidBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
