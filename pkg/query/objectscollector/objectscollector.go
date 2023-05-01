package objectscollector

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/adapters"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/logger"
)

// ObjectsCollector is responsible for collecting flux application resources from all clusters
// It is a wrapper around a generic collector that adapts the records and writes them to
// an store
type ObjectsCollector struct {
	col   collector.Collector
	log   logr.Logger
	store store.StoreWriter
	quit  chan struct{}
	idx   store.IndexWriter
}

func (a *ObjectsCollector) Start() error {
	err := a.col.Start()
	if err != nil {
		return fmt.Errorf("could not start objects collector: %w", err)
	}
	a.log.Info("objects collector started")
	return nil
}

func (a *ObjectsCollector) Stop() error {
	a.quit <- struct{}{}
	err := a.col.Stop()
	if err != nil {
		return fmt.Errorf("could not stop objects collector: %w", err)
	}
	a.log.Info("objects collector stopped")
	return nil
}

func NewObjectsCollector(w store.Store, idx store.IndexWriter, opts collector.CollectorOpts) (*ObjectsCollector, error) {
	if opts.ProcessRecordsFunc == nil {
		opts.ProcessRecordsFunc = defaultProcessRecords
	}

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid collector options: %w", err)
	}

	opts.IndexWriter = idx

	col, err := collector.NewCollector(opts, w)
	if err != nil {
		return nil, fmt.Errorf("cannot create collector: %store", err)
	}

	return &ObjectsCollector{
		col:   col,
		log:   opts.Log.WithName("objects-collector"),
		store: w,
		idx:   idx,
	}, nil
}

func defaultProcessRecords(objectTransactions []models.ObjectTransaction, store store.Store, idx store.IndexWriter, log logr.Logger) error {
	ctx := context.Background()
	upsert := []models.Object{}
	delete := []models.Object{}
	deleteAll := []string{} //holds the cluster names to delete all resources
	debug := log.V(logger.LogLevelDebug)

	for _, objTx := range objectTransactions {
		// Handle delete all tx first as does not hold objects
		if objTx.TransactionType() == models.TransactionTypeDeleteAll {
			deleteAll = append(deleteAll, objTx.ClusterName())
			continue
		}
		gvk := objTx.Object().GetObjectKind().GroupVersionKind()

		o, err := adapters.ToFluxObject(objTx.Object())
		if err != nil {
			log.Error(err, "failed to convert object to flux object")
			continue
		}

		cat, err := adapters.Category(o)
		if err != nil {
			log.Error(err, "failed to get category from flux object")
			continue
		}

		object := models.Object{
			Cluster:    objTx.ClusterName(),
			Name:       objTx.Object().GetName(),
			Namespace:  objTx.Object().GetNamespace(),
			APIGroup:   gvk.Group,
			APIVersion: gvk.Version,
			Kind:       gvk.Kind,
			Status:     string(adapters.Status(o)),
			Message:    adapters.Message(o),
			Category:   cat,
		}

		if objTx.TransactionType() == models.TransactionTypeDelete {
			delete = append(delete, object)
		} else {
			upsert = append(upsert, object)
		}
	}

	if len(upsert) > 0 {
		if err := store.StoreObjects(ctx, upsert); err != nil {
			return fmt.Errorf("failed to store objects: %w", err)
		}

		if err := idx.Add(ctx, upsert); err != nil {
			return fmt.Errorf("failed to index objects: %w", err)
		}
	}

	if len(delete) > 0 {
		if err := store.DeleteObjects(ctx, delete); err != nil {
			return fmt.Errorf("failed to delete objects: %w", err)
		}

		if err := idx.Remove(ctx, delete); err != nil {
			return fmt.Errorf("failed to delete objects from index: %w", err)
		}
	}

	if len(deleteAll) > 0 {
		if err := store.DeleteAllObjects(ctx, deleteAll); err != nil {
			return fmt.Errorf("failed to delete all objects: %w", err)
		}
	}

	debug.Info("objects processed", "upsert", upsert, "delete", delete, "deleteAll", deleteAll)
	return nil
}
