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
	log      logr.Logger
	interval time.Duration
	config   []configuration.ObjectKind
	idx      store.IndexWriter
	store    store.Store
	stop     chan bool
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
		log:      opts.Log,
		interval: opts.Interval,
		config:   opts.Config,
		idx:      opts.Index,
		store:    opts.Store,
	}, nil
}

func (oc *objectCleaner) Start() error {
	ticker := time.Tick(oc.interval)
	stop := make(chan bool, 1)
	oc.stop = stop

	go func() {
		for {
			select {
			case <-ticker:
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
			if k.String() == obj.GroupVersionKind() {
				objKind := oc.config[i]

				if models.IsExpired(objKind.RetentionPolicy, obj) {
					if err := oc.store.DeleteObjects(ctx, []models.Object{obj}); err != nil {
						return fmt.Errorf("could not delete object: %w", err)
					}
				}
			}
		}

	}
	return nil
}
