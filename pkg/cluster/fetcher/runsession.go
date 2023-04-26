package fetcher

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/loft-sh/vcluster/pkg/util/kubeconfig"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	mngrcluster "github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
)

type runSessionFetcher struct {
	log               logr.Logger
	cluster           cluster.Cluster
	scheme            *runtime.Scheme
	isDelegating      bool
	kubeConfigOptions []mngrcluster.KubeConfigOption
	userPrefixes      kube.UserPrefixes
}

func NewRunSessionFetcher(log logr.Logger, hostCluster cluster.Cluster, scheme *runtime.Scheme, isDelegating bool, userPrefixes kube.UserPrefixes, kubeConfigOptions ...mngrcluster.KubeConfigOption) clustersmngr.ClusterFetcher {
	return runSessionFetcher{
		log:               log.WithName("run-session-fetcher"),
		cluster:           hostCluster,
		scheme:            scheme,
		isDelegating:      isDelegating,
		kubeConfigOptions: kubeConfigOptions,
		userPrefixes:      userPrefixes,
	}
}

func (f runSessionFetcher) Fetch(ctx context.Context) ([]cluster.Cluster, error) {
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
	f.log.Info("Found Gitops Run sessions", "clusters", clusterNames)

	return allClusters, nil
}

func (f *runSessionFetcher) runSessions(ctx context.Context) ([]cluster.Cluster, error) {
	clusters := []cluster.Cluster{}

	cl, err := f.cluster.GetServerClient()
	if err != nil {
		return nil, err
	}

	statefulSets := &appsv1.StatefulSetList{}

	if err := cl.List(ctx, statefulSets, client.MatchingLabels(map[string]string{"app": "vcluster", "app.kubernetes.io/part-of": "gitops-run"})); err != nil {
		return nil, err
	}

	for _, ss := range statefulSets.Items {
		var secret corev1.Secret
		err := cl.Get(ctx, client.ObjectKey{Name: "vc-" + ss.Name, Namespace: ss.Namespace}, &secret)
		if err != nil {
			f.log.Error(err, "Couldn't query for gitops run secret", "cluster", ss.Name)
			continue
		}

		configBytes, found := secret.Data[kubeconfig.KubeconfigSecretKey]
		if !found {
			f.log.Error(err, "Couldn't find config in gitops run secret", "cluster", ss.Name)
			continue
		}
		config, err := clientcmd.Load(configBytes)
		if err != nil {
			f.log.Error(err, "Couldn't load gitops run secret", "cluster", ss.Name)
			continue
		}

		clientConfig := clientcmd.NewDefaultClientConfig(*config, nil)

		restConfig, err := clientConfig.ClientConfig()
		if err != nil {
			f.log.Error(err, "Failed to create clientconfig for run session", "cluster", ss.Name)
			continue
		}

		restConfig.Host = fmt.Sprintf("https://%s.%s:443", ss.Name, ss.Namespace)

		cluster, err := cluster.NewSingleCluster(
			types.NamespacedName{
				Name:      ss.Name,
				Namespace: ss.Namespace,
			}.String(),
			restConfig, f.scheme, f.userPrefixes, f.kubeConfigOptions...)
		if err != nil {
			f.log.Error(err, "Failed to connect to run session", "cluster", ss.Name)
			continue
		}

		if f.isDelegating {
			cluster = mngrcluster.NewDelegatingCacheCluster(cluster, restConfig, f.scheme)
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
