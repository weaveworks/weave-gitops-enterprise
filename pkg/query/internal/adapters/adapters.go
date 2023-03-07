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
}

type cRoleAdapter struct {
	*v1.ClusterRole
}

func (c *cRoleAdapter) GetRules() []v1.PolicyRule {
	return c.Rules
}

type roleAdapter struct {
	*v1.Role
}

func (r *roleAdapter) GetRules() []v1.PolicyRule {
	return r.Rules
}

func NewRoleAdapter(obj client.Object) (RoleLike, error) {
	switch o := obj.(type) {
	case *v1.ClusterRole:
		return &cRoleAdapter{o}, nil
	case *v1.Role:
		return &roleAdapter{o}, nil

	default:
		return nil, fmt.Errorf("unknown object type %T", obj)
	}

}

type BindingLike interface {
	client.Object

	GetSubjects() []v1.Subject
}

type cRoleBindingAdapter struct {
	*v1.ClusterRoleBinding
}

func (c *cRoleBindingAdapter) GetSubjects() []v1.Subject {
	return c.Subjects
}

type roleBindingAdapter struct {
	*v1.RoleBinding
}

func (r *roleBindingAdapter) GetSubjects() []v1.Subject {
	return r.Subjects
}

func NewBindingAdapter(obj client.Object) (BindingLike, error) {
	switch o := obj.(type) {
	case *v1.ClusterRoleBinding:
		return &cRoleBindingAdapter{o}, nil
	case *v1.RoleBinding:
		return &roleBindingAdapter{o}, nil

	default:
		return nil, fmt.Errorf("unknown object type %T", obj)
	}

}
