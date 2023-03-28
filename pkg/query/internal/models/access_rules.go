package models

import "strings"

// AccessRule represents a rule that determines if a subject has access to a resource.
// It is not stored in the database and exists as an abstraction over Role/RoleBinding pairs.
type AccessRule struct {
	Cluster           string
	Namespace         string
	AccessibleKinds   []string
	Subjects          []Subject
	ProvidedByRole    string
	ProvidedByBinding string
}

func ContainsWildcard(permissions []string) bool {
	for _, p := range permissions {
		if p == "*" || strings.Contains(p, "*") {
			return true
		}
	}

	return false
}
