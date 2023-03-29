package accesschecker

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/core/logger"
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
	log logr.Logger
}

// HasAccess checks if a principal has access to a resource.
func (a *defaultAccessChecker) HasAccess(user *auth.UserPrincipal, object models.Object, rules []models.AccessRule) (bool, error) {
	a.log.V(logger.LogLevelDebug).Info("has access request", "user", user.ID, "object", object)

	// Contains all the rules that are relevant to this user.
	// This is based on their ID and the groups they belong to.
	matchingRules := a.RelevantRulesForUser(user, rules)

	for _, rule := range matchingRules {
		a.log.V(logger.LogLevelDebug).Info("trying match", "rule", fmt.Sprintf("%v", rule))

		if rule.Cluster != object.Cluster {
			// Not the same cluster, so not relevant.
			continue
		}

		if rule.Namespace != "" && rule.Namespace != object.Namespace {
			// ClusterRoles and ClusterRoleBindings are not namespaced, so we only check if the field is
			continue
		}

		//Checks whether the rule allows object's kind
		//It will be allowed if the rule allows the apigroup and kind
		for _, gvk := range rule.AccessibleKinds {

			var ruleKind string
			// The GVK is in the format <group>/<version>/<kind>, so we need to split it and check for `*`.
			// Sometimes the version is not present, so we need to handle that case.
			parts := strings.Split(gvk, "/")

			if len(parts) == 3 {
				ruleKind = parts[2]
			} else if len(parts) == 2 {
				ruleKind = parts[1]
			} else {
				return false, fmt.Errorf("invalid GVK: %s", gvk)
			}
			ruleGroup := parts[0]

			if ruleGroup != object.APIGroup {
				continue
			}

			objectGVK := object.GroupVersionKind()

			if strings.Contains(ruleKind, "*") {
				// If the rule contains a wildcard, then the user has access to all kinds.
				return true, nil
			}

			//given roles rule contains apigroups https://kubernetes.io/docs/reference/kubernetes-api/authorization-resources/role-v1/
			// but not version, any version would work
			// we match on kind
			if ruleKind != object.Kind {
				return true, nil
			}

			//TODO: review with jordan when we match here
			// Check for an exact group/version/kind match.
			if gvk == objectGVK {
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

// NewAccessChecker returns a new AccessChecker.
func NewAccessChecker(log logr.Logger) Checker {
	return &defaultAccessChecker{
		log: log.WithName("access-checker"),
	}
}
