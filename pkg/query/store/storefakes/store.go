package storefakes

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
)

type FakeStore struct {
	log logr.Logger
}

func (f FakeStore) StoreAccessRules(roles []models.AccessRule) error {
	f.log.Info("faked")
	return nil
}

func (f FakeStore) StoreObjects(ctx context.Context, objects []models.Object) error {
	f.log.Info("faked")
	return nil
}

func (f FakeStore) GetObjects() ([]models.Object, error) {
	f.log.Info("faked")
	return nil, nil
}

func (f FakeStore) GetAccessRules() ([]models.AccessRule, error) {
	f.log.Info("faked ")
	return nil, nil
}

func NewStore(log logr.Logger) FakeStore {
	return FakeStore{log: log}
}

func (f FakeStore) DeleteObject(ctx context.Context, object models.Object) error {
	f.log.Info("faked delete")
	return nil
}

func (f FakeStore) CountObjects(ctx context.Context, kind string) (int64, error) {
	f.log.Info("faked count")
	return 0, nil
}

func (f FakeStore) GetAll(ctx context.Context) ([]models.Object, error) {
	f.log.Info("faked store: get all")
	return []models.Object{}, nil
}

func (f FakeStore) StoreObject(ctx context.Context, object models.Object) (int64, error) {
	f.log.Info("faked store: add")
	return 0, nil
}
