package query

import (
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/memorystore"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
)

// StoreWriter is an interface for storing access rules and objects
type StoreWriter interface {
	StoreAccessRules(roles []models.AccessRule) error
	StoreObjects(objects []models.Object) error
}

// StoreReader is an interface for querying objects
type StoreReader interface {
	Query(groups []string) ([]models.Object, error)
}

func NewStore() StoreWriter {
	return store.NewInMemoryStore()
}
