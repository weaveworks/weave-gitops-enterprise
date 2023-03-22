package namespaces

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	informersv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type NamespacesInformerCache struct {
	namespacesInformer informersv1.NamespaceInformer
}

// NewNamespaceCache registers a namespace cache in the given shared informer
func NewNamespacesInformerCache(factory informers.SharedInformerFactory) (*NamespacesInformerCache, error) {
	namespacesInformer := factory.Core().V1().Namespaces()
	_, err := namespacesInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})
	if err != nil {
		return nil, fmt.Errorf("failed to add event handler for namespaces: %w", err)
	}

	return &NamespacesInformerCache{
		namespacesInformer: namespacesInformer,
	}, nil
}

// List lists all namespaces in cache
func (n *NamespacesInformerCache) List() ([]*v1.Namespace, error) {
	namespaces, err := n.namespacesInformer.Lister().List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces from cache: %w", err)
	}
	return namespaces, nil
}

// CacheSync informs whether the namespace cache was synced or not
func (n *NamespacesInformerCache) CacheSync() cache.InformerSynced {
	return n.namespacesInformer.Informer().HasSynced
}
