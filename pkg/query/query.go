package query

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/core/logger"

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
	Log           logr.Logger
	StoreReader   store.StoreReader
	AccessChecker accesschecker.Checker
	IndexReader   store.IndexReader
}

const (
	OperandIncludes = "includes"
)

func NewQueryService(opts QueryServiceOpts) (QueryService, error) {
	return &qs{
		log:     opts.Log.WithName("query-service"),
		debug:   opts.Log.WithName("query-service").V(logger.LogLevelDebug),
		r:       opts.StoreReader,
		checker: opts.AccessChecker,
		index:   opts.IndexReader,
	}, nil
}

type qs struct {
	log     logr.Logger
	debug   logr.Logger
	r       store.StoreReader
	checker accesschecker.Checker
	index   store.IndexReader
}

type AccessFilter func(principal *auth.UserPrincipal, rules []models.AccessRule, objects []models.Object) []models.Object

func (q *qs) RunQuery(ctx context.Context, query store.Query, opts store.QueryOption) ([]models.Object, error) {
	principal := auth.Principal(ctx)
	if principal == nil {
		return nil, fmt.Errorf("principal not found")
	}
	q.debug.Info("query received", "query", query, "principal", principal.ID)

	// Contains all the rules that are relevant to this user.
	// This is based on their ID and the groups they belong to.
	rules, err := q.r.GetAccessRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting access rules: %w", err)
	}
	rules = q.checker.RelevantRulesForUser(principal, rules)

	iter, err := q.index.Search(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting objects from indexer: %w", err)
	}

	defer iter.Close()

	result := []models.Object{}

	var limit int32

	if opts != nil {
		limit = opts.GetLimit()
	}

	for iter.Next() {
		if limit > 0 && len(result) == int(limit) {
			// Limit is set in the query and reached.
			// If Limit is 0, all objects are returned.
			break
		}

		obj, err := iter.Row()
		if err != nil {
			return nil, fmt.Errorf("error getting row from iterator: %w", err)
		}

		ok, err := q.checker.HasAccess(principal, obj, rules)
		if err != nil {
			q.log.Error(err, "error checking access")
			continue
		}

		if ok {
			result = append(result, obj)
		} else {
			//unauthorised is logged for debugging
			q.debug.Info("unauthorised access", "principal", principal.ID, "object", obj.ID, "rules", rules)
		}
	}

	q.debug.Info("query processed", "query", query, "principal", principal.ID, "numResult", len(result))
	return result, nil
}

func (q *qs) GetAccessRules(ctx context.Context) ([]models.AccessRule, error) {
	return q.r.GetAccessRules(ctx)
}
