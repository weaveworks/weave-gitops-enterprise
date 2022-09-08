package tenancy

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	rbacv1 "k8s.io/api/rbac/v1"
)

func Test_mergePolicyRules(t *testing.T) {
	mergeTests := []struct {
		name   string
		rules1 []rbacv1.PolicyRule
		rules2 []rbacv1.PolicyRule
		want   []rbacv1.PolicyRule
	}{
		{
			name:   "no rules",
			rules1: []rbacv1.PolicyRule{},
			rules2: []rbacv1.PolicyRule{},
			want:   []rbacv1.PolicyRule{},
		},
		{
			name: "no duplication",
			rules1: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"namespaces", "pods"},
					Verbs:     []string{"list", "get"},
				},
			},
			rules2: []rbacv1.PolicyRule{},
			want: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"namespaces", "pods"},
					Verbs:     []string{"get", "list"},
				},
			},
		},
		{
			name: "duplicate resources",
			rules1: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"namespaces", "pods"},
					Verbs:     []string{"list", "get"},
				},
			},
			rules2: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"namespaces", "pods"},
					Verbs:     []string{"create", "delete", "update"},
				},
			},
			want: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"namespaces", "pods"},
					Verbs:     []string{"create", "delete", "get", "list", "update"},
				},
			},
		},
		{
			name: "non-duplicate resources",
			rules1: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"namespaces", "pods"},
					Verbs:     []string{"list", "get"},
				},
			},
			rules2: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"configmaps"},
					Verbs:     []string{"create", "update", "list", "get"},
				},
			},
			want: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"namespaces", "pods"},
					Verbs:     []string{"get", "list"},
				},
				{
					APIGroups: []string{""},
					Resources: []string{"configmaps"},
					Verbs:     []string{"create", "get", "list", "update"},
				},
			},
		},
	}

	for _, tt := range mergeTests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, mergePolicyRules(tt.rules1, tt.rules2)); diff != "" {
				t.Fatalf("failed to merge:\n%s", diff)
			}
		})
	}
}
