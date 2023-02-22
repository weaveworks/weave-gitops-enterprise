package store

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Store
type Store interface {
	StoreWriter
	StoreReader
}

// StoreWriter is an interface for storing access rules and objects
//
//counterfeiter:generate . StoreWriter
type StoreWriter interface {
	StoreAccessRules(ctx context.Context, roles []models.AccessRule) error
	StoreObjects(ctx context.Context, objects []models.Object) error
	DeleteObject(ctx context.Context, object models.Object) error
}

// StoreReader is an interface for querying objects
//
//counterfeiter:generate . StoreReader
type StoreReader interface {
	GetObjects() ([]models.Object, error)
	CountObjects(ctx context.Context, kind string) (int64, error)
	GetAccessRules() ([]models.AccessRule, error)
}

// factory method that by default creates a in memory store
func NewStore(location string, log logr.Logger) (Store, error) {
	return newInMemoryStore(location, log)
}
