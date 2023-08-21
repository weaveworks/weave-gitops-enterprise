package connector

import (
	"context"

	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClusterConnectionOptions holds the options to create the resources with such as the target names and namespace
type ClusterConnectionOptions struct {
	// RemoteClusterContext is the name of the context that we are connecting to
	RemoteClusterContext string

	// ConfigPath is the path of the kubeconfig which contains the contexts
	ConfigPath string

	// ServiceAccountName is the name of the service account to be created in
	// the remote cluster.
	ServiceAccountName string

	// ClusterRoleName is the name of the ClusterRole which will be created in
	// the remote cluster.
	ClusterRoleName string

	// ClusterRoleBindingName is the name of the ClusterRoleBinding which will be created in
	// the remote cluster.
	ClusterRoleBindingName string

	// GitopsClusterName references the GitopsCluster that we want to setup the
	// connection to.
	// This GitopsCluster must reference a Secret, and the Secret that is
	// referenced will be created or updated with the ServiceAccount token that
	// is created in the remote cluster.
	GitopsClusterName types.NamespacedName
}

func getSecretNameForConfig(ctx context.Context, config *rest.Config, options *ClusterConnectionOptions) (string, error) {
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return "", err
	}
	scheme, err := NewGitopsClusterScheme()
	if err != nil {
		return "", err
	}

	return getSecretNameFromCluster(ctx, dynClient, scheme, options.GitopsClusterName)

}

// ConnectCluster connects a cluster to a remote cluster given its name and context
// Given ClusterOptions, a Service account, Cluster Role, Cluster Role binding and secret are created in the remote cluster and token is used to access
func ConnectCluster(ctx context.Context, options *ClusterConnectionOptions) error {
	// Get the context from RemoteClusterContext
	pathOpts := clientcmd.NewDefaultPathOptions()
	pathOpts.LoadingRules.ExplicitPath = options.ConfigPath
	remoteClusterConfig, err := ConfigForContext(ctx, pathOpts, options.RemoteClusterContext)
	if err != nil {
		return err
	}

	secretName, err := getSecretNameForConfig(ctx, remoteClusterConfig, options)
	if err != nil {
		return err
	}

	// ReconcileServiceAccount to create the ServiceAccount/ClusterRole/ClusterRoleBinding/Secret
	kubernetesClient, err := kubernetes.NewForConfig(remoteClusterConfig)
	if err != nil {
		return err
	}
	serviceAccountToken, err := ReconcileServiceAccount(ctx, kubernetesClient, *options)
	if err != nil {
		return err
	}

	// Create or update the referenced secret name with the value from the remote cluster ServiceAccount token.
	newConfig, err := kubeConfigWithToken(ctx, remoteClusterConfig, options.RemoteClusterContext, serviceAccountToken)
	if err != nil {
		return err
	}
	_, err = createOrUpdateGitOpsClusterSecret(ctx, kubernetesClient, secretName, options.GitopsClusterName.Namespace, newConfig)
	if err != nil {
		return err
	}

	return nil
}

// NewGitopsClusterScheme returns a scheme with the GitopsCluster schema
// information registered.
func NewGitopsClusterScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	err := gitopsv1alpha1.AddToScheme(scheme)
	if err != nil {
		return nil, err
	}

	return scheme, nil
}
