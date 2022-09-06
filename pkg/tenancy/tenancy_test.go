package tenancy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	pacv2beta1 "github.com/weaveworks/policy-agent/api/v2beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_CreateTenants(t *testing.T) {
	testCases := []struct {
		name              string
		clusterState      []runtime.Object
		verifications     []verifyFunc
		expectedResources map[client.Object][]client.Object
	}{
		{
			name:         "create tenant with new resources",
			clusterState: []runtime.Object{},
			verifications: []verifyFunc{
				verifyNamespaces(
					setResourceVersion(newNamespace("foo-ns", map[string]string{
						"toolkit.fluxcd.io/tenant": "foo-tenant",
					}), 1),
					setResourceVersion(newNamespace("bar-ns", map[string]string{
						"toolkit.fluxcd.io/tenant": "bar-tenant",
					}), 1),
					setResourceVersion(newNamespace("foobar-ns", map[string]string{
						"toolkit.fluxcd.io/tenant": "bar-tenant",
					}), 1),
				),
				verifyServiceAccounts(
					setResourceVersion(
						newServiceAccount("foo-tenant", "foo-ns", map[string]string{
							"toolkit.fluxcd.io/tenant": "foo-tenant",
						}), 1),
				),
				verifyRoleBindings(
					setResourceVersion(newRoleBinding("foo-tenant", "foo-ns", "foo-tenant", "", map[string]string{
						"toolkit.fluxcd.io/tenant": "foo-tenant",
					}), 1),
				),
				verifyPolicies(
					setResourceVersion(
						testNewAllowedReposPolicy(
							t,
							"bar-tenant",
							[]string{"bar-ns", "foobar-ns"},
							[]AllowedRepository{
								{URL: "https://github.com/testorg/testrepo", Kind: "GitRepository"},
								{URL: "https://github.com/testorg/testinfo", Kind: "GitRepository"},
								{URL: "minio.example.com", Kind: "Bucket"},
								{URL: "https://testorg.github.io/testrepo", Kind: "HelmRepository"}},
							map[string]string{
								"toolkit.fluxcd.io/tenant": "bar-tenant",
							},
						), 1),
					setResourceVersion(
						testNewAllowedClustersPolicy(
							t,
							"bar-tenant",
							[]string{"bar-ns", "foobar-ns"},
							[]AllowedCluster{
								{KubeConfig: "cluster-1-kubeconfig"},
								{KubeConfig: "cluster-2-kubeconfig"},
							},
							map[string]string{
								"toolkit.fluxcd.io/tenant": "bar-tenant",
							},
						), 1),
				),
			},
		},
		{
			name: "update existing tenants",
			clusterState: []runtime.Object{
				// The existing resources do not have labels for the the tenants.
				setResourceVersion(newNamespace("foo-ns", map[string]string{}), 1),
				setResourceVersion(newNamespace("bar-ns", map[string]string{}), 1),
				setResourceVersion(newNamespace("foobar-ns", map[string]string{}), 1),
				setResourceVersion(
					newServiceAccount("foo-tenant", "foo-ns", map[string]string{}), 1),
				setResourceVersion(newRoleBinding("foo-tenant", "foo-ns", "foo-tenant", "", map[string]string{
					"toolkit.fluxcd.io/tenant": "foo-tenant",
				}), 1),
				// The setup version is only for a single tenant, the example
				// file has two.
				setResourceVersion(
					testNewAllowedReposPolicy(
						t,
						"bar-tenant",
						[]string{"bar-ns"},
						[]AllowedRepository{
							{URL: "https://github.com/testorg/testrepo", Kind: "GitRepository"},
							{URL: "https://github.com/testorg/testinfo", Kind: "GitRepository"}},
						map[string]string{
							"toolkit.fluxcd.io/tenant": "bar-tenant",
						},
					), 1),
				setResourceVersion(
					testNewAllowedClustersPolicy(
						t,
						"bar-tenant",
						[]string{"bar-ns", "foobar-ns"},
						[]AllowedCluster{
							{KubeConfig: "cluster-3-kubeconfig"},
						},
						map[string]string{
							"toolkit.fluxcd.io/tenant": "bar-tenant",
						},
					), 1),
			},
			verifications: []verifyFunc{
				verifyNamespaces(
					setResourceVersion(newNamespace("foo-ns", map[string]string{
						"toolkit.fluxcd.io/tenant": "foo-tenant",
					}), 2),
					setResourceVersion(newNamespace("bar-ns", map[string]string{
						"toolkit.fluxcd.io/tenant": "bar-tenant",
					}), 2),
					setResourceVersion(newNamespace("foobar-ns", map[string]string{
						"toolkit.fluxcd.io/tenant": "bar-tenant",
					}), 2),
				),
				verifyServiceAccounts(
					setResourceVersion(
						newServiceAccount("foo-tenant", "foo-ns", map[string]string{
							"toolkit.fluxcd.io/tenant": "foo-tenant",
						}), 2),
				),
				verifyRoleBindings(
					setResourceVersion(newRoleBinding("foo-tenant", "foo-ns", "foo-tenant", "", map[string]string{
						"toolkit.fluxcd.io/tenant": "foo-tenant",
					}), 1),
				),
				verifyPolicies(
					setResourceVersion(
						testNewAllowedReposPolicy(
							t,
							"bar-tenant",
							[]string{"bar-ns", "foobar-ns"},
							[]AllowedRepository{
								{URL: "https://github.com/testorg/testrepo", Kind: "GitRepository"},
								{URL: "https://github.com/testorg/testinfo", Kind: "GitRepository"},
								{URL: "minio.example.com", Kind: "Bucket"},
								{URL: "https://testorg.github.io/testrepo", Kind: "HelmRepository"}},
							map[string]string{
								"toolkit.fluxcd.io/tenant": "bar-tenant",
							},
						), 2),
					setResourceVersion(
						testNewAllowedClustersPolicy(
							t,
							"bar-tenant",
							[]string{"bar-ns", "foobar-ns"},
							[]AllowedCluster{
								{KubeConfig: "cluster-1-kubeconfig"},
								{KubeConfig: "cluster-2-kubeconfig"},
							},
							map[string]string{
								"toolkit.fluxcd.io/tenant": "bar-tenant",
							},
						), 2),
				),
			},
		},
		{
			name: "replace existing rolebindings",
			clusterState: []runtime.Object{
				setResourceVersion(newRoleBinding("foo-tenant", "foo-ns", "foo-tenant", "unknown-cluster-role", map[string]string{
					"toolkit.fluxcd.io/tenant": "foo-tenant",
				}), 1),
			},
			verifications: []verifyFunc{
				verifyRoleBindings(
					setResourceVersion(newRoleBinding("foo-tenant", "foo-ns", "foo-tenant", "", map[string]string{
						"toolkit.fluxcd.io/tenant": "foo-tenant",
					}), 1),
				),
			},
		},
		{
			name: "replace existing role",
			clusterState: []runtime.Object{
				setResourceVersion(newTeamRole("bar-tenant", "bar-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "bar-tenant",
				}, []rbacv1.PolicyRule{{
					APIGroups: []string{""},
					Resources: []string{"pods"},
					Verbs:     []string{"get"},
				}}), 1),
			},
			verifications: []verifyFunc{
				verifyRoles(
					setResourceVersion(newTeamRole("bar-tenant", "bar-ns", map[string]string{
						"toolkit.fluxcd.io/tenant": "bar-tenant",
					}, []rbacv1.PolicyRule{{
						APIGroups: []string{""},
						Resources: []string{"namespaces", "pods"},
						Verbs:     []string{"list", "get"},
					}}), 2),
				),
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

			err = CreateTenants(context.TODO(), &Config{Tenants: tenants.Tenants}, fc, os.Stdout)
			assert.NoError(t, err)

			for _, f := range tt.verifications {
				f(t, fc)
			}
		})
	}
}

func Test_ExportTenants(t *testing.T) {
	testFiles := []struct {
		filename string
	}{
		{"testdata/example.yaml"},
		{"testdata/with_service_account.yaml"},
		{"testdata/with_custom_labels.yaml"},
	}

	for _, tt := range testFiles {
		t.Run(fmt.Sprintf("exporting %s", tt), func(t *testing.T) {
			out := &bytes.Buffer{}

			config, err := Parse(tt.filename)
			if err != nil {
				t.Fatal(err)
			}

			err = ExportTenants(config, out)
			assert.NoError(t, err)

			rendered := out.String()
			expected := readGoldenFile(t, tt.filename+".golden")

			assert.Equal(t, expected, rendered)
		})
	}
}

func TestGenerateTenantResources(t *testing.T) {
	generationTests := []struct {
		name   string
		config *Config
		want   []client.Object
	}{
		{
			name: "simple tenant with one namespace",
			config: &Config{
				Tenants: []Tenant{
					{
						Name: "test-tenant",
						Namespaces: []string{
							"foo-ns",
						},
					},
				},
			},
			want: []client.Object{
				newNamespace("foo-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newServiceAccount("test-tenant", "foo-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newRoleBinding("test-tenant", "foo-ns", "test-tenant", "", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
			},
		},
		{
			name: "simple tenant with two namespaces",
			config: &Config{
				Tenants: []Tenant{
					{
						Name: "test-tenant",
						Namespaces: []string{
							"foo-ns",
							"bar-ns",
						},
					},
				},
			},
			want: []client.Object{
				newNamespace("foo-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newServiceAccount("test-tenant", "foo-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newRoleBinding("test-tenant", "foo-ns", "test-tenant", "", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newNamespace("bar-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newServiceAccount("test-tenant", "bar-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newRoleBinding("test-tenant", "bar-ns", "test-tenant", "", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
			},
		},
		{
			name: "tenant with custom cluster-role",
			config: &Config{
				Tenants: []Tenant{
					{
						Name: "test-tenant",
						Namespaces: []string{
							"foo-ns",
						},
						ClusterRole: "demo-cluster-role",
					},
				},
			},
			want: []client.Object{
				newNamespace("foo-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newServiceAccount("test-tenant", "foo-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newRoleBinding("test-tenant", "foo-ns", "test-tenant", "demo-cluster-role", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
			},
		},
		{
			name: "tenant with additional labels",
			config: &Config{
				Tenants: []Tenant{
					{
						Name: "test-tenant",
						Namespaces: []string{
							"foo-ns",
						},
						Labels: map[string]string{
							"environment": "dev",
							"provisioner": "gitops",
						},
					},
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
				newRoleBinding("test-tenant", "foo-ns", "test-tenant", "cluster-admin", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
					"environment":              "dev",
					"provisioner":              "gitops",
				}),
			},
		},
		{
			name: "tenant option custom service account",
			config: &Config{
				Tenants: []Tenant{
					{
						Name: "test-tenant",
						Namespaces: []string{
							"foo-ns",
							"bar-ns",
						},
					},
				},
				ServiceAccount: &ServiceAccountOptions{
					Name: "default-sa",
				},
			},
			want: []client.Object{
				newNamespace("foo-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newServiceAccount("default-sa", "foo-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newRoleBinding("test-tenant", "foo-ns", "default-sa", "", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newNamespace("bar-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newServiceAccount("default-sa", "bar-ns", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newRoleBinding("test-tenant", "bar-ns", "default-sa", "", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
			},
		},
	}

	for _, tt := range generationTests {
		t.Run(tt.name, func(t *testing.T) {
			resources, err := GenerateTenantResources(tt.config)
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
		{
			name: "tenant with empty teamRBAC rules list",
			tenant: makeTestTenant(t, func(tenant *Tenant) {
				tenant.TeamRBAC = &TenantTeamRBAC{
					GroupNames: []string{"foo-group"},
					Rules:      []rbacv1.PolicyRule{},
				}
			}),
			errorMessages: []string{"must provide group names and team rules in team RBAC"},
		},
		{
			name: "tenant with empty teamRBAC groupNames list",
			tenant: makeTestTenant(t, func(tenant *Tenant) {
				tenant.TeamRBAC = &TenantTeamRBAC{
					GroupNames: []string{},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:     []string{"get"},
							APIGroups: []string{""},
							Resources: []string{"pods"},
						},
					},
				}
			}),
			errorMessages: []string{"must provide group names and team rules in team RBAC"},
		},
	}

	for _, tt := range generationTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GenerateTenantResources(&Config{Tenants: []Tenant{tt.tenant}})

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

	resourceForTenant1, err := GenerateTenantResources(&Config{Tenants: []Tenant{tenant1}})
	assert.NoError(t, err)
	resourceForTenant2, err := GenerateTenantResources(&Config{Tenants: []Tenant{tenant2}})
	assert.NoError(t, err)
	resourceForTenants, err := GenerateTenantResources(&Config{Tenants: []Tenant{tenant1, tenant2}})
	assert.NoError(t, err)
	assert.Equal(t, append(resourceForTenant1, resourceForTenant2...), resourceForTenants)
}

func TestParse(t *testing.T) {
	config, err := Parse("testdata/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(config.Tenants), 2)
	assert.Equal(t, len(config.Tenants[1].Namespaces), 2)
	assert.Equal(t, config.Tenants[1].Namespaces[1], "foobar-ns")
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

	rb := newRoleBinding("test-tenant", "test-namespace", "test-tenant", "", labels)
	assert.Equal(t, rb.Name, "test-tenant")
	assert.Equal(t, rb.Namespace, "test-namespace")
	assert.Equal(t, rb.RoleRef.Name, "cluster-admin")
	assert.Equal(t, rb.Labels["toolkit.fluxcd.io/tenant"], "test-tenant")

	rb = newRoleBinding("test-tenant", "test-namespace", "test-tenant", "test-cluster-role", labels)
	assert.Equal(t, rb.RoleRef.Name, "test-cluster-role")
}

func Test_newTeamRoleBinding(t *testing.T) {
	labels := map[string]string{
		"toolkit.fluxcd.io/tenant": "test-tenant",
	}
	groupNames := []string{"test-group"}

	subjects := []rbacv1.Subject{
		{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Group",
			Name:     groupNames[0],
		},
	}

	rb := newTeamRoleBinding("test-tenant", "test-namespace", groupNames, labels)
	assert.Equal(t, rb.Name, "test-tenant-team-rolebinding")
	assert.Equal(t, rb.Namespace, "test-namespace")
	assert.Equal(t, rb.RoleRef.Name, "test-tenant-team-role")
	assert.Equal(t, rb.Labels["toolkit.fluxcd.io/tenant"], "test-tenant")
	assert.Equal(t, rb.Subjects, subjects)

}

func Test_newTeamRole(t *testing.T) {
	labels := map[string]string{
		"toolkit.fluxcd.io/tenant": "test-tenant",
	}

	rules := []rbacv1.PolicyRule{
		{
			Verbs:     []string{"get"},
			APIGroups: []string{""},
			Resources: []string{"pods"},
		},
	}

	rb := newTeamRole("test-tenant", "test-namespace", labels, rules)
	assert.Equal(t, rb.Name, "test-tenant-team-role")
	assert.Equal(t, rb.Namespace, "test-namespace")
	assert.Equal(t, rb.Labels["toolkit.fluxcd.io/tenant"], "test-tenant")
	assert.Equal(t, rb.Rules, rules)

}

func Test_newAllowedRepositoriesPolicy(t *testing.T) {
	labels := map[string]string{
		"toolkit.fluxcd.io/tenant": "test-tenant",
	}

	namespaces := []string{"test-namespace"}

	pol, err := newAllowedRepositoriesPolicy(
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

	expectedParams := []pacv2beta1.PolicyParameters{
		{
			Name: "git_urls",
			Value: &apiextensionsv1.JSON{
				Raw: val,
			},
			Type: "array",
		},
		{
			Name: "bucket_endpoints",
			Value: &apiextensionsv1.JSON{
				Raw: []byte("null"),
			},
			Type: "array",
		},
		{
			Name: "helm_urls",
			Value: &apiextensionsv1.JSON{
				Raw: []byte("null"),
			},
			Type: "array",
		},
	}

	assert.Equal(t, pol.Name, "weave.policies.tenancy.test-tenant-allowed-repositories")
	assert.Equal(t, pol.Spec.Targets.Namespaces, namespaces)
	assert.Equal(t, pol.Spec.Parameters, expectedParams)
	assert.Equal(t, pol.Labels["toolkit.fluxcd.io/tenant"], "test-tenant")

}

func Test_newAllowedClustersPolicy(t *testing.T) {
	labels := map[string]string{
		"toolkit.fluxcd.io/tenant": "test-tenant",
	}

	namespaces := []string{"test-namespace"}

	pol, err := newAllowedClustersPolicy(
		"test-tenant",
		namespaces,
		[]AllowedCluster{{KubeConfig: "demo-kubeconfig"}},
		labels,
	)
	if err != nil {
		t.Fatal(err)
	}
	val, err := json.Marshal([]string{"demo-kubeconfig"})
	if err != nil {
		t.Fatal(err)
	}
	expectedParams := []pacv2beta1.PolicyParameters{
		{
			Name: "cluster_secrets",
			Value: &apiextensionsv1.JSON{
				Raw: val,
			},
			Type: "array",
		},
	}
	assert.Equal(t, pol.Name, "weave.policies.tenancy.test-tenant-allowed-clusters")
	assert.Equal(t, pol.Spec.Targets.Namespaces, namespaces)
	assert.Equal(t, pol.Spec.Parameters, expectedParams)
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

func verifyNamespaces(ns ...*corev1.Namespace) func(t *testing.T, cl client.Client) {
	return func(t *testing.T, cl client.Client) {
		sort.Slice(ns, func(i, j int) bool { return ns[i].GetName() < ns[j].GetName() })
		namespaces := corev1.NamespaceList{}
		if err := cl.List(context.TODO(), &namespaces); err != nil {
			t.Fatal(err)
		}
		sort.Slice(namespaces.Items, func(i, j int) bool { return namespaces.Items[i].GetName() < namespaces.Items[j].GetName() })
		for i := range ns {
			assert.Equal(t, ns[i], &namespaces.Items[i])
		}
	}
}

func verifyServiceAccounts(sa ...*corev1.ServiceAccount) func(t *testing.T, cl client.Client) {
	return func(t *testing.T, cl client.Client) {
		sort.Slice(sa, func(i, j int) bool { return sa[i].GetName() < sa[j].GetName() })

		accounts := corev1.ServiceAccountList{}
		if err := cl.List(context.TODO(), &accounts, client.InNamespace("foo-ns")); err != nil {
			t.Fatal(err)
		}
		sort.Slice(accounts.Items, func(i, j int) bool { return accounts.Items[i].GetName() < accounts.Items[j].GetName() })
		for i := range sa {
			assert.Equal(t, sa[i], &accounts.Items[i])
		}
	}
}

func verifyRoleBindings(rb ...*rbacv1.RoleBinding) func(t *testing.T, cl client.Client) {
	return func(t *testing.T, cl client.Client) {
		sort.Slice(rb, func(i, j int) bool { return rb[i].GetName() < rb[j].GetName() })
		roleBindings := rbacv1.RoleBindingList{}
		if err := cl.List(context.TODO(), &roleBindings, client.InNamespace("foo-ns")); err != nil {
			t.Fatal(err)
		}

		sort.Slice(roleBindings.Items, func(i, j int) bool { return roleBindings.Items[i].GetName() < roleBindings.Items[j].GetName() })
		for i := range rb {
			assert.Equal(t, rb[i], &roleBindings.Items[i])
		}
	}
}

func verifyRoles(rb ...*rbacv1.Role) func(t *testing.T, cl client.Client) {
	return func(t *testing.T, cl client.Client) {
		sort.Slice(rb, func(i, j int) bool { return rb[i].GetName() < rb[j].GetName() })
		roles := rbacv1.RoleList{}
		if err := cl.List(context.TODO(), &roles, client.InNamespace("bar-ns")); err != nil {
			t.Fatal(err)
		}

		sort.Slice(roles.Items, func(i, j int) bool { return roles.Items[i].GetName() < roles.Items[j].GetName() })
		for i := range rb {
			assert.Equal(t, rb[i], &roles.Items[i])
		}
	}
}

func verifyPolicies(expected ...*pacv2beta1.Policy) func(t *testing.T, cl client.Client) {
	return func(t *testing.T, cl client.Client) {
		sort.Slice(expected, func(i, j int) bool { return expected[i].GetName() < expected[j].GetName() })
		policies := pacv2beta1.PolicyList{}
		if err := cl.List(context.TODO(), &policies); err != nil {
			t.Fatal(err)
		}
		sort.Slice(policies.Items, func(i, j int) bool { return policies.Items[i].GetName() < policies.Items[j].GetName() })

		assert.Equal(t, len(expected), len(policies.Items))
		for i := range policies.Items {
			// This doesn't compare the entirety of the spec, because it contains the
			// complete text of the policy.
			policy := policies.Items[i]
			expectedPolicy := expected[i]

			assert.Equal(t, expectedPolicy.ObjectMeta, policy.ObjectMeta)
			if diff := cmp.Diff(expectedPolicy.Spec.Parameters, policy.Spec.Parameters); diff != "" {
				t.Fatalf("parameters don't match:\n%s", diff)
			}
			assert.Equal(t, expectedPolicy.Spec.Targets, policy.Spec.Targets)
		}
	}
}

type verifyFunc func(t *testing.T, cl client.Client)

func makeTestTenant(t *testing.T, options ...func(*Tenant)) Tenant {
	tenant := &Tenant{
		Name: "foo",
		Namespaces: []string{
			"foo-ns",
		},
	}

	for _, o := range options {
		o(tenant)
	}

	return *tenant
}

func testNewAllowedReposPolicy(t *testing.T, tenantName string, namespaces []string, allowedRepositories []AllowedRepository, labels map[string]string) *pacv2beta1.Policy {
	t.Helper()
	p, err := newAllowedRepositoriesPolicy(tenantName, namespaces, allowedRepositories, labels)
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func testNewAllowedClustersPolicy(t *testing.T, tenantName string, namespaces []string, allowedClusters []AllowedCluster, labels map[string]string) *pacv2beta1.Policy {
	t.Helper()
	p, err := newAllowedClustersPolicy(tenantName, namespaces, allowedClusters, labels)
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func setResourceVersion[T client.Object](obj T, rv int) T {
	obj.SetResourceVersion(fmt.Sprintf("%v", rv))

	return obj
}
