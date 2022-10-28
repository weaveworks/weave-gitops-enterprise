package fetcher

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	vclustercmd "github.com/loft-sh/vcluster/cmd/vclusterctl/cmd"
	"github.com/loft-sh/vcluster/cmd/vclusterctl/log"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type vClusterFetcher struct {
	log     logr.Logger
	cluster cluster.Cluster
	scheme  *runtime.Scheme
}

func NewVClusterFetcher(log logr.Logger, hostCluster cluster.Cluster, scheme *runtime.Scheme) (clustersmngr.ClusterFetcher, error) {
	return vClusterFetcher{
		log:     log.WithName("vcluster-fetcher"),
		cluster: hostCluster,
		scheme:  scheme,
	}, nil
}

func (f vClusterFetcher) Fetch(ctx context.Context) ([]cluster.Cluster, error) {
	clusters := []cluster.Cluster{}
	res, err := f.runSessions(ctx)
	if err != nil {
		f.log.Error(err, "unable to collect GitOps Run sessions")
		return clusters, nil
	}

	allClusters := append(clusters, res...)
	clusterNames := []string{}
	for _, c := range allClusters {
		clusterNames = append(clusterNames, c.GetName())
	}
	f.log.Info("Found vcluster sessions", "clusters", clusterNames)

	return allClusters, nil
}

func (f *vClusterFetcher) runSessions(ctx context.Context) ([]cluster.Cluster, error) {
	clusters := []cluster.Cluster{}

	cl, err := f.cluster.GetServerClient()
	if err != nil {
		return nil, err
	}

	clientSet, err := f.cluster.GetServerClientset()
	if err != nil {
		return nil, err
	}

	statefulSets := &appsv1.StatefulSetList{}

	if err := cl.List(ctx, statefulSets, client.MatchingLabels(map[string]string{"app": "vcluster"})); err != nil {
		return nil, err
	}

	for _, ss := range statefulSets.Items {
		kubeConfig, err := vclustercmd.GetKubeConfig(ctx, clientSet.(*kubernetes.Clientset), ss.Name, ss.Namespace, log.GetInstance())
		f.log.Info("Kube config is ", "kubeconfig", kubeConfig)
		if err != nil {
			f.log.Error(err, "Failed to create kubeconfig from statefulset", "cluster", ss.Name)
			continue
		}

		clientConfig := clientcmd.NewDefaultClientConfig(*kubeConfig, nil)

		restConfig, err := clientConfig.ClientConfig()
		if err != nil {
			f.log.Error(err, "Failed to create clientconfig from statefulset", "cluster", ss.Name)
			continue
		}
		// TODO: this should use a configurable cluster domain
		restConfig.Host = fmt.Sprintf("https://%s.%s:443", ss.Name, ss.Namespace)

		cluster, err := cluster.NewSingleCluster(ss.Name, restConfig, f.scheme)
		if err != nil {
			f.log.Error(err, "Failed to connect to cluster from statefulset", "cluster", ss.Name)
			continue
		}

		clusters = append(clusters, &noAuthCluster{cluster})
	}

	return clusters, nil
}

type noAuthCluster struct {
	cluster.Cluster
}

func (c *noAuthCluster) GetUserClient(user *auth.UserPrincipal) (client.Client, error) {
	return c.Cluster.GetServerClient()
}

func (c *noAuthCluster) GetUserClientset(user *auth.UserPrincipal) (kubernetes.Interface, error) {
	return c.Cluster.GetServerClientset()
}
