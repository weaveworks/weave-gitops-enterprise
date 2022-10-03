package helm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	pb "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos/profiles"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/watcher/cache"
	"k8s.io/apimachinery/pkg/types"
)

const ManagementClusterName = "management"
const ManagementClusterNamespace = "default"

// ObjectReference points to a resource.
type ObjectReference struct {
	Kind       string
	APIVersion string
	Name       string
	Namespace  string
}

type Chart struct {
	Name    string
	Version string
}

// HelmChartIndexer indexs details of Helm charts that have been seen in Helm
// repositories.
type HelmChartIndexer struct {
	CacheDB *sql.DB
}

// NewCache initialises the cache and returns it.
func NewChartIndexer() (*HelmChartIndexer, error) {
	db, err := createMemoryDB()
	if err != nil {
		return nil, err
	}

	return &HelmChartIndexer{
		CacheDB: db,
	}, nil
}

//
// Implement "legacy" helm.Cache interface
//

func (i *HelmChartIndexer) Put(ctx context.Context, helmRepoNamespace, helmRepoName string, value cache.Data) error {
	// FIXME: dummy values
	kind := "HelmRepository"
	apiVersion := "source.toolkit.fluxcd.io/v1beta1"

	for _, chart := range value.Profiles {
		for _, version := range chart.AvailableVersions {
			sqlStatement := `
INSERT INTO helm_charts (name, version, valuesYaml, layer,
	repo_kind, repo_api_version, repo_name, repo_namespace,
	cluster_name, cluster_namespace)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

			_, err := i.CacheDB.ExecContext(
				ctx,
				sqlStatement, chart.Name, version,
				value.Values[chart.Name][version],
				chart.Layer,
				kind, apiVersion, helmRepoName, helmRepoNamespace,
				ManagementClusterName, ManagementClusterNamespace)

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (i *HelmChartIndexer) Delete(ctx context.Context, helmRepoNamespace, helmRepoName string) error {
	sqlStatement := `
DELETE FROM helm_charts
WHERE repo_name = $1 AND repo_namespace = $2`

	_, err := i.CacheDB.ExecContext(ctx, sqlStatement, helmRepoName, helmRepoNamespace)
	return err
}

func (i *HelmChartIndexer) ListProfiles(ctx context.Context, helmRepoNamespace, helmRepoName string) ([]*pb.Profile, error) {
	// select profiles aggregating available versions into a single row
	sqlStatement := `
	select name, group_concat(version) as versions, layer
	from helm_charts
	where repo_name = $1 and repo_namespace = $2 and cluster_name = $3 and cluster_namespace = $4
	group by name`

	rows, err := i.CacheDB.QueryContext(ctx, sqlStatement, helmRepoName, helmRepoNamespace, ManagementClusterName, ManagementClusterNamespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*pb.Profile
	for rows.Next() {
		var p pb.Profile
		var versions string
		var layer sql.NullString
		if err := rows.Scan(&p.Name, &versions, &layer); err != nil {
			return nil, err
		}
		p.AvailableVersions = strings.Split(versions, ",")
		p.AvailableVersions, err = ReverseSemVerSort(p.AvailableVersions)
		if err != nil {
			return nil, fmt.Errorf("parsing template profile %s: %w", p.Name, err)
		}

		if layer.Valid {
			p.Layer = layer.String
		}

		profiles = append(profiles, &p)
	}

	return profiles, nil
}

func (i *HelmChartIndexer) GetProfileValues(ctx context.Context, helmRepoNamespace, helmRepoName, profileName, profileVersion string) ([]byte, error) {
	sqlStatement := `
	select valuesYaml
	from helm_charts
	where name = $1 and version = $2 and repo_name = $3 and repo_namespace = $4 and cluster_name = $5 and cluster_namespace = $6`

	var valuesYaml []byte
	err := i.CacheDB.QueryRowContext(ctx, sqlStatement, profileName, profileVersion, helmRepoName, helmRepoNamespace, ManagementClusterName, ManagementClusterNamespace).Scan(&valuesYaml)

	if err != nil {
		return nil, err
	}

	return valuesYaml, nil
}

func (i *HelmChartIndexer) ListAvailableVersionsForProfile(ctx context.Context, helmRepoNamespace, helmRepoName, profileName string) ([]string, error) {
	sqlStatement := `
	select version
	from helm_charts
	where name = $1 and repo_name = $2 and repo_namespace = $3 and cluster_name = $4 and cluster_namespace = $5`

	rows, err := i.CacheDB.QueryContext(ctx, sqlStatement, profileName, helmRepoName, helmRepoNamespace, ManagementClusterName, ManagementClusterNamespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, nil
}

//
// Future potential interface
//

// AddChart inserts a new chart into helm_charts table.
func (i *HelmChartIndexer) AddChart(ctx context.Context, name, version string, values []byte, clusterRef types.NamespacedName, repoRef ObjectReference) error {
	sqlStatement := `
INSERT INTO helm_charts (name, version, valuesYaml,
	repo_kind, repo_api_version, repo_name, repo_namespace,
	cluster_name, cluster_namespace)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := i.CacheDB.ExecContext(
		ctx,
		sqlStatement, name, version, values,
		repoRef.Kind, repoRef.APIVersion, repoRef.Name, repoRef.Namespace,
		clusterRef.Name, clusterRef.Namespace)

	return err
}

func (i *HelmChartIndexer) Count(ctx context.Context) (int64, error) {
	rows, err := i.CacheDB.QueryContext(ctx, "SELECT COUNT(*) FROM helm_charts")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		var n int64
		if err := rows.Scan(&n); err != nil {
			return 0, err
		}
		count += n
	}

	return count, nil
}

// ListChartsByCluster returns a list of charts filtered by cluster.
func (i *HelmChartIndexer) ListChartsByCluster(ctx context.Context, clusterRef types.NamespacedName) ([]Chart, error) {
	sqlStatement := `
SELECT name, version FROM helm_charts 
WHERE cluster_name = $1 AND cluster_namespace = $2`

	rows, err := i.CacheDB.QueryContext(ctx, sqlStatement, clusterRef.Name, clusterRef.Namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var charts []Chart
	for rows.Next() {
		var chart Chart
		if err := rows.Scan(&chart.Name, &chart.Version); err != nil {
			return nil, err
		}
		charts = append(charts, chart)
	}

	return charts, nil
}

// ListChartsByRepositoryAndCluster returns a list of charts filtered by helm repository and cluster.
func (i *HelmChartIndexer) ListChartsByRepositoryAndCluster(ctx context.Context, repoRef ObjectReference, clusterRef types.NamespacedName) ([]Chart, error) {
	sqlStatement := `
SELECT name, version FROM helm_charts 
WHERE repo_kind = $1 AND repo_api_version = $2 AND repo_name = $3 AND repo_namespace = $4
AND cluster_name = $5 AND cluster_namespace = $6`

	rows, err := i.CacheDB.QueryContext(ctx, sqlStatement, repoRef.Kind, repoRef.APIVersion, repoRef.Name, repoRef.Namespace, clusterRef.Name, clusterRef.Namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var charts []Chart
	for rows.Next() {
		var chart Chart
		if err := rows.Scan(&chart.Name, &chart.Version); err != nil {
			return nil, err
		}
		charts = append(charts, chart)
	}

	return charts, nil
}

func nsn(name, namespace string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: namespace,
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

func applySchema(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS helm_charts (
	name text, version text, valuesYaml blob, layer text,
	repo_kind text, repo_api_version text, repo_name text, repo_namespace text,
	cluster_name text, cluster_namespace text);
	`)
	return err
}

func createMemoryDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		return nil, err
	}
	// From the readme: https://github.com/mattn/go-sqlite3
	db.SetMaxOpenConns(1)

	if err := applySchema(db); err != nil {
		return db, err
	}

	return db, nil
}

func createDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := createMemoryDB()
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
