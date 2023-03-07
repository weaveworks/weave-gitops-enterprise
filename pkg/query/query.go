package query

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

// QueryService is an all-in-one service that handles managing a collector, writing to the store, and responding to queries
type QueryService interface {
	RunQuery(ctx context.Context) ([]models.Object, error)
}

type QueryServiceOpts struct {
	Log         logr.Logger
	StoreReader store.StoreReader
}

func NewQueryService(ctx context.Context, opts QueryServiceOpts) (QueryService, error) {
	return &qs{
		log: opts.Log,
		r:   opts.StoreReader,
	}, nil
}

type qs struct {
	log logr.Logger
	r   store.StoreReader
}

func (q *qs) RunQuery(ctx context.Context) ([]models.Object, error) {
	principal := *auth.Principal(ctx)

	return q.r.Query(principal.Groups)
}
