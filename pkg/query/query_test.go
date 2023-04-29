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
		{
			name: "`or` clause query",
			objects: []models.Object{
				{
					Cluster:    "test-cluster-1",
					Name:       "podinfo",
					Namespace:  "namespace-a",
					Kind:       "Deployment",
					APIGroup:   "apps",
					APIVersion: "v1",
				},
				{
					Cluster:    "test-cluster-2",
					Name:       "podinfo",
					Namespace:  "namespace-b",
					Kind:       "Deployment",
					APIGroup:   "apps",
					APIVersion: "v1",
				},
			},
			query: []store.QueryClause{
				&clause{
					key:     "name",
					value:   "podinfo",
					operand: string(store.OperandEqual),
				},
			},
			opts: &query{
				globalOperand: string(store.GlobalOperandOr),
			},
			want: []string{"podinfo", "podinfo"},
		},
		{
			name: "`or` clause with clusters",
			objects: []models.Object{
				{
					Cluster:    "management",
					Name:       "podinfo",
					Namespace:  "namespace-a",
					Kind:       "Deployment",
					APIGroup:   "apps",
					APIVersion: "v1",
				},
				{
					Cluster:    "management",
					Name:       "podinfo",
					Namespace:  "namespace-b",
					Kind:       "Deployment",
					APIGroup:   "apps",
					APIVersion: "v1",
				},
			},
			query: []store.QueryClause{
				&clause{
					key:     "name",
					value:   "management",
					operand: string(store.OperandEqual),
				},
				&clause{
					key:     "namespace",
					value:   "management",
					operand: string(store.OperandEqual),
				},
				&clause{
					key:     "cluster",
					value:   "management",
					operand: string(store.OperandEqual),
				},
			},
			opts: &query{
				globalOperand: string(store.GlobalOperandOr),
			},
			want: []string{"podinfo", "podinfo"},
		},
		{
			name: "`or` clause with order by",
			objects: []models.Object{
				{
					Cluster:    "management",
					Name:       "podinfo-a",
					Namespace:  "namespace-a",
					Kind:       "A",
					APIGroup:   "apps",
					APIVersion: "v1",
				},
				{
					Cluster:    "management",
					Name:       "podinfo-b",
					Namespace:  "namespace-b",
					Kind:       "B",
					APIGroup:   "apps",
					APIVersion: "v1",
				},
			},
			query: []store.QueryClause{
				&clause{
					key:     "cluster",
					value:   "management",
					operand: string(store.OperandEqual),
				},
			},
			opts: &query{
				orderBy:       "kind desc",
				globalOperand: string(store.GlobalOperandOr),
			},
			want: []string{"podinfo-b", "podinfo-a"},
		},
		{
			name: "scoped query",
			objects: []models.Object{
				{
					Cluster:    "management",
					Name:       "podinfo-a",
					Namespace:  "namespace-a",
					Kind:       "Kustomization",
					APIGroup:   "apps",
					APIVersion: "v1",
				},
				{
					Cluster:    "management",
					Name:       "podinfo-b",
					Namespace:  "namespace-a",
					Kind:       "Kustomization",
					APIGroup:   "apps",
					APIVersion: "v1",
				},
				{
					Cluster:    "management",
					Name:       "podinfo-c",
					Namespace:  "namespace-a",
					Kind:       "HelmRelease",
					APIGroup:   "apps",
					APIVersion: "v1",
				},
			},
			query: []store.QueryClause{
				&clause{
					key:     "cluster",
					value:   "management",
					operand: string(store.OperandEqual),
				},
			},
			opts: &query{
				orderBy:       "kind desc",
				globalOperand: string(store.GlobalOperandAnd),
				scopes:        []string{"Kustomization"},
			},
			want: []string{"podinfo-a", "podinfo-b"},
		},
		{
			name: "scoped query with `or`",
			objects: []models.Object{
				{
					Cluster:    "cluster-a",
					Name:       "podinfo",
					Namespace:  "namespace-a",
					Kind:       "Kustomization",
					APIGroup:   "apps",
					APIVersion: "v1",
				},
				{
					Cluster:    "cluster-a",
					Name:       "podinfo",
					Namespace:  "namespace-b",
					Kind:       "Kustomization",
					APIGroup:   "apps",
					APIVersion: "v1",
				},
				{
					Cluster:    "cluster-b",
					Name:       "podinfo",
					Namespace:  "namespace-c",
					Kind:       "HelmRelease",
					APIGroup:   "apps",
					APIVersion: "v1",
				},
			},
			query: []store.QueryClause{
				&clause{
					key:     "name",
					value:   "podinfo",
					operand: string(store.OperandEqual),
				},
			},
			opts: &query{
				orderBy:       "kind desc",
				globalOperand: string(store.GlobalOperandOr),
				scopes:        []string{"Kustomization"},
			},
			want: []string{"podinfo", "podinfo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			checker := &accesscheckerfakes.FakeChecker{}
			checker.HasAccessReturns(true, nil)

			dir, err := os.MkdirTemp("", "test")
			g.Expect(err).NotTo(HaveOccurred())

			db, err := store.CreateSQLiteDB(dir)
			g.Expect(err).NotTo(HaveOccurred())

			s, err := store.NewSQLiteStore(db, logr.Discard())
			g.Expect(err).NotTo(HaveOccurred())

			q := &qs{
				log:     logr.Discard(),
				debug:   logr.Discard(),
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

func TestQueryIteration(t *testing.T) {
	g := NewGomegaWithT(t)

	checker := &accesscheckerfakes.FakeChecker{}
	checker.HasAccessReturns(true, nil)

	dir, err := os.MkdirTemp("", "test")
	g.Expect(err).NotTo(HaveOccurred())

	db, err := store.CreateSQLiteDB(dir)
	g.Expect(err).NotTo(HaveOccurred())

	s, err := store.NewSQLiteStore(db, logr.Discard())
	g.Expect(err).NotTo(HaveOccurred())

	ctx := auth.WithPrincipal(context.Background(), &auth.UserPrincipal{
		ID: "test",
	})

	objects := []models.Object{
		{
			Cluster:    "test-cluster-1",
			Name:       "obj-1",
			Namespace:  "namespace-a",
			Kind:       "Deployment",
			APIGroup:   "apps",
			APIVersion: "v1",
		},
		{
			Cluster:    "test-cluster-1",
			Name:       "obj-2",
			Namespace:  "namespace-b",
			Kind:       "Deployment",
			APIGroup:   "apps",
			APIVersion: "v1",
		},
		{
			Cluster:    "test-cluster-1",
			Name:       "obj-3",
			Namespace:  "namespace-a",
			Kind:       "Deployment",
			APIGroup:   "apps",
			APIVersion: "v1",
		},
		{
			Cluster:    "test-cluster-1",
			Name:       "obj-4",
			Namespace:  "namespace-a",
			Kind:       "Deployment",
			APIGroup:   "apps",
			APIVersion: "v1",
		},
	}

	g.Expect(store.SeedObjects(db, objects)).To(Succeed())

	q := &qs{
		log:     logr.Discard(),
		debug:   logr.Discard(),
		r:       s,
		checker: checker,
	}

	r, err := db.Model(&models.Object{}).Rows()
	g.Expect(err).NotTo(HaveOccurred())

	var count int

	for r.Next() {
		count += 1
	}

	r.Close()

	g.Expect(count).To(Equal(4))

	checker.HasAccessReturnsOnCall(0, true, nil)
	checker.HasAccessReturnsOnCall(1, false, nil)
	checker.HasAccessReturnsOnCall(2, true, nil)
	checker.HasAccessReturnsOnCall(3, true, nil)

	qy := &query{
		clauses: []clause{
			{
				key:     "cluster",
				value:   "test-cluster-1",
				operand: string(store.OperandEqual),
			},
		},
		limit: 3,
	}

	got, err := q.RunQuery(ctx, []store.QueryClause{&qy.clauses[0]}, qy)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(got).To(HaveLen(3))
}

type query struct {
	clauses       []clause
	offset        int32
	limit         int32
	orderBy       string
	globalOperand string
	scopes        []string
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

func (q *query) GetOrderBy() string {
	return q.orderBy
}

func (q *query) GetGlobalOperand() string {
	return q.globalOperand
}

func (q *query) GetScopedKinds() []string {
	return q.scopes
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
