package accesschecker

import (
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

// AccessChecker is responsible for checking if a subject has access to a resource.
type AccessChecker interface {
	// HasAccess checks if a subject has access to a resource.
	HasAccess(user *auth.UserPrincipal, object models.Object, rules []models.AccessRule) (bool, error)
}

type defaultAccessChecker struct{}

func (a *defaultAccessChecker) HasAccess(user *auth.UserPrincipal, object models.Object, rules []models.AccessRule) (bool, error) {

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

	for _, rule := range matchingRules {
		if rule.Cluster != object.Cluster || rule.Namespace != object.Namespace {
			continue
		}

		for _, kind := range rule.AccessibleKinds {
			if kind == object.Kind && rule.Namespace == object.Namespace {
				return true, nil
			}
		}
	}

	return false, nil
}

// NewAccessChecker returns a new AccessChecker.
func NewAccessChecker() AccessChecker {
	return &defaultAccessChecker{}
}
