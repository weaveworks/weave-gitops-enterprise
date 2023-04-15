package objectscollector

import (
	"context"
	"fmt"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/adapters"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/logger"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ObjectsCollector is responsible for collecting flux application resources from all clusters
// It is a wrapper around a generic collector that adapts the records and writes them to
// an store
type ObjectsCollector struct {
	col   collector.Collector
	log   logr.Logger
	store store.StoreWriter
	quit  chan struct{}
}

func (a *ObjectsCollector) Start(ctx context.Context) error {
	err := a.col.Start(ctx)
	if err != nil {
		return fmt.Errorf("could not start objects collector: %w", err)
	}
	a.log.V(logger.LogLevelDebug).Info("objects collector started")
	return nil
}

func (a *ObjectsCollector) Stop(ctx context.Context) error {
	a.quit <- struct{}{}
	err := a.col.Stop(ctx)
	if err != nil {
		return fmt.Errorf("could not stop objects collector: %w", err)
	}
	a.log.V(logger.LogLevelDebug).Info("objects collector stopped")
	return nil
}

func NewObjectsCollector(w store.Store, opts collector.CollectorOpts) (*ObjectsCollector, error) {

	opts.ObjectKinds = []schema.GroupVersionKind{
		v2beta1.GroupVersion.WithKind("HelmRelease"),
		v1beta2.GroupVersion.WithKind("Kustomization"),
	}

	opts.ProcessRecordsFunc = defaultProcessRecords

	col, err := collector.NewCollector(opts, w)

	if err != nil {
		return nil, fmt.Errorf("cannot create collector: %store", err)
	}
	return &ObjectsCollector{
		col:   col,
		log:   opts.Log.WithName("objects-collector"),
		store: w,
	}, nil
}

func defaultProcessRecords(ctx context.Context, objectTransactions []models.ObjectTransaction, store store.Store, log logr.Logger) error {

	upsert := []models.Object{}
	delete := []models.Object{}
	deleteAll := []string{} //holds the cluster names to delete all resources

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

		object := models.Object{
			Cluster:    objTx.ClusterName(),
			Name:       objTx.Object().GetName(),
			Namespace:  objTx.Object().GetNamespace(),
			APIGroup:   gvk.Group,
			APIVersion: gvk.Version,
			Kind:       gvk.Kind,
			Status:     string(adapters.Status(o)),
			Message:    adapters.Message(o),
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
	}

	if len(delete) > 0 {
		if err := store.DeleteObjects(ctx, delete); err != nil {
			return fmt.Errorf("failed to delete objects: %w", err)
		}
	}

	if len(deleteAll) > 0 {
		if err := store.DeleteAllObjects(ctx, deleteAll); err != nil {
			return fmt.Errorf("failed to delete all objects: %w", err)
		}
	}

	return nil
}
