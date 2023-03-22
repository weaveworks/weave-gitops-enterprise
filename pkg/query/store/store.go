package store

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/models"
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
	DeleteObjects(ctx context.Context, object []models.Object) error
}

type Query interface {
	GetKey() string
	GetOperand() string
	GetValue() string
	GetLimit() int64
	GetOffset() int64
}

// StoreReader is an interface for querying objects
//
//counterfeiter:generate . StoreReader
type StoreReader interface {
	GetObjects(ctx context.Context, q Query) ([]models.Object, error)
	GetAccessRules(ctx context.Context) ([]models.AccessRule, error)
}

type StorageBackend string

const (
	StorageBackendSQLite StorageBackend = "sqlite"
)

// factory method that by default creates a in memory store
func NewStore(backend StorageBackend, uri string, log logr.Logger) (Store, error) {
	switch backend {
	case StorageBackendSQLite:
		db, err := CreateSQLiteDB(uri)
		if err != nil {
			return nil, fmt.Errorf("error creating sqlite db: %w", err)
		}
		return NewSQLiteStore(db)
	default:
		return nil, fmt.Errorf("unknown storage backend: %s", backend)
	}

}
