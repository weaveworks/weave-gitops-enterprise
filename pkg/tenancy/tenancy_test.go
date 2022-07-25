package tenancy

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	pacv2beta1 "github.com/weaveworks/policy-agent/api/v2beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_CreateTenants(t *testing.T) {
	testCases := []struct {
		name         string
		clusterState []runtime.Object
		expected     []client.Object
	}{
		{
			name:         "create tenant with new resources",
			clusterState: []runtime.Object{},
			expected: []client.Object{
				&corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "ServiceAccount",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "foo-tenant",
						Namespace:       "foo-ns",
						ResourceVersion: "1",
						Labels: map[string]string{
							"toolkit.fluxcd.io/tenant": "foo-tenant",
						},
					},
				},
			},
		},
		{
			name: "create tenant with an existing namespace",
			clusterState: []runtime.Object{
				&corev1.Namespace{
					TypeMeta: namespaceTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo-ns",
						Labels: map[string]string{
							"toolkit.fluxcd.io/tenant": "foo-tenant",
						},
					},
				},
			},
			expected: []client.Object{
				&corev1.Namespace{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Namespace",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "foo-ns",
						ResourceVersion: "1000",
						Labels: map[string]string{
							"toolkit.fluxcd.io/tenant": "foo-tenant",
						},
					},
				},
			},
		},
		{
			name: "create tenant with an existing service account",
			clusterState: []runtime.Object{
				&corev1.ServiceAccount{
					TypeMeta: serviceAccountTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-tenant",
						Namespace: "foo-ns",
						Labels: map[string]string{
							"toolkit.fluxcd.io/tenant": "foo-tenant",
						},
					},
				},
			},
			expected: []client.Object{
				&corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "ServiceAccount",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "foo-tenant",
						Namespace:       "foo-ns",
						ResourceVersion: "1000",
						Labels: map[string]string{
							"toolkit.fluxcd.io/tenant": "foo-tenant",
						},
					},
				},
			},
		},
		{
			name: "create tenant with an existing RoleBinding",
			clusterState: []runtime.Object{
				&rbacv1.RoleBinding{
					TypeMeta: roleBindingTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-tenant",
						Namespace: "foo-ns",
						Labels: map[string]string{
							"toolkit.fluxcd.io/tenant": "foo-tenant",
						},
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: "rbac.authorization.k8s.io",
						Kind:     "ClusterRole",
						Name:     "cluster-admin",
					},
					Subjects: []rbacv1.Subject{
						{
							APIGroup: "rbac.authorization.k8s.io",
							Kind:     "User",
							Name:     "gotk:foo-tenant:reconciler",
						},
						{
							Kind:      "ServiceAccount",
							Name:      "foo-tenant",
							Namespace: "foo-ns",
						},
					},
				},
			},
			expected: []client.Object{
				&rbacv1.RoleBinding{
					TypeMeta: roleBindingTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:            "foo-tenant",
						Namespace:       "foo-ns",
						ResourceVersion: "1",
						Labels: map[string]string{
							"toolkit.fluxcd.io/tenant": "foo-tenant",
						},
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: "rbac.authorization.k8s.io",
						Kind:     "ClusterRole",
						Name:     "cluster-admin",
					},
					Subjects: []rbacv1.Subject{
						{
							APIGroup: "rbac.authorization.k8s.io",
							Kind:     "User",
							Name:     "gotk:foo-ns:reconciler",
						},
						{
							Kind:      "ServiceAccount",
							Name:      "foo-tenant",
							Namespace: "foo-ns",
						},
					},
				},
			},
		},
		{
			name: "create tenant with an existing policy",
			clusterState: []runtime.Object{
				&pacv2beta1.Policy{
					TypeMeta: policyTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name: "weave.policies.tenancy.bar-tenant-allowed-repositories",
					},
					Spec: pacv2beta1.PolicySpec{
						ID:          "weave.policies.tenancy.bar-tenant-allowed-repositories",
						Name:        "bar-tenant allowed repositories",
						Category:    "weave.categories.tenancy",
						Severity:    "high",
						Description: "Controls the allowed repositories to be used as sources",
						Targets: pacv2beta1.PolicyTargets{
							Kinds:      policyRepoKinds,
							Namespaces: []string{"bar-ns", "bar"},
						},
						Code: policyCode,
						Tags: []string{"tenancy"},
					},
				},
			},
			expected: []client.Object{
				&pacv2beta1.Policy{
					TypeMeta: policyTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name: "weave.policies.tenancy.bar-tenant-allowed-repositories",
						Labels: map[string]string{
							"toolkit.fluxcd.io/tenant": "bar-tenant",
						},
					},
					Spec: pacv2beta1.PolicySpec{
						Parameters: []pacv2beta1.PolicyParameters{
							{
								Name: "git_urls",
							},
						},
						Targets: pacv2beta1.PolicyTargets{
							Kinds:      policyRepoKinds,
							Namespaces: []string{"bar-ns", "foobar-ns"},
						},
						Code: policyCode,
						Tags: []string{"tenancy"},
					},
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			fc := newFakeClient(t, tt.clusterState...)

			tenants, err := Parse("testdata/example.yaml")
			if err != nil {
				t.Fatal(err)
			}

			err = CreateTenants(context.TODO(), tenants, fc)
			assert.NoError(t, err)

			expectedObj := tt.expected[0]

			if expectedObj.GetObjectKind().GroupVersionKind().Kind == "Namespace" {
				namespaces := corev1.NamespaceList{}
				if err := fc.List(context.TODO(), &namespaces); err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, 3, len(namespaces.Items))

				namespace := namespaces.Items[1]
				expectedNamespace := expectedObj.(*corev1.Namespace)

				assert.Equal(t, expectedNamespace, &namespace)
			} else if expectedObj.GetObjectKind().GroupVersionKind().Kind == "ServiceAccount" {

				accounts := corev1.ServiceAccountList{}
				if err := fc.List(context.TODO(), &accounts, client.InNamespace("foo-ns")); err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, 1, len(accounts.Items))

				account := accounts.Items[0]
				expectedAccount := expectedObj.(*corev1.ServiceAccount)

				assert.Equal(t, expectedAccount, &account)
			} else if expectedObj.GetObjectKind().GroupVersionKind().Kind == "RoleBinding" {
				roleBindings := rbacv1.RoleBindingList{}
				if err := fc.List(context.TODO(), &roleBindings, client.InNamespace("foo-ns")); err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, 1, len(roleBindings.Items))

				roleBinding := roleBindings.Items[0]
				expectedRoleBinding := expectedObj.(*rbacv1.RoleBinding)

				assert.Equal(t, expectedRoleBinding, &roleBinding)
			} else if expectedObj.GetObjectKind().GroupVersionKind().Kind == pacv2beta1.PolicyKind {
				policies := pacv2beta1.PolicyList{}
				if err := fc.List(context.TODO(), &policies); err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, 1, len(policies.Items))
				// This doesn't compare the entirety of the spec, because it contains the
				// complete text of the policy.
				policy := policies.Items[0]
				expectedPolicy := expectedObj.(*pacv2beta1.Policy)

				assert.Equal(t, expectedPolicy.GetLabels(), policy.GetLabels())
				assert.Equal(t, expectedPolicy.Spec.Parameters[0].Name, policy.Spec.Parameters[0].Name)
				assert.Equal(t, expectedPolicy.Spec.Targets, policy.Spec.Targets)
			}
		})
	}
}

func Test_ExportTenants(t *testing.T) {
	out := &bytes.Buffer{}

	tenants, err := Parse("testdata/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	err = ExportTenants(tenants, out)
	assert.NoError(t, err)

	rendered := out.String()
	expected := readGoldenFile(t, "testdata/example.yaml.golden")

	assert.Equal(t, expected, rendered)
}

func TestGenerateTenantResources(t *testing.T) {
	generationTests := []struct {
		name   string
		tenant Tenant
		want   []client.Object
	}{
		{
			name: "simple tenant with one namespace",
			tenant: Tenant{
				Name: "test-tenant",
				Namespaces: []string{
					"foo-ns",
				},
			},
			want: []client.Object{
				newNamespace("foo-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newServiceAccount("test-tenant", "foo-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newRoleBinding("test-tenant", "foo-ns", "", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
			},
		},
		{
			name: "simple tenant with two namespaces",
			tenant: Tenant{
				Name: "test-tenant",
				Namespaces: []string{
					"foo-ns",
					"bar-ns",
				},
			},
			want: []client.Object{
				newNamespace("foo-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newServiceAccount("test-tenant", "foo-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newRoleBinding("test-tenant", "foo-ns", "", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newNamespace("bar-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newServiceAccount("test-tenant", "bar-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newRoleBinding("test-tenant", "bar-ns", "", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
			},
		},
		{
			name: "tenant with custom cluster-role",
			tenant: Tenant{
				Name: "test-tenant",
				Namespaces: []string{
					"foo-ns",
				},
				ClusterRole: "demo-cluster-role",
			},
			want: []client.Object{
				newNamespace("foo-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newServiceAccount("test-tenant", "foo-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newRoleBinding("test-tenant", "foo-ns", "demo-cluster-role", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
			},
		},
		{
			name: "tenant with additional labels",
			tenant: Tenant{
				Name: "test-tenant",
				Namespaces: []string{
					"foo-ns",
				},
				Labels: map[string]string{
					"environment": "dev",
					"provisioner": "gitops",
				},
			},
			want: []client.Object{
				newNamespace("foo-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
					"environment":              "dev",
					"provisioner":              "gitops",
				}),
				newServiceAccount("test-tenant", "foo-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
					"environment":              "dev",
					"provisioner":              "gitops",
				}),
				newRoleBinding("test-tenant", "foo-ns", "cluster-admin", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
					"environment":              "dev",
					"provisioner":              "gitops",
				}),
			},
		},
	}

	for _, tt := range generationTests {
		t.Run(tt.name, func(t *testing.T) {
			resources, err := GenerateTenantResources(tt.tenant)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.want, resources); diff != "" {
				t.Fatalf("failed to generate resources:\n%s", diff)
			}
		})
	}
}

func TestGenerateTenantResources_WithErrors(t *testing.T) {
	generationTests := []struct {
		name          string
		tenant        Tenant
		errorMessages []string
	}{
		{
			name: "simple tenant with no namespace",
			tenant: Tenant{
				Name:       "test-tenant",
				Namespaces: []string{},
			},
			errorMessages: []string{"must provide at least one namespace"},
		},
		{
			name: "tenant with no name",
			tenant: Tenant{
				Namespaces: []string{
					"foo-ns",
				},
			},
			errorMessages: []string{"invalid tenant name"},
		},
		{
			name: "tenant with no name and no namespace",
			tenant: Tenant{
				Namespaces: []string{},
			},
			errorMessages: []string{"invalid tenant name", "must provide at least one namespace"},
		},
	}

	for _, tt := range generationTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GenerateTenantResources(tt.tenant)

			for _, errMessage := range tt.errorMessages {
				assert.ErrorContains(t, err, errMessage)
			}
		})
	}
}

func TestGenerateTenantResources_WithMultipleTenants(t *testing.T) {
	tenant1 := Tenant{
		Name: "foo-tenant",
		Namespaces: []string{
			"foo-ns",
		},
	}
	tenant2 := Tenant{
		Name: "bar-tenant",
		Namespaces: []string{
			"foo-ns",
		},
	}

	resourceForTenant1, err := GenerateTenantResources(tenant1)
	assert.NoError(t, err)
	resourceForTenant2, err := GenerateTenantResources(tenant2)
	assert.NoError(t, err)
	resourceForTenants, err := GenerateTenantResources(tenant1, tenant2)
	assert.NoError(t, err)
	assert.Equal(t, append(resourceForTenant1, resourceForTenant2...), resourceForTenants)
}

func TestParse(t *testing.T) {
	tenants, err := Parse("testdata/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(tenants), 2)
	assert.Equal(t, len(tenants[1].Namespaces), 2)
	assert.Equal(t, tenants[1].Namespaces[1], "foobar-ns")
}

func Test_newNamespace(t *testing.T) {
	labels := map[string]string{
		"toolkit.fluxcd.io/tenant": "test-tenant",
	}

	ns := newNamespace("foo-ns", labels)
	assert.Equal(t, ns.Labels["toolkit.fluxcd.io/tenant"], "test-tenant")
}

func Test_newServiceAccount(t *testing.T) {
	labels := map[string]string{
		"toolkit.fluxcd.io/tenant": "test-tenant",
	}

	sa := newServiceAccount("test-tenant", "test-namespace", labels)
	assert.Equal(t, sa.Name, "test-tenant")
	assert.Equal(t, sa.Namespace, "test-namespace")
	assert.Equal(t, sa.Labels["toolkit.fluxcd.io/tenant"], "test-tenant")
}

func Test_newRoleBinding(t *testing.T) {
	labels := map[string]string{
		"toolkit.fluxcd.io/tenant": "test-tenant",
	}

	rb := newRoleBinding("test-tenant", "test-namespace", "", labels)
	assert.Equal(t, rb.Name, "test-tenant")
	assert.Equal(t, rb.Namespace, "test-namespace")
	assert.Equal(t, rb.RoleRef.Name, "cluster-admin")
	assert.Equal(t, rb.Labels["toolkit.fluxcd.io/tenant"], "test-tenant")

	rb = newRoleBinding("test-tenant", "test-namespace", "test-cluster-role", labels)
	assert.Equal(t, rb.RoleRef.Name, "test-cluster-role")
}

func Test_newPolicy(t *testing.T) {
	labels := map[string]string{
		"toolkit.fluxcd.io/tenant": "test-tenant",
	}

	namespaces := []string{"test-namespace"}

	pol, err := newPolicy(
		"test-tenant",
		namespaces,
		[]AllowedRepository{{URL: "https://github.com/testorg/testrepo", Kind: "GitRepository"}},
		labels,
	)
	if err != nil {
		t.Fatal(err)
	}
	val, err := json.Marshal([]string{"https://github.com/testorg/testrepo"})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, pol.Name, "weave.policies.tenancy.test-tenant-allowed-repositories")
	assert.Equal(t, pol.Spec.Targets.Namespaces, namespaces)
	assert.Equal(t, pol.Spec.Parameters[0].Value.Raw, val)
	assert.Equal(t, pol.Spec.Parameters[0].Name, "git_urls")
	assert.Equal(t, pol.Labels["toolkit.fluxcd.io/tenant"], "test-tenant")

}

func readGoldenFile(t *testing.T, filename string) string {
	t.Helper()

	b, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	return string(b)
}

func newFakeClient(t *testing.T, objs ...runtime.Object) client.Client {
	t.Helper()

	scheme := runtime.NewScheme()

	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	if err := pacv2beta1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objs...).
		Build()
}
