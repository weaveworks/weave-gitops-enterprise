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

	weaveCluster := capiv1.Cluster{}
	lib.Log.Info("Getting weavecluster", "cluster", name)
	err = cl.Get(ctx, client.ObjectKey{
		Namespace: lib.Namespace,
		Name:      name,
	}, &weaveCluster)
	if err != nil {
		lib.Log.Error(err, "Failed to get weavecluster", "cluster", name)
		return nil, fmt.Errorf("error getting weavecluster %s/%s: %s", lib.Namespace, name, err)
	}
	lib.Log.Info("Got weavecluster", "cluster", name)

	return &weaveCluster, nil
}

func (lib *CRDLibrary) List(ctx context.Context) (map[string]*capiv1.Cluster, error) {
	lib.Log.Info("Getting client from context")
	cl, err := lib.ClientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	lib.Log.Info("Querying namespace for WeaveCluster resources", "namespace", lib.Namespace)
	weaveClusterList := capiv1.ClusterList{}
	err = cl.List(ctx, &weaveClusterList, client.InNamespace(lib.Namespace))
	if err != nil {
		return nil, fmt.Errorf("error getting weaveclusters: %s", err)
	}
	lib.Log.Info("Got weaveclusters", "numberOfClusters", len(weaveClusterList.Items))

	result := map[string]*capiv1.Cluster{}
	for i, ct := range weaveClusterList.Items {
		result[ct.ObjectMeta.Name] = &weaveClusterList.Items[i]
	}
	return result, nil
}
