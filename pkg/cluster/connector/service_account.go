package connector

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

func newClusterRole(name, namespace string, rules []rbacv1.PolicyRule) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{Kind: "ClusterRole", APIVersion: "rbac.authorization.k8s.io/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			// Namespace: namespace,
		},
		Rules: rules,
	}

}

func newClusterRoleBinding(name, namespace, roleName, serviceAccountName string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			// Namespace: namespace,
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

// ReconcileServiceAccount accepts a client and the name for a service account.
// A new Service account is created, if one with same name exists that will be used
// A new cluster role and cluster role binding are created, if already existing those will be used
func ReconcileServiceAccount(ctx context.Context, client kubernetes.Interface, serviceAccountName string) ([]byte, error) {
	log := logr.Logger{}
	// When creating service account, create ClusterRole and ClusterRoleBinding
	namespace := corev1.NamespaceDefault //TODO get from options

	serviceAccount, err := client.CoreV1().ServiceAccounts(namespace).Create(ctx, &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: namespace,
		},
	}, metav1.CreateOptions{})

	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return nil, err
		} else {
			serviceAccount, err = client.CoreV1().ServiceAccounts(namespace).Get(ctx, serviceAccountName, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
		}

	}

	// Create ClusterRole
	clusterRoleName := serviceAccount.Name + "-cluster-role"
	clusterRoleObj := newClusterRole(clusterRoleName, namespace, []rbacv1.PolicyRule{
		{
			APIGroups: []string{"*"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		},
	})

	clusterRole, err := client.RbacV1().ClusterRoles().Create(ctx, clusterRoleObj, metav1.CreateOptions{})
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return nil, err
		} else {
			clusterRole, err = client.RbacV1().ClusterRoles().Get(ctx, clusterRoleName, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			if len(clusterRoleObj.Rules) != len(clusterRole.Rules) || !reflect.DeepEqual(clusterRole.Rules, clusterRoleObj.Rules) {
				log.Info("cluster role already exists with a different set of rules", "clusterRole", clusterRole.Name)
			}
		}
	}

	// Create  cluster role binding
	clusterRoleBindingName := serviceAccount.Name + "-cluster-role-binding"
	clusterRoleBindingObj := newClusterRoleBinding(clusterRoleBindingName, namespace, clusterRole.Name, serviceAccountName)
	_, err = client.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBindingObj, metav1.CreateOptions{})
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return nil, err
		} else {
			clusterRoleBinding, err := client.RbacV1().ClusterRoleBindings().Get(ctx, clusterRoleBindingName, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			log.Info("cluster role binding already exists", "clusterRoleBinding", clusterRoleBinding.Name)

		}
	}

	// Create secret
	secretName := serviceAccountName + "-token"
	secretObj := newServiceAccountTokenSecret(secretName, serviceAccountName, namespace)
	_, err = client.CoreV1().Secrets(namespace).Create(ctx, secretObj, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	err = wait.PollUntilContextTimeout(ctx, time.Second, 10*time.Second, true, func(ctx context.Context) (done bool, err error) {
		secret, err := client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if secret.Data != nil && secret.Data["token"] != nil {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	token := secret.Data["token"]
	if token == nil {
		return nil, fmt.Errorf("secret %s/%s was not populated with a token", namespace, secretName)
	}
	return token, nil

}
