package rolecollector

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/client-go/rest"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/clusters"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/adapters"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
)

func NewRoleCollector(w store.Store, mgr clusters.Subscriber, sa collector.ImpersonateServiceAccount, log logr.Logger) (collector.Collector, error) {
	incoming := make(chan []models.ObjectTransaction)

	newWatcher := func(clusterName string, config *rest.Config) (collector.Starter, error) {
		return collector.NewWatcher(clusterName, config, configuration.SupportedRbacKinds, incoming, log)
	}

	deleteWatcher := func(clusterName string) error {
		tx := collector.NewDeleteAllTransaction(clusterName)
		return processRecords([]models.ObjectTransaction{tx}, w, log)
	}

	go func() {
		for updates := range incoming {
			if err := processRecords(updates, w, log); err != nil {
				log.Error(err, "could not process records")
			}
		}
	}()

	opts := collector.CollectorOpts{
		Name:            "rbac",
		Log:             log,
		NewWatcherFunc:  newWatcher,
		StopWatcherFunc: deleteWatcher,
		Clusters:        mgr,
		ServiceAccount:  sa,
	}

	col, err := collector.NewCollector(opts)

	if err != nil {
		return nil, fmt.Errorf("cannot create collector: %w", err)
	}
	return col, nil
}

func processRecords(objectTransactions []models.ObjectTransaction, store store.Store, log logr.Logger) error {
	ctx := context.Background()
	deleteAll := []string{}

	roles := []models.Role{}
	rolesToDelete := []models.Role{}

	bindings := []models.RoleBinding{}
	bindingsToDelete := []models.RoleBinding{}

	for _, obj := range objectTransactions {

		log.Info("processing object tx", "tx", obj.ClusterName())

		// Handle delete all tx first as does not hold objects
		if obj.TransactionType() == models.TransactionTypeDeleteAll {
			deleteAll = append(deleteAll, obj.ClusterName())
			continue
		}

		kind := obj.Object().GetObjectKind().GroupVersionKind().Kind

		if kind == "ClusterRole" || kind == "Role" {
			role, err := adapters.NewRoleAdapter(obj.ClusterName(), kind, obj.Object().Raw())
			if err != nil {
				return fmt.Errorf("cannot create role: %w", err)
			}

			if obj.TransactionType() == models.TransactionTypeDelete {
				rolesToDelete = append(rolesToDelete, role.ToModel())
				continue
			}

			// Explorer should support aggregated clusteroles.
			// Related issue: https://github.com/weaveworks/weave-gitops-enterprise/issues/3443
			if len(role.GetRules()) == 0 {
				// Certain roles have no policy rules for some reason.
				// Possibly related to the rbac.authorization.k8s.io/aggregate-to-gitops-reader label?
				continue
			}

			roles = append(roles, role.ToModel())
		}

		if kind == "ClusterRoleBinding" || kind == "RoleBinding" {
			binding, err := adapters.NewBindingAdapter(obj.ClusterName(), obj.Object().Raw())
			if err != nil {
				return fmt.Errorf("cannot create binding: %w", err)
			}

			if obj.TransactionType() == models.TransactionTypeDelete {
				bindingsToDelete = append(bindingsToDelete, binding.ToModel())
				continue
			}

			bindings = append(bindings, binding.ToModel())
		}
	}

	if len(roles) > 0 {
		if err := store.StoreRoles(ctx, roles); err != nil {
			return fmt.Errorf("cannot store roles: %w", err)
		}
	}

	if len(rolesToDelete) > 0 {
		if err := store.DeleteRoles(ctx, rolesToDelete); err != nil {
			return fmt.Errorf("cannot delete roles: %w", err)
		}
	}

	if len(bindings) > 0 {
		if err := store.StoreRoleBindings(ctx, bindings); err != nil {
			return fmt.Errorf("cannot store role bindings: %w", err)
		}
	}

	if len(bindingsToDelete) > 0 {
		if err := store.DeleteRoleBindings(ctx, bindingsToDelete); err != nil {
			return fmt.Errorf("cannot delete role bindings: %w", err)
		}
	}

	if len(deleteAll) > 0 {
		if err := store.DeleteAllRoles(ctx, deleteAll); err != nil {
			return fmt.Errorf("failed to delete all roles: %w", err)
		}
		if err := store.DeleteAllRoleBindings(ctx, deleteAll); err != nil {
			return fmt.Errorf("failed to delete all role bindings: %w", err)
		}
	}

	log.Info("roles processed", "roles-upsert", roles, "roles-delete", rolesToDelete, "rolebindings-upsert", bindings, "rolebindings-delete", bindingsToDelete, "deleteAll", deleteAll)
	return nil
}
