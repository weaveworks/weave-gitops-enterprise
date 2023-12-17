package tenantscollector

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/clusters"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	TenantLabel = "toolkit.fluxcd.io/tenant"
)

var NamespaceKind = configuration.ObjectKind{
	Gvk: v1.SchemeGroupVersion.WithKind("Namespace"),
	NewClientObjectFunc: func() client.Object {
		return &v1.Namespace{}
	},
	Labels:   []string{TenantLabel},
	Category: "",
	AddToSchemeFunc: func(scheme *runtime.Scheme) error {
		return v1.AddToScheme(scheme)
	},
}

func NewTenantsCollector(s store.Store, mgr clusters.Subscriber, sa collector.ImpersonateServiceAccount, log logr.Logger) (collector.Collector, error) {
	incoming := make(chan []models.ObjectTransaction)

	tc := &tenantsCollector{
		log:   log,
		store: s,
	}

	kinds := []configuration.ObjectKind{
		NamespaceKind,
	}

	newWatcher := func(clusterName string, config *rest.Config) (collector.Starter, error) {
		return collector.NewWatcher(clusterName, config, kinds, incoming, log)
	}

	opts := collector.CollectorOpts{
		Name:           "tenants",
		Log:            log,
		NewWatcherFunc: newWatcher,
		Clusters:       mgr,
		ServiceAccount: sa,
	}

	col, err := collector.NewCollector(opts)
	if err != nil {
		return nil, fmt.Errorf("cannot create tenants collector: %w", err)
	}

	tc.Collector = col

	go func() {
		for r := range incoming {
			if err := tc.update(context.Background(), r); err != nil {
				tc.log.Error(err, "could not update tenants")
			}
		}
	}()

	return tc, nil
}

type tenantsCollector struct {
	collector.Collector

	log   logr.Logger
	store store.Store
}

func (tc *tenantsCollector) update(ctx context.Context, txs []models.ObjectTransaction) error {
	toAdd := []models.Tenant{}
	toDelete := []models.Tenant{}

	for _, tx := range txs {
		tenant := tx.Object().GetLabels()[TenantLabel]

		if tenant == "" {
			// This is a namespace without a tenant label, so we
			// don't care about it.
			continue
		}

		if tx.TransactionType() == models.TransactionTypeDelete {
			toDelete = append(toDelete, models.Tenant{
				ClusterName: tx.ClusterName(),
				Name:        tenant,
				Namespace:   tx.Object().GetName(),
			})
			continue
		}

		if tx.TransactionType() == models.TransactionTypeDeleteAll {
			allTenants, err := tc.store.GetTenants(ctx)
			if err != nil {
				return fmt.Errorf("error getting all tenants: %w", err)
			}

			for _, t := range allTenants {
				if t.ClusterName == tx.ClusterName() && t.Namespace == tx.Object().GetName() {
					toDelete = append(toDelete, t)
				}
			}
			continue

		}

		toAdd = append(toAdd, models.Tenant{
			ClusterName: tx.ClusterName(),
			Name:        tenant,
			Namespace:   tx.Object().GetName(),
		})
	}

	if err := tc.store.StoreTenants(ctx, toAdd); err != nil {
		return fmt.Errorf("error storing tenants: %w", err)
	}

	if len(toDelete) > 0 {
		if err := tc.store.DeleteTenants(ctx, toDelete); err != nil {
			return fmt.Errorf("error deleting tenants: %w", err)
		}
	}

	return nil
}
