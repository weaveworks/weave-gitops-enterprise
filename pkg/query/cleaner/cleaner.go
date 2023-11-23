package cleaner

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/cleaner/metrics"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
)

type ObjectCleaner interface {
	Start() error
	Stop() error
}

type objectCleaner struct {
	log              logr.Logger
	ticker           *time.Ticker
	config           []configuration.ObjectKind
	idx              store.IndexWriter
	store            store.Store
	stop             chan bool
	status           string
	lastStatusChange time.Time
}

type CleanerOpts struct {
	Log      logr.Logger
	Interval time.Duration
	Config   []configuration.ObjectKind
	Store    store.Store
	Index    store.IndexWriter
}

const (
	CleanerStarting = "starting"
	CleanerStarted  = "started"
	CleanerStopped  = "stopped"
)

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
	oc.setStatus(CleanerStarting)
	stop := make(chan bool, 1)
	oc.stop = stop

	go func() {
		oc.setStatus(CleanerStarted)

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
	oc.setStatus(CleanerStopped)

	return nil
}

// setStatus sets cleaner status and records it as a metric.
func (oc *objectCleaner) setStatus(s string) {
	if oc.status != "" {
		metrics.CleanerWatcherDecrease(oc.status)
	}

	oc.lastStatusChange = time.Now()
	oc.status = s

	if oc.status != CleanerStopped {
		metrics.CleanerWatcherIncrease(oc.status)
	}
}

func recordCleanerMetrics(start time.Time, err error) {
	metrics.CleanerAddInflightRequests(-1)

	label := metrics.SuccessLabel
	if err != nil {
		label = metrics.FailedLabel
	}

	metrics.CleanerSetLatency(label, time.Since(start))
}

func (oc *objectCleaner) removeOldObjects(ctx context.Context) (retErr error) {
	// metrics
	metrics.CleanerAddInflightRequests(1)
	defer recordCleanerMetrics(time.Now(), retErr)

	iter, err := oc.store.GetAllObjects(ctx)

	if err != nil {
		return fmt.Errorf("could not get all objects: %w", err)
	}

	all, err := iter.All()

	if err != nil {
		return fmt.Errorf("could not iterate over objects: %w", err)
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
						oc.log.Error(err, "could not delete object with ID: %s", obj.ID)
					}

					if err := oc.idx.Remove(ctx, remove); err != nil {
						oc.log.Error(err, "could not delete object with ID: %s from index", obj.ID)
					}
				}
			}
		}

	}
	return nil
}
