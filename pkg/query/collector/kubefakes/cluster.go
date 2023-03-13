package kubefakes

import (
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type fakeCluster struct {
	namme  types.NamespacedName
	cfg    *rest.Config
	log    logr.Logger
	client client.Client
}

func (f fakeCluster) GetName() string {
	f.log.Info("faked")
	return f.namme.Name
}

func (f fakeCluster) GetHost() string {
	f.log.Info("faked")
	return f.namme.Name
}

func (f fakeCluster) GetServerClient() (client.Client, error) {
	f.log.Info("faked")
	return f.client, nil
}

func (f fakeCluster) GetUserClient(principal *auth.UserPrincipal) (client.Client, error) {
	f.log.Info("faked")
	return f.client, nil
}

func (f fakeCluster) GetServerClientset() (kubernetes.Interface, error) {
	f.log.Info("faked")
	return nil, nil
}

func (f fakeCluster) GetUserClientset(principal *auth.UserPrincipal) (kubernetes.Interface, error) {
	f.log.Info("faked")
	return nil, nil
}

func (f fakeCluster) GetServerConfig() (*rest.Config, error) {
	f.log.Info("faked server config")
	return f.cfg, nil
}

func NewCluster(name types.NamespacedName, cfg *rest.Config, runtimeClient client.Client, log logr.Logger) (cluster.Cluster, error) {
	return fakeCluster{
		namme:  name,
		cfg:    cfg,
		log:    log,
		client: runtimeClient,
	}, nil
}
