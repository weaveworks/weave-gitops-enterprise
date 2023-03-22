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

type defaultAccessChecker struct{}

// HasAccess checks if a principal has access to a resource.
func (a *defaultAccessChecker) HasAccess(user *auth.UserPrincipal, object models.Object, rules []models.AccessRule) (bool, error) {
	// Contains all the rules that are relevant to this user.
	// This is based on their ID and the groups they belong to.
	matchingRules := a.RelevantRulesForUser(user, rules)

	for _, rule := range matchingRules {
		if rule.Cluster != object.Cluster {
			// Not the same cluster, so not relevant.
			continue
		}

		if rule.Namespace != "" && rule.Namespace != object.Namespace {
			// ClusterRoles and ClusterRoleBindings are not namespaced, so we only check if the field is
			continue
		}

		for _, gvk := range rule.AccessibleKinds {

			var kind string
			// The GVK is in the format <group>/<version>/<kind>, so we need to split it and check for `*`.
			// Sometimes the version is not present, so we need to handle that case.
			parts := strings.Split(gvk, "/")
			if len(parts) == 3 {
				kind = parts[2]
			} else if len(parts) == 2 {
				kind = parts[1]
			} else {
				return false, fmt.Errorf("invalid GVK: %s", gvk)
			}

			if strings.Contains(kind, "*") {
				// If the rule contains a wildcard, then the user has access to all kinds.
				return true, nil
			}

			// Check for an exact group/version/kind match.
			if gvk == object.GroupVersionKind() {
				return true, nil
			}
		}
	}

	return false, nil
}

func (a *defaultAccessChecker) RelevantRulesForUser(user *auth.UserPrincipal, rules []models.AccessRule) []models.AccessRule {
	matchingRules := []models.AccessRule{}

	for _, rule := range rules {
		for _, subject := range rule.Subjects {
			if subject.Kind == "User" && subject.Name == user.ID {
				matchingRules = append(matchingRules, rule)
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

// NewAccessChecker returns a new AccessChecker.
func NewAccessChecker() Checker {
	return &defaultAccessChecker{}
}
