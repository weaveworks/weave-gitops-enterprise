package connector

import (
	"context"
	"encoding/json"
	"fmt"

	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// getSecretNameFromCluster gets the secret name from the secretref of a gitops cluster given its name and namespace if found
func getSecretNameFromCluster(client *dynamic.DynamicClient, clusterName, namespace string) (string, error) {
	resource := gitopsv1alpha1.GroupVersion.WithResource("gitopscluster")
	fmt.Printf("error %v", client.Resource(gitopsv1alpha1.SchemeBuilder.GroupVersion.WithResource("gitopscluster")))
	u, err := client.Resource(resource).Namespace(namespace).Get(context.Background(), clusterName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	gitopsCluster := capiv1_protos.GitopsCluster{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), &gitopsCluster)
	if err != nil {
		return "", err
	}

	secretName := gitopsCluster.SecretRef.Name
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
