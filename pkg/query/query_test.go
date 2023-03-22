package query

import (
	"context"
	"os"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/accesschecker/accesscheckerfakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func TestRunQuery(t *testing.T) {
	tests := []struct {
		name    string
		objects []models.Object
		query   store.Query
		want    []string
	}{
		{
			name:  "get all objects",
			query: &query{},
			objects: []models.Object{
				{
					Cluster:    "test-cluster",
					Name:       "someName",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
				{
					Cluster:    "test-cluster",
					Name:       "otherName",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
			},
			want: []string{"someName", "otherName"},
		},
		{
			name: "get objects by cluster",
			query: &query{
				key:     "cluster",
				value:   "test-cluster",
				operand: string(store.OperandEqual),
			},
			objects: []models.Object{
				{
					Cluster:    "test-cluster",
					Name:       "obj-cluster-1",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
				{
					Cluster:    "test-cluster-2",
					Name:       "obj-cluster-2",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
			},
			want: []string{"obj-cluster-1"},
		},
		{
			name: "pagination - no offset",
			query: &query{
				limit:  1,
				offset: 0,
			},
			objects: []models.Object{
				{
					Cluster:    "test-cluster",
					Name:       "obj-cluster-1",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
				{
					Cluster:    "test-cluster-2",
					Name:       "obj-cluster-2",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
			},
			want: []string{"obj-cluster-1"},
		},
		{
			name: "pagination - with offset",
			query: &query{
				limit:  1,
				offset: 1,
			},
			objects: []models.Object{
				{
					Cluster:    "test-cluster",
					Name:       "obj-cluster-1",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
				{
					Cluster:    "test-cluster-2",
					Name:       "obj-cluster-2",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
			},
			want: []string{"obj-cluster-2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			checker := &accesscheckerfakes.FakeChecker{}
			checker.HasAccessReturns(true, nil)

			dir, err := os.MkdirTemp("", "test")
			db, err := store.CreateSQLiteDB(dir)
			g.Expect(err).NotTo(HaveOccurred())

			s, err := store.NewSQLiteStore(db)
			g.Expect(err).NotTo(HaveOccurred())

			q := &qs{
				log:     logr.Discard(),
				r:       s,
				checker: checker,
			}

			g.Expect(store.SeedObjects(db, tt.objects)).To(Succeed())

			ctx := auth.WithPrincipal(context.Background(), &auth.UserPrincipal{
				ID: "test",
				Groups: []string{
					"group-a",
				},
			})

			got, err := q.RunQuery(ctx, tt.query)
			g.Expect(err).NotTo(HaveOccurred())

			names := []string{}

			for _, o := range got {
				names = append(names, o.Name)
			}

			g.Expect(names).To(Equal(tt.want))
		})
	}

}

type query struct {
	key     string
	value   string
	operand string
	offset  int64
	limit   int64
}

func (q *query) GetKey() string {
	return q.key
}

func (q *query) GetOperand() string {
	return q.operand
}

func (q *query) GetValue() string {
	return q.value
}

func (q *query) GetOffset() int64 {
	return q.offset
}

func (q *query) GetLimit() int64 {
	return q.limit
}
