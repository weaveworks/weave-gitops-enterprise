package helmfakes

import (
	"context"
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"k8s.io/apimachinery/pkg/types"
)

// NewFakeChartCache returns a new FakeChartCache
// Great for testing!
// Supports returning errors for each method
// Use the hander ClusterRefToString to get the a key for a chart in the cache
func NewFakeChartCache(opts ...func(*FakeChartCache)) FakeChartCache {
	fc := FakeChartCache{
		Charts: make(map[string][]helm.Chart),
	}
	for _, o := range opts {
		o(&fc)
	}

	return fc
}

func WithCharts(key string, charts []helm.Chart) func(*FakeChartCache) {
	return func(fc *FakeChartCache) {
		fc.Charts[key] = charts
	}
}

// fake cache implementation
type FakeChartCache struct {
	Charts                         map[string][]helm.Chart
	AddChartError                  error
	DeleteError                    error
	DeleteAllChartsForClusterError error
}

func (fc FakeChartCache) AddChart(ctx context.Context, name, version, kind, layer string, clusterRef types.NamespacedName, repoRef helm.ObjectReference) error {
	if fc.AddChartError != nil {
		return fc.AddChartError
	}

	k := ClusterRefToString(repoRef, clusterRef)
	fmt.Printf("Adding chart %s to cache with key %s\n", name, k)
	fc.Charts[k] = append(
		fc.Charts[k],
		helm.Chart{
			Name:    name,
			Version: version,
			Layer:   layer,
			Kind:    kind,
		},
	)
	return nil
}

func (fc FakeChartCache) Delete(ctx context.Context, repoRef helm.ObjectReference, clusterRef types.NamespacedName) error {
	if fc.DeleteError != nil {
		return fc.DeleteError
	}

	k := ClusterRefToString(repoRef, clusterRef)
	delete(fc.Charts, k)
	return nil
}

func (fc FakeChartCache) DeleteAllChartsForCluster(ctx context.Context, clusterRef types.NamespacedName) error {
	if fc.DeleteAllChartsForClusterError != nil {
		return fc.DeleteAllChartsForClusterError
	}
	// delete all the keys that end with the clusterRef
	for k := range fc.Charts {
		// if the key ends with the clusterRef, delete it
		if strings.HasSuffix(k, fmt.Sprintf("%s_%s", clusterRef.Name, clusterRef.Namespace)) {
			delete(fc.Charts, k)
		}
	}
	return nil
}

func ClusterRefToString(or helm.ObjectReference, cr types.NamespacedName) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s", or.Kind, or.Name, or.Namespace, cr.Name, cr.Namespace)
}
