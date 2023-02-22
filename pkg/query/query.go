package query

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

// QueryService is an all-in-one service that handles managing a collector, writing to the store, and responding to queries
type QueryService interface {
	RunQuery(ctx context.Context, q []Query) ([]models.Object, error)
	GetAccessRules(ctx context.Context) ([]models.AccessRule, error)
}

type QueryServiceOpts struct {
	Log         logr.Logger
	StoreReader store.StoreReader
}

const (
	OperandIncludes = "includes"
)

type Query interface {
	GetKey() string
	GetOperand() string
	GetValue() string
}

func NewQueryService(ctx context.Context, opts QueryServiceOpts) (QueryService, error) {
	return &qs{
		log:    opts.Log,
		r:      opts.StoreReader,
		filter: defaultAccessFilter,
	}, nil
}

type qs struct {
	log    logr.Logger
	r      store.StoreReader
	filter AccessFilter
}

type AccessFilter func(principal *auth.UserPrincipal, rules []models.AccessRule, objects []models.Object) []models.Object

func (q *qs) RunQuery(ctx context.Context, opts []Query) ([]models.Object, error) {
	principal := auth.Principal(ctx)

	allObjects, err := q.r.GetObjects()
	if err != nil {
		return nil, fmt.Errorf("error getting objects from store: %w", err)
	}

	rules, err := q.r.GetAccessRules()
	if err != nil {
		return nil, fmt.Errorf("error getting access rules: %w", err)
	}

	result := q.filter(principal, rules, allObjects)

	return result, nil
}

func (q *qs) GetAccessRules(ctx context.Context) ([]models.AccessRule, error) {
	return q.r.GetAccessRules()
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
