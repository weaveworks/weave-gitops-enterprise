package testutils

import (
	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewHelmRelease creates a test helm release out of the parameters.It uses a decorator pattern to add custom configuration.
func NewHelmRelease(name string, namespace string, opts ...func(*v2beta1.HelmRelease)) *v2beta1.HelmRelease {
	helmRelease := &v2beta1.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v2beta1.GroupVersion.Version,
			Kind:       v2beta1.HelmReleaseKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	for _, opt := range opts {
		opt(helmRelease)
	}

	return helmRelease
}

var (
	roleTypeMeta = typeMeta("Role", "rbac.authorization.k8s.io/v1")
)

func typeMeta(kind, apiVersion string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       kind,
		APIVersion: apiVersion,
	}
}

// NewRole creates a test helm release out of the parameters.It uses a decorator pattern to add custom configuration.
func NewRole(name string, namespace string, opts ...func(*rbacv1.Role)) *rbacv1.Role {

	role := &rbacv1.Role{
		TypeMeta: roleTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"kustomize.toolkit.fluxcd.io"},
				Resources: []string{"kustomizations"},
				Verbs:     []string{"get"},
			},
		},
	}

	for _, opt := range opts {
		opt(role)
	}

	return role
}

func NewObjectTransaction(clusterName string, object client.Object, txType models.TransactionType) models.ObjectTransaction {
	return objectTransaction{
		clusterName:     clusterName,
		object:          object,
		transactionType: txType,
	}
}

type objectTransaction struct {
	clusterName     string
	object          client.Object
	transactionType models.TransactionType
}

func (r objectTransaction) ClusterName() string {
	return r.clusterName
}

func (r objectTransaction) Object() client.Object {
	return r.object
}

func (r objectTransaction) TransactionType() models.TransactionType {
	return r.transactionType
}

func (r objectTransaction) RetentionPolicy() configuration.RetentionPolicy {
	return 0
}
