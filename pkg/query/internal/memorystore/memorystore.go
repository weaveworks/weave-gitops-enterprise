package memorystore

import (
	"fmt"
	"sync"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
)

type MemoryStore struct {
	m           *sync.Mutex
	accessRules []models.AccessRule
	objects     map[string]models.Object
}

func NewInMemoryStore() *MemoryStore {
	return &MemoryStore{
		m:           &sync.Mutex{},
		accessRules: []models.AccessRule{},
		objects:     map[string]models.Object{},
	}
}

func (s *MemoryStore) StoreAccessRules(rules []models.AccessRule) error {
	s.m.Lock()
	defer s.m.Unlock()

	s.accessRules = append(s.accessRules, rules...)

	return nil
}

func (s *MemoryStore) StoreObjects(objects []models.Object) error {
	s.m.Lock()
	defer s.m.Unlock()

	for _, o := range objects {
		s.objects[toKey(o)] = o
	}
	return nil
}

func (s *MemoryStore) GetObjects() ([]models.Object, error) {
	s.m.Lock()
	defer s.m.Unlock()

	list := []models.Object{}

	for _, v := range s.objects {
		list = append(list, v)
	}

	return list, nil
}

func (s *MemoryStore) GetAccessRules() ([]models.AccessRule, error) {
	s.m.Lock()
	defer s.m.Unlock()

	return s.accessRules, nil
}

func toKey(o models.Object) string {
	return fmt.Sprintf("%s/%s/%s/%s", o.Cluster, o.Namespace, o.Kind, o.Name)
}
