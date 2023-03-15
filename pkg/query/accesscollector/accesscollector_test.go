package accesscollector

import (
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
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

var log logr.Logger
var g *WithT

func TestNewAccessRulesCollector(t *testing.T) {
	g = NewWithT(t)
	log = testr.New(t)

	fakeStore := &storefakes.FakeStore{}

	tests := []struct {
		name       string
		store      store.Store
		options    collector.CollectorOpts
		errPattern string
	}{
		{
			name: "can create access collector with valid arguments",
			options: collector.CollectorOpts{
				Log: log,
			},
			store:      fakeStore,
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			accessRulesCollector, err := NewAccessRulesCollector(tt.store, tt.options)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(accessRulesCollector).NotTo(BeNil())
		})
	}

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
			//arc, _, _ := createAccessRulesCollector(t)

			objs := []models.ObjectRecord{}

			for _, o := range tt.objs {
				fc := &collectorfakes.FakeObjectRecord{}
				fc.ClusterNameReturns("test-cluster")

				fc.ObjectReturns(o)
				objs = append(objs, fc)
			}

			result, err := handleRulesReceived(objs)
			assert.NoError(t, err)

			assert.Equal(t, tt.expected, result)

		})
	}
}

func createAccessRulesCollector(t *testing.T) (*AccessRulesCollector, chan []models.ObjectRecord, *storefakes.FakeStoreWriter) {
	store := &storefakes.FakeStoreWriter{}

	col := &collectorfakes.FakeCollector{}
	ch := make(chan []models.ObjectRecord)
	col.StartReturns(nil)

	arc := &AccessRulesCollector{
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
