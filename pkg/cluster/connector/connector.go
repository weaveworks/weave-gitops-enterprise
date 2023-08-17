package connector

import (
	"context"

	"github.com/go-logr/logr"
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ClusterConnectionOptions holds the options to create the resources with such as the target names and namespace
type ClusterConnectionOptions struct {
	// RemoteClusterContext is the name of the context that we are connecting to
	RemoteClusterContext string

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
	GitopsClusterName runtime.NamespacedName
}

func ConnectCluster(ctx context.Context, logger logr.Logger, options *ClusterConnectionOptions) error {
	// 1. Get the gitopsCluster secret name
	//   If this fails, error appropriately, differentiate between cluster not
	//   found, and cluster does not reference a secret.
	// 2. Get the context from RemoteClusterContext - error if it doesn't exist.
	// 3. Create the ClusterRole/ClusterRoleBinding/ServiceAccount/Secret.
	// 4. Wait for the secret to be populated and get the value
	// 5. Create or update the referenced secret name with the value from the
	// remote cluster ServiceAccount token.
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
