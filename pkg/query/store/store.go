package store

import (
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/memorystore"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . StoreWriter

// StoreWriter is an interface for storing access rules and objects
type StoreWriter interface {
	StoreReader
	StoreAccessRules(roles []models.AccessRule) error
	StoreObjects(objects []models.Object) error
}

//counterfeiter:generate . StoreReader

// StoreReader is an interface for querying objects
type StoreReader interface {
	Query(groups []string) ([]models.Object, error)
}

func NewStore() StoreWriter {
	return store.NewInMemoryStore()
}
