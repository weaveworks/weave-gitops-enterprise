package query

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
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

// TestRunQuery runs a set of test cases for acceptance on the querying logic.
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
			query: &query{filters: []string{"cluster:my-cluster"}},

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
			opts:  &query{limit: 1, offset: 0, orderBy: "name", descending: false},
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
				limit:      1,
				offset:     1,
				orderBy:    "name",
				descending: false,
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
					Namespace:  "bravo",
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
			query: &query{
				terms:   "",
				filters: []string{"kind:Kind1", "namespace:bravo"},
			},
			opts: &query{
				orderBy:    "name",
				descending: false,
			},
			want: []string{"bar"},
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
		{
			name: "by namespace",
			objects: []models.Object{
				{
					Cluster:    "management",
					Name:       "my-app",
					Namespace:  "namespace-a",
					Kind:       "Deployment",
					APIGroup:   "apps",
					APIVersion: "v1",
				},
				{
					Cluster:    "management",
					Name:       "other-thing",
					Namespace:  "namespace-b",
					Kind:       "Deployment",
					APIGroup:   "apps",
					APIVersion: "v1",
				},
			},
			query: &query{filters: []string{"namespace:namespace-a"}},
			want:  []string{"my-app"},
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
			query: &query{},
			opts: &query{
				orderBy:    "kind",
				descending: true,
			},
			want: []string{"podinfo-b", "podinfo-a"},
		},
		{
			name: "order by name asc",
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
				orderBy:    "name",
				descending: false,
			},
			want: []string{"podinfo-a", "podinfo-b"},
		},
		{
			name: "order by name desc",
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
				orderBy:    "name",
				descending: true,
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
			query: &query{filters: []string{"kind:Kustomization"}},
			opts:  &query{orderBy: "name", descending: false},
			want:  []string{"podinfo-a", "podinfo-b"},
		},
		{
			name: "complex composite query",
			objects: []models.Object{
				{
					Cluster:    "management",
					Name:       "podinfo-a",
					Namespace:  "namespace-a",
					Kind:       "HelmChart",
					APIGroup:   "apps",
					APIVersion: "v1",
				},
				{
					Cluster:    "management",
					Name:       "podinfo-b",
					Namespace:  "namespace-a",
					Kind:       "HelmRepository",
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
				{
					Cluster:    "management",
					Name:       "podinfo-d",
					Namespace:  "namespace-b",
					Kind:       "HelmRelease",
					APIGroup:   "apps",
					APIVersion: "v1",
				},
			},
			query: &query{filters: []string{"kind:/(HelmChart|HelmRepository)/", "namespace:namespace-a"}},
			opts:  &query{orderBy: "name"},
			want:  []string{"podinfo-a", "podinfo-b"},
		},
		{
			name: "uniqe hits only",
			objects: []models.Object{
				{
					Cluster:    "management",
					Name:       "some-name",
					Namespace:  "namespace-1",
					Kind:       "HelmChart",
					APIGroup:   "apps",
					APIVersion: "v1",
					Unstructured: toUnstructured(&sourcev1beta2.HelmChart{
						ObjectMeta: v1.ObjectMeta{
							Name:      "some-name",
							Namespace: "namespace-1",
						},
						Spec: sourcev1beta2.HelmChartSpec{
							Chart: "some-name",
							SourceRef: sourcev1beta2.LocalHelmChartSourceReference{
								Kind: "HelmRepository",
								Name: "some-name",
							},
						},
					}),
				},
			},
			query: &query{terms: "", filters: []string{}},
			opts:  &query{orderBy: "name", descending: true},
			want:  []string{"some-name"},
		},
		{
			name: "exact name matches",
			objects: []models.Object{
				{
					Cluster:    "management",
					Name:       "kube-prometheus-stack",
					Namespace:  "namespace-1",
					Kind:       "HelmChart",
					APIGroup:   "apps",
					APIVersion: "v1",
					Category:   configuration.CategoryAutomation,
				},
			},
			query: &query{terms: "kube-prometheus-stack", filters: []string{}},
			opts:  &query{orderBy: "name", descending: true, filters: []string{"category:automation"}},
			want:  []string{"kube-prometheus-stack"},
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

			idx, err := store.NewIndexer(s, idxDir, logr.Discard())
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

// TestRunQuery_ErrorScenarios injects errors to ensure that querying is tolerant to errors where possible.
func TestRunQuery_ErrorScenarios(t *testing.T) {
	t.Run("should be tolerant to inconsistency between indexer and datastore", func(t *testing.T) {
		objects := []models.Object{
			{
				Cluster:    "test-cluster-1",
				Name:       "podinfo1",
				Namespace:  "namespace-a",
				Kind:       "Deployment",
				APIGroup:   "apps",
				APIVersion: "v1",
			},
			{
				Cluster:    "test-cluster-2",
				Name:       "podinfo2",
				Namespace:  "namespace-b",
				Kind:       "Deployment",
				APIGroup:   "apps",
				APIVersion: "v1",
			},
		}
		want := []string{"podinfo1"}
		g := NewGomegaWithT(t)

		dir, err := os.MkdirTemp("", "test")
		g.Expect(err).NotTo(HaveOccurred())

		db, err := store.CreateSQLiteDB(dir)
		g.Expect(err).NotTo(HaveOccurred())

		s, err := store.NewSQLiteStore(db, logr.Discard())
		g.Expect(err).NotTo(HaveOccurred())

		idxDir, err := os.MkdirTemp("", "indexer-test")
		g.Expect(err).NotTo(HaveOccurred())

		idx, err := store.NewIndexer(s, idxDir, logr.Discard())
		g.Expect(err).NotTo(HaveOccurred())

		q := &qs{
			log:        logr.Discard(),
			debug:      logr.Discard(),
			r:          s,
			index:      idx,
			authorizer: allowAll,
		}

		g.Expect(idx.Add(context.Background(), objects)).To(Succeed())

		//force inconsistency by just adding one element in the datastore
		g.Expect(store.SeedObjects(db, objects[:1])).To(Succeed())

		ctx := auth.WithPrincipal(context.Background(), &auth.UserPrincipal{
			ID: "test",
			Groups: []string{
				"group-a",
			},
		})

		got, err := q.RunQuery(ctx, &query{terms: ""}, nil)
		g.Expect(err).NotTo(HaveOccurred())

		names := []string{}

		for _, o := range got {
			names = append(names, o.Name)
		}

		g.Expect(names).To(Equal(want))

	})
}
func TestQueryIteration(t *testing.T) {
	g := NewGomegaWithT(t)

	dir, err := os.MkdirTemp("", "test")
	g.Expect(err).NotTo(HaveOccurred())

	db, err := store.CreateSQLiteDB(dir)
	g.Expect(err).NotTo(HaveOccurred())

	s, err := store.NewSQLiteStore(db, logr.Discard())
	g.Expect(err).NotTo(HaveOccurred())

	idx, err := store.NewIndexer(s, dir, logr.Discard())
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

func TestQueryOrdering_Realistic(t *testing.T) {
	g := NewGomegaWithT(t)

	dir, err := os.MkdirTemp("", "test")
	g.Expect(err).NotTo(HaveOccurred())

	db, err := store.CreateSQLiteDB(dir)
	g.Expect(err).NotTo(HaveOccurred())

	s, err := store.NewSQLiteStore(db, logr.Discard())
	g.Expect(err).NotTo(HaveOccurred())

	idx, err := store.NewIndexer(s, dir, logr.Discard())
	g.Expect(err).NotTo(HaveOccurred())

	ctx := auth.WithPrincipal(context.Background(), &auth.UserPrincipal{
		ID: "test",
	})

	ex, err := os.Getwd()
	g.Expect(err).NotTo(HaveOccurred())

	data, err := os.ReadFile(ex + "/../../test/utils/data/explorer/sort_fixture.json")
	g.Expect(err).NotTo(HaveOccurred())

	objects := []models.Object{}

	if err := json.Unmarshal(data, &objects); err != nil {
		t.Fatal(err)
	}

	g.Expect(store.SeedObjects(db, objects)).To(Succeed())
	g.Expect(idx.Add(context.Background(), objects)).To(Succeed())

	q := &qs{
		log:        logr.Discard(),
		debug:      logr.Discard(),
		r:          s,
		index:      idx,
		authorizer: allowAll,
	}

	t.Run("query by name", func(t *testing.T) {

		qy := &query{
			orderBy: "name",
		}

		got, err := q.RunQuery(ctx, qy, qy)
		g.Expect(err).NotTo(HaveOccurred())

		expected := []string{
			"flux-dashboards",
			"flux-system",
			"flux-system",
			"kube-prometheus-stack",
			"kube-prometheus-stack",
			"monitoring-config",
			"podinfo",
			"podinfo",
			"podinfo",
			"podinfo",
		}

		actual := []string{}
		for _, o := range got {
			actual = append(actual, o.Name)
		}

		diff := cmp.Diff(expected, actual)

		if diff != "" {
			t.Fatalf("unexpected result (-want +got):\n%s", diff)
		}
	})

	t.Run("query by score if order selected", func(t *testing.T) {

		qy := &query{
			terms: "flux-system",
		}

		got, err := q.RunQuery(ctx, qy, qy)
		g.Expect(err).NotTo(HaveOccurred())

		expected := []string{
			"flux-system",
			"flux-system",
			"monitoring-config",
			"kube-prometheus-stack",
		}

		actual := []string{}
		for _, o := range got {
			actual = append(actual, o.Name)
		}

		diff := cmp.Diff(expected, actual)

		if diff != "" {
			t.Fatalf("unexpected result (-want +got):\n%s", diff)
		}
	})

}

type query struct {
	terms      string
	filters    []string
	offset     int32
	limit      int32
	orderBy    string
	descending bool
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

func (q *query) GetDescending() bool {
	return q.descending
}

func toUnstructured(obj client.Object) json.RawMessage {
	data, _ := json.Marshal(obj)
	return data
}
