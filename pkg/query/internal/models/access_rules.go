package models

import (
	"fmt"
	"strings"
)

// AccessRule represents a rule that determines if a subject has access to a resource.
// It is not stored in the database and exists as an abstraction over Role/RoleBinding pairs.
type AccessRule struct {
	Cluster   string
	Namespace string
	//TODO should this be renamed to GR from GVR?
	AccessibleKinds         []string
	Subjects                []Subject
	ProvidedByRole          string
	ProvidedByBinding       string
	AccessibleResourceNames []string
}

// String returns a string version of the access rule that includes cluster/namespace/rolebinding/role
// to mainly use in the context of auditing and debugging
func (r *AccessRule) String() interface{} {
	return fmt.Sprintf("%s/%s/%s/%s", r.Cluster, r.Namespace, r.ProvidedByBinding, r.ProvidedByRole)
}

func ContainsWildcard(permissions []string) bool {
	for _, p := range permissions {
		if p == "*" || strings.Contains(p, "*") {
			return true
		}
	}

	return false
}
