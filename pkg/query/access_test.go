package query

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/models"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

// We test access rules here to get to get coverage on the store logic as well as the query service.
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
					APIGroup:   "example.com",
					APIVersion: "v1",
					Kind:       "somekind",
					Name:       "somename",
				},
				{
					Cluster:    "cluster-a",
					Namespace:  "ns-b",
					APIGroup:   "example.com",
					APIVersion: "v1",
					Kind:       "somekind",
					Name:       "somename",
				},
			},
			roles: []models.Role{{
				Name:      "role-a",
				Cluster:   "cluster-a",
				Namespace: "ns-a",
				Kind:      "Role",
				PolicyRules: []models.PolicyRule{{
					APIGroups: strings.Join([]string{"example.com/v1"}, ","),
					Resources: strings.Join([]string{"somekind"}, ","),
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
					APIGroup:   "example.com",
					APIVersion: "v1",
					Kind:       "somekind",
					Name:       "somename",
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
					APIGroup:   "example.com",
					APIVersion: "v1",
					Kind:       "somekind",
					Name:       "somename",
				},
			},
			roles: []models.Role{{
				Name:      "role-a",
				Cluster:   "cluster-a",
				Namespace: "",
				Kind:      "ClusterRole",
				PolicyRules: []models.PolicyRule{{
					APIGroups: strings.Join([]string{"example.com/v1"}, ","),
					Resources: strings.Join([]string{"somekind"}, ","),
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
					APIGroup:   "example.com",
					APIVersion: "v1",
					Kind:       "somekind",
					Name:       "somename",
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
				},
			},
			roles: []models.Role{
				{
					Name:      "role-a",
					Cluster:   "cluster-a",
					Namespace: "",
					Kind:      "ClusterRole",
					PolicyRules: []models.PolicyRule{{
						APIGroups: strings.Join([]string{"example.com/v1"}, ","),
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
			name: "cluster roles with unspecified api version",

			user: auth.NewUserPrincipal(auth.ID("some-user"), auth.Groups([]string{"group-a"})),
			objects: []models.Object{
				{
					Cluster:    "cluster-a",
					Namespace:  "ns-a",
					APIGroup:   "example.com",
					APIVersion: "v1",
					Kind:       "somekind",
					Name:       "somename",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			ctx := auth.WithPrincipal(context.Background(), tt.user)

			dir, err := os.MkdirTemp("", "test")
			g.Expect(err).NotTo(HaveOccurred())

			store, err := store.NewStore(store.StorageBackendSQLite, store.StoreOpts{Url: dir})
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(store.StoreObjects(context.Background(), tt.objects)).To(Succeed())
			g.Expect(store.StoreRoles(context.Background(), tt.roles)).To(Succeed())
			g.Expect(store.StoreRoleBindings(context.Background(), tt.bindings)).To(Succeed())

			qs, err := NewQueryService(ctx, QueryServiceOpts{
				Log:   logr.Discard(),
				Store: store,
			})

			assert.NoError(t, err)

			actual, err := qs.RunQuery(ctx, nil, nil)
			assert.NoError(t, err)

			opt := cmpopts.IgnoreFields(models.Object{}, "ID", "CreatedAt", "UpdatedAt", "DeletedAt")

			diff := cmp.Diff(tt.expected, actual, opt)

			if diff != "" {
				t.Errorf("RunQuery() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
