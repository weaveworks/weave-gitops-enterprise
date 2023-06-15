package store

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/metrics"

	bleve "github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
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
	ListFacets(ctx context.Context) (Facets, error)
}

var indexFile = "index.db"
var filterFields = []string{"cluster", "namespace", "kind"}

func NewIndexer(s Store, path string) (Indexer, error) {
	idxFileLocation := filepath.Join(path, indexFile)
	mapping := bleve.NewIndexMapping()

	addFieldMappings(mapping, filterFields)

	index, err := bleve.New(idxFileLocation, mapping)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}

	return &bleveIndexer{
		idx:   index,
		store: s,
		recorder: metrics.NewRecorder(true, "explorer_indexer"),
	}, nil
}

var facetSuffix = ".facet"

func addFieldMappings(index *mapping.IndexMappingImpl, fields []string) {
	objMapping := bleve.NewDocumentMapping()

	for _, field := range fields {
		// This mapping allows us to do query-string queries on the field.
		// For example, we can do `cluster:foo` to get all objects in the `foo` cluster.
		mapping := bleve.NewTextFieldMapping()
		mapping.Analyzer = "keyword"
		objMapping.AddFieldMappingsAt(field, mapping)

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
	recorder metrics.Recorder
}

func (i *bleveIndexer) Add(ctx context.Context, objects []models.Object) error {
	start := time.Now()
	batch := i.idx.NewBatch()
	const action = "add"
	
	i.recorder.InflightRequests(action, 1)
	
	var err error
	defer func() {
		i.recorder.SetStoreLatency(action, time.Since(start))
		i.recorder.InflightRequests(action, -1)
		if err != nil {
			i.recorder.IncRequestCounter(action, "error")
			return
		}
		i.recorder.IncRequestCounter(action, "success")
	}()

	for _, obj := range objects {
		err = batch.Index(obj.GetID(), obj)
		if err != nil {
			return fmt.Errorf("failed to index object: %w", err)
		}
	}

	// The bleve Index is not updated until the batch is executed, so we need to run Batch() before we report the process is success.
	err = i.idx.Batch(batch)
	if err != nil {
		return fmt.Errorf("failed to batch: %w", err)
	}

	return nil
}

func (i *bleveIndexer) Remove(ctx context.Context, objects []models.Object) error {
	for _, obj := range objects {
		if err := i.idx.Delete(obj.GetID()); err != nil {
			return fmt.Errorf("failed to delete object: %w", err)
		}
	}

	return nil
}

func (i *bleveIndexer) RemoveByQuery(ctx context.Context, q string) error {
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

func (i *bleveIndexer) Search(ctx context.Context, q Query, opts QueryOption) (Iterator, error) {
	// Match all by default.
	// Conjunction queries will return results that match all of the subqueries.
	query := bleve.NewConjunctionQuery(bleve.NewMatchAllQuery())

	terms := q.GetTerms()

	if terms != "" {
		tq := bleve.NewMatchQuery(terms)
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

	sortBy := "name"
	tmpl := "-%v"

	if opts != nil {
		// `-` reverses the order
		if !opts.GetAscending() {
			tmpl = "%v"
		}

		sort := opts.GetOrderBy()
		if sort != "" {
			sortBy = sort
		}

		if opts.GetOffset() > 0 {
			req.From = int(opts.GetOffset())
		}

	}

	req.SortBy([]string{fmt.Sprintf(tmpl, sortBy)})

	searchResults, err := i.idx.Search(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search for objects: %w", err)
	}

	iter := &indexerIterator{
		result: searchResults,
		s:      i.store,
		mu:     sync.Mutex{},
		index:  -1,
		opts:   opts,
	}

	return iter, nil
}

func (i *bleveIndexer) ListFacets(ctx context.Context) (Facets, error) {
	query := bleve.NewMatchAllQuery()

	req := bleve.NewSearchRequest(query)

	for _, f := range filterFields {
		req.AddFacet(f, bleve.NewFacetRequest(f+facetSuffix, 100))
	}

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
	}

	return facets, nil
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

func (i *indexerIterator) Close() error {
	return nil
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
