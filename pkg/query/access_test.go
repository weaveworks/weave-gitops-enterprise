package query

import (
	"context"
	"os"
	"strings"
	"testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/utils/testutils"

	"github.com/alecthomas/assert"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/rbac"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

// We test access rules here to get coverage on the store logic as well as the query service.
// Mocking the store here wouldn't really be testing anything.
func TestRunQuery_AccessRules(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		objects   []models.Object
		roles     []models.Role
		bindings  []models.RoleBinding
		user      *auth.UserPrincipal
		expected  []models.Object
	}{
		{
			name: "namespaced roles + groups",
			user: auth.NewUserPrincipal(auth.ID("some-user"), auth.Groups([]string{"group-a"})),
			objects: []models.Object{
				{
					Cluster:    "cluster-a",
					Namespace:  "ns-a",
					APIGroup:   helmv2.GroupVersion.Group,
					APIVersion: helmv2.GroupVersion.Version,
					Kind:       helmv2.HelmReleaseKind,
					Name:       "somename",
					Category:   models.CategoryAutomation,
				},
				{
					Cluster:    "cluster-a",
					Namespace:  "ns-b",
					APIGroup:   helmv2.GroupVersion.Group,
					APIVersion: helmv2.GroupVersion.Version,
					Kind:       helmv2.HelmReleaseKind,
					Name:       "somename",
					Category:   models.CategoryAutomation,
				},
			},
			roles: []models.Role{{
				Name:      "role-a",
				Cluster:   "cluster-a",
				Namespace: "ns-a",
				Kind:      "Role",
				PolicyRules: []models.PolicyRule{{
					APIGroups: strings.Join([]string{helmv2.GroupVersion.Group}, ","),
					Resources: strings.Join([]string{"helmreleases"}, ","),
					Verbs:     strings.Join([]string{"get", "list", "watch"}, ","),
				}},
			}},
			bindings: []models.RoleBinding{{
				Cluster:   "cluster-a",
				Name:      "binding-a",
				Namespace: "ns-a",
				Kind:      "RoleBinding",
				Subjects: []models.Subject{{
					Kind: "Group",
					Name: "group-a",
				}},
				RoleRefName: "role-a",
				RoleRefKind: "Role",
			}},
			expected: []models.Object{
				{
					Cluster:    "cluster-a",
					Namespace:  "ns-a",
					APIGroup:   helmv2.GroupVersion.Group,
					APIVersion: helmv2.GroupVersion.Version,
					Kind:       helmv2.HelmReleaseKind,
					Name:       "somename",
					Category:   models.CategoryAutomation,
				},
			},
		},
		{
			name: "non-namespaced roles + users",
			user: auth.NewUserPrincipal(auth.ID("some-user"), auth.Groups([]string{"group-a"})),
			objects: []models.Object{
				{
					Cluster:    "cluster-a",
					Namespace:  "ns-a",
					APIGroup:   helmv2.GroupVersion.Group,
					APIVersion: helmv2.GroupVersion.Version,
					Kind:       helmv2.HelmReleaseKind,
					Name:       "somename",
					Category:   models.CategoryAutomation,
				},
			},
			roles: []models.Role{{
				Name:      "role-a",
				Cluster:   "cluster-a",
				Namespace: "",
				Kind:      "ClusterRole",
				PolicyRules: []models.PolicyRule{{
					APIGroups: strings.Join([]string{helmv2.GroupVersion.Group}, ","),
					Resources: strings.Join([]string{"helmreleases"}, ","),
					Verbs:     strings.Join([]string{"get", "list", "watch"}, ","),
				}},
			}},
			bindings: []models.RoleBinding{{
				Cluster:   "cluster-a",
				Name:      "binding-a",
				Namespace: "",
				Kind:      "ClusterRoleBinding",
				Subjects: []models.Subject{{
					Kind: "User",
					Name: "some-user",
				}},
				RoleRefName: "role-a",
				RoleRefKind: "ClusterRole",
			}},
			expected: []models.Object{
				{
					Cluster:    "cluster-a",
					Namespace:  "ns-a",
					APIGroup:   helmv2.GroupVersion.Group,
					APIVersion: helmv2.GroupVersion.Version,
					Kind:       helmv2.HelmReleaseKind,
					Name:       "somename",
					Category:   models.CategoryAutomation,
				},
			},
		},
		{
			name: "cluster roles with wildcard",
			user: auth.NewUserPrincipal(auth.ID("some-user"), auth.Groups([]string{"group-a"})),
			objects: []models.Object{
				{
					Cluster:    "cluster-a",
					Namespace:  "ns-a",
					APIGroup:   "example.com",
					APIVersion: "v1",
					Kind:       "somekind",
					Name:       "somename",
					Category:   models.CategoryAutomation,
				},
			},
			roles: []models.Role{
				{
					Name:      "role-a",
					Cluster:   "cluster-a",
					Namespace: "",
					Kind:      "ClusterRole",
					PolicyRules: []models.PolicyRule{{
						APIGroups: strings.Join([]string{"example.com"}, ","),
						Resources: strings.Join([]string{"*"}, ","),
						Verbs:     strings.Join([]string{"get", "list", "watch"}, ","),
					}},
				},
			},
			bindings: []models.RoleBinding{{
				Cluster:   "cluster-a",
				Name:      "binding-a",
				Namespace: "",
				Kind:      "ClusterRoleBinding",
				Subjects: []models.Subject{{
					Kind: "User",
					Name: "some-user",
				}},
				RoleRefName: "role-a",
				RoleRefKind: "ClusterRole",
			}},
			expected: []models.Object{
				{
					Cluster:    "cluster-a",
					Namespace:  "ns-a",
					APIGroup:   "example.com",
					APIVersion: "v1",
					Kind:       "somekind",
					Name:       "somename",
				},
			},
		},
		{
			name: "cluster roles with unspecified api version with wildcard",
			user: auth.NewUserPrincipal(auth.ID("some-user"), auth.Groups([]string{"group-a"})),
			objects: []models.Object{
				{
					Cluster:    "cluster-a",
					Namespace:  "ns-a",
					APIGroup:   "example.com",
					APIVersion: "v1",
					Kind:       "somekind",
					Name:       "somename",
					Category:   models.CategoryAutomation,
				},
			},
			roles: []models.Role{
				{
					Name:      "role-a",
					Cluster:   "cluster-a",
					Namespace: "",
					Kind:      "ClusterRole",
					PolicyRules: []models.PolicyRule{{
						APIGroups: strings.Join([]string{"example.com"}, ","),
						Resources: strings.Join([]string{"*"}, ","),
						Verbs:     strings.Join([]string{"get", "list", "watch"}, ","),
					}},
				},
			},
			bindings: []models.RoleBinding{{
				Cluster:   "cluster-a",
				Name:      "binding-a",
				Namespace: "",
				Kind:      "ClusterRoleBinding",
				Subjects: []models.Subject{{
					Kind: "User",
					Name: "some-user",
				}},
				RoleRefName: "role-a",
				RoleRefKind: "ClusterRole",
			}},
			expected: []models.Object{
				{
					Cluster:    "cluster-a",
					Namespace:  "ns-a",
					APIGroup:   "example.com",
					APIVersion: "v1",
					Kind:       "somekind",
					Name:       "somename",
				},
			},
		},
		{
			name: "cluster roles with supported resource",
			user: auth.NewUserPrincipal(auth.ID("wego"), auth.Groups([]string{"group-a"})),
			objects: []models.Object{
				{
					Cluster:    "flux-system/leaf-cluster-1",
					Namespace:  "flux-stress",
					APIGroup:   helmv2.GroupVersion.Group,
					APIVersion: helmv2.GroupVersion.Version,
					Kind:       helmv2.HelmReleaseKind,
					Name:       "nginx-113",
					Category:   models.CategoryAutomation,
				},
			},
			roles: []models.Role{
				{
					Name:      "wego-cluster-role",
					Cluster:   "flux-system/leaf-cluster-1",
					Namespace: "",
					Kind:      "ClusterRole",
					PolicyRules: []models.PolicyRule{{
						APIGroups: strings.Join([]string{helmv2.GroupVersion.Group}, ","),
						Resources: strings.Join([]string{"helmreleases"}, ","),
						Verbs:     strings.Join([]string{"get", "list", "patch"}, ","),
					}},
				},
			},
			bindings: []models.RoleBinding{{
				Cluster:   "flux-system/leaf-cluster-1",
				Name:      "wego-cluster-role",
				Namespace: "",
				Kind:      "ClusterRoleBinding",
				Subjects: []models.Subject{{
					Kind: "User",
					Name: "wego",
				}},
				RoleRefName: "wego-cluster-role",
				RoleRefKind: "ClusterRole",
			}},
			expected: []models.Object{
				{
					Cluster:    "flux-system/leaf-cluster-1",
					Namespace:  "flux-stress",
					APIGroup:   helmv2.GroupVersion.Group,
					APIVersion: helmv2.GroupVersion.Version,
					Kind:       helmv2.HelmReleaseKind,
					Name:       "nginx-113",
				},
			},
		},
		{
			name: "deny for unsupported kind",
			user: auth.NewUserPrincipal(auth.ID("wego"), auth.Groups([]string{"group-a"})),
			objects: []models.Object{
				{
					Cluster:    "flux-system/leaf-cluster-1",
					Namespace:  "flux-stress",
					APIGroup:   "apiGroup",
					APIVersion: "v1",
					Kind:       "notSupportedKind",
					Name:       "nginx-113",
					Category:   models.CategoryAutomation,
				},
			},
			roles: []models.Role{
				{
					Name:      "wego-cluster-role",
					Cluster:   "flux-system/leaf-cluster-1",
					Namespace: "",
					Kind:      "ClusterRole",
					PolicyRules: []models.PolicyRule{{
						APIGroups: strings.Join([]string{helmv2.GroupVersion.String()}, ","),
						Resources: strings.Join([]string{"helmreleases"}, ","),
						Verbs:     strings.Join([]string{"get", "list", "patch"}, ","),
					}},
				},
			},
			bindings: []models.RoleBinding{{
				Cluster:   "flux-system/leaf-cluster-1",
				Name:      "wego-cluster-role",
				Namespace: "",
				Kind:      "ClusterRoleBinding",
				Subjects: []models.Subject{{
					Kind: "User",
					Name: "wego",
				}},
				RoleRefName: "wego-cluster-role",
				RoleRefKind: "ClusterRole",
			}},
			expected: []models.Object{},
		},
		{
			name: "policy rule with wildcard",
			user: auth.NewUserPrincipal(auth.ID("some-user"), auth.Groups([]string{"group-a"})),
			roles: []models.Role{
				{
					Name:      "role-a",
					Cluster:   "cluster-a",
					Namespace: "",
					Kind:      "ClusterRole",
					PolicyRules: []models.PolicyRule{
						{
							APIGroups: strings.Join([]string{"example.com"}, ","),
							Resources: strings.Join([]string{"*"}, ","),
							Verbs:     strings.Join([]string{"get", "list", "watch"}, ","),
						},
					},
				},
			},
			bindings: []models.RoleBinding{{
				Cluster:   "cluster-a",
				Name:      "binding-a",
				Namespace: "",
				Kind:      "ClusterRoleBinding",
				Subjects: []models.Subject{{
					Kind: "User",
					Name: "some-user",
				}},
				RoleRefName: "role-a",
				RoleRefKind: "ClusterRole",
			}},
			objects: []models.Object{
				{
					Cluster:    "cluster-a",
					Namespace:  "ns-a",
					APIGroup:   "example.com",
					APIVersion: "v1",
					Kind:       "somekind",
					Name:       "somename",
					Category:   models.CategoryAutomation,
				},
				{
					Cluster:    "cluster-a",
					Namespace:  "ns-a",
					APIGroup:   "otherorg.com",
					APIVersion: "v1",
					Kind:       "otherkind",
					Name:       "othername",
					Category:   models.CategoryAutomation,
				},
			},
			expected: []models.Object{
				{
					Cluster:    "cluster-a",
					Namespace:  "ns-a",
					APIGroup:   "example.com",
					APIVersion: "v1",
					Kind:       "somekind",
					Name:       "somename",
					Category:   models.CategoryAutomation,
				},
			},
		},

		{
			name: "rule with resource name",
			user: auth.NewUserPrincipal(auth.ID("some-user"), auth.Groups([]string{"group-a"})),
			objects: []models.Object{
				{
					Cluster:    "cluster-a",
					Namespace:  "ns-a",
					APIGroup:   helmv2.GroupVersion.Group,
					APIVersion: helmv2.GroupVersion.Version,
					Kind:       helmv2.HelmReleaseKind,
					Name:       "somename",
					Category:   models.CategoryAutomation,
				},
				{
					Cluster:    "cluster-a",
					Namespace:  "ns-a",
					APIGroup:   helmv2.GroupVersion.Group,
					APIVersion: helmv2.GroupVersion.Version,
					Kind:       helmv2.HelmReleaseKind,
					Name:       "othername",
					Category:   models.CategoryAutomation,
				},
			},
			roles: []models.Role{
				{
					Name:      "role-a",
					Cluster:   "cluster-a",
					Namespace: "ns-a",
					Kind:      "Role",
					PolicyRules: []models.PolicyRule{{
						APIGroups:     strings.Join([]string{helmv2.GroupVersion.Group}, ","),
						Resources:     strings.Join([]string{"helmreleases"}, ","),
						Verbs:         strings.Join([]string{"get", "list", "watch"}, ","),
						ResourceNames: strings.Join([]string{"somename"}, ","),
					}},
				},
			},
			bindings: []models.RoleBinding{{
				Cluster:   "cluster-a",
				Name:      "binding-a",
				Namespace: "ns-a",
				Kind:      "RoleBinding",
				Subjects: []models.Subject{{
					Kind: "Group",
					Name: "group-a",
				}},
				RoleRefName: "role-a",
				RoleRefKind: "Role",
			}},
			expected: []models.Object{
				{
					Cluster:    "cluster-a",
					Namespace:  "ns-a",
					APIGroup:   helmv2.GroupVersion.Group,
					APIVersion: helmv2.GroupVersion.Version,
					Kind:       helmv2.HelmReleaseKind,
					Name:       "somename",
					Category:   models.CategoryAutomation,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			ctx := auth.WithPrincipal(context.Background(), tt.user)

			dir, err := os.MkdirTemp("", "test")
			g.Expect(err).NotTo(HaveOccurred())

			s, err := store.NewStore(store.StorageBackendSQLite, dir, logr.Discard())
			g.Expect(err).NotTo(HaveOccurred())

			indexer, err := store.NewIndexer(s, dir, logr.Discard())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(s.StoreObjects(context.Background(), tt.objects)).To(Succeed())
			g.Expect(s.StoreRoles(context.Background(), tt.roles)).To(Succeed())
			g.Expect(s.StoreRoleBindings(context.Background(), tt.bindings)).To(Succeed())

			g.Expect(indexer.Add(context.Background(), tt.objects)).To(Succeed())

			//create gvks and resources configuration
			kindToResourceMap, err := testutils.CreateDefaultResourceKindMap()
			assert.NoError(t, err)

			authz := rbac.NewAuthorizer(kindToResourceMap)

			qs, err := NewQueryService(QueryServiceOpts{
				Log:         logr.Discard(),
				StoreReader: s,
				IndexReader: indexer,
				Authorizer:  authz,
			})

			assert.NoError(t, err)

			actual, err := qs.RunQuery(ctx, &query{}, nil)
			assert.NoError(t, err)

			opt := cmpopts.IgnoreFields(models.Object{}, "ID", "CreatedAt", "UpdatedAt", "DeletedAt", "Category")

			diff := cmp.Diff(tt.expected, actual, opt)

			if diff != "" {
				t.Errorf("RunQuery() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
