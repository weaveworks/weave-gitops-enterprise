package utils

import (
	"context"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// GetKubernetesClient creates a kuberentes client from the default kubeconfig
func GetKubernetesClient() (*kubernetes.Clientset, error) {
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

// GetSecret get secret values from kubernetes
func GetSecret(secretNamespace, secretName string) (*corev1.Secret, error) {
	// Create a new Kubernetes client using the config.
	clientset, err := GetKubernetesClient()
	if err != nil {
		return nil, err
	}

	// Fetch the secret from the Kubernetes cluster.
	secret, err := clientset.CoreV1().Secrets(secretNamespace).Get(context.TODO(), secretName, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return secret, nil
}

// CreateSecret create a kubernetes secret
func CreateSecret(secretName string, secretNamespace string, secretData map[string][]byte) error {
	// Create a new Kubernetes client using the config.
	clientset, err := GetKubernetesClient()
	if err != nil {
		return err
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
	if err != nil {
		return err
	}
	return nil
}

// DeleteSecret delete a kubernetes secret
func DeleteSecret(secretName string, secretNamespace string) error {
	// Create a new Kubernetes client using the config.
	clientset, err := GetKubernetesClient()
	if err != nil {
		return err
	}

	err = clientset.CoreV1().Secrets(secretNamespace).Delete(context.TODO(), secretName, v1.DeleteOptions{
		TypeMeta: v1.TypeMeta{},
	})
	if err != nil {
		return err
	}
	return nil
}
