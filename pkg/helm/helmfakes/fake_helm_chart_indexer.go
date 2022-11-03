package helmfakes

import (
	"context"
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"k8s.io/apimachinery/pkg/types"
)

func MakeFakeCache(opts ...func(*FakeChartCache)) FakeChartCache {
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
	Charts map[string][]helm.Chart
}

func (fc FakeChartCache) AddChart(ctx context.Context, name, version, kind, layer string, clusterRef types.NamespacedName, repoRef helm.ObjectReference) error {
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
	k := ClusterRefToString(repoRef, clusterRef)
	delete(fc.Charts, k)
	return nil
}

func (fc FakeChartCache) DeleteAllChartsForCluster(ctx context.Context, clusterRef types.NamespacedName) error {
	// delete all the keys that end with the clusterRef
	for k := range fc.Charts {
		// if the key ends with the clusterRef, delete it
		if strings.HasSuffix(k, fmt.Sprintf("%s_%s", clusterRef.Name, clusterRef.Namespace)) {
			delete(fc.Charts, k)
		}
	}
	return nil
}

// fake erroring cache implementation
type FakeErroringChartCache struct {
	AddError       error
	DeleteError    error
	DeleteAllError error
}

func (fc FakeErroringChartCache) AddChart(ctx context.Context, name, version, kind, layer string, clusterRef types.NamespacedName, repoRef helm.ObjectReference) error {
	return fc.AddError
}

func (fc FakeErroringChartCache) Delete(ctx context.Context, repoRef helm.ObjectReference, clusterRef types.NamespacedName) error {
	return fc.DeleteError
}

func (fc FakeErroringChartCache) DeleteAllChartsForCluster(ctx context.Context, clusterRef types.NamespacedName) error {
	return fc.DeleteAllError
}

func ClusterRefToString(or helm.ObjectReference, cr types.NamespacedName) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s", or.Kind, or.Name, or.Namespace, cr.Name, cr.Namespace)
}
