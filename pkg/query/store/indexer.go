package store

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	bleve "github.com/blevesearch/bleve/v2"
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
	Add(ctx context.Context, objects []models.Object) error
	Remove(ctx context.Context, objects []models.Object) error
}

//counterfeiter:generate . IndexReader
type IndexReader interface {
	Search(ctx context.Context, query Query, opts QueryOption) (Iterator, error)
}

var indexFile = "index.db"

func NewIndexer(s Store, path string) (Indexer, error) {
	idxFileLocation := filepath.Join(path, indexFile)
	mapping := bleve.NewIndexMapping()

	index, err := bleve.New(idxFileLocation, mapping)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}

	return &bleveIndexer{
		idx:   index,
		store: s,
	}, nil
}

type bleveIndexer struct {
	idx   bleve.Index
	store Store
}

func (i *bleveIndexer) Add(ctx context.Context, objects []models.Object) error {
	batch := i.idx.NewBatch()

	for _, obj := range objects {
		if err := batch.Index(obj.GetID(), obj); err != nil {
			return fmt.Errorf("failed to index object: %w", err)
		}
	}

	return i.idx.Batch(batch)
}

func (i *bleveIndexer) Remove(ctx context.Context, objects []models.Object) error {
	for _, obj := range objects {
		if err := i.idx.Delete(obj.GetID()); err != nil {
			return fmt.Errorf("failed to delete object: %w", err)
		}
	}

	return nil
}

func (i *bleveIndexer) Search(ctx context.Context, q Query, opts QueryOption) (Iterator, error) {
	if q == "" {
		return i.store.GetObjects(ctx, nil, opts)
	}

	query := bleve.NewQueryStringQuery(string(q))

	req := bleve.NewSearchRequest(query)

	if opts != nil {
		// `-` reverses the order
		req.SortBy([]string{fmt.Sprintf("-%v", opts.GetOrderBy())})
	}

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
