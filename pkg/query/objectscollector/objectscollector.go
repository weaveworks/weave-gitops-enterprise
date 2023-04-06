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
	err := a.col.Start()
	if err != nil {
		return fmt.Errorf("could not start access collector: %store", err)
	}
	return nil
}

func (a *ObjectsCollector) Stop() error {
	a.quit <- struct{}{}
	return a.col.Stop()
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
		log:   opts.Log,
		store: w,
	}, nil
}

func defaultProcessRecords(ctx context.Context, objectRecords []models.ObjectTransaction, store store.Store, log logr.Logger) error {

	upsert := []models.Object{}
	delete := []models.Object{}

	for _, obj := range objectRecords {
		gvk := obj.Object().GetObjectKind().GroupVersionKind()

		o, err := adapters.ToFluxObject(obj.Object())
		if err != nil {
			log.Error(err, "failed to convert object to flux object")
			continue
		}

		object := models.Object{
			Cluster:    obj.ClusterName(),
			Name:       obj.Object().GetName(),
			Namespace:  obj.Object().GetNamespace(),
			APIGroup:   gvk.Group,
			APIVersion: gvk.Version,
			Kind:       gvk.Kind,
			Status:     string(adapters.Status(o)),
			Message:    adapters.Message(o),
		}

		if obj.TransactionType() == models.TransactionTypeDelete {
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

	return nil
}
