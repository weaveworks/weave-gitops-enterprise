package fakes

import (
	"context"
	"github.com/enekofb/collector/pkg/cluster/store"
	"github.com/go-logr/logr"
)

type FakeStore struct {
	log logr.Logger
}

func NewStore(log logr.Logger) FakeStore {
	return FakeStore{log: log}
}

func (f FakeStore) Delete(ctx context.Context, document store.Document) error {
	f.log.Info("faked delete")
	return nil
}

func (f FakeStore) Count(ctx context.Context, kind string) (int64, error) {
	f.log.Info("faked count")
	return 0, nil
}

func (f FakeStore) GetAll(ctx context.Context) ([]store.Document, error) {
	f.log.Info("faked store: get all")
	return []store.Document{}, nil
}

func (f FakeStore) Add(ctx context.Context, document store.Document) (int64, error) {
	f.log.Info("faked store: add")
	return 0, nil
}
