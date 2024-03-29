package store

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/metrics"

	bleve "github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Indexer
type Indexer interface {
	IndexReader
	IndexWriter
}

//counterfeiter:generate . IndexWriter
type IndexWriter interface {
	// Add adds the given objects to the index.
	Add(ctx context.Context, objects []models.Object) error
	// Remove removes the given objects from the index.
	Remove(ctx context.Context, objects []models.Object) error
	// RemoveByQuery removes objects from the index that match the given query.
	// This is useful for removing objects for a given cluster, namespace, etc.
	RemoveByQuery(ctx context.Context, q string) error
}

type Facets map[string][]string

//counterfeiter:generate . IndexReader
type IndexReader interface {
	// Search searches the index for objects that match the given query.
	// The filters are applied, followed by the terms of the query,
	// so terms search WITHIN a set of filtered objects.
	Search(ctx context.Context, query Query, opts QueryOption) (Iterator, error)
	// ListFacets returns a map of facets and their values.
	// Facets can be used to build a filtering UI or to see what values are available for a given field.
	ListFacets(ctx context.Context, category configuration.ObjectCategory) (Facets, error)
}

var indexFile = "index.db"
var commonFields = []string{"cluster", "namespace", "kind"}

func NewIndexer(s Store, path string, log logr.Logger) (Indexer, error) {
	idxFileLocation := filepath.Join(path, indexFile)
	mapping := bleve.NewIndexMapping()

	addFieldMappings(mapping, commonFields)

	index, err := bleve.New(idxFileLocation, mapping)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}

	return &bleveIndexer{
		idx:   index,
		store: s,
		log:   log,
	}, nil
}

var facetSuffix = ".facet"

func addFieldMappings(index *mapping.IndexMappingImpl, fields []string) {
	objMapping := bleve.NewDocumentMapping()

	nameFieldMapping := bleve.NewTextFieldMapping()
	nameFieldMapping.Analyzer = "keyword"
	objMapping.AddFieldMappingsAt("name", nameFieldMapping)

	// do not index unstructured as for of the main index to get more realistic score
	unstructuredMapping := bleve.NewDocumentDisabledMapping()
	objMapping.AddSubDocumentMapping("unstructured", unstructuredMapping)

	for _, field := range commonFields {
		// This mapping allows us to do query-string queries on the field.
		// For example, we can do `cluster:foo` to get all objects in the `foo` cluster.
		fieldMapping := bleve.NewTextFieldMapping()
		fieldMapping.Analyzer = "keyword"
		objMapping.AddFieldMappingsAt(field, fieldMapping)

		// This adds the facets so the UI can be built around the correct values,
		// without changing how things are searched.
		facetMapping := bleve.NewTextFieldMapping()
		facetMapping.Name = field + facetSuffix
		// Setting analyzer to keyword gives us the exact value of the field.
		facetMapping.Analyzer = "keyword"
		objMapping.AddFieldMappingsAt(field, facetMapping)
	}

	index.AddDocumentMapping("object", objMapping)
}

type bleveIndexer struct {
	idx   bleve.Index
	store Store
	log   logr.Logger
}

// We want the index to contain our raw JSON objects,
// but adding the JSON to the index ID messes up the facets.
// So we prepend the ID with a prefix, and then store the JSON under that ID.
const unstructuredSuffix = "_unstructured"

func (i *bleveIndexer) Add(ctx context.Context, objects []models.Object) (err error) {
	// metrics
	metrics.IndexerAddInflightRequests(metrics.AddAction, 1)
	defer recordIndexerMetrics(metrics.AddAction, time.Now(), err)

	batch := i.idx.NewBatch()

	for _, obj := range objects {
		if err := batch.Index(obj.GetID(), obj); err != nil {
			i.log.Error(err, "failed to index object", "object", obj.GetID())
			continue
		}

		if obj.Unstructured != nil {
			var data interface{}

			if err := json.Unmarshal(obj.Unstructured, &data); err != nil {
				i.log.Error(err, "failed to unmarshal object", "object", obj.GetID())
				continue
			}

			if err := batch.Index(obj.GetID()+unstructuredSuffix, data); err != nil {
				i.log.Error(err, "failed to index unstructured object", "object", obj.GetID())
				continue
			}
		}
	}

	return i.idx.Batch(batch)
}

func (i *bleveIndexer) Remove(ctx context.Context, objects []models.Object) (err error) {
	// metrics
	metrics.IndexerAddInflightRequests(metrics.RemoveAction, 1)
	defer recordIndexerMetrics(metrics.RemoveAction, time.Now(), err)

	for _, obj := range objects {
		if err := i.idx.Delete(obj.GetID()); err != nil {
			return fmt.Errorf("failed to delete object: %w", err)
		}
	}

	return nil
}

func (i *bleveIndexer) RemoveByQuery(ctx context.Context, q string) (err error) {
	// metrics
	metrics.IndexerAddInflightRequests(metrics.RemoveByQueryAction, 1)
	defer recordIndexerMetrics(metrics.RemoveByQueryAction, time.Now(), err)

	query := bleve.NewQueryStringQuery(q)
	req := bleve.NewSearchRequest(query)

	result, err := i.idx.Search(req)
	if err != nil {
		return fmt.Errorf("failed to search index: %w", err)
	}

	for _, hit := range result.Hits {
		if err := i.idx.Delete(hit.ID); err != nil {
			return fmt.Errorf("failed to delete object: %w", err)
		}
	}

	return nil
}

func (i *bleveIndexer) Search(ctx context.Context, q Query, opts QueryOption) (it Iterator, err error) {
	// metrics
	metrics.IndexerAddInflightRequests(metrics.SearchAction, 1)
	defer recordIndexerMetrics(metrics.SearchAction, time.Now(), err)

	// Match all by default.
	// Conjunction queries will return results that match all the sub-queries.
	query := bleve.NewConjunctionQuery(bleve.NewMatchAllQuery())

	terms := q.GetTerms()

	if terms != "" {
		tq := bleve.NewTermQuery(terms)
		query.AddQuery(tq)
	}

	filters := q.GetFilters()

	if len(filters) > 0 {
		// Prepend a `+` to each filter to make it a required term.
		// This gives us the AND between categories, and OR within categories.
		str := "+"
		str += strings.Join(q.GetFilters(), " +")

		qs := bleve.NewQueryStringQuery(str)

		query.AddQuery(qs)
	}

	req := bleve.NewSearchRequest(query)

	count, err := i.idx.DocCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get document count: %w", err)
	}

	// We return all results, because we are filtering out objects AFTER the search with our RBAC rules.
	// There will be cases where we return a slice of objects based on .GetLimit(), but those
	// objects will be filtered out by the accesschecker.
	// The query iterator will handle limiting the page size.
	req.Size = int(count)

	orders := search.SortOrder{}

	if opts != nil {
		sort := opts.GetOrderBy()
		if sort != "" {
			sf := &search.SortField{
				Field: sort,
				Type:  search.SortFieldAsString,
				// Desc behaves oddly in bleve. Setting .Desc to `true` will reverse the default order (descending).
				Desc: opts.GetDescending(),
			}

			orders = append(orders, sf)
		}

		if opts.GetOffset() > 0 {
			req.From = int(opts.GetOffset())
		}
	}

	// We order by score here so that we can get the most relevant results first.
	orders = append(orders, &search.SortScore{
		Desc: true,
	})

	req.SortByCustom(orders)

	searchResults, err := i.idx.Search(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search for objects: %w", err)
	}

	// Strip the `_unstructured` suffix from the ID so we can get the object from the store.
	for i, hit := range searchResults.Hits {
		if strings.Contains(hit.ID, unstructuredSuffix) {
			hit.ID = strings.Replace(hit.ID, "_unstructured", "", 1)
		}

		searchResults.Hits[i] = hit
	}

	// Ensure hits are unique
	seen := map[string]bool{}
	uniqueHits := []*search.DocumentMatch{}
	for _, hit := range searchResults.Hits {
		if _, ok := seen[hit.ID]; !ok {
			uniqueHits = append(uniqueHits, hit)
			seen[hit.ID] = true
		}
	}

	searchResults.Hits = uniqueHits

	iter := &indexerIterator{
		result: searchResults,
		s:      i.store,
		mu:     sync.Mutex{},
		index:  -1,
		opts:   opts,
	}

	return iter, nil
}

func recordIndexerMetrics(action string, start time.Time, err error) {
	metrics.IndexerAddInflightRequests(action, -1)
	if err != nil {
		metrics.IndexerSetLatency(action, metrics.FailedLabel, time.Since(start))
		return
	}
	metrics.IndexerSetLatency(action, metrics.SuccessLabel, time.Since(start))
}

func (i *bleveIndexer) ListFacets(ctx context.Context, category configuration.ObjectCategory) (fcs Facets, err error) {
	// metrics
	metrics.IndexerAddInflightRequests(metrics.ListFacetsAction, 1)
	defer recordIndexerMetrics(metrics.ListFacetsAction, time.Now(), err)

	query := bleve.NewMatchAllQuery()
	req := bleve.NewSearchRequest(query)

	addDefaultFacets(req, category)

	searchResults, err := i.idx.Search(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search for objects: %w", err)
	}

	facets := map[string][]string{}
	for k, v := range searchResults.Facets {
		facets[k] = []string{}

		for _, t := range v.Terms.Terms() {
			facets[k] = append(facets[k], t.Term)
		}

		// Remove empty facets
		if len(facets[k]) == 0 {
			delete(facets, k)
		}
	}

	return facets, nil
}

// addDefaultFacets adds a set of defaault facets to facets search requests. Default facets are comprised of a set
// of common fields like cluster and set of objectkind specific fields like labels
func addDefaultFacets(req *bleve.SearchRequest, cat configuration.ObjectCategory) {
	// adding facets for common fields
	for _, f := range commonFields {
		req.AddFacet(f, bleve.NewFacetRequest(f+facetSuffix, 100))
	}

	// adding facets for labels
	for _, objectKind := range configuration.SupportedObjectKinds {
		if cat != "" && objectKind.Category != cat {
			continue
		}
		for _, label := range objectKind.Labels {
			labelFacet := fmt.Sprintf("labels.%s", label)
			req.AddFacet(labelFacet, bleve.NewFacetRequest(labelFacet, 100))
		}
	}
}

type indexerIterator struct {
	result *bleve.SearchResult
	mu     sync.Mutex
	index  int
	s      Store
	opts   QueryOption
}

func (i *indexerIterator) Next() bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.index++
	return i.index < len(i.result.Hits)
}

func (i *indexerIterator) Row() (models.Object, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	result := i.result.Hits[i.index]

	id := result.ID

	return i.s.GetObjectByID(context.Background(), id)
}

func (i *indexerIterator) All() ([]models.Object, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	ids := []string{}

	for _, hit := range i.result.Hits {
		ids = append(ids, hit.ID)
	}

	iter, err := i.s.GetObjects(context.Background(), ids, i.opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get objects: %w", err)
	}

	return iter.All()
}

func (i *indexerIterator) Page(count int, offset int) ([]models.Object, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	ids := []string{}

	numHits := i.result.Hits.Len()
	upper := offset + count
	if upper > numHits {
		upper = numHits
	}

	for index := offset; index < upper; index++ {
		ids = append(ids, i.result.Hits[index].ID)
	}

	if len(ids) == 0 {
		return []models.Object{}, nil
	}

	iter, err := i.s.GetObjects(context.Background(), ids, i.opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get objects: %w", err)
	}

	return iter.All()
}

func (i *indexerIterator) Close() error {
	return nil
}
