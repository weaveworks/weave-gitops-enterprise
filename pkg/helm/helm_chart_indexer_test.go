package helm

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
)

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

	err = indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", "layer-0",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"))
	assert.NoError(t, err)

	err = indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", "layer-0",
		nsn("cluster2", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"))
	assert.NoError(t, err)

	charts, err := indexer.ListChartsByRepositoryAndCluster(context.TODO(),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"),
		nsn("cluster1", "clusters"), "chart")

	assert.NoError(t, err)
	assert.Equal(t, 2, len(charts))
	assert.Equal(t, "redis", charts[0].Name)
	assert.Equal(t, "1.0.2", charts[0].Version)
	assert.Equal(t, "nginx", charts[1].Name)
	assert.Equal(t, "1.0.1", charts[1].Version)
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
