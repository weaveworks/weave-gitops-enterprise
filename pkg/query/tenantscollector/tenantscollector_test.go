package tenantscollector

import (
	"context"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/collectorfakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/utils/testutils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTenantsCollector(t *testing.T) {

	tests := []struct {
		name            string
		namespaces      []*v1.Namespace
		transactionType models.TransactionType
		expected        []models.Tenant
		seed            []models.Tenant
	}{
		{
			name:            "works with an empty list of transactions",
			namespaces:      nil,
			transactionType: models.TransactionTypeUpsert,
			expected:        []models.Tenant{},
		},
		{
			name: "adds a tenant",
			namespaces: []*v1.Namespace{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
						Labels: map[string]string{
							TenantLabel: "my-tenant",
						},
					},
				},
			},
			transactionType: models.TransactionTypeUpsert,
			expected: []models.Tenant{
				{
					Name:        "my-tenant",
					ClusterName: "my-cluster",
					Namespace:   "test",
				},
			},
		},
		{
			name: "ignores namespaces without a tenant label",
			namespaces: []*v1.Namespace{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
				},
			},
			transactionType: models.TransactionTypeUpsert,
			expected:        []models.Tenant{},
		},
		{
			name: "removes a tenant",
			namespaces: []*v1.Namespace{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
						Labels: map[string]string{
							TenantLabel: "my-tenant",
						},
					},
				},
			},
			transactionType: models.TransactionTypeDelete,
			expected:        []models.Tenant{},
			seed: []models.Tenant{
				{
					Name:        "my-tenant",
					ClusterName: "my-cluster",
					Namespace:   "test",
				},
			},
		},
		{
			name: "removes all tenants",
			namespaces: []*v1.Namespace{{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						TenantLabel: "my-tenant",
					},
				},
			}},
			transactionType: models.TransactionTypeDeleteAll,
			expected:        []models.Tenant{},
			seed: []models.Tenant{{
				Name:        "my-tenant",
				ClusterName: "my-cluster",
				Namespace:   "test",
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			col := &collectorfakes.FakeCollector{}
			s, err := createTestStore()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if (test.seed) != nil {
				if err := s.StoreTenants(context.Background(), test.seed); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			tc := &tenantsCollector{
				log:       logr.Discard(),
				store:     s,
				Collector: col,
			}

			txs := []models.ObjectTransaction{}
			for _, ns := range test.namespaces {
				txs = append(txs, testutils.NewObjectTransaction("my-cluster", ns, test.transactionType))
			}

			if err := tc.update(context.Background(), txs); err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			actual, err := s.GetTenants(context.Background())
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			diff := cmp.Diff(test.expected, actual,
				cmpopts.IgnoreUnexported(models.Tenant{}),
				cmpopts.IgnoreFields(models.Tenant{}, "ID", "CreatedAt", "UpdatedAt", "DeletedAt"),
			)

			if diff != "" {
				t.Errorf("unexpected diff: %s", diff)
			}
		})
	}

}

func createTestStore() (store.Store, error) {
	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		return nil, err
	}

	return store.NewStore(store.StorageBackendSQLite, dir, logr.Discard())
}
