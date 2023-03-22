package accesscollector

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/models"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAccessRuleCollector(t *testing.T) {
	// TODO: we need to test upserting and deleting access rules
}

func TestAccessLogic(t *testing.T) {

	tests := []struct {
		name        string
		objs        []client.Object
		expected    []models.AccessRule
		clusterName string
	}{

		{
			name:        "ClusterRole, one rule",
			clusterName: "test-cluster",
			objs: makeClusterRolePair("test", []v1.PolicyRule{{
				APIGroups: []string{"somegroup"},
				Verbs:     []string{"get", "list", "watch"},
				Resources: []string{"somekind"},
			}}),

			expected: []models.AccessRule{{
				Cluster:         "test-cluster",
				Principal:       "test",
				AccessibleKinds: []string{"somegroup/somekind"},
			}},
		},
		{
			name:        "ClusterRole, many rules",
			clusterName: "test-cluster",
			objs: makeClusterRolePair("test", []v1.PolicyRule{
				{
					APIGroups: []string{"somegroup"},
					Verbs:     []string{"get", "list", "watch"},
					Resources: []string{"somekind"},
				},
				{
					APIGroups: []string{"othergroup"},
					Verbs:     []string{"get", "list", "watch"},
					Resources: []string{"otherkind"},
				},
			}),

			expected: []models.AccessRule{{
				Cluster:         "test-cluster",
				Principal:       "test",
				AccessibleKinds: []string{"somegroup/somekind", "othergroup/otherkind"},
			}},
		},
		{
			name:        "ClusterRole, no list",
			clusterName: "test-cluster",
			objs: makeClusterRolePair("test", []v1.PolicyRule{
				{
					APIGroups: []string{"somegroup"},
					Verbs:     []string{"get", "watch"},
					Resources: []string{"somekind"},
				},
			}),
			expected: []models.AccessRule{{
				Cluster:         "test-cluster",
				Principal:       "test",
				AccessibleKinds: []string{},
			}},
		},
		{
			name:        "ClusterRole, * api groups",
			clusterName: "test-cluster",
			objs: makeClusterRolePair("test", []v1.PolicyRule{{
				APIGroups: []string{"*"},
				Verbs:     []string{"get", "list", "watch"},
				Resources: []string{"somekind", "otherkind"},
			}}),

			expected: []models.AccessRule{{
				Cluster:   "test-cluster",
				Principal: "test",
				// Weird case here. Someone would have to have a completely open API server to hit this.
				AccessibleKinds: []string{
					"*/somekind",
					"*/otherkind",
				},
			}},
		},
		{
			name:        "ClusterRole, * api groups, no list",
			clusterName: "test-cluster",

			objs: makeClusterRolePair("test", []v1.PolicyRule{{
				APIGroups: []string{"*"},
				Verbs:     []string{"get"},
				Resources: []string{"somekind", "otherkind"},
			}}),
			expected: []models.AccessRule{{
				Cluster:         "test-cluster",
				Principal:       "test",
				AccessibleKinds: []string{},
			}},
		},
		{
			name:        "ClusterRole, * resources",
			clusterName: "test-cluster",
			objs: makeClusterRolePair("test", []v1.PolicyRule{{
				APIGroups: []string{"someapigroup.example.com"},
				Verbs:     []string{"get", "watch"},
				Resources: []string{"*"},
			}}),

			expected: []models.AccessRule{{
				Cluster:   "test-cluster",
				Principal: "test",
				// Weird case here. Someone would have to have a completely open API server to hit this.
				AccessibleKinds: []string{
					"someapigroup.example.com/*",
				},
			}},
		},
		{
			name:        "ClusterRole, * verbs",
			clusterName: "test-cluster",
			objs: makeClusterRolePair("test", []v1.PolicyRule{{
				APIGroups: []string{"someapigroup.example.com"},
				Verbs:     []string{"*"},
				Resources: []string{"somekind", "otherkind"},
			}}),

			expected: []models.AccessRule{{
				Cluster:   "test-cluster",
				Principal: "test",
				AccessibleKinds: []string{
					"someapigroup.example.com/somekind",
					"someapigroup.example.com/otherkind",
				},
			}},
		},
		{
			name:        "Role, many rules",
			clusterName: "test-cluster",
			objs: makeRolePair("test", "test-namespace", []v1.PolicyRule{
				{
					APIGroups: []string{"somegroup"},
					Verbs:     []string{"get", "list", "watch"},
					Resources: []string{"somekind"},
				},
				{
					APIGroups: []string{"othergroup"},
					Verbs:     []string{"get", "list", "watch"},
					Resources: []string{"otherkind"},
				},
				{
					APIGroups: []string{"coolgroup"},
					// no list
					Verbs:     []string{"get", "watch"},
					Resources: []string{"coolkind"},
				},
			}),

			expected: []models.AccessRule{{
				Cluster:   "test-cluster",
				Principal: "test",
				Namespace: "test-namespace",
				AccessibleKinds: []string{
					"somegroup/somekind",
					"othergroup/otherkind",
				},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objs := []models.ObjectTransaction{}

			for _, o := range tt.objs {
				fc := &fakeObjTranaction{clusterName: tt.clusterName, obj: o}

				objs = append(objs, fc)
			}

			// Deep in the handleRulesReceived function, we use a map[string]bool{}, which does not guarantee order.
			// We need to sort the slice to ensure the test is deterministic.
			opt := cmp.Transformer("Sort", func(in []models.AccessRule) []models.AccessRule {
				out := append([]models.AccessRule{}, in...) // Copy input to avoid mutating it
				for _, r := range out {
					sort.Strings(r.AccessibleKinds)
				}
				return out
			})

			result, _, err := handleRulesReceived(objs)
			assert.NoError(t, err)

			diff := cmp.Diff(tt.expected, result, opt)

			if diff != "" {
				t.Errorf("Unexpected result: %s", diff)
			}

		})
	}
}

type fakeObjTranaction struct {
	obj             client.Object
	clusterName     string
	transactionType models.TransactionType
}

func (f *fakeObjTranaction) Object() client.Object {
	return f.obj
}

func (f *fakeObjTranaction) ClusterName() string {
	return f.clusterName
}

func (f *fakeObjTranaction) TransactionType() models.TransactionType {
	return f.transactionType
}

func makeClusterRolePair(name string, rules []v1.PolicyRule) []client.Object {
	return []client.Object{
		&v1.ClusterRole{
			TypeMeta: metav1.TypeMeta{
				Kind: "ClusterRole",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Rules: rules,
		}, &v1.ClusterRoleBinding{
			TypeMeta: metav1.TypeMeta{
				Kind: "ClusterRoleBinding",
			},
			Subjects: []v1.Subject{{
				Kind: "User",
				Name: "test",
			}},
			RoleRef: v1.RoleRef{
				Kind: "ClusterRole",
				Name: name,
			},
		},
	}
}

func makeRolePair(name string, namespace string, rules []v1.PolicyRule) []client.Object {
	return []client.Object{
		&v1.Role{
			TypeMeta: metav1.TypeMeta{
				Kind: "Role",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Rules: rules,
		}, &v1.RoleBinding{
			TypeMeta: metav1.TypeMeta{
				Kind: "RoleBinding",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
			},
			Subjects: []v1.Subject{{
				Kind: "User",
				Name: "test",
			}},
			RoleRef: v1.RoleRef{
				Kind: "Role",
				Name: name,
			},
		},
	}
}
