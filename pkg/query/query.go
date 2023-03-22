package query

import (
	"context"
	"fmt"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/models"

	"github.com/go-logr/logr"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

// QueryService is an all-in-one service that handles managing a collector, writing to the store, and responding to queries
type QueryService interface {
	RunQuery(ctx context.Context, q store.Query) ([]models.Object, error)
	GetAccessRules(ctx context.Context) ([]models.AccessRule, error)
}

type StoreService interface {
	StoreAccessRules(ctx context.Context, rules []models.AccessRule) error
	StoreObjects(ctx context.Context, objs []models.Object) error
}

type QueryServiceOpts struct {
	Log         logr.Logger
	StoreReader store.StoreReader
}

type StoreServiceOpts struct {
	Log         logr.Logger
	StoreWriter store.StoreWriter
}

const (
	OperandIncludes = "includes"
)

func NewQueryService(ctx context.Context, opts QueryServiceOpts) (QueryService, error) {
	return &qs{
		log:    opts.Log,
		r:      opts.StoreReader,
		filter: defaultAccessFilter,
	}, nil
}

func NewStoreService(ctx context.Context, opts StoreServiceOpts) (StoreService, error) {
	return &ss{
		log: opts.Log,
		w:   opts.StoreWriter,
	}, nil
}

type qs struct {
	log    logr.Logger
	r      store.StoreReader
	filter AccessFilter
}

type ss struct {
	log logr.Logger
	w   store.StoreWriter
}

func (s ss) StoreObjects(ctx context.Context, objs []models.Object) error {
	err := s.w.StoreObjects(ctx, objs)
	if err != nil {
		return fmt.Errorf("error writting objects to store: %w", err)
	}
	return nil
}

func (s ss) StoreAccessRules(ctx context.Context, rules []models.AccessRule) error {
	err := s.w.StoreAccessRules(ctx, rules)
	if err != nil {
		return fmt.Errorf("cannot store acess rules: %w", err)
	}
	return nil
}

type AccessFilter func(principal *auth.UserPrincipal, rules []models.AccessRule, objects []models.Object) []models.Object

func (q *qs) RunQuery(ctx context.Context, opts store.Query) ([]models.Object, error) {
	principal := auth.Principal(ctx)

	allObjects, err := q.r.GetObjects(ctx, opts)

	if err != nil {
		return nil, fmt.Errorf("error getting objects from store: %w", err)
	}

	rules, err := q.r.GetAccessRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting access rules: %w", err)
	}

	result := q.filter(principal, rules, allObjects)

	return result, nil
}

func (q *qs) GetAccessRules(ctx context.Context) ([]models.AccessRule, error) {
	return q.r.GetAccessRules(ctx)
}

func defaultAccessFilter(user *auth.UserPrincipal, rules []models.AccessRule, objects []models.Object) []models.Object {
	ruleLookup := map[string]bool{}

	result := []models.Object{}

	for _, rule := range rules {
		for _, kind := range rule.AccessibleKinds {
			key := toKey(rule.Cluster, rule.Namespace, kind, rule.Principal)
			ruleLookup[key] = true
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
