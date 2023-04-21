package accesschecker

import (
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// Checker is responsible for checking if a subject has access to a resource.
//
//counterfeiter:generate . Checker
type Checker interface {
	// HasAccess checks if a subject has access to a resource.
	HasAccess(user *auth.UserPrincipal, object models.Object, rules []models.AccessRule) (bool, error)
	// RelevantRulesForUser returns all the AccessRules that are relevant to a user.
	// This is based on their ID and the groups they belong to.
	// Useful for debugging mostly.
	RelevantRulesForUser(user *auth.UserPrincipal, rules []models.AccessRule) []models.AccessRule
}

type defaultAccessChecker struct {
	kindByResourceMap map[string]string
}

// HasAccess checks if a principal has access to an object given principal access rules
func (a *defaultAccessChecker) HasAccess(user *auth.UserPrincipal, object models.Object, rules []models.AccessRule) (bool, error) {
	for _, rule := range rules {
		if rule.Cluster != object.Cluster {
			// Not the same cluster, so not relevant.
			continue
		}

		if rule.Namespace != "" && rule.Namespace != object.Namespace {
			// ClusterRoles and ClusterRoleBindings are not namespaced, so we only check if the field is
			continue
		}

		// A RBAC policyRule includes a set of <ApiGroup,Resource> https://kubernetes.io/docs/reference/kubernetes-api/authorization-resources/role-v1/
		// It will allow access if both apiGroups and resources allow access
		for _, gr := range rule.AccessibleKinds {
			var resourceName string
			// The GVK is in the format <group>/<version>/<kind>, so we need to split it and check for `*`.
			// Sometimes the version is not present, so we need to handle that case.
			parts := strings.Split(gr, "/")

			if len(parts) == 3 {
				resourceName = parts[2]
			} else if len(parts) == 2 {
				resourceName = parts[1]
			} else {
				return false, fmt.Errorf("invalid GVK: %s", gr)
			}
			ruleGroup := parts[0]

			//apigroups should be the same
			if ruleGroup != object.APIGroup {
				continue
			}

			//wildcard allows any
			if strings.Contains(resourceName, "*") {
				return true, nil
			}
			//find whether resource allows kind
			if a.kindByResourceMap[resourceName] == object.Kind {
				return true, nil
			}
		}
	}
	return false, nil
}

func (a *defaultAccessChecker) RelevantRulesForUser(user *auth.UserPrincipal, rules []models.AccessRule) []models.AccessRule {
	matchingRules := []models.AccessRule{}

	for _, rule := range rules {
		if rule.AccessibleKinds == nil || len(rule.AccessibleKinds) == 0 {
			// Not sure how this rule got created, but it doesn't provide any kinds, so ignore.
			continue
		}

		for _, subject := range rule.Subjects {
			if subject.Kind == "User" && subject.Name == user.ID {
				matchingRules = append(matchingRules, rule)
				continue
			}

			for _, group := range user.Groups {
				if subject.Kind == "Group" && subject.Name == group {
					matchingRules = append(matchingRules, rule)
				}
			}
		}
	}

	return matchingRules
}

// NewAccessChecker returns a new AccessChecker configured with a set of allowed resources
// and kinds it could check access to
func NewAccessChecker(kindByResourceMap map[string]string) (Checker, error) {
	return &defaultAccessChecker{
		kindByResourceMap: kindByResourceMap,
	}, nil
}
