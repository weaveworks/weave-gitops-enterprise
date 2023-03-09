package adapters

import (
	"fmt"

	v1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RoleLike is an interface that represents a role or cluster role
// Tried this with generics but it didn't work out.
type RoleLike interface {
	client.Object

	GetRules() []v1.PolicyRule
	GetClusterName() string
}

type cRoleAdapter struct {
	*v1.ClusterRole
	ClusterName string
}

func (c *cRoleAdapter) GetRules() []v1.PolicyRule {
	return c.Rules
}

func (c *cRoleAdapter) GetClusterName() string {
	return c.ClusterName
}

type roleAdapter struct {
	*v1.Role
	ClusterName string
}

func (r *roleAdapter) GetRules() []v1.PolicyRule {
	return r.Rules
}

func (r *roleAdapter) GetClusterName() string {
	return r.ClusterName
}

func NewRoleAdapter(clusterName string, obj client.Object) (RoleLike, error) {
	switch o := obj.(type) {
	case *v1.ClusterRole:
		return &cRoleAdapter{o, clusterName}, nil
	case *v1.Role:
		return &roleAdapter{o, clusterName}, nil

	default:
		return nil, fmt.Errorf("unknown object type %T", obj)
	}

}

type BindingLike interface {
	client.Object

	GetSubjects() []v1.Subject
	GetClusterName() string
	GetRoleRef() v1.RoleRef
}

type cRoleBindingAdapter struct {
	*v1.ClusterRoleBinding
	ClusterName string
}

func (c *cRoleBindingAdapter) GetSubjects() []v1.Subject {
	return c.Subjects
}

func (c *cRoleBindingAdapter) GetClusterName() string {
	return c.ClusterName
}

func (c *cRoleBindingAdapter) GetRoleRef() v1.RoleRef {
	return c.RoleRef
}

type roleBindingAdapter struct {
	*v1.RoleBinding
	ClusterName string
}

func (r *roleBindingAdapter) GetSubjects() []v1.Subject {
	return r.Subjects
}

func (r *roleBindingAdapter) GetClusterName() string {
	return r.ClusterName
}

func (r *roleBindingAdapter) GetRoleRef() v1.RoleRef {
	return r.RoleRef
}

func NewBindingAdapter(clusterName string, obj client.Object) (BindingLike, error) {
	switch o := obj.(type) {
	case *v1.ClusterRoleBinding:
		return &cRoleBindingAdapter{o, clusterName}, nil
	case *v1.RoleBinding:
		return &roleBindingAdapter{o, clusterName}, nil

	default:
		return nil, fmt.Errorf("unknown object type %T", obj)
	}

}
