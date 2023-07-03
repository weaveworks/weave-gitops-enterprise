package query

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/core/logger"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

// QueryService is an all-in-one service that handles managing a collector, writing to the store, and responding to queries
type QueryService interface {
	RunQuery(ctx context.Context, q store.Query, opts store.QueryOption) ([]models.Object, error)
	ListFacets(ctx context.Context) (store.Facets, error)
	GetAccessRules(ctx context.Context) ([]models.AccessRule, error)
}

// Authorizer creates an authorization predicate when given a cluster name.
type Authorizer interface {
	ObjectAuthorizer(roles []models.Role, rolebindings []models.RoleBinding, principal *auth.UserPrincipal, cluster string) func(models.Object) (bool, error)
}

type QueryServiceOpts struct {
	Log         logr.Logger
	StoreReader store.StoreReader
	IndexReader store.IndexReader
	Authorizer  Authorizer
}

func (o QueryServiceOpts) Validate() error {
	if o.StoreReader == nil {
		return fmt.Errorf("store reader is required")
	}
	if o.IndexReader == nil {
		return fmt.Errorf("index reader is required")
	}
	if o.Authorizer == nil {
		return fmt.Errorf("authorizer is required")
	}
	return nil
}

const (
	OperandIncludes = "includes"
)

func NewQueryService(opts QueryServiceOpts) (QueryService, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	return &qs{
		log:        opts.Log.WithName("query-service"),
		debug:      opts.Log.WithName("query-service").V(logger.LogLevelDebug),
		r:          opts.StoreReader,
		index:      opts.IndexReader,
		authorizer: opts.Authorizer,
	}, nil
}

type qs struct {
	log        logr.Logger
	debug      logr.Logger
	r          store.StoreReader
	index      store.IndexReader
	authorizer Authorizer
}

func (q *qs) RunQuery(ctx context.Context, query store.Query, opts store.QueryOption) ([]models.Object, error) {
	principal := auth.Principal(ctx)
	if principal == nil {
		return nil, fmt.Errorf("principal not found")
	}
	q.debug.Info("query received", "query", query, "principal", principal.ID)

	roles, err := q.r.GetRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching access rules from the store: %w", err)
	}
	bindings, err := q.r.GetRoleBindings(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching access rules from the store: %w", err)
	}

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

	// keep track of any cluster authorize predicate we might need again
	perClusterAllowed := map[string](func(models.Object) (bool, error)){}

	for iter.Next() {
		if limit > 0 && len(result) == int(limit) {
			// Limit is set in the query and reached.
			// If Limit is 0, all objects are returned.
			break
		}

		obj, err := iter.Row()
		if err != nil {
			q.log.Error(err, "error getting row from iterator")
			continue
		}

		cluster := obj.Cluster
		allow, ok := perClusterAllowed[cluster]
		if !ok {
			allow = q.authorizer.ObjectAuthorizer(roles, bindings, principal, cluster)
			perClusterAllowed[cluster] = allow
		}

		ok, err = allow(obj)
		if err != nil {
			q.log.Error(err, "error checking access")
			continue
		}

		if ok {
			result = append(result, obj)
		} else {
			//unauthorised is logged for debugging
			q.debug.Info("unauthorised access", "principal", principal.ID, "object", obj.ID)
		}
	}

	q.debug.Info("query processed", "query", query, "principal", principal.ID, "numResult", len(result))
	return result, nil
}

func (q *qs) GetAccessRules(ctx context.Context) ([]models.AccessRule, error) {
	return q.r.GetAccessRules(ctx)
}

func (q *qs) ListFacets(ctx context.Context) (store.Facets, error) {
	return q.index.ListFacets(ctx)
}
