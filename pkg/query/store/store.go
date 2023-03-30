package store

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"gorm.io/gorm"
	"k8s.io/kubectl/pkg/util/slice"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Store
type Store interface {
	StoreWriter
	StoreReader
}

// StoreWriter is an interface for storing access rules and objects
//
//counterfeiter:generate . StoreWriter
type StoreWriter interface {
	StoreRoles(ctx context.Context, roles []models.Role) error
	StoreRoleBindings(ctx context.Context, roleBindings []models.RoleBinding) error
	StoreObjects(ctx context.Context, objects []models.Object) error
	DeleteObjects(ctx context.Context, object []models.Object) error
	DeleteRoles(ctx context.Context, roles []models.Role) error
	DeleteRoleBindings(ctx context.Context, roleBindings []models.RoleBinding) error
}

type QueryOperand string

const (
	OperandEqual    QueryOperand = "equal"
	OperandNotEqual QueryOperand = "not_equal"
)

type GlobalOperand string

const (
	GlobalOperandAnd GlobalOperand = "and"
	GlobalOperandOr  GlobalOperand = "or"
)

type Query []QueryClause

type QueryOption interface {
	GetLimit() int32
	GetOffset() int32
	GetOrderBy() string
	GetGlobalOperand() string
}

type QueryClause interface {
	GetKey() string
	GetOperand() string
	GetValue() string
}

// StoreReader is an interface for querying objects
//
//counterfeiter:generate . StoreReader
type StoreReader interface {
	GetObjects(ctx context.Context, q Query, opts QueryOption) ([]models.Object, error)
	GetAccessRules(ctx context.Context) ([]models.AccessRule, error)
}

type StorageBackend string

const (
	StorageBackendSQLite StorageBackend = "sqlite"
)

// factory method that by default creates a in memory store
func NewStore(backend StorageBackend, uri string, log logr.Logger) (Store, error) {
	switch backend {
	case StorageBackendSQLite:
		db, err := CreateSQLiteDB(uri)
		if err != nil {
			return nil, fmt.Errorf("error creating sqlite db: %w", err)
		}
		return NewSQLiteStore(db, log)
	default:
		return nil, fmt.Errorf("unknown storage backend: %s", backend)
	}

}

var DefaultVerbsRequiredForAccess = []string{"list"}

// DeriveAcceessRules computes the access rules for a given set of roles and role bindings.
// This is implemented as a helper function to keep this logic testable and storage backend agnostic.
func DeriveAccessRules(roles []models.Role, bindings []models.RoleBinding) []models.AccessRule {
	accessRules := []models.AccessRule{}

	// Figure out the binding/role pairs
	for _, role := range roles {
		for _, binding := range bindings {
			if bindingRoleMatch(binding, role) {
				rule := convertToAccessRule(role.Cluster, role, binding, DefaultVerbsRequiredForAccess)
				accessRules = append(accessRules, rule)
			}
		}

	}
	return accessRules
}

func bindingRoleMatch(binding models.RoleBinding, role models.Role) bool {
	if binding.Cluster != role.Cluster {
		return false
	}

	if binding.Namespace != role.Namespace {
		return false
	}

	if binding.RoleRefKind != role.Kind {
		return false
	}

	if binding.RoleRefName != role.Name {
		return false
	}

	return true
}

func convertToAccessRule(clusterName string, role models.Role, binding models.RoleBinding, requiredVerbs []string) models.AccessRule {
	rules := role.PolicyRules

	derivedAccess := map[string]map[string]bool{}

	// {wego.weave.works: {Application: true, Source: true}}
	for _, rule := range rules {
		for _, apiGroup := range models.SplitRuleData(rule.APIGroups) {

			if _, ok := derivedAccess[apiGroup]; !ok {
				derivedAccess[apiGroup] = map[string]bool{}
			}

			rList := models.SplitRuleData(rule.Resources)
			if models.ContainsWildcard(rList) {
				derivedAccess[apiGroup]["*"] = true
			}

			vList := models.SplitRuleData(rule.Verbs)
			if models.ContainsWildcard(vList) || hasVerbs(vList, requiredVerbs) {
				for _, resource := range rList {
					derivedAccess[apiGroup][resource] = true
				}
			}
		}
	}

	accessibleKinds := []string{}
	for group, resources := range derivedAccess {
		for k, v := range resources {
			if v {
				accessibleKinds = append(accessibleKinds, fmt.Sprintf("%s/%s", group, k))
			}
		}
	}

	return models.AccessRule{
		Cluster:           clusterName,
		Namespace:         role.Namespace,
		AccessibleKinds:   accessibleKinds,
		Subjects:          binding.Subjects,
		ProvidedByRole:    fmt.Sprintf("%s/%s", role.Kind, role.Name),
		ProvidedByBinding: fmt.Sprintf("%s/%s", binding.Kind, binding.Name),
	}
}

func hasVerbs(a, b []string) bool {
	for _, v := range b {
		if models.ContainsWildcard(a) {
			return true
		}
		if slice.ContainsString(a, v, nil) {
			return true
		}
	}

	return false
}

func SeedObjects(db *gorm.DB, rows []models.Object) error {
	withID := []models.Object{}

	for _, o := range rows {
		o.ID = o.GetID()
		withID = append(withID, o)
	}
	result := db.Create(&withID)

	return result.Error
}
