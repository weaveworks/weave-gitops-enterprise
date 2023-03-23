package models

import "strings"

type AccessRule struct {
	Cluster         string
	Namespace       string
	AccessibleKinds []string
	Subjects        []Subject
}

func ContainsWildcard(permissions []string) bool {
	for _, p := range permissions {
		if p == "*" || strings.Contains(p, "*") {
			return true
		}
	}

	return false
}
