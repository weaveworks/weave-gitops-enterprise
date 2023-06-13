package retainer

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
)

type RetentionManager interface {
	Start() error
}

type retentionManager struct {
	log      logr.Logger
	interval time.Duration
	config   []configuration.ObjectKind
	idx      store.IndexWriter
	store    store.Store
}

func NewRetentionManager() RetentionManager {
	return &retentionManager{}
}

func (r *retentionManager) Start() error {

	ticker := time.Tick(r.interval)

	// Start a goroutine to execute the job
	go func() {
		for {
			<-ticker
			// Call your job function here
			if err := r.removeOldObjects(context.Background()); err != nil {
				r.log.Error(err, "could not remove old objects")
			}
		}
	}()

	return nil
}

func (r *retentionManager) removeOldObjects(ctx context.Context) error {
	iter, err := r.store.GetAllObjects(ctx)

	if err != nil {
		return fmt.Errorf("could not get all objects: %w", err)
	}

	all, err := iter.All()

	if err != nil {
		return fmt.Errorf("could not get all objects: %w", err)
	}

	for _, obj := range all {

		for i, k := range r.config {
			if k.String() == obj.GroupVersionKind() {
				objKind := r.config[i]

				if models.IsExpired(objKind.RetentionPolicy, obj) {

				}
			}
		}

	}
	return nil
}
