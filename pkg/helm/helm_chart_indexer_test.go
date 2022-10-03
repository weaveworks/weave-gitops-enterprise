package helm

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	pb "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos/profiles"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/watcher/cache"
)

// legacy interface

func TestDelete(t *testing.T) {
	db := createDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	if err := indexer.AddChart(context.TODO(), "redis", "1.0.1", nil,
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "weave-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}
	if err := indexer.AddChart(context.TODO(), "nginx", "1.0.1", nil,
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}

	if err := indexer.Delete(context.TODO(), "team-ns", "weave-charts"); err != nil {
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

func TestGetProfileValues(t *testing.T) {
	db := createDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	if err := indexer.AddChart(context.TODO(), "redis", "1.0.1", []byte("hi: there"),
		nsn(ManagementClusterName, ManagementClusterNamespace),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}

	values, err := indexer.GetProfileValues(context.TODO(), "team-ns", "bitnami-charts", "redis", "1.0.1")
	if err != nil {
		t.Fatal(err)
	}

	if string(values) != "hi: there" {
		t.Fatalf("got %s, want 'hi: there'", string(values))
	}

	_, err = indexer.GetProfileValues(context.TODO(), "team-ns", "bitnami-charts", "missing", "1.0.1")
	if err != sql.ErrNoRows {
		t.Fatalf("got %v, want sql.ErrNoRows", err)
	}
}

func TestListProfiles(t *testing.T) {
	db := createDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	data := cache.Data{
		Profiles: []*pb.Profile{
			{
				Name:              "redis",
				AvailableVersions: []string{"1.0.1"},
				Layer:             "layer-0",
			},
			{
				Name:              "nginx",
				AvailableVersions: []string{"1.0.2", "1.0.3"},
				Layer:             "layer-1",
			},
		},
		Values: cache.ValueMap{
			"redis": {
				"1.0.1": []byte("hi: there"),
			},
			"nginx": {
				"1.0.2": []byte("hi: there"),
				"1.0.3": []byte("hi: there"),
			},
		},
	}

	err := indexer.Put(context.TODO(), "team-ns", "bitnami-charts", data)
	if err != nil {
		t.Fatal(err)
	}

	profiles, err := indexer.ListProfiles(context.TODO(), "team-ns", "bitnami-charts")
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 2 {
		t.Fatalf("got %d, want 2", len(profiles))
	}

	for _, profile := range profiles {
		if profile.Name == "nginx" {
			if profile.Layer != "layer-1" {
				t.Fatalf("got %s, want layer-1", profile.Layer)
			}
			if len(profile.AvailableVersions) != 2 {
				t.Fatalf("got %d, want 2", len(profile.AvailableVersions))
			}
		}
	}
}

func TestListAvailableVersionsForProfile(t *testing.T) {
	db := createDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	if err := indexer.AddChart(context.TODO(), "redis", "1.0.1", nil,
		nsn(ManagementClusterName, ManagementClusterNamespace),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}
	if err := indexer.AddChart(context.TODO(), "redis", "1.0.2", nil,
		nsn(ManagementClusterName, ManagementClusterNamespace),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}
	if err := indexer.AddChart(context.TODO(), "nginx", "1.0.3", nil,
		nsn(ManagementClusterName, ManagementClusterNamespace),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}

	versions, err := indexer.ListAvailableVersionsForProfile(context.TODO(), "team-ns", "bitnami-charts", "redis")
	if err != nil {
		t.Fatal(err)
	}

	if len(versions) != 2 {
		t.Fatalf("got %d, want 2", len(versions))
	}
}

func TestHelmChartIndex(t *testing.T) {
	db := createDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	if err := indexer.AddChart(context.TODO(), "redis", "1.0.1", nil,
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
	db := createDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	if err := indexer.AddChart(context.TODO(), "redis", "1.0.1", nil,
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}
	if err := indexer.AddChart(context.TODO(), "nginx", "1.0.1", nil,
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}
	if err := indexer.AddChart(context.TODO(), "nginx", "1.0.1", nil,
		nsn("cluster2", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}

	charts, err := indexer.ListChartsByCluster(context.TODO(), nsn("cluster1", "clusters"))
	if err != nil {
		t.Fatal(err)
	}
	if len(charts) != 2 {
		t.Fatalf("got %d, want 2", len(charts))
	}
}

func TestListChartsByRepositoryAndCluster(t *testing.T) {
	db := createDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	if err := indexer.AddChart(context.TODO(), "redis", "1.0.1", nil,
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}
	if err := indexer.AddChart(context.TODO(), "nginx", "1.0.1", nil,
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}
	if err := indexer.AddChart(context.TODO(), "nginx", "1.0.1", nil,
		nsn("cluster2", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}

	charts, err := indexer.ListChartsByRepositoryAndCluster(context.TODO(),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"),
		nsn("cluster1", "clusters"))
	if err != nil {
		t.Fatal(err)
	}
	if len(charts) != 2 {
		t.Fatalf("got %d, want 2", len(charts))
	}
}
