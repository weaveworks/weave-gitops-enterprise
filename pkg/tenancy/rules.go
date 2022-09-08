package tenancy

import (
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type mergeRule struct {
	verbs     sets.String
	apiGroups []string
	resources []string
}

func mergePolicyRules(rules1, rules2 []rbacv1.PolicyRule) []rbacv1.PolicyRule {
	resKey := func(r rbacv1.PolicyRule) string {
		return strings.Join(r.APIGroups, ",") + ":" + strings.Join(r.Resources, ",")
	}

	// map of resource to merges
	knownResources := map[string]mergeRule{}
	for _, r1 := range rules1 {
		key := resKey(r1)
		merge, ok := knownResources[key]
		if !ok {
			merge = mergeRule{apiGroups: r1.APIGroups, resources: r1.Resources, verbs: sets.NewString()}
		}
		merge.verbs.Insert(r1.Verbs...)
		knownResources[key] = merge
	}
	for _, r2 := range rules2 {
		key := resKey(r2)
		merge, ok := knownResources[key]
		if !ok {
			merge = mergeRule{apiGroups: r2.APIGroups, resources: r2.Resources, verbs: sets.NewString()}
		}
		merge.verbs.Insert(r2.Verbs...)
		knownResources[key] = merge
	}

	result := []rbacv1.PolicyRule{}
	for _, mr := range knownResources {
		result = append(result, rbacv1.PolicyRule{APIGroups: mr.apiGroups, Resources: mr.resources, Verbs: mr.verbs.List()})
	}

	return result
}
