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

// GetKubernetesClient creates a kuberentes client from the default kubeconfig.
func GetKubernetesClient(kubeconfig string) (*kubernetes.Clientset, error) {
	if kubeconfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// GetSecret get secret values from kubernetes.
func GetSecret(name string, namespace string, clientset kubernetes.Interface) (*corev1.Secret, error) {
	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return secret, nil
}

// CreateSecret create a kubernetes secret.
func CreateSecret(name string, namespace string, data map[string][]byte, clientset kubernetes.Interface) error {
	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}

	if _, err := clientset.CoreV1().Secrets(namespace).Create(context.Background(), secret, v1.CreateOptions{
		TypeMeta: secret.TypeMeta,
	}); err != nil {
		return err
	}

	return nil
}

// DeleteSecret delete a kubernetes secret.
func DeleteSecret(name string, namespace string, clientset kubernetes.Interface) error {
	if err := clientset.CoreV1().Secrets(namespace).Delete(context.Background(), name, v1.DeleteOptions{
		TypeMeta: v1.TypeMeta{},
	}); err != nil {
		return err
	}

	return nil
}
