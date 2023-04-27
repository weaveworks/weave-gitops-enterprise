package mgmtfetcher

import (
	"fmt"

	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	typedauth "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
)

type UserConfigAuth struct {
	mngcluster cluster.Cluster
}

func NewUserConfigAuth(cfg *rest.Config, mgmtCluster string, userPrefixes kube.UserPrefixes) (*UserConfigAuth, error) {
	mngcluster, err := cluster.NewSingleCluster(
		mgmtCluster,
		cfg,
		nil,
		userPrefixes,
	)

	if err != nil {
		return nil, err
	}

	return &UserConfigAuth{
		mngcluster: mngcluster,
	}, nil
}

func (u *UserConfigAuth) Get(user *auth.UserPrincipal) (typedauth.AuthorizationV1Interface, error) {
	cs, err := u.mngcluster.GetUserClientset(user)
	if err != nil {
		return nil, fmt.Errorf("error making authorization clientset: %w", err)
	}

	return cs.AuthorizationV1(), nil

}
