package mgmtfetcher

import (
	"fmt"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	typedauth "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
)

type UserConfigAuth struct {
	mngcluster clustersmngr.Cluster
}

func NewUserConfigAuth(cfg *rest.Config, managementCluster types.NamespacedName) *UserConfigAuth {
	mngcluster := clustersmngr.Cluster{
		Name:        managementCluster.Name,
		Server:      cfg.Host,
		BearerToken: cfg.BearerToken,
		TLSConfig:   cfg.TLSClientConfig,
	}

	return &UserConfigAuth{
		mngcluster: mngcluster,
	}
}

func (u *UserConfigAuth) Get(user *auth.UserPrincipal) (typedauth.AuthorizationV1Interface, error) {
	cfg, err := clustersmngr.ClientConfigWithUser(user)(u.mngcluster)
	if err != nil {
		return nil, fmt.Errorf("failed to create management cluster client config: %w", err)
	}

	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error making authorization clientset: %w", err)
	}

	return cs.AuthorizationV1(), nil

}
