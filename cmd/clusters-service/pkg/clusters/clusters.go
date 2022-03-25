package clusters

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClusterGetter implementations get clusters by name.
type ClusterGetter interface {
	Get(ctx context.Context, name string) (*capiv1.Cluster, error)
}

// ClusterLister implementations list clusters from a Library.
type ClusterLister interface {
	List(ctx context.Context) (map[string]*capiv1.Cluster, error)
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

func (lib *CRDLibrary) Get(ctx context.Context, name string) (*capiv1.Cluster, error) {
	lib.Log.Info("Getting client from context")
	cl, err := lib.ClientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	cluster := capiv1.Cluster{}
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

func (lib *CRDLibrary) List(ctx context.Context) (map[string]*capiv1.Cluster, error) {
	lib.Log.Info("Getting client from context")
	cl, err := lib.ClientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	lib.Log.Info("Querying namespace for Cluster resources", "namespace", lib.Namespace)
	clusterList := capiv1.ClusterList{}
	err = cl.List(ctx, &clusterList, client.InNamespace(lib.Namespace))
	if err != nil {
		return nil, fmt.Errorf("error getting clusters: %s", err)
	}
	lib.Log.Info("Got clusters", "numberOfClusters", len(clusterList.Items))

	result := map[string]*capiv1.Cluster{}
	for i, ct := range clusterList.Items {
		result[ct.ObjectMeta.Name] = &clusterList.Items[i]
	}
	return result, nil
}
