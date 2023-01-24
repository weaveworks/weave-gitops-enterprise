package helm

import (
	"context"

	_ "github.com/mattn/go-sqlite3"
	"k8s.io/apimachinery/pkg/types"
)

// LayerAnnotation specifies profile application order.
// Profiles are sorted by layer and those at a higher "layer" are only installed after
// lower layers have successfully installed and started.
const LayerAnnotation = "weave.works/layer"

// ChartsCacheWriter is the "writing" interface to the cache, used by the reconciler etc
type ChartsCacheWriter interface {
	AddChart(ctx context.Context, name, version, kind, layer string, clusterRef types.NamespacedName, repoRef ObjectReference) error
	Delete(ctx context.Context, repoRef ObjectReference, clusterRef types.NamespacedName) error
	DeleteAllChartsForCluster(ctx context.Context, clusterRef types.NamespacedName) error
}

// ChartsCacheReader is the "reading" interface to the cache, used by api etc
type ChartsCacheReader interface {
	ListChartsByRepositoryAndCluster(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, kind string) ([]Chart, error)
	IsKnownChart(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, chart Chart) (bool, error)
	GetChartValues(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, chart Chart) ([]byte, error)
	UpdateValuesYaml(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, chart Chart, valuesYaml []byte) error
	GetLatestVersion(ctx context.Context, clusterRef, repoRef types.NamespacedName, name string) (string, error)
	GetLayer(ctx context.Context, clusterRef, repoRef types.NamespacedName, name, version string) (string, error)
}

type ChartsCache interface {
	ChartsCacheReader
	ChartsCacheWriter
}

// ObjectReference points to a resource.
type ObjectReference struct {
	Kind       string
	APIVersion string
	Name       string
	Namespace  string
}

// Chart holds the name and version of a chart.
type Chart struct {
	Name    string
	Version string
	Kind    string
	Layer   string
}

// Implementation of ChartsCache that does nothing.
type NilCache struct{}

func (n NilCache) AddChart(ctx context.Context, name, version, kind, layer string, clusterRef types.NamespacedName, repoRef ObjectReference) error {
	return nil
}

func (n NilCache) Delete(ctx context.Context, repoRef ObjectReference, clusterRef types.NamespacedName) error {
	return nil
}

func (n NilCache) DeleteAllChartsForCluster(ctx context.Context, clusterRef types.NamespacedName) error {
	return nil
}

func (n NilCache) ListChartsByRepositoryAndCluster(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, kind string) ([]Chart, error) {
	return nil, nil
}

func (n NilCache) IsKnownChart(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, chart Chart) (bool, error) {
	return false, nil
}

func (n NilCache) GetChartValues(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, chart Chart) ([]byte, error) {
	return nil, nil
}

func (n NilCache) UpdateValuesYaml(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, chart Chart, valuesYaml []byte) error {
	return nil
}

func (n NilCache) GetLatestVersion(ctx context.Context, clusterRef, repoRef types.NamespacedName, name string) (string, error) {
	return "", nil
}

func (n NilCache) GetLayer(ctx context.Context, clusterRef, repoRef types.NamespacedName, name, version string) (string, error) {
	return "", nil
}
