package helm

import (
	"context"
	"errors"
	"fmt"

	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/apimachinery/pkg/types"
)

type HelmLocalChartsCacheReader struct {
	index *repo.IndexFile
}

func NewHelmChartLocalCache(index *repo.IndexFile) HelmLocalChartsCacheReader {
	return HelmLocalChartsCacheReader{
		index: index,
	}
}

// implements ChartsCacheReader

func (c HelmLocalChartsCacheReader) ListChartsByRepositoryAndCluster(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, kind string) ([]Chart, error) {
	var charts []Chart

	for _, chart := range c.index.Entries {
		for _, version := range chart {
			charts = append(charts, Chart{
				Name:    version.Name,
				Version: version.Version,
				Kind:    kind,
				Layer:   version.Annotations[LayerAnnotation],
			})
		}
	}

	return charts, nil
}

func (c HelmLocalChartsCacheReader) IsKnownChart(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, chart Chart) (bool, error) {
	for _, version := range c.index.Entries[chart.Name] {
		if version.Version == chart.Version {
			return true, nil
		}
	}

	return false, nil
}

func (c HelmLocalChartsCacheReader) GetLatestVersion(ctx context.Context, clusterRef, repoRef types.NamespacedName, name string) (string, error) {
	versions := []string{}
	for _, v := range c.index.Entries[name] {
		versions = append(versions, v.Version)
	}

	sorted, err := ReverseSemVerSort(versions)
	if err != nil {
		return "", fmt.Errorf("retrieving latest version %s: %w", name, err)
	}

	return sorted[0], nil
}

func (c HelmLocalChartsCacheReader) GetLayer(ctx context.Context, clusterRef, repoRef types.NamespacedName, name, version string) (string, error) {
	versions := c.index.Entries[name]

	for _, v := range versions {
		if v.Version == version {
			return v.Annotations[LayerAnnotation], nil
		}
	}

	return "", nil
}

// not implmented, this does not support reading values.yaml
func (c HelmLocalChartsCacheReader) GetChartValues(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, chart Chart) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// not implmented, this does not support reading values.yaml
func (c HelmLocalChartsCacheReader) UpdateValuesYaml(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, chart Chart, valuesYaml []byte) error {
	return errors.New("not implemented")
}
