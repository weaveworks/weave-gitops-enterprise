package helm

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"k8s.io/apimachinery/pkg/types"
)

// legacy interface

func TestDelete(t *testing.T) {
	db := testCreateDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	if err := indexer.AddChart(context.TODO(), "redis", "1.0.1", "chart", nil,
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "weave-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}
	if err := indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", nil,
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}

	if err := indexer.Delete(context.TODO(), objref("", "", "weave-charts", "team-ns"), nsn("cluster1", "clusters")); err != nil {
		t.Fatal(err)
	}

	count, err := indexer.Count(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Fatalf("got %d, want 1", count)
	}

}

func TestHelmChartIndex(t *testing.T) {
	db := testCreateDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	if err := indexer.AddChart(context.TODO(), "redis", "1.0.1", "chart", nil,
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}
	count, err := indexer.Count(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("got %d, want 1", count)
	}
}

func TestListChartsByCluster(t *testing.T) {
	db := testCreateDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	if err := indexer.AddChart(context.TODO(), "redis", "1.0.1", "chart", nil,
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}
	if err := indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", nil,
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}
	if err := indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", nil,
		nsn("cluster2", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}

	charts, err := indexer.ListChartsByCluster(context.TODO(), nsn("cluster1", "clusters"), "chart")
	if err != nil {
		t.Fatal(err)
	}
	if len(charts) != 2 {
		t.Fatalf("got %d, want 2", len(charts))
	}
}

func TestListChartsByRepositoryAndCluster(t *testing.T) {
	db := testCreateDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	if err := indexer.AddChart(context.TODO(), "redis", "1.0.1", "chart", nil,
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}
	if err := indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", nil,
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}
	if err := indexer.AddChart(context.TODO(), "nginx", "1.0.1", "chart", nil,
		nsn("cluster2", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}

	charts, err := indexer.ListChartsByRepositoryAndCluster(context.TODO(),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"),
		nsn("cluster1", "clusters"), "chart")
	if err != nil {
		t.Fatal(err)
	}
	if len(charts) != 2 {
		t.Fatalf("got %d, want 2", len(charts))
	}
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
	if err != nil {
		t.Fatal(err)
	}

	return db
}
