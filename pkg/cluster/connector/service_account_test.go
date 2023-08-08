package connector

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestReconcileServiceAccount(t *testing.T) {
	var tests = []struct {
		name                   string
		existingResources      []runtime.Object
		serviceAccountName     string
		expectedServiceAccount v1.ServiceAccount
		expectedResources      map[string]runtime.Object // Should include expected ServiceAccount,ClusterRole, ClusterRoleBinding
		expectedError          error
	}{
		{
			"create new service account",
			nil,
			"test-service-account",
			v1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service-account",
					Namespace: corev1.NamespaceDefault,
				},
			},
			map[string]runtime.Object{
				"ServiceAccount": &v1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-account",
						Namespace: corev1.NamespaceDefault,
					},
				},
				"ClusterRole": newClusterRole("test-service-account-cluster-role", corev1.NamespaceDefault, []rbacv1.PolicyRule{

					{
						APIGroups: []string{"*"},
						Resources: []string{"*"},
						Verbs:     []string{"*"},
					},
				}),
				"ClusterRoleBinding": newClusterRoleBinding("test-service-account-cluster-role-binding", corev1.NamespaceDefault, "test-service-account-cluster-role", "test-service-account"),
			},
			nil,
		},
		{
			"existing service account",
			[]runtime.Object{
				&v1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-account",
						Namespace: corev1.NamespaceDefault,
					},
				},
			},
			"test-service-account",
			v1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service-account",
					Namespace: corev1.NamespaceDefault,
				},
			},
			map[string]runtime.Object{
				"ServiceAccount": &v1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-account",
						Namespace: corev1.NamespaceDefault,
					},
				},
				"ClusterRole": newClusterRole("test-service-account-cluster-role", corev1.NamespaceDefault, []rbacv1.PolicyRule{
					{
						APIGroups: []string{"*"},
						Resources: []string{"*"},
						Verbs:     []string{"*"},
					},
				}),
				"ClusterRoleBinding": newClusterRoleBinding("test-service-account-cluster-role-binding", corev1.NamespaceDefault, "test-service-account-cluster-role", "test-service-account"),
			},
			nil,
		},
		{
			"existing cluster role with different rules than expected",
			[]runtime.Object{
				&rbacv1.ClusterRole{
					TypeMeta: metav1.TypeMeta{Kind: "ClusterRole", APIVersion: "rbac.authorization.k8s.io/v1"},
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-service-account-cluster-role",
					},
					Rules: []rbacv1.PolicyRule{
						{
							APIGroups: []string{"*"},
						},
					},
				},
			},
			"test-service-account",
			v1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service-account",
					Namespace: corev1.NamespaceDefault,
				},
			},
			map[string]runtime.Object{
				"ServiceAccount": &v1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-account",
						Namespace: corev1.NamespaceDefault,
					},
				},
				"ClusterRole": newClusterRole("test-service-account-cluster-role", corev1.NamespaceDefault, []rbacv1.PolicyRule{
					{
						APIGroups: []string{"*"},
					},
				}),
				"ClusterRoleBinding": newClusterRoleBinding("test-service-account-cluster-role-binding", corev1.NamespaceDefault, "test-service-account-cluster-role", "test-service-account"),
			},
			nil,
		},
		{
			"existing cluster role binding",
			[]runtime.Object{
				newClusterRoleBinding("test-service-account-cluster-role-binding", corev1.NamespaceDefault, "existing-cluster-role", "test-service-account"),
			},
			"test-service-account",
			v1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service-account",
					Namespace: corev1.NamespaceDefault,
				},
			},
			map[string]runtime.Object{
				"ServiceAccount": &v1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-account",
						Namespace: corev1.NamespaceDefault,
					},
				},
				"ClusterRole": newClusterRole("test-service-account-cluster-role", corev1.NamespaceDefault, []rbacv1.PolicyRule{
					{
						APIGroups: []string{"*"},
						Resources: []string{"*"},
						Verbs:     []string{"*"},
					},
				}),
				"ClusterRoleBinding": newClusterRoleBinding("test-service-account-cluster-role-binding", corev1.NamespaceDefault, "existing-cluster-role", "test-service-account"),
			},
			nil,
		},
		{
			"existing cluster role and cluster role binding",
			[]runtime.Object{
				newClusterRole("test-service-account-cluster-role", corev1.NamespaceDefault, []rbacv1.PolicyRule{
					{
						APIGroups: []string{"*"},
					},
				}),
				newClusterRoleBinding("test-service-account-cluster-role-binding", corev1.NamespaceDefault, "test-service-account-cluster-role", "test-service-account"),
			},
			"test-service-account",
			v1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service-account",
					Namespace: corev1.NamespaceDefault,
				},
			},
			map[string]runtime.Object{
				"ServiceAccount": &v1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-account",
						Namespace: corev1.NamespaceDefault,
					},
				},
				"ClusterRole": newClusterRole("test-service-account-cluster-role", corev1.NamespaceDefault, []rbacv1.PolicyRule{
					{
						APIGroups: []string{"*"},
					},
				}),
				"ClusterRoleBinding": newClusterRoleBinding("test-service-account-cluster-role-binding", corev1.NamespaceDefault, "test-service-account-cluster-role", "test-service-account"),
			},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remoteClientSet := fake.NewSimpleClientset()

			// This is artificial where it populates the token of the secret as kubernetes isn't running. Kubernetes should populate it once the secret is created.
			go func(secretName, namespace string, token []byte) {
				if err := wait.PollUntilContextTimeout(context.Background(), time.Second, 10*time.Second, true, func(ctx context.Context) (done bool, err error) {
					secret, err := remoteClientSet.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
					if err != nil {
						if apierrors.IsNotFound(err) {
							return false, nil
						}
						return false, err
					}

					secret.Data = map[string][]byte{
						"token": token,
					}
					if _, err := remoteClientSet.CoreV1().Secrets(namespace).Update(context.Background(), secret, metav1.UpdateOptions{}); err != nil {
						return false, err
					}

					return true, nil
				}); err != nil {
					t.Logf("failed to update secret with token: %s", err)
				}
			}(tt.serviceAccountName+"-token", corev1.NamespaceDefault, []byte("usertest"))

			err := addFakeResources(remoteClientSet, tt.existingResources)
			if err != nil {
				t.Errorf("error adding resources: %v", err)
			}
			// Reconcile Service account
			saToken, err := ReconcileServiceAccount(context.Background(), remoteClientSet, "test-service-account")
			assert.NoError(t, err)
			assert.Equal(t, []byte("usertest"), saToken, "service account token doesn't match expected")

			// Verify Service account created/exists
			serviceAccount, err := remoteClientSet.CoreV1().ServiceAccounts(corev1.NamespaceDefault).Get(context.Background(), "test-service-account", metav1.GetOptions{})
			assert.NoError(t, err)
			expectedServiceAccount := tt.expectedResources["ServiceAccount"].(*v1.ServiceAccount)
			assert.Equal(t, expectedServiceAccount, serviceAccount, "service account found doesn't match expected")

			// Verify ClusterRole created/exists
			expectedClusterRole := tt.expectedResources["ClusterRole"].(*rbacv1.ClusterRole)
			clusterRole, err := remoteClientSet.RbacV1().ClusterRoles().Get(context.Background(), tt.serviceAccountName+"-cluster-role", metav1.GetOptions{})
			assert.NoError(t, err)
			assert.Equal(t, expectedClusterRole, clusterRole, "cluster role found doesn't match expected")

			// Verfiy ClusterRoleBinding created/exists
			expectedClusterRoleBinding := tt.expectedResources["ClusterRoleBinding"].(*rbacv1.ClusterRoleBinding)
			clusterRoleBinding, err := remoteClientSet.RbacV1().ClusterRoleBindings().Get(context.Background(), tt.serviceAccountName+"-cluster-role-binding", metav1.GetOptions{})
			assert.NoError(t, err)
			assert.Equal(t, expectedClusterRoleBinding, clusterRoleBinding, "cluster role found doesn't match expected")

			// Verify Secret created with populated token(token is fake in test)
			expectedSecret := newServiceAccountTokenSecret(tt.serviceAccountName+"-token", tt.serviceAccountName, corev1.NamespaceDefault)
			expectedSecret.Data = map[string][]byte{
				"token": []byte("usertest"),
			}

			secret, err := remoteClientSet.CoreV1().Secrets(corev1.NamespaceDefault).Get(context.Background(), tt.serviceAccountName+"-token", metav1.GetOptions{})
			assert.NoError(t, err)
			assert.Equal(t, expectedSecret, secret, "secret found doesn't match expected")
		})
	}

}

func TestGetServiceAccount(t *testing.T) {
	var tests = []struct {
		name               string
		serviceAccountName string
		serviceAccounts    []string
		expected           v1.ServiceAccount
	}{
		{
			"get exisiting service account",
			"test-service-account",
			[]string{
				"test-service-account",
			},
			v1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service-account",
					Namespace: corev1.NamespaceDefault,
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

			resServiceAccount, err := cli.CoreV1().ServiceAccounts(corev1.NamespaceDefault).Get(context.Background(), "test-service-account", metav1.GetOptions{})
			if err != nil {
				t.Errorf("Error getting service account: %v", err)
			}
			assert.Equal(t, tt.expected, *resServiceAccount, "service account found doesn't match expected")

		})
	}

}

// Add resources of different types to the client based on the type of the resource
// Valid resources: v1.ServiceAccount, rbacv1.ClusterRole, rbacv1.ClusterRoleBinding, v1.Secret
func addFakeResources(client kubernetes.Interface, resources []runtime.Object) error {
	if resources == nil {
		return nil
	}
	for _, resource := range resources {
		switch reflect.TypeOf(resource) {

		case reflect.TypeOf(&v1.ServiceAccount{}):
			_, err := client.CoreV1().ServiceAccounts(corev1.NamespaceDefault).Create(context.Background(), resource.(*v1.ServiceAccount), metav1.CreateOptions{})
			if err != nil {
				return err
			}
		case reflect.TypeOf(&rbacv1.ClusterRole{}):
			_, err := client.RbacV1().ClusterRoles().Create(context.Background(), resource.(*rbacv1.ClusterRole), metav1.CreateOptions{})
			if err != nil {
				return err
			}
		case reflect.TypeOf(&rbacv1.ClusterRoleBinding{}):
			_, err := client.RbacV1().ClusterRoleBindings().Create(context.Background(), resource.(*rbacv1.ClusterRoleBinding), metav1.CreateOptions{})
			if err != nil {
				return err
			}
		case reflect.TypeOf(&v1.Secret{}):
			_, err := client.CoreV1().Secrets(corev1.NamespaceDefault).Create(context.Background(), resource.(*v1.Secret), metav1.CreateOptions{})
			if err != nil {
				return err
			}

		}

	}
	return nil
}

func addFakeServiceAccounts(client kubernetes.Interface, serviceAccounts []string) error {
	if serviceAccounts == nil {
		return nil
	}
	for _, serviceAccountName := range serviceAccounts {
		serviceAccount := &v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceAccountName,
				Namespace: corev1.NamespaceDefault,
			},
		}

		_, err := client.CoreV1().ServiceAccounts(serviceAccount.Namespace).Create(context.Background(), serviceAccount, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
