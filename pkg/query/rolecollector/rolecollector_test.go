package rolecollector

import (
	"testing"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/models"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRoleCollector(t *testing.T) {
	// TODO: we need to test upserting and deleting access rules

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
