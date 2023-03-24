package rolecollector

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/adapters"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var DefaultVerbsRequiredForAccess = []string{"list"}

// RoleCollector is responsible for collecting access rules from all clusters.
// It is a wrapper around a Collector that converts the received objects to AccessRules.
// It writes the received rules to a StoreWriter.
type RoleCollector struct {
	col       collector.Collector
	log       logr.Logger
	converter runtime.UnstructuredConverter
	w         store.StoreWriter
	verbs     []string
	quit      chan struct{}
}

func (a *RoleCollector) Start(ctx context.Context) error {
	err := a.col.Start()
	if err != nil {
		return fmt.Errorf("could not start access collector: %w", err)
	}
	return nil
}

func (a *RoleCollector) Stop() error {
	a.quit <- struct{}{}
	return a.col.Stop()
}

func NewRoleCollector(w store.Store, opts collector.CollectorOpts) (*RoleCollector, error) {
	opts.ObjectKinds = []schema.GroupVersionKind{
		rbacv1.SchemeGroupVersion.WithKind("ClusterRole"),
		rbacv1.SchemeGroupVersion.WithKind("Role"),
		rbacv1.SchemeGroupVersion.WithKind("ClusterRoleBinding"),
		rbacv1.SchemeGroupVersion.WithKind("RoleBinding"),
	}

	opts.ProcessRecordsFunc = defaultProcessRecords

	col, err := collector.NewCollector(opts, w)

	if err != nil {
		return nil, fmt.Errorf("cannot create collector: %w", err)
	}
	return &RoleCollector{
		col:       col,
		log:       opts.Log,
		converter: runtime.DefaultUnstructuredConverter,
		w:         w,
		verbs:     DefaultVerbsRequiredForAccess,
	}, nil
}

func defaultProcessRecords(ctx context.Context, objectRecords []models.ObjectTransaction, store store.Store, log logr.Logger) error {
	roles := []models.Role{}
	rolesToDelete := []models.Role{}

	bindings := []models.RoleBinding{}
	bindingsToDelete := []models.RoleBinding{}

	for _, obj := range objectRecords {
		kind := obj.Object().GetObjectKind().GroupVersionKind().Kind

		if kind == "ClusterRole" || kind == "Role" {
			role, err := adapters.NewRoleAdapter(obj.ClusterName(), obj.Object())
			if err != nil {
				return fmt.Errorf("cannot create role: %w", err)
			}

			if len(role.GetRules()) == 0 {
				// Certain roles have no policy rules for some reason.
				// Possibly related to the rbac.authorization.k8s.io/aggregate-to-gitops-reader label?
				continue
			}

			if obj.TransactionType() == models.TransactionTypeDelete {
				rolesToDelete = append(rolesToDelete, role.ToModel())
				continue
			}

			roles = append(roles, role.ToModel())
		}

		if kind == "ClusterRoleBinding" || kind == "RoleBinding" {
			binding, err := adapters.NewBindingAdapter(obj.ClusterName(), obj.Object())
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

	return nil
}

func (a *RoleCollector) Watch(cluster cluster.Cluster, objectsChannel chan []models.ObjectTransaction, ctx context.Context, log logr.Logger) error {
	return a.col.Watch(cluster, objectsChannel, ctx, log)
}

func (a *RoleCollector) Status(cluster cluster.Cluster) (string, error) {
	return a.col.Status(cluster)
}
