package query

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

// QueryService is an all-in-one service that handles managing a collector, writing to the store, and responding to queries
type QueryService interface {
	Start() error
	Stop()
	RunQuery(ctx context.Context) ([]models.Object, error)
}

type QueryServiceOpts struct {
	Log         logr.Logger
	Collector   collector.Collector
	StoreWriter store.StoreWriter
	StoreReader store.StoreReader
}

func NewQueryService(ctx context.Context, opts QueryServiceOpts) (QueryService, error) {
	return &qs{
		log:       opts.Log,
		collector: opts.Collector,
		w:         opts.StoreWriter,
		r:         opts.StoreReader,
	}, nil
}

type qs struct {
	log       logr.Logger
	collector collector.Collector
	w         store.StoreWriter
	r         store.StoreReader
	done      chan (bool)
}

func (q *qs) Start() error {
	q.log.Info("starting query service")
	// q.done = make(chan bool)

	// go func() {
	// 	ch, error := q.collector.Start()
	// 	if error != nil {
	// 		q.log.Error(error, "failed to start collector")
	// 		return
	// 	}

	// 	for {
	// 		select {
	// 		case o := <-ch:
	// 			result := []models.Object{}

	// 			for _, ob := range o {
	// 				result = append(result, convertK8sToModelObject(ob.Object()))
	// 			}

	// 			if err := q.w.StoreObjects(result); err != nil {
	// 				q.log.Error(err, "failed to store objects")
	// 			}

	// 		case <-q.done:
	// 			return
	// 		}
	// 	}
	// }()

	return nil
}

func (q *qs) Stop() {
	q.log.Info("stopping query service")
	if err := q.collector.Stop(); err != nil {
		q.log.Error(err, "failed to stop collector")
	}
	q.done <- true
	close(q.done)
}

func (q *qs) RunQuery(ctx context.Context) ([]models.Object, error) {
	principal := *auth.Principal(ctx)

	return q.r.Query(principal.Groups)
}
