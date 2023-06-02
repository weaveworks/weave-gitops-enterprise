package query

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

// In here, we don't care about checking access. The following lets us
// go straight through authorization.

type predicateAuthz struct {
	predicate func(models.Object) (bool, error)
}

var allowAll Authorizer = predicateAuthz{
	predicate: func(models.Object) (bool, error) {
		return true, nil
	},
}

func (c predicateAuthz) ObjectAuthorizer([]models.Role, []models.RoleBinding, *auth.UserPrincipal, string) func(models.Object) (bool, error) {
	return c.predicate
}

func TestRunQuery(t *testing.T) {
	tests := []struct {
		name    string
		objects []models.Object
		query   *query
		opts    store.QueryOption
		want    []string
	}{
		{
			name:  "get all objects",
			query: &query{terms: ""},
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
			query: &query{filters: []string{"+cluster:my-cluster"}},

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
			opts:  &query{limit: 1, offset: 0, orderBy: "name", ascending: false},
			query: &query{},
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
			name:  "pagination - with offset",
			query: &query{},
			opts: &query{
				limit:     1,
				offset:    1,
				orderBy:   "name",
				ascending: false,
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
		// {
		// 	name: "composite query",
		// 	objects: []models.Object{
		// 		{
		// 			Cluster:    "test-cluster-1",
		// 			Name:       "foo",
		// 			Namespace:  "alpha",
		// 			Kind:       "Kind1",
		// 			APIGroup:   "example.com",
		// 			APIVersion: "v1",
		// 		},
		// 		{
		// 			Cluster:    "test-cluster-2",
		// 			Name:       "bar",
		// 			Namespace:  "bravo",
		// 			Kind:       "Kind1",
		// 			APIGroup:   "example.com",
		// 			APIVersion: "v1",
		// 		},
		// 		{
		// 			Cluster:    "test-cluster-3",
		// 			Name:       "baz",
		// 			Namespace:  "bravo",
		// 			Kind:       "Kind2",
		// 			APIGroup:   "example.com",
		// 			APIVersion: "v1",
		// 		},
		// 		{
		// 			Cluster:    "test-cluster-3",
		// 			Name:       "bang",
		// 			Namespace:  "delta",
		// 			Kind:       "Kind1",
		// 			APIGroup:   "example.com",
		// 			APIVersion: "v1",
		// 		},
		// 	},
		// 	query: &query{
		// 		terms:   "",
		// 		filters: []string{"kind:Kind1", "namespace:bravo"},
		// 	},
		// 	opts: &query{
		// 		orderBy:   "name",
		// 		ascending: false,
		// 	},
		// 	want: []string{"bar", "baz"},
		// },
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
				{
					Cluster:    "test-cluster-3",
					Name:       "foo",
					Namespace:  "namespace-b",
					Kind:       "Deployment",
					APIGroup:   "apps",
					APIVersion: "v1",
				},
			},
			query: &query{terms: "podinfo"},
			want:  []string{"podinfo", "podinfo"},
		},
		// {
		// 	name: "by namespace",
		// 	objects: []models.Object{
		// 		{
		// 			Cluster:    "management",
		// 			Name:       "my-app",
		// 			Namespace:  "namespace-a",
		// 			Kind:       "Deployment",
		// 			APIGroup:   "apps",
		// 			APIVersion: "v1",
		// 		},
		// 		{
		// 			Cluster:    "management",
		// 			Name:       "other-thing",
		// 			Namespace:  "namespace-b",
		// 			Kind:       "Deployment",
		// 			APIGroup:   "apps",
		// 			APIVersion: "v1",
		// 		},
		// 	},
		// 	query: &query{filters: []string{"+namespace:namespace-a"}},
		// 	want:  []string{"my-app"},
		// },
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
			query: &query{},
			opts: &query{
				orderBy:   "kind",
				ascending: true,
			},
			want: []string{"podinfo-b", "podinfo-a"},
		},
		{
			name: "by kind",
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
			query: &query{filters: []string{"+kind:kustomization"}},
			opts:  &query{orderBy: "name", ascending: true},
			want:  []string{"podinfo-a", "podinfo-b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

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
				log:        logr.Discard(),
				debug:      logr.Discard(),
				r:          s,
				index:      idx,
				authorizer: allowAll,
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

			g.Expect(names).To(Equal(tt.want), fmt.Sprintf("terms: %s, filters: %s", tt.query.terms, tt.query.filters))
		})
	}

}

func TestQueryIteration(t *testing.T) {
	g := NewGomegaWithT(t)

	dir, err := os.MkdirTemp("", "test")
	g.Expect(err).NotTo(HaveOccurred())

	db, err := store.CreateSQLiteDB(dir)
	g.Expect(err).NotTo(HaveOccurred())

	s, err := store.NewSQLiteStore(db, logr.Discard())
	g.Expect(err).NotTo(HaveOccurred())

	idx, err := store.NewIndexer(s, dir)
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
	g.Expect(idx.Add(context.Background(), objects)).To(Succeed())

	// Verify that the "raw" data has the four items
	r, err := db.Model(&models.Object{}).Rows()
	g.Expect(err).NotTo(HaveOccurred())

	var count int

	for r.Next() {
		count += 1
	}

	r.Close()

	g.Expect(count).To(Equal(4))

	dropNamespaceB := predicateAuthz{
		predicate: func(obj models.Object) (bool, error) {
			return obj.Namespace != "namespace-b", nil
		},
	}

	// Now check that the query does not get the "unauthorized"
	// object, but still gets the desired number.
	q := &qs{
		log:        logr.Discard(),
		debug:      logr.Discard(),
		r:          s,
		index:      idx,
		authorizer: dropNamespaceB,
	}

	qy := &query{
		terms: "",
		limit: 3,
	}

	got, err := q.RunQuery(ctx, qy, qy)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(got).To(HaveLen(3))
	g.Expect(got).To(HaveEach(HaveField("Namespace", "namespace-a")), "all be in namespace-a")
}

type query struct {
	terms     string
	filters   []string
	offset    int32
	limit     int32
	orderBy   string
	ascending bool
}

func (q *query) GetTerms() string {
	return q.terms
}

func (q *query) GetFilters() []string {
	return q.filters
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
