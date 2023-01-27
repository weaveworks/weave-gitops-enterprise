package helm

import (
	"context"
	"testing"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/apimachinery/pkg/types"
)

func TestIndexGetLatestVersion(t *testing.T) {
	index := makeTestIndex()
	cache := NewHelmIndexFileReader(index)
	clusterRef := types.NamespacedName{}

	latest, err := cache.GetLatestVersion(context.TODO(), clusterRef, types.NamespacedName{}, "chart1")
	if err != nil {
		t.Fatal(err)
	}

	if latest != "1.0.1" {
		t.Fatal("unexpected latest version", latest)
	}
}

func TestIndexGetLayer(t *testing.T) {
	index := makeTestIndex()
	cache := NewHelmIndexFileReader(index)
	clusterRef := types.NamespacedName{}

	layer, err := cache.GetLayer(context.TODO(), clusterRef, types.NamespacedName{}, "chart2", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	if layer != "layer-0" {
		t.Fatal("unexpected layer", layer)
	}
}

func makeTestIndex() *repo.IndexFile {
	return &repo.IndexFile{
		Entries: map[string]repo.ChartVersions{
			"chart1": {
				makeTestChartVersion("chart1", "1.0.0", nil),
				makeTestChartVersion("chart1", "1.0.1", nil),
			},
			"chart2": {
				makeTestChartVersion("chart2", "1.0.0", map[string]string{
					LayerAnnotation: "layer-0",
				}),
				makeTestChartVersion("chart2", "1.0.1", nil),
			},
		},
	}
}

func makeTestChartVersion(name, version string, annotations map[string]string) *repo.ChartVersion {
	return &repo.ChartVersion{
		Metadata: &chart.Metadata{
			Name:        name,
			Version:     version,
			Annotations: annotations,
		},
	}
}
