package fetcher

import (
	"context"
	"fmt"
	"strings"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/go-logr/logr"
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	mngr "github.com/weaveworks/weave-gitops/core/clustersmngr"
	mngrcluster "github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/logger"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	v1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	dataKey     = "value"
	yamlDataKey = "value.yaml"
)

type gitopsClusterFetcher struct {
	log               logr.Logger
	cluster           mngrcluster.Cluster
	scheme            *runtime.Scheme
	namespace         string
	isDelegating      bool
	kubeConfigOptions []mngrcluster.KubeConfigOption
}

func NewGitopsClusterFetcher(log logr.Logger, managementCluster mngrcluster.Cluster, namespace string, scheme *runtime.Scheme, isDelegating bool, kubeConfigOptions ...mngrcluster.KubeConfigOption) mngr.ClusterFetcher {
	return gitopsClusterFetcher{
		log:               log.WithName("gitops-cluster-fetcher"),
		cluster:           managementCluster,
		scheme:            scheme,
		namespace:         namespace,
		isDelegating:      isDelegating,
		kubeConfigOptions: kubeConfigOptions,
	}
}

// ToClusterName takes a nice type.NamespacedName and returns
// a string that the MC-fetcher understands
// e.g.
// - {Name: "foo", Namespace: "bar"} -> "bar/foo"
// - {Name: "foo"} -> "foo"
// (ManagementCluster doesn't have a namespace)
func ToClusterName(cluster types.NamespacedName) string {
	if cluster.Namespace == "" {
		return cluster.Name
	}

	return cluster.String()
}

// FromClusterName takes a string that the MC-fetcher understands
// and returns a nice type.NamespacedName
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

func (f gitopsClusterFetcher) Fetch(ctx context.Context) ([]mngrcluster.Cluster, error) {
	clusters := []mngrcluster.Cluster{}

	res, err := f.leafClusters(ctx)
	if err != nil {
		f.log.Error(err, "unable to collect GitOps Clusters")
		return clusters, nil
	}

	allClusters := append(clusters, res...)
	clusterNames := []string{}
	for _, c := range allClusters {
		clusterNames = append(clusterNames, c.GetName())
	}
	f.log.Info("Found clusters", "clusters", clusterNames)

	return allClusters, nil
}

func (f gitopsClusterFetcher) leafClusters(ctx context.Context) ([]mngrcluster.Cluster, error) {
	clusters := []mngrcluster.Cluster{}

	cl, err := f.cluster.GetServerClient()
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
			f.log.V(logger.LogLevelDebug).Info("Ignoring GitOps Cluster, no secret ref found", "cluster", cluster.Name)
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
			f.log.V(logger.LogLevelDebug).Info("Ignoring GitOps Cluster, no data found", "cluster", cluster.Name)
			continue
		}

		restCfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(data))
		if err != nil {
			f.log.Error(err, "unable to create kubeconfig from GitOps Cluster secret data", "cluster", cluster.Name)

			continue
		}

		leafCluster, err := mngrcluster.NewSingleCluster(
			types.NamespacedName{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
			}.String(),
			restCfg,
			f.scheme,
			kube.UserPrefixes{},
			f.kubeConfigOptions...,
		)
		// TODO: the DefaultKubeConfigOptions will throw an error if the cluster can't be reached
		// This has moved here, so we won't even return unreachable clusters - is that acceptable?
		if err != nil {
			f.log.Error(err, "unable to create cluster object from GitOps Cluster secret data", "cluster", cluster.Name)

			continue
		}

		if f.isDelegating {
			leafCluster = mngrcluster.NewDelegatingCacheCluster(leafCluster, restCfg, f.scheme)
		}

		clusters = append(clusters, leafCluster)
	}

	return clusters, nil
}

func isReady(cluster gitopsv1alpha1.GitopsCluster) bool {
	return apimeta.IsStatusConditionTrue(cluster.GetConditions(), meta.ReadyCondition)
}

func hasConnectivity(cluster gitopsv1alpha1.GitopsCluster) bool {
	return apimeta.IsStatusConditionTrue(cluster.GetConditions(), gitopsv1alpha1.ClusterConnectivity)
}
