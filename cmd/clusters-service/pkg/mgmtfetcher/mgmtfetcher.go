package mgmtfetcher

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/namespaces"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/api/core/v1"
	typedauth "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NamespacesCache interface {
	List() ([]*v1.Namespace, error)
}

type UserAuthClientGetter interface {
	Get(user *auth.UserPrincipal) (typedauth.AuthorizationV1Interface, error)
}

type ManagementCrossNamespacesFetcher struct {
	UsersResourcesNamespaces *namespaces.UsersResourcesNamespaces
	usersCacheLock           sync.Map
	clientGetter             kube.ClientGetter
	namespacesCache          NamespacesCache
	authClientGetter         UserAuthClientGetter
}

type returnListFactory func() client.ObjectList

type NamespacedList struct {
	Namespace string
	List      client.ObjectList
	Error     error
}

// @TODO use p.Hash when merged in core
func hash(p *auth.UserPrincipal) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%s/%s/%v", p.ID, p.Token(), p.Groups)))
	return hex.EncodeToString(hash[:])
}

// NewManagementCrossNamespacesFetcher returns a fetcher that lists resources across namespeaces on management cluster
// based on the current user's permissions
func NewManagementCrossNamespacesFetcher(namespacesCache NamespacesCache, ClientGetter kube.ClientGetter, authClientGetter UserAuthClientGetter) *ManagementCrossNamespacesFetcher {
	return &ManagementCrossNamespacesFetcher{
		UsersResourcesNamespaces: namespaces.NewUsersResourcesNamespaces(),
		clientGetter:             ClientGetter,
		namespacesCache:          namespacesCache,
		authClientGetter:         authClientGetter,
	}
}

func (m *ManagementCrossNamespacesFetcher) lockUserCache(user *auth.UserPrincipal) *sync.Mutex {
	actual, _ := m.usersCacheLock.LoadOrStore(hash(user), &sync.Mutex{})
	lock := actual.(*sync.Mutex)
	lock.Lock()
	return lock
}

func (m *ManagementCrossNamespacesFetcher) getUserNamespaces(ctx context.Context, user *auth.UserPrincipal, resourceKind string) ([]string, error) {
	userLock := m.lockUserCache(user)
	defer userLock.Unlock()
	namespaces, err := m.namespacesCache.List()
	if err != nil {
		return nil, err
	}
	var userNamespaces []string
	userNamespaces, found := m.UsersResourcesNamespaces.Get(hash(user), resourceKind)

	fmt.Printf("found %v, userNamespaces: %v", found, userNamespaces)

	authClientSet, err := m.authClientGetter.Get(user)
	if err != nil {
		return nil, err
	}

	if !found {
		err := m.UsersResourcesNamespaces.Build(ctx, hash(user), authClientSet, namespaces)
		if err != nil {
			return nil, err
		}
		userNamespaces, found = m.UsersResourcesNamespaces.Get(hash(user), resourceKind)
		if !found {
			return nil, fmt.Errorf("unsupported resource kind: %s", resourceKind)
		}
		fmt.Printf("userNamespaces: %v, userHash: %v", userNamespaces, hash(user))
	}

	return userNamespaces, nil
}

// Fetch list the specified resource across all the namespaces that the user has access to this resource on
func (m *ManagementCrossNamespacesFetcher) Fetch(ctx context.Context, resourceKind string, fn returnListFactory) ([]NamespacedList, error) {
	// @TODO handle pagination across multiple namespaces
	user := auth.Principal(ctx)
	c, err := m.clientGetter.Client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	userNamespaces, err := m.getUserNamespaces(ctx, user, resourceKind)
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}
	var namespacedList []NamespacedList
	var listAggrLock sync.Mutex

	for _, namespace := range userNamespaces {
		wg.Add(1)

		go func(nsName string) {
			defer wg.Done()
			opts := []client.ListOption{client.InNamespace(nsName)}

			res := fn()

			err := c.List(ctx, res, opts...)
			listAggrLock.Lock()
			defer listAggrLock.Unlock()
			namespacedList = append(namespacedList, NamespacedList{
				Namespace: nsName,
				List:      res,
				Error:     err,
			})
		}(namespace)
	}
	wg.Wait()

	return namespacedList, nil
}
