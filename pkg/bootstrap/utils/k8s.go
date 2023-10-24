package utils

import (
	"context"
	"flag"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"
	k8s_config "sigs.k8s.io/controller-runtime/pkg/client/config"
)

// GetKubernetesHttp creates a kuberentes client from the default kubeconfig.
func GetKubernetesHttp(kubeconfig string) (*kube.KubeHTTP, error) {
	if kubeconfig != "" {
		err := flag.CommandLine.Set("kubeconfig", kubeconfig)
		if err != nil {
			return nil, err
		}
		flag.Parse()
		k8s_config.RegisterFlags(flag.CommandLine)
	}

	config, err := k8s_config.GetConfig()
	if err != nil {
		return nil, err
	}

	return kube.NewKubeHTTPClientWithConfig(config, config.Host)
}

// GetSecret get secret values from kubernetes.
func GetSecret(client k8s_client.Client, name string, namespace string) (*corev1.Secret, error) {
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
func CreateSecret(client k8s_client.Client, name string, namespace string, data map[string][]byte) error {
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
func DeleteSecret(client k8s_client.Client, name string, namespace string) error {
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
