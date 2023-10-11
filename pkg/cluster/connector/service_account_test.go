package connector

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestReconcileServiceAccount(t *testing.T) {
	var tests = []struct {
		name                   string
		existingResources      []runtime.Object
		serviceAccountName     string
		expectedServiceAccount corev1.ServiceAccount
		expectedResources      map[string]runtime.Object // Should include expected ServiceAccount,ClusterRole, ClusterRoleBinding
	}{
		{
			"create new service account",
			[]runtime.Object{
				newClusterRole("cluster-admin", corev1.NamespaceDefault, []rbacv1.PolicyRule{

					{
						APIGroups: []string{"*"},
						Resources: []string{"*"},
						Verbs:     []string{"*"},
					},
				}),
			},
			"test-service-account",
			corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service-account",
					Namespace: corev1.NamespaceDefault,
				},
			},
			map[string]runtime.Object{
				"ServiceAccount": &corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-account",
						Namespace: corev1.NamespaceDefault,
					},
				},

				"ClusterRoleBinding": newClusterRoleBinding("test-service-account-cluster-role-binding", corev1.NamespaceDefault, "cluster-admin", "test-service-account"),
			},
		},
		{
			"existing service account",
			[]runtime.Object{
				&corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-account",
						Namespace: corev1.NamespaceDefault,
					},
				},
				newClusterRole("cluster-admin", corev1.NamespaceDefault, []rbacv1.PolicyRule{

					{
						APIGroups: []string{"*"},
						Resources: []string{"*"},
						Verbs:     []string{"*"},
					},
				}),
			},
			"test-service-account",
			corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service-account",
					Namespace: corev1.NamespaceDefault,
				},
			},
			map[string]runtime.Object{
				"ServiceAccount": &corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-account",
						Namespace: corev1.NamespaceDefault,
					},
				},
				"ClusterRoleBinding": newClusterRoleBinding("test-service-account-cluster-role-binding", corev1.NamespaceDefault, "cluster-admin", "test-service-account"),
			},
		},
		{
			"existing cluster role binding",
			[]runtime.Object{
				newClusterRoleBinding("test-service-account-cluster-role-binding", corev1.NamespaceDefault, "cluster-admin", "test-service-account"),
				newClusterRole("cluster-admin", corev1.NamespaceDefault, []rbacv1.PolicyRule{

					{
						APIGroups: []string{"*"},
						Resources: []string{"*"},
						Verbs:     []string{"*"},
					},
				}),
			},
			"test-service-account",
			corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service-account",
					Namespace: corev1.NamespaceDefault,
				},
			},
			map[string]runtime.Object{
				"ServiceAccount": &corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-account",
						Namespace: corev1.NamespaceDefault,
					},
				},
				"ClusterRoleBinding": newClusterRoleBinding("test-service-account-cluster-role-binding", corev1.NamespaceDefault, "cluster-admin", "test-service-account"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remoteClientSet := fake.NewSimpleClientset()

			setupFakeSecretToken(t, remoteClientSet, tt.serviceAccountName+"-token", corev1.NamespaceDefault, []byte("usertest"))

			addFakeResources(t, remoteClientSet, tt.existingResources...)

			// Reconcile Service account
			clusterConnectionOpts := ClusterConnectionOptions{
				ServiceAccountName:     tt.serviceAccountName,
				ClusterRoleName:        tt.serviceAccountName + "-cluster-role",
				ClusterRoleBindingName: tt.serviceAccountName + "-cluster-role-binding",
				GitopsClusterName:      types.NamespacedName{Namespace: corev1.NamespaceDefault},
			}

			saToken, err := ReconcileServiceAccount(context.Background(), remoteClientSet, clusterConnectionOpts)
			assert.NoError(t, err)
			assert.Equal(t, []byte("usertest"), saToken, "service account token doesn't match expected")

			// Verify Service account created/exists
			serviceAccount, err := remoteClientSet.CoreV1().ServiceAccounts(clusterConnectionOpts.GitopsClusterName.Namespace).Get(context.Background(), "test-service-account", metav1.GetOptions{})
			assert.NoError(t, err)
			expectedServiceAccount := tt.expectedResources["ServiceAccount"].(*corev1.ServiceAccount)
			assert.Equal(t, expectedServiceAccount, serviceAccount, "service account found doesn't match expected")

			// Verify ClusterRoleBinding created/exists
			expectedClusterRoleBinding := tt.expectedResources["ClusterRoleBinding"].(*rbacv1.ClusterRoleBinding)
			clusterRoleBinding, err := remoteClientSet.RbacV1().ClusterRoleBindings().Get(context.Background(), tt.serviceAccountName+"-cluster-role-binding", metav1.GetOptions{})
			assert.NoError(t, err)
			assert.Equal(t, expectedClusterRoleBinding, clusterRoleBinding, "cluster role found doesn't match expected")

			// Verify Secret created with populated token(token is fake in test)
			expectedSecret := newServiceAccountTokenSecret(tt.serviceAccountName+"-token", tt.serviceAccountName, clusterConnectionOpts.GitopsClusterName.Namespace)
			expectedSecret.Data = map[string][]byte{
				"token": []byte("usertest"),
			}

			secret, err := remoteClientSet.CoreV1().Secrets(clusterConnectionOpts.GitopsClusterName.Namespace).Get(context.Background(), tt.serviceAccountName+"-token", metav1.GetOptions{})
			assert.NoError(t, err)
			assert.Equal(t, expectedSecret, secret, "secret found doesn't match expected")
		})
	}

}

// setupFakeSecretToken populates the token of the secret with the given token
// This is artificial where it populates the token of the secret as kubernetes isn't running. Kubernetes should populate it once the secret is created.
func setupFakeSecretToken(t *testing.T, client kubernetes.Interface, secretName, namespace string, token []byte) {
	go func(secretName, namespace string, token []byte) {
		if err := wait.PollUntilContextTimeout(context.Background(), time.Second, 10*time.Second, true, func(ctx context.Context) (done bool, err error) {
			secret, err := client.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					return false, nil
				}
				return false, err
			}

			secret.Data = map[string][]byte{
				"token": token,
			}
			if _, err := client.CoreV1().Secrets(namespace).Update(context.Background(), secret, metav1.UpdateOptions{}); err != nil {
				return false, err
			}

			return true, nil
		}); err != nil {
			t.Logf("failed to update secret with token: %s", err)
		}
	}(secretName, namespace, token)

}

func TestGetServiceAccount(t *testing.T) {
	var tests = []struct {
		name               string
		serviceAccountName string
		serviceAccounts    []string
		expected           corev1.ServiceAccount
	}{
		{
			"get existing service account",
			"test-service-account",
			[]string{
				"test-service-account",
			},
			corev1.ServiceAccount{
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

func TestCheckServiceAccountName(t *testing.T) {
	var tests = []struct {
		name               string
		serviceAccountName string
		existingResources  []runtime.Object
		expectedError      string
	}{
		{
			"check existing service account name matching label cluster-controller and service account name",
			"test-service-account",
			[]runtime.Object{
				&corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-account",
						Namespace: corev1.NamespaceDefault,
						Labels: map[string]string{
							"app.kubernetes.io/managed-by": "cluster-connector",
						},
					},
				},
			},
			"",
		},
		{
			"check existing service account name not matching label cluster-controller and name provided",
			"test-service-account",
			[]runtime.Object{
				&corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "other-service-account",
						Namespace: corev1.NamespaceDefault,
						Labels: map[string]string{
							"app.kubernetes.io/managed-by": "cluster-connector",
						},
					},
				},
			},
			"service account not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remoteClientSet := fake.NewSimpleClientset()
			clusterConnectionOpts := ClusterConnectionOptions{
				GitopsClusterName:  types.NamespacedName{Namespace: corev1.NamespaceDefault},
				ServiceAccountName: tt.serviceAccountName,
			}

			addFakeResources(t, remoteClientSet, tt.existingResources...)

			req, _ := labels.NewRequirement("app.kubernetes.io/managed-by", selection.Equals, []string{"cluster-connector"})
			selector := labels.NewSelector()
			selector = selector.Add(*req)

			err := checkServiceAccountName(context.Background(), remoteClientSet, &clusterConnectionOpts, selector)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.expectedError)
			}

		})
	}

}

func TestCheckClusterRoleBindingName(t *testing.T) {
	var tests = []struct {
		name                   string
		clusterRoleBindingName string
		existingResources      []runtime.Object
		expectedError          string
	}{
		{
			"get existing cluster role binding name matching label cluster-controller",
			"test-service-account-cluster-role-binding",
			[]runtime.Object{
				newClusterRoleBinding("test-service-account-cluster-role-binding", corev1.NamespaceDefault, "cluster-admin", "test-service-account"),
			},
			"",
		},
		{
			"check existing cluster role binding name not matching label cluster-controller and name provided",
			"test-service-account-cluster-role-binding",
			[]runtime.Object{
				newClusterRoleBinding("other-service-account-cluster-role-binding", corev1.NamespaceDefault, "cluster-admin", "test-service-account"),
			},
			"cluster role binding not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remoteClientSet := fake.NewSimpleClientset()
			clusterConnectionOpts := ClusterConnectionOptions{
				GitopsClusterName:      types.NamespacedName{Namespace: corev1.NamespaceDefault},
				ClusterRoleBindingName: tt.clusterRoleBindingName,
			}

			addFakeResources(t, remoteClientSet, tt.existingResources...)

			req, _ := labels.NewRequirement("app.kubernetes.io/managed-by", selection.Equals, []string{"cluster-connector"})
			selector := labels.NewSelector()
			selector = selector.Add(*req)

			err := checkClusterRoleBindingName(context.Background(), remoteClientSet, &clusterConnectionOpts, selector)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.expectedError)
			}
		})
	}

}

func TestDeleteServiceAccountResources(t *testing.T) {
	var tests = []struct {
		name                   string
		existingResources      []runtime.Object
		serviceAccountName     string
		clusterRoleBindingName string
		expectedResources      map[string]runtime.Object
	}{
		{
			"delete service account resources",
			[]runtime.Object{
				&corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-account",
						Namespace: corev1.NamespaceDefault,
					},
				},
				newClusterRoleBinding("test-service-account-cluster-role-binding", corev1.NamespaceDefault, "cluster-admin", "test-service-account"),
				newServiceAccountTokenSecret("test-service-account-token", "test-service-account", corev1.NamespaceDefault),
			},
			"test-service-account",
			"test-service-account-cluster-role-binding",
			map[string]runtime.Object{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remoteClientSet := fake.NewSimpleClientset()
			clusterConnectionOpts := ClusterConnectionOptions{
				ServiceAccountName:     tt.serviceAccountName,
				ClusterRoleBindingName: tt.clusterRoleBindingName,
				GitopsClusterName:      types.NamespacedName{Namespace: corev1.NamespaceDefault},
			}

			addFakeResources(t, remoteClientSet, tt.existingResources...)

			err := DeleteServiceAccountResources(context.Background(), remoteClientSet, clusterConnectionOpts)
			assert.NoError(t, err)

			// verify service account deleted
			_, err = remoteClientSet.CoreV1().ServiceAccounts(clusterConnectionOpts.GitopsClusterName.Namespace).Get(context.Background(), tt.serviceAccountName, metav1.GetOptions{})
			assert.Error(t, err)
			assert.ErrorContains(t, err, apierrors.NewNotFound(corev1.Resource("serviceaccounts"), tt.serviceAccountName).Error())

			// Verify ClusterRoleBinding deleted
			_, err = remoteClientSet.RbacV1().ClusterRoleBindings().Get(context.Background(), tt.clusterRoleBindingName, metav1.GetOptions{})
			assert.Error(t, err)
			assert.ErrorContains(t, err, apierrors.NewNotFound(rbacv1.Resource("clusterrolebindings"), tt.clusterRoleBindingName).Error())

			// Verify Secret deleted
			_, err = remoteClientSet.CoreV1().Secrets(clusterConnectionOpts.GitopsClusterName.Namespace).Get(context.Background(), tt.serviceAccountName+"-token", metav1.GetOptions{})
			assert.Error(t, err)
			assert.ErrorContains(t, err, apierrors.NewNotFound(corev1.Resource("secrets"), tt.serviceAccountName+"-token").Error())

		})
	}

}

// Add resources of different types to the client based on the type of the resource
// Valid resources: corev1.ServiceAccount, rbacv1.ClusterRole, rbacv1.ClusterRoleBinding, v1.Secret
func addFakeResources(t *testing.T, client kubernetes.Interface, resources ...runtime.Object) {
	for _, resource := range resources {
		switch resource := resource.(type) {
		case *corev1.ServiceAccount:
			_, err := client.CoreV1().ServiceAccounts(corev1.NamespaceDefault).Create(context.Background(), resource, metav1.CreateOptions{})
			assert.NoError(t, err)
		case *rbacv1.ClusterRole:
			_, err := client.RbacV1().ClusterRoles().Create(context.Background(), resource, metav1.CreateOptions{})
			assert.NoError(t, err)
		case *rbacv1.ClusterRoleBinding:
			_, err := client.RbacV1().ClusterRoleBindings().Create(context.Background(), resource, metav1.CreateOptions{})
			assert.NoError(t, err)
		case *corev1.Secret:
			_, err := client.CoreV1().Secrets(corev1.NamespaceDefault).Create(context.Background(), resource, metav1.CreateOptions{})
			assert.NoError(t, err)
		default:
			t.Fatalf("invalid resource type %s", resource)

		}

	}
}

func addFakeServiceAccounts(client kubernetes.Interface, serviceAccounts []string) error {
	if serviceAccounts == nil {
		return nil
	}
	for _, serviceAccountName := range serviceAccounts {
		serviceAccount := &corev1.ServiceAccount{
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
