package accesschecker

import (
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func RelevantRulesForUser(user *auth.UserPrincipal, rules []models.AccessRule) []models.AccessRule {
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
