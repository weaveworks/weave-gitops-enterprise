package connector

import (
	"context"
	"fmt"

	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/core/logger"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// getSecretNameFromCluster gets the secret name from the secretref of a
// GitopsCluster given its name and namespace if found.
func getSecretNameFromCluster(ctx context.Context, client dynamic.Interface, scheme *runtime.Scheme, clusterName types.NamespacedName) (string, error) {
	lgr := log.FromContext(ctx)
	resource := gitopsv1alpha1.GroupVersion.WithResource("gitopsclusters")
	u, err := client.Resource(resource).Namespace(clusterName.Namespace).Get(ctx, clusterName.Name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get GitopsCluster %s: %w", clusterName, err)
	}
	lgr.V(logger.LogLevelDebug).Info("remote gitopscluster found", "gitopscluster", clusterName.Name)

	gitopsCluster, err := unstructuredToGitopsCluster(scheme, u)
	if err != nil {
		return "", fmt.Errorf("failed to load GitopsCluster %s: %w", clusterName, err)
	}

	secretName := gitopsCluster.Spec.SecretRef.Name
	if secretName == "" {
		return "", fmt.Errorf("failed to find referenced secret in gitopscluster %s", clusterName)
	}
	lgr.V(logger.LogLevelDebug).Info("referenced secret name found in gitops cluster", "gitopscluster", clusterName, "secret", secretName)

	return secretName, nil
}

// createOrUpdateGitOpsClusterSecret updates/creates the secret with the kubeconfig data given the secret name and namespace of the secret
func createOrUpdateGitOpsClusterSecret(ctx context.Context, client kubernetes.Interface, secretName, namespace string, config *clientcmdapi.Config) (*v1.Secret, error) {
	lgr := log.FromContext(ctx)
	configBytes, err := clientcmd.Write(*config)
	if err != nil {
		return nil, err
	}

	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
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
		secret, err = client.CoreV1().Secrets(namespace).Create(ctx, newSecretObj, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
		lgr.V(logger.LogLevelDebug).Info("new secret with kubeconfig data created", "secret", secretName)

	}

	secret.Data["value"] = configBytes
	updatedSecret, err := client.CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	lgr.V(logger.LogLevelDebug).Info("secret updated with kubeconfig data successfully")

	return updatedSecret, nil

}

func unstructuredToGitopsCluster(scheme *runtime.Scheme, uns *unstructured.Unstructured) (*gitopsv1alpha1.GitopsCluster, error) {
	newObj, err := scheme.New(uns.GetObjectKind().GroupVersionKind())
	if err != nil {
		return nil, err
	}

	return newObj.(*gitopsv1alpha1.GitopsCluster), scheme.Convert(uns, newObj, nil)
}
