package connector

import (
	"context"
	"fmt"
	"reflect"
	"time"

	// "github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ReconcileServiceAccount accepts a client and the name for a service account.
// A new Service account is created, if one with same name exists that will be used
// A new cluster role and cluster role binding are created, if already existing those will be used
// returns the token of the secret created for the service account
func ReconcileServiceAccount(ctx context.Context, client kubernetes.Interface, clusterConnectionOpts ClusterConnectionOptions) ([]byte, error) {
	namespace := clusterConnectionOpts.GitopsClusterName.Namespace

	err := createServiceAccount(ctx, client, clusterConnectionOpts)
	if err != nil {
		return nil, err
	}

	err = createClusterRole(ctx, client, clusterConnectionOpts)
	if err != nil {
		return nil, err
	}

	err = createClusterRoleBinding(ctx, client, clusterConnectionOpts)
	if err != nil {
		return nil, err
	}

	secret, err := createSecret(ctx, client, clusterConnectionOpts)
	if err != nil {
		return nil, err
	}

	// wait for token to be populated in secret
	err = wait.PollUntilContextTimeout(ctx, time.Second, 10*time.Second, true, func(ctx context.Context) (done bool, err error) {
		logger := log.FromContext(ctx)
		logger.Info("waiting for service account secret token to be populated...")
		secret, err := client.CoreV1().Secrets(namespace).Get(ctx, secret.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if secret.Data != nil && secret.Data["token"] != nil {
			logger.Info("service account secret token populated", "secret", secret.Name)
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	// Get populated secret
	secret, err = client.CoreV1().Secrets(namespace).Get(ctx, secret.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	token := secret.Data["token"]
	if token == nil {
		return nil, fmt.Errorf("secret %s/%s was not populated with a token", namespace, secret.Name)
	}
	return token, nil

}

func newClusterRole(name, namespace string, rules []rbacv1.PolicyRule) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{Kind: "ClusterRole", APIVersion: "rbac.authorization.k8s.io/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Rules: rules,
	}

}

func newClusterRoleBinding(name, namespace, roleName, serviceAccountName string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     roleName,
		},
	}
}

func newServiceAccountTokenSecret(name, serviceAccountName, namespace string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				corev1.ServiceAccountNameKey: serviceAccountName,
			},
		},
		Type: corev1.SecretTypeServiceAccountToken,
	}

}

func createServiceAccount(ctx context.Context, client kubernetes.Interface, clusterConnectionOpts ClusterConnectionOptions) error {
	logger := log.FromContext(ctx)
	serviceAccountName := clusterConnectionOpts.ServiceAccountName
	namespace := clusterConnectionOpts.GitopsClusterName.Namespace

	_, err := client.CoreV1().ServiceAccounts(namespace).Create(ctx, &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: namespace,
		},
	}, metav1.CreateOptions{})

	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		_, err = client.CoreV1().ServiceAccounts(namespace).Get(ctx, serviceAccountName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		logger.Info("service account already exists", "serviceaccount", serviceAccountName)
	} else {
		logger.Info("service account created successfully!", "serviceaccount", serviceAccountName)
	}
	return nil

}

func createClusterRole(ctx context.Context, client kubernetes.Interface, clusterConnectionOpts ClusterConnectionOptions) error {
	logger := log.FromContext(ctx)
	namespace := clusterConnectionOpts.GitopsClusterName.Namespace
	clusterRoleName := clusterConnectionOpts.ClusterRoleName

	clusterAccessRules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{"*"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		},
	}
	clusterRoleObj := newClusterRole(clusterRoleName, namespace, clusterAccessRules)

	_, err := client.RbacV1().ClusterRoles().Create(ctx, clusterRoleObj, metav1.CreateOptions{})
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		} else {
			clusterRole, err := client.RbacV1().ClusterRoles().Get(ctx, clusterRoleName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(clusterRole.Rules, clusterRoleObj.Rules) {
				logger.Info("cluster role already exists with a different set of rules", "clusterRole", clusterRole.Name)
			}
		}
	} else {
		logger.Info("cluster role created successfully!", "clusterrole", clusterRoleName)
	}
	return nil
}

func createClusterRoleBinding(ctx context.Context, client kubernetes.Interface, clusterConnectionOpts ClusterConnectionOptions) error {
	logger := log.FromContext(ctx)
	serviceAccountName := clusterConnectionOpts.ServiceAccountName
	clusterRoleName := clusterConnectionOpts.ClusterRoleName
	clusterRoleBindingName := clusterConnectionOpts.ClusterRoleBindingName
	namespace := clusterConnectionOpts.GitopsClusterName.Namespace

	clusterRoleBindingObj := newClusterRoleBinding(clusterRoleBindingName, namespace, clusterRoleName, serviceAccountName)
	_, err := client.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBindingObj, metav1.CreateOptions{})
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		} else {
			clusterRoleBinding, err := client.RbacV1().ClusterRoleBindings().Get(ctx, clusterRoleBindingName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			logger.Info("cluster role binding already exists", "clusterRoleBinding", clusterRoleBinding.Name)
		}
	} else {
		logger.Info("cluster role binding created successfully!", "clusterrolebinding", clusterRoleBindingName)
	}
	return nil
}

func createSecret(ctx context.Context, client kubernetes.Interface, clusterConnectionOpts ClusterConnectionOptions) (*corev1.Secret, error) {
	logger := log.FromContext(ctx)
	serviceAccountName := clusterConnectionOpts.ServiceAccountName
	namespace := clusterConnectionOpts.GitopsClusterName.Namespace

	secretName := serviceAccountName + "-token"
	secretObj := newServiceAccountTokenSecret(secretName, serviceAccountName, namespace)
	secret, err := client.CoreV1().Secrets(namespace).Create(ctx, secretObj, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	logger.Info("service account secret created successfully!")

	return secret, nil

}
