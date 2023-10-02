package connector

import (
	"context"
	"fmt"
	"maps"

	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/core/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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
	scheme, err := NewGitopsClusterScheme()
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

	err = addOptionsToGitOpsClusterLabel(ctx, hubClusterConfig, options.GitopsClusterName, options)
	if err != nil {
		return err
	}

	lgr.V(logger.LogLevelInfo).Info("Successfully connected cluster", "cluster", options.GitopsClusterName)

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

// addOptionsToGitOpsClusterLabel updates the GitopsCluster with the new labels
func addOptionsToGitOpsClusterLabel(ctx context.Context, config *rest.Config, clusterName types.NamespacedName, options *ClusterConnectionOptions) error {
	lgr := log.FromContext(ctx)
	client, scheme, err := getDynClientAndScheme(config)
	if err != nil {
		return err
	}

	newLabels := map[string]string{
		"clusters.weave.works/connect-cluster-service-account":      options.ServiceAccountName,
		"clusters.weave.works/connect-cluster-cluster-role-binding": options.ClusterRoleBindingName,
	}

	resource := gitopsv1alpha1.GroupVersion.WithResource("gitopsclusters")
	u, err := client.Resource(resource).Namespace(clusterName.Namespace).Get(ctx, clusterName.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get GitopsCluster %s: %w", clusterName, err)
	}

	gitopsCluster, err := unstructuredToGitopsCluster(scheme, u)
	if err != nil {
		return err
	}
	if gitopsCluster.Labels != nil {
		maps.Copy(gitopsCluster.Labels, newLabels)
	} else {
		gitopsCluster.Labels = newLabels
	}

	newUnstructured := unstructured.Unstructured{}
	err = scheme.Convert(gitopsCluster, &newUnstructured, nil)
	if err != nil {
		return err
	}

	_, err = client.Resource(resource).Namespace(clusterName.Namespace).Update(ctx, &newUnstructured, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	lgr.V(logger.LogLevelDebug).Info("Updated Gitopscluster with cluster connection options", "cluster", clusterName.Name)

	return nil

}
