package utils

import (
	"context"
	"os"
	"path/filepath"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"
)

// GetKubernetesClient creates a kuberentes client from the default kubeconfig.
func GetKubernetesClient(kubeconfig string) (k8s_client.Client, error) {
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

	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		sourcev1.AddToScheme,
		kustomizev1.AddToScheme,
	}

	err = schemeBuilder.AddToScheme(scheme)
	if err != nil {
		return nil, err
	}

	client, err := k8s_client.New(config, k8s_client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	return client, nil
}

// GetSecret get secret values from kubernetes.
func GetSecret(name string, namespace string, client k8s_client.Client) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	err := client.Get(context.Background(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, secret, &k8s_client.GetOptions{})

	if err != nil {
		return nil, err
	}

	return secret, nil
}

// CreateSecret create a kubernetes secret.
func CreateSecret(name string, namespace string, data map[string][]byte, client k8s_client.Client) error {
	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}

	err := client.Create(context.Background(), secret, &k8s_client.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

// DeleteSecret delete a kubernetes secret.
func DeleteSecret(name string, namespace string, client k8s_client.Client) error {
	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := client.Delete(context.Background(), secret, &k8s_client.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}
