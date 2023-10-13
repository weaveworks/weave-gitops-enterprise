package connector

import (
	"context"

	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/core/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/log"
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

func getDynClientAndScheme(config *rest.Config) (dynamic.Interface, *runtime.Scheme, error) {
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	scheme, err := newGitopsClusterScheme()
	if err != nil {
		return nil, nil, err
	}
	return dynClient, scheme, nil
}

func getSecretNameForConfig(ctx context.Context, config *rest.Config, options *ClusterConnectionOptions) (string, error) {
	dynClient, scheme, err := getDynClientAndScheme(config)
	if err != nil {
		return "", err
	}

	return getSecretNameFromCluster(ctx, dynClient, scheme, options.GitopsClusterName)

}

// ConnectCluster connects a cluster to a spoke cluster given its name and context
// Given ClusterOptions, a Service account, Cluster Role, Cluster Role binding and secret are created in the remote cluster and token is used to access
func ConnectCluster(ctx context.Context, options *ClusterConnectionOptions) error {
	lgr := log.FromContext(ctx)
	pathOpts := clientcmd.NewDefaultPathOptions()
	pathOpts.LoadingRules.ExplicitPath = options.ConfigPath

	// load hub kubeconfig
	hubClusterConfig, err := configForContext(ctx, pathOpts, "")
	if err != nil {
		return err
	}
	// Get the context from SpokeClusterContext
	spokeClusterConfig, err := configForContext(ctx, pathOpts, options.RemoteClusterContext)
	if err != nil {
		return err
	}
	secretName, err := getSecretNameForConfig(ctx, hubClusterConfig, options)
	if err != nil {
		return err
	}

	// ReconcileServiceAccount to create the ServiceAccount/ClusterRole/ClusterRoleBinding/Secret
	spokeKubernetesClient, err := kubernetes.NewForConfig(spokeClusterConfig)
	if err != nil {
		return err
	}
	serviceAccountToken, err := ReconcileServiceAccount(ctx, spokeKubernetesClient, *options)
	if err != nil {
		return err
	}

	// Create or update the referenced secret name with the value from the remote cluster ServiceAccount token.
	newConfig, err := kubeConfigWithToken(ctx, spokeClusterConfig, options.RemoteClusterContext, serviceAccountToken)
	if err != nil {
		return err
	}

	hubKubernetesClient, err := kubernetes.NewForConfig(hubClusterConfig)
	if err != nil {
		return err
	}
	_, err = createOrUpdateGitOpsClusterSecret(ctx, hubKubernetesClient, secretName, options.GitopsClusterName.Namespace, newConfig)
	if err != nil {
		return err
	}

	lgr.V(logger.LogLevelInfo).Info("Successfully connected cluster", "cluster", options.GitopsClusterName)

	return nil
}

// newGitopsClusterScheme returns a scheme with the GitopsCluster schema
// information registered.
func newGitopsClusterScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	err := gitopsv1alpha1.AddToScheme(scheme)
	if err != nil {
		return nil, err
	}

	return scheme, nil
}

func deleteGitOpsClusterSecret(ctx context.Context, client kubernetes.Interface, secretName, namespace string) error {
	lgr := log.FromContext(ctx)
	err := client.CoreV1().Secrets(namespace).Delete(ctx, secretName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	lgr.V(logger.LogLevelDebug).Info("gitops cluster secret deleted successfully!", "secret", secretName, "namespace", namespace)
	return nil
}

// DisconnectCluster disconnects a cluster from a spoke cluster given its name and context
// The Service account, Cluster Role binding and secret are deleted in the remote cluster and secret containing token in hub cluster is deleted
func DisconnectCluster(ctx context.Context, options *ClusterConnectionOptions) error {
	lgr := log.FromContext(ctx)
	pathOpts := clientcmd.NewDefaultPathOptions()
	pathOpts.LoadingRules.ExplicitPath = options.ConfigPath

	// load hub kubeconfig
	hubClusterConfig, err := configForContext(ctx, pathOpts, "")
	if err != nil {
		return err
	}

	// Get the context from SpokeClusterContext
	spokeClusterConfig, err := configForContext(ctx, pathOpts, options.RemoteClusterContext)
	if err != nil {
		return err
	}
	secretName, err := getSecretNameForConfig(ctx, hubClusterConfig, options)
	if err != nil {
		return err
	}

	spokeKubernetesClient, err := kubernetes.NewForConfig(spokeClusterConfig)
	if err != nil {
		return err
	}

	managedbyReq, err := labels.NewRequirement("app.kubernetes.io/managed-by", selection.Equals, []string{managedByLabelName})
	if err != nil {
		return err
	}

	selector := labels.NewSelector()
	selector = selector.Add(*managedbyReq)
	err = checkServiceAccountName(ctx, spokeKubernetesClient, options, selector)
	if err != nil {
		return err
	}
	err = checkClusterRoleBindingName(ctx, spokeKubernetesClient, options, selector)
	if err != nil {
		return err
	}

	err = deleteServiceAccountResources(ctx, spokeKubernetesClient, *options)
	if err != nil {
		return err
	}

	hubKubernetesClient, err := kubernetes.NewForConfig(hubClusterConfig)
	if err != nil {
		return err
	}
	err = deleteGitOpsClusterSecret(ctx, hubKubernetesClient, secretName, options.GitopsClusterName.Namespace)
	if err != nil {
		return err
	}

	lgr.V(logger.LogLevelInfo).Info("Successfully disconnected cluster and deleted resources", "cluster", options.GitopsClusterName)

	return nil
}
