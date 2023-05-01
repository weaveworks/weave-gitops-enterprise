package query

import (
	"context"
	"fmt"
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
			query: "",
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
			name:  "get objects by cluster",
			query: "my-cluster",

			objects: []models.Object{
				{
					Cluster:    "my-cluster",
					Name:       "obj-1",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
				{
					Cluster:    "b",
					Name:       "obj-2",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
			},
			want: []string{"obj-1"},
		},
		{
			name:  "pagination - no offset",
			opts:  &query{limit: 1, offset: 0},
			query: "",
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
					Name:       "foo",
					Namespace:  "alpha",
					Kind:       "Kind1",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
				{
					Cluster:    "test-cluster-2",
					Name:       "bar",
					Namespace:  "bravo",
					Kind:       "Kind1",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
				{
					Cluster:    "test-cluster-3",
					Name:       "baz",
					Namespace:  "charlie",
					Kind:       "Kind2",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
				{
					Cluster:    "test-cluster-3",
					Name:       "bang",
					Namespace:  "delta",
					Kind:       "Kind1",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
			},
			query: "+kind:Kind1 +namespace:bravo",
			opts:  &query{orderBy: "name"},
			want:  []string{"bar"},
		},
		{
			name: "across clusters",
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
			query: "podinfo",
			want:  []string{"podinfo", "podinfo"},
		},
		{
			name: "by namespace",
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
			query: "namespace:namespace-a",
			want:  []string{"podinfo", "podinfo"},
		},
		{
			name: "order by",
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
			query: "management",
			opts: &query{
				orderBy: "kind",
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
			query: "+kind:Kustomization",
			opts:  &query{orderBy: "name"},
			want:  []string{"podinfo-a", "podinfo-b"},
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

			idxDir, err := os.MkdirTemp("", "indexer-test")
			g.Expect(err).NotTo(HaveOccurred())

			idx, err := store.NewIndexer(s, idxDir)
			g.Expect(err).NotTo(HaveOccurred())

			q := &qs{
				log:     logr.Discard(),
				debug:   logr.Discard(),
				r:       s,
				checker: checker,
				index:   idx,
			}

			g.Expect(store.SeedObjects(db, tt.objects)).To(Succeed())

			g.Expect(idx.Add(context.Background(), tt.objects)).To(Succeed())

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

			g.Expect(names).To(Equal(tt.want), fmt.Sprintf("query: %s", tt.query))
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
		q:     "cluster:test-cluster-1",
		limit: 3,
	}

	got, err := q.RunQuery(ctx, qy.GetQuery(), qy)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(got).To(HaveLen(3))
}

type query struct {
	q         string
	offset    int32
	limit     int32
	orderBy   string
	ascending bool
}

func (q *query) GetQuery() store.Query {
	return store.Query(q.q)
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

func (q *query) GetAscending() bool {
	return q.ascending
}
