package query

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/accesschecker"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

// QueryService is an all-in-one service that handles managing a collector, writing to the store, and responding to queries
type QueryService interface {
	RunQuery(ctx context.Context, q store.Query, opts store.QueryOption) ([]models.Object, error)
	GetAccessRules(ctx context.Context) ([]models.AccessRule, error)
}

type QueryServiceOpts struct {
	Log         logr.Logger
	StoreReader store.StoreReader
}

const (
	OperandIncludes = "includes"
)

func NewQueryService(ctx context.Context, opts QueryServiceOpts) (QueryService, error) {
	return &qs{
		log:     opts.Log,
		r:       opts.StoreReader,
		checker: accesschecker.NewAccessChecker(opts.Log),
	}, nil
}

type qs struct {
	log     logr.Logger
	r       store.StoreReader
	checker accesschecker.Checker
}

type AccessFilter func(principal *auth.UserPrincipal, rules []models.AccessRule, objects []models.Object) []models.Object

func (q *qs) RunQuery(ctx context.Context, query store.Query, opts store.QueryOption) ([]models.Object, error) {
	principal := auth.Principal(ctx)

	if principal == nil {
		return nil, fmt.Errorf("principal not found")
	}

	allObjects, err := q.r.GetObjects(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting objects from store: %w", err)
	}

	rules, err := q.r.GetAccessRules(ctx)
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
	return q.r.GetAccessRules(ctx)
}
