package query

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/accesschecker"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/models"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

// QueryService is an all-in-one service that handles managing a collector, writing to the store, and responding to queries
type QueryService interface {
	store.StoreWriter
	RunQuery(ctx context.Context, q store.Query, opts store.QueryOption) ([]models.Object, error)
	GetAccessRules(ctx context.Context) ([]models.AccessRule, error)
}

type QueryServiceOpts struct {
	Log   logr.Logger
	Store store.Store
}

const (
	OperandIncludes = "includes"
)

func NewQueryService(ctx context.Context, opts QueryServiceOpts) (QueryService, error) {
	return &qs{
		log:     opts.Log,
		store:   opts.Store,
		checker: accesschecker.NewAccessChecker(),
	}, nil
}

type qs struct {
	log     logr.Logger
	store   store.Store
	checker accesschecker.Checker
}

func (s qs) DeleteRoles(ctx context.Context, roles []models.Role) error {
	err := s.store.DeleteRoles(ctx, roles)
	if err != nil {
		return fmt.Errorf("error writting roles to delete: %w", err)
	}
	return nil
}

func (s qs) DeleteRoleBindings(ctx context.Context, roleBindings []models.RoleBinding) error {
	err := s.store.DeleteRoleBindings(ctx, roleBindings)
	if err != nil {
		return fmt.Errorf("error writting rolebindings to delete: %w", err)
	}
	return nil
}

func (s qs) DeleteObjects(ctx context.Context, objs []models.Object) error {
	err := s.store.DeleteObjects(ctx, objs)
	if err != nil {
		return fmt.Errorf("error writting objects to delete: %w", err)
	}
	return nil
}

func (s qs) StoreObjects(ctx context.Context, objs []models.Object) error {
	err := s.store.StoreObjects(ctx, objs)
	if err != nil {
		return fmt.Errorf("error writting objects to store: %w", err)
	}
	return nil
}

func (s qs) StoreRoles(ctx context.Context, roles []models.Role) error {
	err := s.store.StoreRoles(ctx, roles)
	if err != nil {
		return fmt.Errorf("cannot store roles: %w", err)
	}
	return nil
}

func (s qs) StoreRoleBindings(ctx context.Context, roleBindings []models.RoleBinding) error {
	err := s.store.StoreRoleBindings(ctx, roleBindings)
	if err != nil {
		return fmt.Errorf("cannot store role bindings: %w", err)
	}
	return nil
}

type AccessFilter func(principal *auth.UserPrincipal, rules []models.AccessRule, objects []models.Object) []models.Object

func (q *qs) RunQuery(ctx context.Context, query store.Query, opts store.QueryOption) ([]models.Object, error) {
	principal := auth.Principal(ctx)

	if principal == nil {
		return nil, fmt.Errorf("principal not found")
	}

	allObjects, err := q.store.GetObjects(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting objects from store: %w", err)
	}

	rules, err := q.store.GetAccessRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting access rules: %w", err)
	}

	result := []models.Object{}
	for _, obj := range allObjects {
		ok, err := q.checker.HasAccess(principal, obj, rules)
		if err != nil {
			q.log.Error(err, "error checking access")
			continue
		}

		if ok {
			result = append(result, obj)
		}
	}

	return result, nil
}

func (q *qs) GetAccessRules(ctx context.Context) ([]models.AccessRule, error) {
	return q.store.GetAccessRules(ctx)
}

func defaultAccessFilter(user *auth.UserPrincipal, rules []models.AccessRule, objects []models.Object) []models.Object {
	ruleLookup := map[string]bool{}

	result := []models.Object{}

	for _, rule := range rules {
		for _, s := range rule.Subjects {
			for _, kind := range rule.AccessibleKinds {
				fmt.Println(rule.Namespace)
				key := toKey(rule.Cluster, rule.Namespace, kind, s.Name)
				ruleLookup[key] = true
			}
		}
	}

	for _, o := range objects {
		hasAccess := false
		for _, group := range user.Groups {
			key := toKey(o.Cluster, o.Namespace, o.Kind, group)
			if ruleLookup[key] {
				hasAccess = true
				break
			}
		}

		idKey := toKey(o.Cluster, o.Namespace, o.Kind, user.ID)

		if ruleLookup[idKey] {
			hasAccess = true
		}

		if hasAccess {
			result = append(result, o)
		}

	}

	return result
}

func toKey(cluster, namespace, kind, role string) string {
	return fmt.Sprintf("%s/%s/%s/%s", cluster, namespace, kind, role)
}
