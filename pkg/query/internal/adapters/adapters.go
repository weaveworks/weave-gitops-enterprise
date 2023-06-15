package adapters

import (
	"fmt"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RoleLike is an interface that represents a role or cluster role
// Tried this with generics but it didn't work out.
type RoleLike interface {
	client.Object

	GetRules() []v1.PolicyRule
	GetClusterName() string
	GetSubjects() []v1.Subject
	ToModel() models.Role
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

func (c *cRoleAdapter) ToModel() models.Role {
	rules := []models.PolicyRule{}

	for _, r := range c.Rules {
		rules = append(rules, models.PolicyRule{
			APIGroups:     models.JoinRuleData(r.APIGroups),
			Resources:     models.JoinRuleData(r.Resources),
			Verbs:         models.JoinRuleData(r.Verbs),
			ResourceNames: models.JoinRuleData(r.ResourceNames),
		})
	}

	return models.Role{
		Cluster:     c.ClusterName,
		Namespace:   c.GetNamespace(),
		Kind:        c.Kind,
		Name:        c.Name,
		PolicyRules: rules,
	}
}

func (c *cRoleAdapter) GetSubjects() []v1.Subject {
	return []v1.Subject{{
		Kind: "ClusterRole",
		Name: c.Name,
	}}
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

func (r *roleAdapter) ToModel() models.Role {
	rules := []models.PolicyRule{}

	for _, r := range r.Rules {
		rules = append(rules, models.PolicyRule{
			APIGroups:     models.JoinRuleData(r.APIGroups),
			Resources:     models.JoinRuleData(r.Resources),
			Verbs:         models.JoinRuleData(r.Verbs),
			ResourceNames: models.JoinRuleData(r.ResourceNames),
		})
	}

	return models.Role{
		Cluster:     r.ClusterName,
		Namespace:   r.Namespace,
		Kind:        r.Kind,
		Name:        r.Name,
		PolicyRules: rules,
	}
}

func (r *roleAdapter) GetSubjects() []v1.Subject {
	return []v1.Subject{{
		Kind:      "Role",
		Name:      r.Name,
		Namespace: r.Namespace,
	}}
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
	ToModel() models.RoleBinding
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

func (c *cRoleBindingAdapter) ToModel() models.RoleBinding {
	subjects := []models.Subject{}

	for _, s := range c.Subjects {
		subjects = append(subjects, models.Subject{
			Kind:      s.Kind,
			Name:      s.Name,
			Namespace: s.Namespace,
			APIGroup:  s.APIGroup,
		})
	}

	return models.RoleBinding{
		Cluster:     c.ClusterName,
		Namespace:   c.Namespace,
		Kind:        c.Kind,
		Name:        c.Name,
		RoleRefName: c.RoleRef.Name,
		RoleRefKind: c.RoleRef.Kind,
		Subjects:    subjects,
	}
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

func (r *roleBindingAdapter) ToModel() models.RoleBinding {
	subjects := []models.Subject{}

	for _, s := range r.Subjects {
		subjects = append(subjects, models.Subject{
			Kind:      s.Kind,
			Name:      s.Name,
			Namespace: s.Namespace,
			APIGroup:  s.APIGroup,
		})
	}

	rb := models.RoleBinding{
		Cluster:     r.ClusterName,
		Namespace:   r.Namespace,
		Kind:        r.Kind,
		Name:        r.Name,
		Subjects:    subjects,
		RoleRefName: r.RoleRef.Name,
		RoleRefKind: r.RoleRef.Kind,
	}

	return rb
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

type EventLike interface {
	client.Object
	GetConditions() []metav1.Condition
}

type eventAdapter struct {
	*core.Event
}

func (ea *eventAdapter) GetConditions() []metav1.Condition {
	cond := metav1.Condition{
		Type:    string(NoStatus),
		Message: ea.Message,
		Status:  "True",
	}
	return []metav1.Condition{cond}
}
