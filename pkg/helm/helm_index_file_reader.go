package helm

import (
	"context"
	"fmt"

	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/apimachinery/pkg/types"
)

type HelmIndexFileReader struct {
	index *repo.IndexFile
}

// Returns a new HelmIndexFileReader a ChartsCacheReader implementation
// that reads from a Helm IndexFile. As we're only reading from a single
// index file, we ignore the repoRef and clusterRef parameters provided
// to most methods.
func NewHelmIndexFileReader(index *repo.IndexFile) *HelmIndexFileReader {
	return &HelmIndexFileReader{
		index: index,
	}
}

// GetLatestVersion returns the latest version of the chart. The repoRef
// and clusterRef parameters are ignored.
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

// GetLayer returns the layer of the chart. The repoRef
// and clusterRef parameters are ignored.
func (c HelmIndexFileReader) GetLayer(ctx context.Context, clusterRef, repoRef types.NamespacedName, name, version string) (string, error) {
	versions := c.index.Entries[name]

	for _, v := range versions {
		if v.Version == version {
			return v.Annotations[LayerAnnotation], nil
		}
	}

	return "", nil
}
