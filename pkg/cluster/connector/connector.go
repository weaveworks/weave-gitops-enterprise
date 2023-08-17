package connector

import (
	"context"
	"encoding/json"
	"fmt"

	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// getSecretNameFromCluster gets the secret name from the secretref of a
// GitopsCluster given its name and namespace if found.
func getSecretNameFromCluster(ctx context.Context, client dynamic.Interface, scheme *runtime.Scheme, clusterName, namespace string) (string, error) {
	resource := gitopsv1alpha1.GroupVersion.WithResource("gitopsclusters")
	u, err := client.Resource(resource).Namespace(namespace).Get(ctx, clusterName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	gitopsCluster, err := unstructuredToGitopsCluster(scheme, u)
	if err != nil {
		return "", fmt.Errorf("failed to load GitopsCluster %s/%s: %w", namespace, clusterName, err)
	}

	secretName := gitopsCluster.Spec.SecretRef.Name
	if secretName != "" {
		return secretName, nil
	}

	return secretName, nil
}

// secretWithKubeconfig updates/creates the secret with the kubeconfig data given the secret name and namespace of the secret
func secretWithKubeconfig(client kubernetes.Interface, secretName, namespace string, config *clientcmdapi.Config) (*v1.Secret, error) {
	configBytes, err := json.Marshal(config)
	// configStr, err := clientcmd.NewClientConfigFromBytes(configBytes)
	if err != nil {
		return nil, err
	}

	secret, err := client.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, err
		}
		newSecretObj := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: namespace,
			},
			Data: map[string][]byte{
				"value": configBytes,
			},
		}
		secret, err = client.CoreV1().Secrets(namespace).Create(context.Background(), newSecretObj, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	}

	secret.Data["value"] = configBytes
	updatedSecret, err := client.CoreV1().Secrets(namespace).Update(context.Background(), secret, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return updatedSecret, nil

}

func unstructuredToGitopsCluster(scheme *runtime.Scheme, uns *unstructured.Unstructured) (*gitopsv1alpha1.GitopsCluster, error) {
	newObj, err := scheme.New(uns.GetObjectKind().GroupVersionKind())
	if err != nil {
		return nil, err
	}

	return newObj.(*gitopsv1alpha1.GitopsCluster), scheme.Convert(uns, newObj, nil)
}
