package namespaces

import (
	"context"
	"fmt"
	"time"

	"github.com/cheshir/ttlcache"
	v1 "k8s.io/api/core/v1"
	typedauth "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

const (
	cacheTTL            = time.Minute
	cacheResolutionTime = 30 * time.Second
)

type UsersResourcesNamespaces struct {
	cache *ttlcache.Cache
}

func NewUsersResourcesNamespaces() *UsersResourcesNamespaces {
	return &UsersResourcesNamespaces{
		cache: ttlcache.New(cacheResolutionTime),
	}
}

func (n *UsersResourcesNamespaces) Get(userID, kind string) ([]string, bool) {
	if val, found := n.cache.Get(n.cacheKey(userID, kind)); found {
		return val.([]string), true
	}

	return nil, false
}

func (n *UsersResourcesNamespaces) Set(userID, kind string, namespaces []string) {
	n.cache.Set(n.cacheKey(userID, kind), namespaces, cacheTTL)
}

func (n *UsersResourcesNamespaces) cacheKey(userID, kind string) uint64 {
	return ttlcache.StringKey(fmt.Sprintf("%s:%s", userID, kind))
}

func (n *UsersResourcesNamespaces) Build(ctx context.Context, userID string, client typedauth.AuthorizationV1Interface, namespaces []*v1.Namespace) error {
	resourceNamespaces, err := buildCache(ctx, client, namespaces)
	if err != nil {
		return err
	}

	for kind, nsList := range resourceNamespaces {
		n.Set(userID, kind, nsList)
	}
	return nil
}
