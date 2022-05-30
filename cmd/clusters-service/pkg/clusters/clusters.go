package clusters

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClusterGetter implementations get clusters by name.
type ClusterGetter interface {
	Get(ctx context.Context, name string) (*gitopsv1alpha1.GitopsCluster, error)
}

// ClusterLister implementations list clusters from a Library.
type ClusterLister interface {
	List(ctx context.Context, listOptions client.ListOptions) (map[string]*gitopsv1alpha1.GitopsCluster, string, error)
}

// Library represents a library of Clusters indexed by name.
type Library interface {
	ClusterGetter
	ClusterLister
}

type CRDLibrary struct {
	Log          logr.Logger
	ClientGetter kube.ClientGetter
	Namespace    string
}

func (lib *CRDLibrary) Get(ctx context.Context, name string) (*gitopsv1alpha1.GitopsCluster, error) {
	lib.Log.Info("Getting client from context")
	cl, err := lib.ClientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	cluster := gitopsv1alpha1.GitopsCluster{}
	lib.Log.Info("Getting cluster", "cluster", name)
	err = cl.Get(ctx, client.ObjectKey{
		Namespace: lib.Namespace,
		Name:      name,
	}, &cluster)
	if err != nil {
		lib.Log.Error(err, "Failed to get cluster", "cluster", name)
		return nil, fmt.Errorf("error getting cluster %s/%s: %s", lib.Namespace, name, err)
	}
	lib.Log.Info("Got cluster", "cluster", name)

	return &cluster, nil
}

func (lib *CRDLibrary) List(ctx context.Context, listOptions client.ListOptions) (map[string]*gitopsv1alpha1.GitopsCluster, string, error) {
	lib.Log.Info("Getting client from context")
	cl, err := lib.ClientGetter.Client(ctx)
	if err != nil {
		return nil, "", err
	}

	lib.Log.Info("Querying namespace for Cluster resources", "namespace", lib.Namespace)

	clusterList := gitopsv1alpha1.GitopsClusterList{}
	err = cl.List(ctx, &clusterList, client.InNamespace(lib.Namespace), &listOptions)
	if err != nil {
		return nil, "", fmt.Errorf("error getting clusters: %s", err)
	}

	lib.Log.Info("Got clusters", "numberOfClusters", len(clusterList.Items))

	nextPageToken := clusterList.GetContinue()
	result := map[string]*gitopsv1alpha1.GitopsCluster{}
	for i, ct := range clusterList.Items {
		result[ct.ObjectMeta.Name] = &clusterList.Items[i]
	}
	return result, nextPageToken, nil
}
