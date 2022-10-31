package fetcher

import (
	"context"
	"fmt"
	"strings"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/go-logr/logr"
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	mngr "github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	v1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
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
	cluster      types.NamespacedName
}

func NewMultiClusterFetcher(log logr.Logger, config *rest.Config, cg kube.ClientGetter, namespace, mgmtCluster string) (mngr.ClusterFetcher, error) {
	return multiClusterFetcher{
		log:          log.WithName("multi-cluster-fetcher"),
		cfg:          config,
		clientGetter: cg,
		namespace:    namespace,
		cluster:      types.NamespacedName{Name: mgmtCluster},
	}, nil
}

// ToClusterName takes a types.NamespacedName and returns the name of the cluster
// ManagementCluster doesn't have a namespace
func ToClusterName(cluster types.NamespacedName) string {
	if cluster.Namespace == "" {
		return cluster.Name
	}

	return cluster.String()
}

// Take a nice type.NamespacedName and return a string that the MC-fetcher understands
// e.g.
// - {Name: "foo", Namespace: "bar"} -> "bar/foo"
// - {Name: "foo"} -> "foo"
// (ManagementCluster doesn't have a namespace)
func FromClusterName(clusterName string) types.NamespacedName {
	parts := strings.Split(clusterName, "/")
	if len(parts) == 1 {
		return types.NamespacedName{
			Name: parts[0],
		}
	}

	return types.NamespacedName{
		Namespace: parts[0],
		Name:      parts[1],
	}
}

// IsManagementCluster returns true if the cluster is the management cluster
// Provide the name of the management cluster and a ref to another cluster
func IsManagementCluster(mgmtClusterName string, cluster types.NamespacedName) bool {
	return cluster.Namespace == "" && mgmtClusterName == cluster.Name
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
		Name:        f.cluster.Name,
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
		if !isReady(cluster) || !hasConnectivity(cluster) {
			continue
		}

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
				Name: types.NamespacedName{
					Name:      cluster.Name,
					Namespace: cluster.Namespace,
				}.String(),
				Server:      restCfg.Host,
				BearerToken: restCfg.BearerToken,
				TLSConfig:   restCfg.TLSClientConfig,
			})
	}

	return clusters, nil
}

func isReady(cluster gitopsv1alpha1.GitopsCluster) bool {
	return apimeta.IsStatusConditionTrue(cluster.GetConditions(), meta.ReadyCondition)
}

func hasConnectivity(cluster gitopsv1alpha1.GitopsCluster) bool {
	return apimeta.IsStatusConditionTrue(cluster.GetConditions(), gitopsv1alpha1.ClusterConnectivity)
}
