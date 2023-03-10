package store

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . StoreWriter
type Store interface {
	StoreWriter
	StoreReader
}

// StoreWriter is an interface for storing access rules and objects
type StoreWriter interface {
	StoreAccessRules(roles []models.AccessRule) error
	StoreObjects(objects []models.Object) error
	Add(ctx context.Context, document Document) (int64, error)
	Delete(ctx context.Context, document Document) error
}

//counterfeiter:generate . StoreReader

// StoreReader is an interface for querying objects
type StoreReader interface {
	GetObjects() ([]models.Object, error)
	GetAccessRules() ([]models.AccessRule, error)
	Count(ctx context.Context, kind string) (int64, error)
}

// factory method that by default creates a in memory store
func NewStore(location string, log logr.Logger) (Store, error) {
	return newInMemoryStore(location, log)
}
