package helm

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
)

var _ ChartsCache = (*HelmChartIndexer)(nil)

func TestUpdateValuesYaml(t *testing.T) {
	// create a test db
	db := testCreateDB(t)

	// create a test indexer
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	// add a chart
	err := indexer.AddChart(context.TODO(), "redis", "1.0.1", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"))
	assert.NoError(t, err)

	// update the values.yaml
	err = indexer.UpdateValuesYaml(context.TODO(),
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"),
		Chart{Name: "redis", Version: "1.0.1"},
		[]byte("test"))
	assert.NoError(t, err)

	// get the values
	values, err := indexer.GetChartValues(context.TODO(),
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"),
		Chart{"redis", "1.0.1", "chart", "layer-0"},
	)
	assert.NoError(t, err)

	// check the values
	assert.Equal(t, []byte("test"), values)
}

func TestGetValuesFromDB(t *testing.T) {
	// create a test db
	db := testCreateDB(t)

	// create a test indexer
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	// test getting values from a non-existent chart
	values, err := indexer.GetChartValues(context.TODO(),
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"),
		Chart{"redis", "1.0.1", "chart", "layer-0"},
	)
	assert.NoError(t, err)
	assert.Nil(t, values)

	// add a chart
	err = indexer.AddChart(context.TODO(), "redis", "1.0.1", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"))
	assert.NoError(t, err)

	// update the values.yaml
	err = indexer.UpdateValuesYaml(context.TODO(),
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"),
		Chart{Name: "redis", Version: "1.0.1"},
		[]byte("test"))
	assert.NoError(t, err)

	// get the values
	values, err = indexer.GetChartValues(context.TODO(),
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"),
		Chart{"redis", "1.0.1", "chart", "layer-0"},
	)
	assert.NoError(t, err)

	// check the values
	assert.Equal(t, []byte("test"), values)

}

func TestIsKnownChart(t *testing.T) {
	// create a test db
	db := testCreateDB(t)

	// create a test indexer
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	// see if a chart is known
	known, err := indexer.IsKnownChart(context.TODO(),
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"),
		Chart{"redis", "1.0.1", "chart", ""})
	assert.NoError(t, err)

	// should not be known
	assert.False(t, known)

	// add a chart
	err = indexer.AddChart(context.TODO(), "redis", "1.0.1", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"))
	assert.NoError(t, err)

	// see if a chart is known
	known, err = indexer.IsKnownChart(context.TODO(),
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"),
		Chart{"redis", "1.0.1", "chart", ""})
	assert.NoError(t, err)

	// should be known
	assert.True(t, known)
}

func TestHelmChartIndexer_Delete(t *testing.T) {
	db := testCreateDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	err := indexer.AddChart(context.TODO(), "redis", "1.0.1", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "weave-charts", "team-ns"))
	assert.NoError(t, err)

	err = indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", "layer-1",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"))
	assert.NoError(t, err)

	err = indexer.Delete(context.TODO(), objref("", "", "weave-charts", "team-ns"), nsn("cluster1", "clusters"))
	assert.NoError(t, err)

	chart, err := indexer.ListChartsByCluster(context.TODO(), nsn("cluster1", "clusters"), "chart")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(chart))
	assert.Equal(t, "nginx", chart[0].Name)
}

func TestHelmChartIndexer_RemoveChart(t *testing.T) {
	db := testCreateDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	err := indexer.AddChart(context.TODO(), "redis", "1.0.1", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "weave-charts", "team-ns"))
	assert.NoError(t, err)

	err = indexer.AddChart(context.TODO(), "redis", "1.0.2", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "weave-charts", "team-ns"))
	assert.NoError(t, err)

	err = indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", "layer-1",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "weave-charts", "team-ns"))
	assert.NoError(t, err)

	err = indexer.RemoveChart(context.TODO(), "redis", "1.0.1", nsn("cluster1", "clusters"), objref("HelmRepository", "", "weave-charts", "team-ns"))
	assert.NoError(t, err)

	charts, err := indexer.ListChartsByCluster(context.TODO(), nsn("cluster1", "clusters"), "chart")
	assert.NoError(t, err)

	want := []Chart{
		{Name: "redis", Version: "1.0.2"},
		{Name: "nginx", Version: "1.0.1"},
	}
	assert.Equal(t, want, charts)
}

func TestHelmChartIndexer_Count(t *testing.T) {
	db := testCreateDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	count, err := indexer.Count(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	err = indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", "",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"))
	assert.NoError(t, err)

	count, err = indexer.Count(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	err = indexer.AddChart(context.TODO(), "redis", "1.0.1", "chart", "",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"))
	assert.NoError(t, err)

	count, err = indexer.Count(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestHelmChartIndexer_ListChartsByCluster(t *testing.T) {
	db := testCreateDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	err := indexer.AddChart(context.TODO(), "redis", "1.0.1", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"))
	assert.NoError(t, err)

	err = indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", "layer-1",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"))
	assert.NoError(t, err)

	err = indexer.AddChart(context.TODO(), "nginx", "1.0.2", "chart", "layer-1",
		nsn("cluster2", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"))
	assert.NoError(t, err)

	charts, err := indexer.ListChartsByCluster(context.TODO(), nsn("cluster1", "clusters"), "chart")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(charts))
	assert.Equal(t, "redis", charts[0].Name)
	assert.Equal(t, "1.0.1", charts[1].Version)
}

func TestHelmChartIndexer_ListChartsByRepositoryAndCluster(t *testing.T) {
	db := testCreateDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	err := indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "foo-charts", "team-ns"))

	assert.NoError(t, err)

	err = indexer.AddChart(context.TODO(), "redis", "1.0.2", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"))
	assert.NoError(t, err)

	err = indexer.AddChart(context.TODO(), "nginx", "1.0.3", "profile", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"))
	assert.NoError(t, err)

	err = indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"))
	assert.NoError(t, err)

	err = indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", "layer-0",
		nsn("cluster2", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"))
	assert.NoError(t, err)

	charts, err := indexer.ListChartsByRepositoryAndCluster(context.TODO(),
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"),
		"chart")

	assert.NoError(t, err)
	assert.Equal(t, 2, len(charts))
	assert.Equal(t, "redis", charts[0].Name)
	assert.Equal(t, "1.0.2", charts[0].Version)
	assert.Equal(t, "nginx", charts[1].Name)
	assert.Equal(t, "1.0.1", charts[1].Version)
}

func TestAddChart(t *testing.T) {
	db := testCreateDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	err := indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "foo-charts", "team-ns"))
	assert.NoError(t, err)

	err = indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "foo-charts", "team-ns"))
	assert.NoError(t, err)

	charts, err := indexer.ListChartsByRepositoryAndCluster(context.TODO(),
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "foo-charts", "team-ns"),
		"chart")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(charts))
}

func TestGetLatestVersion(t *testing.T) {
	db := testCreateDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	err := indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "foo-charts", "team-ns"))
	assert.NoError(t, err)

	err = indexer.AddChart(context.TODO(), "nginx", "1.0.2", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "foo-charts", "team-ns"))
	assert.NoError(t, err)

	err = indexer.AddChart(context.TODO(), "nginx", "1.0.3", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "foo-charts", "team-ns"))
	assert.NoError(t, err)

	version, err := indexer.GetLatestVersion(context.TODO(),
		nsn("cluster1", "clusters"),
		nsn("foo-charts", "team-ns"),
		"nginx")
	assert.NoError(t, err)
	assert.Equal(t, "1.0.3", version)
}

func TestGetLatestVersion_NotFound(t *testing.T) {
	db := testCreateDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	err := indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "foo-charts", "team-ns"))
	assert.NoError(t, err)

	version, err := indexer.GetLatestVersion(context.TODO(),
		nsn("cluster1", "clusters"),
		nsn("foo-charts", "team-ns"),
		"redis")
	assert.NoError(t, err)
	assert.Equal(t, "", version)
}

func TestGetLayer(t *testing.T) {
	db := testCreateDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	err := indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "foo-charts", "team-ns"))
	assert.NoError(t, err)

	layer, err := indexer.GetLayer(context.TODO(),
		nsn("cluster1", "clusters"),
		nsn("foo-charts", "team-ns"),
		"nginx",
		"1.0.1")
	assert.NoError(t, err)
	assert.Equal(t, "layer-0", layer)
}

func TestGetLayer_NotFound(t *testing.T) {
	db := testCreateDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	err := indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "foo-charts", "team-ns"))
	assert.NoError(t, err)

	layer, err := indexer.GetLayer(context.TODO(),
		nsn("cluster1", "clusters"),
		nsn("foo-charts", "team-ns"),
		"redis",
		"1.0.1")
	assert.NoError(t, err)
	assert.Equal(t, "", layer)
}

func TestDeleteAllChartsForCluster(t *testing.T) {
	db := testCreateDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	// add a chart to cluster1
	err := indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "foo-charts", "team-ns"))
	assert.NoError(t, err)

	// add a chart to cluster2
	err = indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", "layer-0",
		nsn("cluster2", "clusters"),
		objref("HelmRepository", "", "foo-charts", "team-ns"))
	assert.NoError(t, err)

	// delete all charts for cluster1
	err = indexer.DeleteAllChartsForCluster(context.TODO(), nsn("cluster1", "clusters"))
	assert.NoError(t, err)

	// check that no charts are returned for cluster1
	charts, err := indexer.ListChartsByCluster(context.TODO(), nsn("cluster1", "clusters"), "chart")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(charts))

	// check that only the chart for cluster2 is left
	charts, err = indexer.ListChartsByCluster(context.TODO(), nsn("cluster2", "clusters"), "chart")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(charts))
}

func objref(kind, apiVersion, name, namespace string) ObjectReference {
	return ObjectReference{
		Kind:       kind,
		APIVersion: apiVersion,
		Name:       name,
		Namespace:  namespace,
	}
}

func nsn(name, namespace string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
}

func testCreateDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	// From the readme: https://github.com/mattn/go-sqlite3
	db.SetMaxOpenConns(1)

	if err := applySchema(db); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	})

	return db
}
