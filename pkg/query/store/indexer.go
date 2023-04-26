package store

import (
	"context"
	"fmt"
	"sync"

	bleve "github.com/blevesearch/bleve/v2"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
)

type indexerWriter struct {
	Store
	mapping bleve.Index
	mt      sync.Mutex
}

func NewIndexer(s Store) (Store, error) {
	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("index.db", mapping)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}

	return &indexerWriter{
		Store:   s,
		mapping: index,
	}, nil
}

func (i *indexerWriter) add(ctx context.Context, objects []models.Object) error {
	for _, obj := range objects {
		if err := i.mapping.Index(obj.GetID(), obj); err != nil {
			return fmt.Errorf("failed to index object: %w", err)
		}
	}

	return nil
}

func (i *indexerWriter) GetObjects(ctx context.Context, q Query, opts QueryOption) (Iterator, error) {
	if len(q) == 0 {
		return i.Store.GetObjects(ctx, q, opts)
	}

	val := q[0].GetValue()

	query := bleve.NewMatchQuery(val)
	search := bleve.NewSearchRequest(query)
	searchResults, err := i.mapping.Search(search)
	if err != nil {
		return nil, fmt.Errorf("failed to search for objects: %w", err)
	}

	iter := &indexerIterator{
		result: searchResults,
		s:      i.Store,
		mu:     sync.Mutex{},
		index:  -1,
	}

	return iter, nil
}

func (w *indexerWriter) StoreObjects(ctx context.Context, objects []models.Object) error {
	if err := w.add(ctx, objects); err != nil {
		return fmt.Errorf("failed to add objects to indexer: %w", err)
	}

	return w.Store.StoreObjects(ctx, objects)
}

type indexerIterator struct {
	result *bleve.SearchResult
	mu     sync.Mutex
	index  int
	s      Store
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

	result := []models.Object{}
	for _, hit := range i.result.Hits {
		obj, err := i.s.GetObjectByID(context.Background(), hit.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get object by ID: %w", err)
		}

		result = append(result, obj)
	}

	return result, nil
}
