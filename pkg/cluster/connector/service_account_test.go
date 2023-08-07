package connector

import (
	"context"
	"fmt"
	syslog "log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

// var (
// 	clusterRoleTypeMeta        = typeMeta("ClusterRole", "rbac.authorization.k8s.io/v1")
// 	clusterRoleBindingTypeMeta = typeMeta("ClusterRoleBinding", "rbac.authorization.k8s.io/v1")
// )

func createClient(kubeconfig *rest.Config) (*kubernetes.Clientset, error) {
	// new kubernetes client
	client := kubernetes.NewForConfigOrDie(kubeconfig)
	return client, nil
}

func newClusterRole(name string, rules []rbacv1.PolicyRule) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{Kind: "ClusterRole", APIVersion: "rbac.authorization.k8s.io/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Rules: rules,
	}

}

// Create secret creates secret if not existance
// wait for it to be populated
// ensure timeouts if failures and recovery
// return token
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

func newClusterRoleBinding(name, namespace, roleKind, roleName, serviceAccountName string) *rbacv1.ClusterRoleBinding {
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
			Kind:     roleKind,
			Name:     roleName,
		},
	}
}

// ReconcileServiceAccount accepts a client and the name for a service account.
//
// It creates the ServiceAccount, and if the service account exists, this is not
// an error.
func ReconcileServiceAccount(ctx context.Context, client kubernetes.Interface, serviceAccountName string) error {
	// When creating service account, create ClusterRole and ClusterRoleBinding
	namespace := "default" //TODO get from options

	serviceAccount, err := client.CoreV1().ServiceAccounts(namespace).Create(ctx, &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": serviceAccountName,
			},
		},
	}, metav1.CreateOptions{})
	// TODO check for api error already exists, if exists then it's same as getting it
	if err != nil {
		// if Service account already exists then no need to recreate it
		if apierrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}

	fmt.Printf("Service Account created: %v\n", serviceAccountName)
	fmt.Printf("Service Account created: %v\n", serviceAccount.Name)

	// Create ClusterRole
	clusterRoleName := serviceAccountName + "-cluster-role" // TODO Rename
	clusterRoleObj := newClusterRole(clusterRoleName, []rbacv1.PolicyRule{
		{
			APIGroups: []string{"*"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		},
	})

	clusterRole, err := client.RbacV1().ClusterRoles().Create(ctx, clusterRoleObj, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("Cluster Role created: %v\n", clusterRoleName) // TODO remove/ change to log

	// Create  cluster role binding
	clusterRoleBindingName := serviceAccountName + "-cluster-role-binding" // TODO rename
	clusterRoleBindingObj := newClusterRoleBinding(clusterRoleBindingName, namespace, clusterRole.Kind, clusterRole.Name, serviceAccountName)
	clusterRoleBinding, err := client.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBindingObj, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("Cluster RoleBinding created: %v\n", clusterRoleBindingName)  // TODO remove/change to log
	fmt.Printf("Cluster RoleBinding created: %v\n", clusterRoleBinding.Name) // TODO remove

	// Create secret
	secretName := serviceAccountName + "-token"
	secretObj := newServiceAccountTokenSecret(secretName, serviceAccountName, namespace)
	_, err = client.CoreV1().Secrets(namespace).Create(ctx, secretObj, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	err = wait.PollUntilContextTimeout(ctx, time.Second, 10*time.Second, true, func(ctx context.Context) (done bool, err error) {
		secret, err := client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		syslog.Printf("RANA!!! the secret = %#v", secret)
		if secret.Data != nil && secret.Data["token"] != nil {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// token := secret.Data["token"]
	fmt.Printf("Secret created: %v\n", secret.Name) // TODO remove

	return nil
}

// GetServiceAccount looks up service account with given name, if found returns it. If not found throws error
func GetServiceAccount(ctx context.Context, client kubernetes.Interface, serviceAccountName string, namespace string) (*v1.ServiceAccount, error) {
	serviceAccount, err := client.CoreV1().ServiceAccounts(namespace).Get(ctx, serviceAccountName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return serviceAccount, nil
}

// Test with non-existing SA
// Test with existing SA
// Look for the fake client in client-go
func TestReconcileServiceAccount(t *testing.T) {
	// Call ReconcileServiceAccount with a fake client and service name
	// If it doesn't fail, load the ServiceAccount with the name, it should exist
	var tests = []struct {
		name string
		// client             corev1.CoreV1Interface
		serviceAccountName     string
		expectedServiceAccount v1.ServiceAccount
		expectedError          error
	}{
		{
			"create new service account",
			"test-service-account",
			v1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service-account",
					Namespace: "default",
					Labels: map[string]string{
						"app": "test-service-account",
					},
				},
			},
			nil,
		},
		{
			"existing service account", // TODO
			"test-service-account",
			v1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service-account",
					Namespace: "default",
					Labels: map[string]string{
						"app": "test-service-account",
					},
				},
			},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := fake.NewSimpleClientset()
			// client := cli.CoreV1()

			ReconcileServiceAccount(context.Background(), cli, "test-service-account")

			serviceAccount, err := GetServiceAccount(context.Background(), cli, "test-service-account", "default")
			assert.NoError(t, err)
			t.Logf("Service account retrieved: %v", serviceAccount)

			// TODO
			// get ccluster role, cluster role binding and verify

			assert.Equal(t, &tt.expectedServiceAccount, serviceAccount, "service account found doesn't match expected")

			expectedClusterRole := newClusterRole(tt.serviceAccountName+"-cluster-role", []rbacv1.PolicyRule{
				{
					APIGroups: []string{"*"},
					Resources: []string{"*"},
					Verbs:     []string{"*"},
				},
			})
			clusterRole, err := cli.RbacV1().ClusterRoles().Get(context.Background(), tt.serviceAccountName+"-cluster-role", metav1.GetOptions{})
			assert.NoError(t, err)
			assert.Equal(t, expectedClusterRole, clusterRole, "cluster role found doesn't match expected")

			expectedClusterRoleBinding := newClusterRoleBinding(tt.serviceAccountName+"-cluster-role-binding", "default", clusterRole.Kind, clusterRole.Name, tt.serviceAccountName)
			clusterRoleBinding, err := cli.RbacV1().ClusterRoleBindings().Get(context.Background(), tt.serviceAccountName+"-cluster-role-binding", metav1.GetOptions{})
			assert.NoError(t, err)
			assert.Equal(t, expectedClusterRoleBinding, clusterRoleBinding, "cluster role found doesn't match expected")

			expectedSecret := newServiceAccountTokenSecret(tt.serviceAccountName+"-token", tt.serviceAccountName, "default")
			expectedSecret.Data = map[string][]byte{
				"token": []byte("usertest"), //TODO
			}

			// time.Sleep(10 * time.Second)
			// go func(secretName, namespace string, token []byte) {
			// 	t.Logf("In go routine")

			// 	secret, _ := cli.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
			// 	secret.Data = map[string][]byte{
			// 		"token": token,
			// 	}
			// 	cli.CoreV1().Secrets("default").Update(context.Background(), secret, metav1.UpdateOptions{})

			// }(tt.serviceAccountName+"-token", "default", []byte("usertest"))

			// secret, err := cli.CoreV1().Secrets("default").Get(context.Background(), tt.serviceAccountName+"-token", metav1.GetOptions{})
			// assert.NoError(t, err)
			// assert.Equal(t, expectedSecret, secret, "secret found doesn't match expected")
		})
	}

}

func TestGetServiceAccount(t *testing.T) {
	// log := logr.Logger{}
	var tests = []struct {
		name               string
		serviceAccountName string
		serviceAccounts    []*v1.ServiceAccount
		expected           v1.ServiceAccount
	}{
		{
			"get exisiting service account",
			"test-service-account",
			[]*v1.ServiceAccount{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-account",
						Namespace: "default",
						Labels: map[string]string{
							"app": "test-service-account",
						},
					},
				},
			},
			v1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service-account",
					Namespace: "default",
					Labels: map[string]string{
						"app": "test-service-account",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := fake.NewSimpleClientset()
			err := addFakeServiceAccounts(cli, tt.serviceAccounts)
			if err != nil {
				t.Errorf("Error adding service accounts: %v", err)
			}

			resServiceAccount, err := GetServiceAccount(context.Background(), cli, "test-service-account", "default")
			t.Logf("service account: %v", resServiceAccount)

			if err != nil {
				t.Errorf("Error getting service account: %v", err)
			}
			// assert.Equal(t, tt.expected, resServiceAccount, "service account found doesn't match expected")

		})
	}

}

func addFakeServiceAccounts(client kubernetes.Interface, serviceAccounts []*v1.ServiceAccount) error {
	for _, serviceAccount := range serviceAccounts {
		_, err := client.CoreV1().ServiceAccounts(serviceAccount.Namespace).Create(context.Background(), serviceAccount, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
