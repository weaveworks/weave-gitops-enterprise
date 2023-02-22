package applicationscollector

import (
	"context"
	"fmt"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var DefaultVerbsRequiredForAccess = []string{"list"}

// ApplicationsCollector is responsible for collecting flux application resources from all clusters
// It is a wrapper around a generic collector that adapts the records and writes them to
// an store
type ApplicationsCollector struct {
	col   collector.Collector
	log   logr.Logger
	store store.StoreWriter
	quit  chan struct{}
}

func (a *ApplicationsCollector) Start(ctx context.Context) error {
	err := a.col.Start()
	if err != nil {
		return fmt.Errorf("could not start access collector: %store", err)
	}
	return nil
}

func (a *ApplicationsCollector) Stop() error {
	a.quit <- struct{}{}
	return a.col.Stop()
}

func NewApplicationsCollector(w store.Store, opts collector.CollectorOpts) (*ApplicationsCollector, error) {

	opts.ObjectKinds = []schema.GroupVersionKind{
		v2beta1.GroupVersion.WithKind("HelmRelease"),
		v1beta2.GroupVersion.WithKind("Kustomization"),
	}

	col, err := collector.NewCollector(opts, w, defaultProcessRecords, nil)

	if err != nil {
		return nil, fmt.Errorf("cannot create collector: %store", err)
	}
	return &ApplicationsCollector{
		col:   col,
		log:   opts.Log,
		store: w,
	}, nil
}

func defaultProcessRecords(ctx context.Context, objectRecords []models.ObjectRecord, store store.Store, log logr.Logger) error {
	objects, err := adaptObjects(objectRecords)
	if err != nil {
		return fmt.Errorf("cannot adapt object: %store", err)
	}
	if err := store.StoreObjects(ctx, objects); err != nil {
		return fmt.Errorf("cannot store object: %store", err)
	}
	return nil
}

// TODO: allow to overwrite the function
// default adapt function
func adaptObjects(objectRecords []models.ObjectRecord) ([]models.Object, error) {

	objects := []models.Object{}

	for _, objectRecord := range objectRecords {
		object := models.Object{
			Cluster:   objectRecord.ClusterName(),
			Name:      objectRecord.Object().GetName(),
			Namespace: objectRecord.Object().GetNamespace(),
			Kind:      objectRecord.Object().GetObjectKind().GroupVersionKind().Kind,
			Operation: "not available",
			Status:    "not available",
			Message:   "not available",
		}
		objects = append(objects, object)
	}

	return objects, nil

}
