package fetcher

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	mngr "github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	dataKey     = "value"
	yamlDataKey = "value.yaml"
)

type multiClusterFetcher struct {
	log          logr.Logger
	cfg          *rest.Config
	clientGetter kube.ClientGetter
	namespace    string
}

func NewMultiClusterFetcher(log logr.Logger, config *rest.Config, cg kube.ClientGetter, namespace string) (mngr.ClusterFetcher, error) {
	return multiClusterFetcher{
		log:          log.WithName("multi-cluster-fetcher"),
		cfg:          config,
		clientGetter: cg,
		namespace:    namespace,
	}, nil
}

func (f multiClusterFetcher) Fetch(ctx context.Context) ([]mngr.Cluster, error) {
	clusters := []mngr.Cluster{f.self()}

	res, err := f.leafClusters(ctx)
	if err != nil {
		f.log.Error(err, "unable to collect GitOps Clusters")
		return clusters, nil
	}

	allClusters := append(clusters, res...)
	clusterNames := []string{}
	for _, c := range allClusters {
		clusterNames = append(clusterNames, c.Name)
	}
	f.log.Info("Found clusters", "clusters", clusterNames)

	return allClusters, nil
}

func (f *multiClusterFetcher) self() mngr.Cluster {
	return mngr.Cluster{
		Name:        mngr.DefaultCluster,
		Server:      f.cfg.Host,
		BearerToken: f.cfg.BearerToken,
		TLSConfig:   f.cfg.TLSClientConfig,
	}
}

func (f multiClusterFetcher) leafClusters(ctx context.Context) ([]mngr.Cluster, error) {
	clusters := []mngr.Cluster{}

	cl, err := f.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	goClusters := &gitopsv1alpha1.GitopsClusterList{}

	if err := cl.List(ctx, goClusters, client.InNamespace(f.namespace)); err != nil {
		return nil, err
	}

	for _, cluster := range goClusters.Items {
		var secretRef string

		if cluster.Spec.CAPIClusterRef != nil {
			secretRef = fmt.Sprintf("%s-kubeconfig", cluster.Spec.CAPIClusterRef.Name)
		}

		if secretRef == "" && cluster.Spec.SecretRef != nil {
			secretRef = cluster.Spec.SecretRef.Name
		}

		if secretRef == "" {
			continue
		}

		key := types.NamespacedName{
			Name:      secretRef,
			Namespace: cluster.Namespace,
		}

		var secret v1.Secret
		if err := cl.Get(ctx, key, &secret); err != nil {
			f.log.Error(err, "unable to fetch secret for GitOps Cluster", "cluster", cluster.Name)

			continue
		}

		var data []byte

		for k := range secret.Data {
			if k == dataKey || k == yamlDataKey {
				data = secret.Data[k]

				break
			}
		}

		if len(data) == 0 {
			continue
		}

		restCfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(data))
		if err != nil {
			f.log.Error(err, "unable to create kubconfig from GitOps Cluster secret data", "cluster", cluster.Name)

			continue
		}

		clusters = append(clusters,
			mngr.Cluster{
				Name:        cluster.Name,
				Server:      restCfg.Host,
				BearerToken: restCfg.BearerToken,
				TLSConfig:   restCfg.TLSClientConfig,
			})
	}

	return clusters, nil
}
