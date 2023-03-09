package accesscollector

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	collectorfakes "github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/collectorfakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/storefakes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAccessRulesCollector(t *testing.T) {

	t.Run("Start", func(t *testing.T) {
		store := &storefakes.FakeStoreWriter{}

		cr := &v1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Rules: []v1.PolicyRule{{
				APIGroups: []string{"somegroup"},
				Verbs:     []string{"get", "list", "watch"},
				Resources: []string{"somekind"},
			}},
		}

		col := &collectorfakes.FakeCollector{}
		ch := make(chan []collector.ObjectRecord)
		col.StartReturns(ch, nil)
		arc := &accessRulesCollector{
			log:       logr.Discard(),
			col:       col,
			w:         store,
			converter: runtime.DefaultUnstructuredConverter,
			verbs:     DefaultVerbsRequiredForAccess,
		}
		t.Run("stores access rules", func(t *testing.T) {
			arc.Start()

			fc := &collectorfakes.FakeObjectRecord{}
			fc.ClusterNameReturns("test-cluster")

			fc.ObjectReturns(cr)

			objs := []collector.ObjectRecord{fc}

			ch <- objs

			expected := models.AccessRule{
				Cluster:         "test-cluster",
				Principal:       "test",
				AccessibleKinds: []string{"somegroup/somekind"},
			}

			assert.Equal(t, store.StoreAccessRulesCallCount(), 1)
			assert.Equal(t, expected, store.StoreAccessRulesArgsForCall(0)[0])
		})
	})
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
			arc, _, _ := createAccessRulesCollector(t)

			objs := []collector.ObjectRecord{}

			for _, o := range tt.objs {
				fc := &collectorfakes.FakeObjectRecord{}
				fc.ClusterNameReturns("test-cluster")

				fc.ObjectReturns(o)
				objs = append(objs, fc)
			}

			result, err := arc.handleRulesReceived(objs)
			assert.NoError(t, err)

			assert.Equal(t, tt.expected, result)

		})
	}
}

func createAccessRulesCollector(t *testing.T) (*accessRulesCollector, chan []collector.ObjectRecord, *storefakes.FakeStoreWriter) {
	store := &storefakes.FakeStoreWriter{}

	col := &collectorfakes.FakeCollector{}
	ch := make(chan []collector.ObjectRecord)
	col.StartReturns(ch, nil)

	arc := &accessRulesCollector{
		log:       logr.Discard(),
		col:       col,
		w:         store,
		converter: runtime.DefaultUnstructuredConverter,
		verbs:     DefaultVerbsRequiredForAccess,
	}

	return arc, ch, store
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
