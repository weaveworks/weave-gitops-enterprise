package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/metrics"
	indexermetrics "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/metrics"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"

	bleve "github.com/blevesearch/bleve/v2"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
)

func TestIndexer_Metrics(t *testing.T) {
	g := NewGomegaWithT(t)

	indexermetrics.IndexerLatencyHistogram.Reset()
	indexermetrics.IndexerInflightRequests.Reset()

	metrics.NewPrometheusServer(metrics.Options{
		ServerAddress: "localhost:8080",
	})

	metricsUrl := "http://localhost:8080/metrics"

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

	err = idx.Add(context.Background(), objects)
	g.Expect(err).NotTo(HaveOccurred())

	err = s.StoreObjects(context.Background(), objects)
	g.Expect(err).NotTo(HaveOccurred())

	t.Run("should have Search instrumented", func(t *testing.T) {

		it, err := idx.Search(context.Background(), query{}, nil)
		g.Expect(err).NotTo(HaveOccurred())

		wantMetrics := []string{
			`indexer_inflight_requests{action="Search"} 0`,
			`indexer_latency_seconds_bucket{action="Search",status="success",le="0.01"} 1`,
		}
		assertMetrics(g, metricsUrl, wantMetrics)
		t.Cleanup(func() {
			err := it.Close()
			if err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Run("should have ListFacets instrumented", func(t *testing.T) {
		_, err := idx.ListFacets(context.Background())
		g.Expect(err).NotTo(HaveOccurred())

		wantMetrics := []string{
			`indexer_inflight_requests{action="ListFacets"} 0`,
			`indexer_latency_seconds_bucket{action="ListFacets",status="success",le="0.01"} 1`,
		}
		assertMetrics(g, metricsUrl, wantMetrics)
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

type query struct{}

func (q query) GetTerms() string {
	return ""
}

func (q query) GetFilters() []string {
	return []string{}
}
