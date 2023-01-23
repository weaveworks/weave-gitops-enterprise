package helm

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/apimachinery/pkg/types"
)

type HelmIndexFileReader struct {
	index *repo.IndexFile
}

func NewHelmIndexFileReader(index *repo.IndexFile) *HelmIndexFileReader {
	return &HelmIndexFileReader{
		index: index,
	}
}

func (c HelmIndexFileReader) ListChartsByRepositoryAndCluster(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, kind string) ([]Chart, error) {
	var charts []Chart

	for _, chart := range c.index.Entries {
		for _, version := range chart {
			charts = append(charts, Chart{
				Name:    version.Name,
				Version: version.Version,
				Layer:   version.Annotations[LayerAnnotation],
			})
		}
	}

	sort.Slice(charts, func(i, j int) bool {
		return charts[i].Name < charts[j].Name
	})

	return charts, nil
}

func (c HelmIndexFileReader) IsKnownChart(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, chart Chart) (bool, error) {
	for _, version := range c.index.Entries[chart.Name] {
		if version.Version == chart.Version {
			return true, nil
		}
	}

	return false, nil
}

func (c HelmIndexFileReader) GetLatestVersion(ctx context.Context, clusterRef, repoRef types.NamespacedName, name string) (string, error) {
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

func (c HelmIndexFileReader) GetLayer(ctx context.Context, clusterRef, repoRef types.NamespacedName, name, version string) (string, error) {
	versions := c.index.Entries[name]

	for _, v := range versions {
		if v.Version == version {
			return v.Annotations[LayerAnnotation], nil
		}
	}

	return "", nil
}

// not implmented, this does not support reading values.yaml
func (c HelmIndexFileReader) GetChartValues(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, chart Chart) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// not implmented, this does not support reading values.yaml
func (c HelmIndexFileReader) UpdateValuesYaml(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, chart Chart, valuesYaml []byte) error {
	return errors.New("not implemented")
}
