package objectscollector

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/client-go/rest"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/clusters"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/adapters"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/logger"
)

func NewObjectsCollector(w store.Store, idx store.IndexWriter, mgr clusters.Subscriber, sa collector.ImpersonateServiceAccount, kinds []configuration.ObjectKind, log logr.Logger) (collector.Collector, error) {
	incoming := make(chan []models.ObjectTransaction)
	go func() {
		for tx := range incoming {
			if err := processRecords(tx, w, idx, log); err != nil {
				log.Error(err, "could not process records")
			}
		}
	}()

	newWatcher := func(config *rest.Config, clusterName string) (collector.Watcher, error) {
		return collector.DefaultNewWatcher(config, sa, clusterName, incoming, kinds, log)
	}

	opts := collector.CollectorOpts{
		Log:            log,
		NewWatcherFunc: newWatcher,
		Clusters:       mgr,
	}

	col, err := collector.NewCollector(opts)
	if err != nil {
		return nil, fmt.Errorf("cannot create collector: %store", err)
	}

	return col, nil
}

func processRecords(objectTransactions []models.ObjectTransaction, store store.Store, idx store.IndexWriter, log logr.Logger) error {
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

		cat, err := adapters.Category(objTx.Object())
		if err != nil {
			log.Error(err, "failed to get category from flux object")
			continue
		}

		var deletedAt time.Time
		ts := objTx.Object().GetCreationTimestamp()
		if !ts.IsZero() {
			deletedAt = ts.Time
		}

		object := models.Object{
			Cluster:             objTx.ClusterName(),
			Name:                objTx.Object().GetName(),
			Namespace:           objTx.Object().GetNamespace(),
			APIGroup:            gvk.Group,
			APIVersion:          gvk.Version,
			Kind:                gvk.Kind,
			Status:              string(adapters.Status(o)),
			Message:             adapters.Message(o),
			Category:            cat,
			KubernetesDeletedAt: deletedAt,
		}

		if objTx.TransactionType() == models.TransactionTypeDelete {
			// We want to retain some objects longer than kubernetes does.
			// Objects like Events get removed in 1h by default on some cloud providers.
			// Users want to be able to see these events for longer than that.
			if !models.IsExpired(*objTx.RetentionPolicy(), object) {
				debug.Info("object is not expired, skipping", "object", object)
				// We need to upsert here to catch the kubernetes deleted at timestamp
				upsert = append(upsert, object)
				continue
			}
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

		for _, cluster := range deleteAll {
			if err := idx.RemoveByQuery(ctx, fmt.Sprintf("+cluster:%s", cluster)); err != nil {
				return fmt.Errorf("failed to delete all objects for cluster %q: %w", cluster, err)
			}
		}
	}

	debug.Info("objects processed", "upsert", upsert, "delete", delete, "deleteAll", deleteAll)
	return nil
}
