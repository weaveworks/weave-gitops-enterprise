package store

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"testing"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/monitoring/metrics"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	indexermetrics "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/metrics"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"

	bleve "github.com/blevesearch/bleve/v2"
	gapiv1 "github.com/weaveworks/templates-controller/apis/gitops/v1alpha2"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
)

func TestIndexer_Metrics(t *testing.T) {
	g := NewGomegaWithT(t)

	indexermetrics.IndexerLatencyHistogram.Reset()
	indexermetrics.IndexerInflightRequests.Reset()

	_, h := metrics.NewDefaultPrometheusHandler()
	ts := httptest.NewServer(h)
	defer ts.Close()

	idxFileLocation := filepath.Join(t.TempDir(), indexFile)
	mapping := bleve.NewIndexMapping()

	index, err := bleve.New(idxFileLocation, mapping)
	g.Expect(err).NotTo(HaveOccurred())

	s, err := NewStore(StorageBackendSQLite, t.TempDir(), logr.Discard())
	g.Expect(err).NotTo(HaveOccurred())

	idx := &bleveIndexer{
		idx:   index,
		store: s,
	}

	objects := []models.Object{
		{
			Cluster:    "management",
			Kind:       "Namespace",
			Name:       "anyName",
			APIGroup:   "anyGroup",
			APIVersion: "anyVersion",
			Category:   "automation",
			Namespace:  "anyNamespace",
		},
		{
			Cluster:    "othercluster",
			Kind:       "Namespace",
			Name:       "otherName",
			APIGroup:   "anyGroup",
			APIVersion: "anyVersion",
			Category:   "automation",
			Namespace:  "anyNamespace",
		},
	}

	t.Run("should have Add instrumented", func(t *testing.T) {

		err = idx.Add(context.Background(), objects)
		g.Expect(err).NotTo(HaveOccurred())

		wantMetrics := []string{
			`indexer_inflight_requests{action="Add"} 0`,
			`indexer_latency_seconds_count{action="Add",status="success"} 1`,
		}
		assertMetrics(g, ts, wantMetrics)
	})

	t.Run("should have Remove instrumented", func(t *testing.T) {

		err = idx.Remove(context.Background(), objects)
		g.Expect(err).NotTo(HaveOccurred())

		wantMetrics := []string{
			`indexer_inflight_requests{action="Remove"} 0`,
			`indexer_latency_seconds_count{action="Remove",status="success"} 1`,
		}
		assertMetrics(g, ts, wantMetrics)
	})

	t.Run("should have RemoveByQuery instrumented", func(t *testing.T) {

		err = idx.RemoveByQuery(context.Background(), "+cluster:management")
		g.Expect(err).NotTo(HaveOccurred())

		wantMetrics := []string{
			`indexer_inflight_requests{action="RemoveByQuery"} 0`,
			`indexer_latency_seconds_count{action="RemoveByQuery",status="success"} 1`,
		}
		assertMetrics(g, ts, wantMetrics)
	})

	t.Run("should have Search instrumented", func(t *testing.T) {

		it, err := idx.Search(context.Background(), query{}, nil)
		g.Expect(err).NotTo(HaveOccurred())

		wantMetrics := []string{
			`indexer_inflight_requests{action="Search"} 0`,
			`indexer_latency_seconds_count{action="Search",status="success"} 1`,
		}
		assertMetrics(g, ts, wantMetrics)
		t.Cleanup(func() {
			err := it.Close()
			if err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Run("should have ListFacets instrumented", func(t *testing.T) {
		_, err := idx.ListFacets(context.Background(), configuration.CategoryAutomation)
		g.Expect(err).NotTo(HaveOccurred())

		wantMetrics := []string{
			`indexer_inflight_requests{action="ListFacets"} 0`,
			`indexer_latency_seconds_count{action="ListFacets",status="success"} 1`,
		}
		assertMetrics(g, ts, wantMetrics)
	})
}

func TestIndexer_RemoveByQuery(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name     string
		query    string
		objects  []models.Object
		expected []string
	}{
		{
			name:  "removes by cluster",
			query: "+cluster:management",
			objects: []models.Object{
				{
					Cluster:    "management",
					Kind:       "Namespace",
					Name:       "anyName",
					APIGroup:   "anyGroup",
					APIVersion: "anyVersion",
					Category:   "automation",
					Namespace:  "anyNamespace",
				},
				{
					Cluster:    "othercluster",
					Kind:       "Namespace",
					Name:       "otherName",
					APIGroup:   "anyGroup",
					APIVersion: "anyVersion",
					Category:   "automation",
					Namespace:  "anyNamespace",
				},
			},
			expected: []string{"otherName"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idxFileLocation := filepath.Join(t.TempDir(), indexFile)
			mapping := bleve.NewIndexMapping()

			index, err := bleve.New(idxFileLocation, mapping)
			g.Expect(err).NotTo(HaveOccurred())

			s, err := NewStore(StorageBackendSQLite, t.TempDir(), logr.Discard())
			g.Expect(err).NotTo(HaveOccurred())

			idx := &bleveIndexer{
				idx:   index,
				store: s,
			}

			err = idx.Add(context.Background(), tt.objects)
			g.Expect(err).NotTo(HaveOccurred())

			err = s.StoreObjects(context.Background(), tt.objects)
			g.Expect(err).NotTo(HaveOccurred())

			iter, err := idx.Search(context.Background(), query{}, nil)
			g.Expect(err).NotTo(HaveOccurred())

			// Ensure things got written to the index
			all, err := iter.All()
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(len(all)).To(Equal(2))

			err = idx.RemoveByQuery(context.Background(), tt.query)
			g.Expect(err).NotTo(HaveOccurred())

			iter, err = idx.Search(context.Background(), query{}, nil)
			g.Expect(err).NotTo(HaveOccurred())

			all, err = iter.All()
			g.Expect(err).NotTo(HaveOccurred())

			names := []string{}
			for _, obj := range all {
				names = append(names, obj.Name)
			}

			g.Expect(names).To(Equal(tt.expected))

		})
	}
}

func TestIndexer_RemoveByQueryWithPagination(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name     string
		query    string
		objects  []models.Object
		expected []string
	}{
		{
			name:  "removes by cluster",
			query: "+cluster:management",
			objects: []models.Object{
				{
					Cluster:    "management",
					Kind:       "Namespace",
					Name:       "name-1",
					APIGroup:   "anyGroup",
					APIVersion: "anyVersion",
					Category:   "automation",
					Namespace:  "anyNamespace",
				},
				{
					Cluster:    "othercluster",
					Kind:       "Namespace",
					Name:       "name-2",
					APIGroup:   "anyGroup",
					APIVersion: "anyVersion",
					Category:   "automation",
					Namespace:  "anyNamespace",
				},
				{
					Cluster:    "management",
					Kind:       "Namespace",
					Name:       "name-3",
					APIGroup:   "anyGroup",
					APIVersion: "anyVersion",
					Category:   "automation",
					Namespace:  "anyNamespace",
				},
				{
					Cluster:    "management",
					Kind:       "Namespace",
					Name:       "name-4",
					APIGroup:   "anyGroup",
					APIVersion: "anyVersion",
					Category:   "automation",
					Namespace:  "anyNamespace",
				},
				{
					Cluster:    "othercluster",
					Kind:       "Namespace",
					Name:       "name-5",
					APIGroup:   "anyGroup",
					APIVersion: "anyVersion",
					Category:   "automation",
					Namespace:  "anyNamespace",
				},
				{
					Cluster:    "management",
					Kind:       "Namespace",
					Name:       "name-6",
					APIGroup:   "anyGroup",
					APIVersion: "anyVersion",
					Category:   "automation",
					Namespace:  "anyNamespace",
				},
				{
					Cluster:    "othercluster",
					Kind:       "Namespace",
					Name:       "name-7",
					APIGroup:   "anyGroup",
					APIVersion: "anyVersion",
					Category:   "automation",
					Namespace:  "anyNamespace",
				},
			},
			expected: []string{"name-2", "name-5", "name-7"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idxFileLocation := filepath.Join(t.TempDir(), indexFile)
			mapping := bleve.NewIndexMapping()

			index, err := bleve.New(idxFileLocation, mapping)
			g.Expect(err).NotTo(HaveOccurred())

			s, err := NewStore(StorageBackendSQLite, t.TempDir(), logr.Discard())
			g.Expect(err).NotTo(HaveOccurred())

			idx := &bleveIndexer{
				idx:   index,
				store: s,
			}

			err = idx.Add(context.Background(), tt.objects)
			g.Expect(err).NotTo(HaveOccurred())

			err = s.StoreObjects(context.Background(), tt.objects)
			g.Expect(err).NotTo(HaveOccurred())

			// Ensure things got written to the index.
			// Iterate through all pages without an initial offset.
			iter, err := idx.Search(context.Background(), query{}, nil)
			g.Expect(err).NotTo(HaveOccurred())

			pageObjects, err := iter.Page(3, 0)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(len(pageObjects)).To(Equal(3))

			pageObjects, err = iter.Page(3, 3)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(len(pageObjects)).To(Equal(3))

			pageObjects, err = iter.Page(3, 6)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(len(pageObjects)).To(Equal(1))

			// Iterate through all pages with an initial offset.
			iter, err = idx.Search(context.Background(), query{}, nil)
			g.Expect(err).NotTo(HaveOccurred())

			pageObjects, err = iter.Page(2, 2)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(len(pageObjects)).To(Equal(2))

			pageObjects, err = iter.Page(2, 4)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(len(pageObjects)).To(Equal(2))

			pageObjects, err = iter.Page(2, 6)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(len(pageObjects)).To(Equal(1))

			// Check that required objects were removed.
			err = idx.RemoveByQuery(context.Background(), tt.query)
			g.Expect(err).NotTo(HaveOccurred())

			iter, err = idx.Search(context.Background(), query{}, nil)
			g.Expect(err).NotTo(HaveOccurred())

			all, err := iter.All()
			g.Expect(err).NotTo(HaveOccurred())

			names := []string{}
			for _, obj := range all {
				names = append(names, obj.Name)
			}

			g.Expect(names).To(Equal(tt.expected))

		})
	}
}

type query struct{}

func (q query) GetTerms() string {
	return ""
}

func (q query) GetFilters() []string {
	return []string{}
}

func TestListFacets(t *testing.T) {
	g := NewGomegaWithT(t)

	kustGvk := kustomizev1.GroupVersion.WithKind(kustomizev1.KustomizationKind)
	templateGvk := gapiv1.GroupVersion.WithKind(gapiv1.Kind)

	tests := []struct {
		name              string
		objects           []models.Object
		expected          Facets
		requestedCategory configuration.ObjectCategory
	}{
		{
			name: "adds default facets",
			objects: []models.Object{
				{
					Cluster:    "management",
					Kind:       kustGvk.Kind,
					Name:       "name-1",
					Category:   configuration.CategoryAutomation,
					Namespace:  "my-namespace",
					APIGroup:   kustGvk.Group,
					APIVersion: kustGvk.Version,
				},
			},
			expected: Facets{
				"cluster":   []string{"management"},
				"namespace": []string{"my-namespace"},
				"kind":      []string{"Kustomization"},
			},
		},
		{
			name: "adds facets from object values",
			objects: []models.Object{
				{
					Cluster:    "management",
					Kind:       kustGvk.Kind,
					Name:       "somename",
					Category:   configuration.CategoryAutomation,
					Namespace:  "ns-1",
					APIGroup:   kustGvk.Group,
					APIVersion: kustGvk.Version,
				},
				{
					Cluster:    "management",
					Kind:       kustGvk.Kind,
					Name:       "othername",
					Category:   configuration.CategoryAutomation,
					Namespace:  "ns-2",
					APIGroup:   kustGvk.Group,
					APIVersion: kustGvk.Version,
				},
			},
			expected: Facets{
				"cluster":   []string{"management"},
				"namespace": []string{"ns-1", "ns-2"},
				"kind":      []string{"Kustomization"},
			},
		},
		{
			name: "adds facets from labels",
			objects: []models.Object{
				{
					Cluster:    "management",
					Kind:       templateGvk.Kind,
					Name:       "somename-two",
					Category:   configuration.CategoryTemplate,
					Namespace:  "ns-1",
					APIGroup:   templateGvk.Group,
					APIVersion: templateGvk.Version,
					Labels: map[string]string{
						"weave.works/template-type": "cluster",
					},
				},
			},
			expected: Facets{
				"cluster":                          []string{"management"},
				"namespace":                        []string{"ns-1"},
				"kind":                             []string{"GitOpsTemplate"},
				"labels.weave.works/template-type": []string{"cluster"},
			},
		},
		{
			name: "does not show facets for irrelevant categories",
			objects: []models.Object{
				{
					Cluster:    "management",
					Kind:       kustGvk.Kind,
					Name:       "somename-one",
					Category:   configuration.CategoryAutomation,
					Namespace:  "ns-1",
					APIGroup:   kustGvk.Group,
					APIVersion: kustGvk.Version,
				},
				{
					Cluster:    "management",
					Kind:       templateGvk.Kind,
					Name:       "somename-two",
					Category:   configuration.CategoryTemplate,
					Namespace:  "ns-1",
					APIGroup:   kustGvk.Group,
					APIVersion: kustGvk.Version,
					Labels: map[string]string{
						"weave.works/template-type": "cluster",
					},
				},
			},
			expected: Facets{
				"cluster":   []string{"management"},
				"namespace": []string{"ns-1"},
				"kind":      []string{"GitOpsTemplate", "Kustomization"},
			},
			requestedCategory: configuration.CategoryAutomation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idxFileLocation := filepath.Join(t.TempDir(), indexFile)

			idx, err := NewIndexer(nil, idxFileLocation, logr.Discard())
			g.Expect(err).NotTo(HaveOccurred())

			err = idx.Add(context.Background(), tt.objects)
			g.Expect(err).NotTo(HaveOccurred())

			defer func() {
				err = idx.Remove(context.Background(), tt.objects)
				g.Expect(err).NotTo(HaveOccurred())
			}()

			facets, err := idx.ListFacets(context.Background(), tt.requestedCategory)

			diff := cmp.Diff(tt.expected, facets)

			if diff != "" {
				t.Fatalf("unexpected facets (-want +got):\n%s", diff)
			}
		})
	}
}
