package cleaner

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
)

type ObjectCleaner interface {
	Start() error
	Stop() error
}

type objectCleaner struct {
	log    logr.Logger
	ticker *time.Ticker
	config []configuration.ObjectKind
	idx    store.IndexWriter
	store  store.Store
	stop   chan bool
}

type CleanerOpts struct {
	Log      logr.Logger
	Interval time.Duration
	Config   []configuration.ObjectKind
	Store    store.Store
	Index    store.IndexWriter
}

func NewObjectCleaner(opts CleanerOpts) (ObjectCleaner, error) {
	return &objectCleaner{
		log:    opts.Log,
		config: opts.Config,
		idx:    opts.Index,
		store:  opts.Store,
		ticker: time.NewTicker(opts.Interval),
	}, nil
}

func (oc *objectCleaner) Start() error {
	stop := make(chan bool, 1)
	oc.stop = stop

	go func() {
		for {
			select {
			case <-oc.ticker.C:
				if err := oc.removeOldObjects(context.Background()); err != nil {
					oc.log.Error(err, "could not remove old objects")
				}
			case <-stop:
				return
			}

		}
	}()

	return nil
}

func (oc *objectCleaner) Stop() error {
	oc.stop <- true

	return nil
}

func (oc *objectCleaner) removeOldObjects(ctx context.Context) error {
	iter, err := oc.store.GetAllObjects(ctx)

	if err != nil {
		return fmt.Errorf("could not get all objects: %w", err)
	}

	all, err := iter.All()

	if err != nil {
		return fmt.Errorf("could not get all objects: %w", err)
	}

	for _, obj := range all {
		for i, k := range oc.config {

			kind := fmt.Sprintf("%s/%s", k.Gvk.GroupVersion().String(), k.Gvk.Kind)
			gvk := obj.GroupVersionKind()
			if kind == gvk {
				objKind := oc.config[i]

				if models.IsExpired(objKind.RetentionPolicy, obj) {
					remove := []models.Object{obj}

					if err := oc.store.DeleteObjects(ctx, remove); err != nil {
						oc.log.Error(err, "could not delete object")
					}

					if err := oc.idx.Remove(ctx, remove); err != nil {
						oc.log.Error(err, "could not delete object from index")
					}
				}
			}
		}

	}
	return nil
}
