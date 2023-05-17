package helmfakes

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"k8s.io/apimachinery/pkg/types"
)

// NewFakeChartCache returns a new FakeChartCache
// Great for testing!
// Supports returning errors for each method
// Use the helper ClusterRefToString to get the a key for a chart in the cache
func NewFakeChartCache(opts ...func(*FakeChartCache)) FakeChartCache {
	fc := FakeChartCache{
		Charts:      make(map[string][]helm.Chart),
		ChartValues: make(map[string][]byte),
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

func WithValues(key string, values []byte) func(*FakeChartCache) {
	return func(fc *FakeChartCache) {
		fc.ChartValues[key] = []byte(base64.StdEncoding.EncodeToString(values))
	}
}

// fake cache implementation
type FakeChartCache struct {
	Charts                         map[string][]helm.Chart
	ChartValues                    map[string][]byte
	AddChartError                  error
	RemoveChartError               error
	DeleteError                    error
	DeleteAllChartsForClusterError error
}

func (fc FakeChartCache) AddChart(ctx context.Context, name, version, kind, layer string, clusterRef types.NamespacedName, repoRef helm.ObjectReference) error {
	if fc.AddChartError != nil {
		return fc.AddChartError
	}

	// the database-based version guards against duplicate data
	// using the name and version as the guard values
	k := ClusterRefToString(repoRef, clusterRef)
	exists := func(hc []helm.Chart) bool {
		for _, c := range hc {
			if c.Name == name && c.Version == version {
				return true
			}
		}

		return false
	}(fc.Charts[k])

	if exists {
		return nil
	}

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

func (fc FakeChartCache) RemoveChart(ctx context.Context, name, version string, clusterRef types.NamespacedName, repoRef helm.ObjectReference) error {
	if fc.RemoveChartError != nil {
		return fc.RemoveChartError
	}

	k := ClusterRefToString(repoRef, clusterRef)
	fmt.Printf("Removing chart %s from cache with key %s\n", name, k)

	charts := fc.Charts[k]
	var updated []helm.Chart
	for _, v := range charts {
		if !(v.Name == name && v.Version == version) {
			updated = append(updated, v)
		}
	}

	fc.Charts[k] = updated

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

// Reader interface

func (fc FakeChartCache) ListChartsByRepositoryAndCluster(ctx context.Context, clusterRef types.NamespacedName, repoRef helm.ObjectReference, kind string) ([]helm.Chart, error) {
	charts, ok := fc.Charts[ClusterRefToString(repoRef, clusterRef)]
	if !ok {
		return nil, errors.New("no charts found")
	}
	// filter by kind
	var filtered []helm.Chart
	for _, c := range charts {
		if c.Kind == kind {
			filtered = append(filtered, c)
		}
	}
	return filtered, nil
}
func (fc FakeChartCache) IsKnownChart(ctx context.Context, clusterRef types.NamespacedName, repoRef helm.ObjectReference, chart helm.Chart) (bool, error) {
	charts, ok := fc.Charts[ClusterRefToString(repoRef, clusterRef)]
	if !ok {
		return false, nil
	}
	for _, c := range charts {
		if c.Name == chart.Name && c.Version == chart.Version {
			return true, nil
		}
	}
	return false, nil
}
func (fc FakeChartCache) GetChartValues(ctx context.Context, clusterRef types.NamespacedName, repoRef helm.ObjectReference, chart helm.Chart) ([]byte, error) {
	if values, ok := fc.ChartValues[ChartRefToString(repoRef, clusterRef, chart)]; ok {
		return values, nil
	}
	return nil, nil
}
func (fc FakeChartCache) UpdateValuesYaml(ctx context.Context, clusterRef types.NamespacedName, repoRef helm.ObjectReference, chart helm.Chart, valuesYaml []byte) error {
	fc.ChartValues[ChartRefToString(repoRef, clusterRef, chart)] = valuesYaml
	return nil
}

func (fc FakeChartCache) GetLatestVersion(ctx context.Context, clusterRef, repoRef types.NamespacedName, name string) (string, error) {
	repoObjRef := helm.ObjectReference{
		Name:      repoRef.Name,
		Namespace: repoRef.Namespace,
	}

	charts, ok := fc.Charts[ClusterRefToString(repoObjRef, clusterRef)]
	if !ok {
		return "", errors.New("no charts found")
	}
	var versions []string
	for _, c := range charts {
		if c.Name == name {
			versions = append(versions, c.Version)
		}
	}
	if len(versions) == 0 {
		return "", errors.New("no charts found")
	}
	// sort by version
	sorted, err := helm.ReverseSemVerSort(versions)
	if err != nil {
		return "", err
	}
	// return the latest version
	return sorted[0], nil
}

func (fc FakeChartCache) GetLayer(ctx context.Context, clusterRef, repoRef types.NamespacedName, name, version string) (string, error) {
	repoObjRef := helm.ObjectReference{
		Name:      repoRef.Name,
		Namespace: repoRef.Namespace,
	}

	charts, ok := fc.Charts[ClusterRefToString(repoObjRef, clusterRef)]
	if !ok {
		return "", errors.New("no charts found")
	}
	for _, c := range charts {
		if c.Name == name && c.Version == version {
			return c.Layer, nil
		}
	}
	return "", nil
}

func ChartRefToString(or helm.ObjectReference, cr types.NamespacedName, c helm.Chart) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s_%s_%s", or.Kind, or.Name, or.Namespace, cr.Name, cr.Namespace, c.Name, c.Version)
}

func ClusterRefToString(or helm.ObjectReference, cr types.NamespacedName) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s", or.Kind, or.Name, or.Namespace, cr.Name, cr.Namespace)
}
