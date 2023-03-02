package query

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/memorystore"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

type QueryService interface {
	Start() error
	Stop()
	RunQuery(ctx context.Context) ([]models.Object, error)
}

func NewQueryService(ctx context.Context, log logr.Logger, col Collector, w StoreWriter, r StoreReader, rulesTicker *time.Ticker, objectsTicker *time.Ticker) (QueryService, error) {
	return &qs{
		log:           log,
		collector:     col,
		storeWriter:   w,
		rulesTicker:   rulesTicker,
		objectsTicker: objectsTicker,
		storeReader:   r,
	}, nil
}

func NewInMemoryStore() StoreWriter {
	return store.NewInMemoryStore()
}

type qs struct {
	log           logr.Logger
	collector     Collector
	storeWriter   StoreWriter
	storeReader   StoreReader
	rulesTicker   *time.Ticker
	objectsTicker *time.Ticker
	done          chan (bool)
}

func (q *qs) Start() error {
	q.log.Info("starting query service")
	q.done = make(chan bool)
	go func() {
		for {
			select {
			case <-q.rulesTicker.C:
				accessRules, err := q.collector.CollectAccessRules()
				if err != nil {
					q.log.Error(err, "failed to collect access rules")
				}

				if err := q.storeWriter.StoreAccessRules(accessRules); err != nil {
					q.log.Error(err, "failed to store access rules")
				}

			case <-q.done:
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-q.objectsTicker.C:
				fmt.Println("tick")
				objs, err := q.collector.CollectObjects()
				if err != nil {
					q.log.Error(err, "failed to collect objects")
				}

				fmt.Println(objs)
				if err := q.storeWriter.StoreObjects(objs); err != nil {
					q.log.Error(err, "failed to store objects")
				}

			case <-q.done:
				return
			}
		}
	}()

	return nil
}

func (q *qs) Stop() {
	q.log.Info("stopping query service")
	q.rulesTicker.Stop()
	q.done <- true
}

func (q *qs) RunQuery(ctx context.Context) ([]models.Object, error) {
	principal := *auth.Principal(ctx)

	return q.storeReader.Query(principal.Groups)
}
