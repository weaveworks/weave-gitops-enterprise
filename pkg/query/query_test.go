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
		opts    store.QueryOption
		want    []string
	}{
		{
			name:  "get all objects",
			query: []store.QueryClause{},
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

			query: []store.QueryClause{&clause{
				key:     "cluster",
				value:   "test-cluster",
				operand: string(store.OperandEqual),
			}},

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
			name:  "pagination - no offset",
			opts:  &query{limit: 1, offset: 0},
			query: []store.QueryClause{},
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
			opts: &query{
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
		{
			name: "composite query",
			objects: []models.Object{
				{
					Cluster:    "test-cluster-1",
					Name:       "obj-a",
					Namespace:  "namespace-b",
					Kind:       "Kind1",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
				{
					Cluster:    "test-cluster-2",
					Name:       "obj-b",
					Namespace:  "namespace-b",
					Kind:       "Kind1",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
				{
					Cluster:    "test-cluster-3",
					Name:       "obj-c",
					Namespace:  "namespace-b",
					Kind:       "Kind2",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
				{
					Cluster:    "test-cluster-3",
					Name:       "obj-d",
					Namespace:  "namespace-c",
					Kind:       "Kind1",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
			},
			query: []store.QueryClause{
				&clause{
					key:     "kind",
					value:   "Kind1",
					operand: string(store.OperandEqual),
				},
				&clause{
					key:     "namespace",
					value:   "namespace-b",
					operand: string(store.OperandEqual),
				},
			},
			want: []string{"obj-a", "obj-b"},
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

			s, err := store.NewSQLiteStore(db, logr.Discard())
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

			got, err := q.RunQuery(ctx, tt.query, tt.opts)
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
	clauses []clause
	offset  int32
	limit   int32
}

func (q *query) GetQuery() []store.QueryClause {
	clauses := []store.QueryClause{}

	for _, c := range q.clauses {
		clauses = append(clauses, &c)
	}

	return clauses
}

func (q *query) GetOffset() int32 {
	return q.offset
}

func (q *query) GetLimit() int32 {
	return q.limit
}

type clause struct {
	key     string
	operand string
	value   string
}

func (c *clause) GetKey() string {
	return c.key
}

func (c *clause) GetOperand() string {
	return c.operand
}

func (c *clause) GetValue() string {
	return c.value
}
