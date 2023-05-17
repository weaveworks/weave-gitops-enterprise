package helm

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/Masterminds/semver"
	_ "github.com/mattn/go-sqlite3"
	"k8s.io/apimachinery/pkg/types"
)

// dbFile is the name of the sqlite3 database file
const dbFile = "charts.db"

// HelmChartIndexer indexs details of Helm charts that have been seen in Helm
// repositories.
type HelmChartIndexer struct {
	CacheDB *sql.DB
	Cluster types.NamespacedName
}

// NewCache initialises the cache and returns it.
func NewChartIndexer(cacheLocation, mgmtCluster string) (*HelmChartIndexer, error) {
	db, err := createDB(cacheLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache database: %w", err)
	}

	return &HelmChartIndexer{
		CacheDB: db,
		Cluster: types.NamespacedName{
			Name: mgmtCluster,
		},
	}, nil
}

// AddChart inserts a new chart into helm_charts table.
func (i *HelmChartIndexer) AddChart(ctx context.Context, name, version, kind, layer string, clusterRef types.NamespacedName, repoRef ObjectReference) error {
	sqlStatement := `
INSERT INTO helm_charts (name, version, kind, layer,
	repo_kind, repo_api_version, repo_name, repo_namespace,
	cluster_name, cluster_namespace)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	found, err := i.IsKnownChart(ctx, clusterRef, repoRef, Chart{Name: name, Version: version})
	if err != nil {
		return fmt.Errorf("failed to check if chart is known: %w", err)
	}
	if found {
		return nil
	}

	_, err = i.CacheDB.ExecContext(
		ctx,
		sqlStatement, name, version, kind, layer,
		repoRef.Kind, repoRef.APIVersion, repoRef.Name, repoRef.Namespace,
		clusterRef.Name, clusterRef.Namespace)

	return err
}

// RemoveChart removes a chart from the helm_charts table if it exists.
func (i *HelmChartIndexer) RemoveChart(ctx context.Context, name, version string, clusterRef types.NamespacedName, repoRef ObjectReference) error {
	sqlStatement := `
DELETE FROM helm_charts WHERE name = $1 AND version = $2
AND repo_name = $3 AND repo_namespace = $4
AND cluster_name = $5 AND cluster_namespace = $6
`
	_, err := i.CacheDB.ExecContext(
		ctx,
		sqlStatement, name, version,
		repoRef.Name, repoRef.Namespace,
		clusterRef.Name, clusterRef.Namespace)

	return err
}

// IsKnownChart returns true if the chart in in a repo is known
func (i *HelmChartIndexer) IsKnownChart(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, chart Chart) (bool, error) {
	sqlStatement := `
SELECT COUNT(*) FROM helm_charts 
WHERE name = $1 AND version = $2
AND repo_name = $3 AND repo_namespace = $4
AND cluster_name = $5 AND cluster_namespace = $6`

	rows, err := i.CacheDB.QueryContext(ctx, sqlStatement, chart.Name, chart.Version, repoRef.Name, repoRef.Namespace, clusterRef.Name, clusterRef.Namespace)
	if err != nil {
		return false, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	// return the count we get back
	if rows.Next() {
		var count int64
		if err := rows.Scan(&count); err != nil {
			return false, fmt.Errorf("failed to scan database: %w", err)
		}
		return count > 0, nil
	}

	// we didn't get any rows back, so something went wrong
	return false, fmt.Errorf("no rows returned")
}

// GetValuesYaml returns the values.yaml for a chart in a repo
func (i *HelmChartIndexer) GetChartValues(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, chart Chart) ([]byte, error) {
	sqlStatement := `
SELECT valuesYaml FROM helm_charts 
WHERE name = $1 AND version = $2
AND repo_name = $3 AND repo_namespace = $4
AND cluster_name = $5 AND cluster_namespace = $6`

	rows, err := i.CacheDB.QueryContext(ctx, sqlStatement, chart.Name, chart.Version, repoRef.Name, repoRef.Namespace, clusterRef.Name, clusterRef.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	// If there are no rows, then the chart is not known
	if !rows.Next() {
		return nil, nil
	}

	// valuesYaml might be null
	var valuesYaml []byte
	if err := rows.Scan(&valuesYaml); err != nil {
		return nil, fmt.Errorf("failed to scan database: %w", err)
	}

	return valuesYaml, nil
}

// UpdateValuesYaml updates the values.yaml for a chart in a repo
func (i *HelmChartIndexer) UpdateValuesYaml(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, chart Chart, valuesYaml []byte) error {
	sqlStatement := `
UPDATE helm_charts SET valuesYaml = $1
WHERE name = $2 AND version = $3
AND repo_name = $4 AND repo_namespace = $5
AND cluster_name = $6 AND cluster_namespace = $7`

	_, err := i.CacheDB.ExecContext(ctx, sqlStatement, valuesYaml, chart.Name, chart.Version, repoRef.Name, repoRef.Namespace, clusterRef.Name, clusterRef.Namespace)
	return err
}

func (i *HelmChartIndexer) Count(ctx context.Context) (int64, error) {
	rows, err := i.CacheDB.QueryContext(ctx, "SELECT COUNT(*) FROM helm_charts")
	if err != nil {
		return 0, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		var n int64
		if err := rows.Scan(&n); err != nil {
			return 0, fmt.Errorf("failed to scan database: %w", err)
		}
		count += n
	}

	return count, nil
}

// ListChartsByCluster returns a list of charts filtered by cluster and kind (chart/profile).
func (i *HelmChartIndexer) ListChartsByCluster(ctx context.Context, clusterRef types.NamespacedName, kind string) ([]Chart, error) {
	sqlStatement := `
SELECT name, version FROM helm_charts 
WHERE cluster_name = $1 AND cluster_namespace = $2`

	if kind != "" {
		sqlStatement += " AND kind = $3"
	}

	rows, err := i.CacheDB.QueryContext(ctx, sqlStatement, clusterRef.Name, clusterRef.Namespace, kind)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var charts []Chart
	for rows.Next() {
		var chart Chart
		if err := rows.Scan(&chart.Name, &chart.Version); err != nil {
			return nil, fmt.Errorf("failed to scan database: %w", err)
		}
		charts = append(charts, chart)
	}

	return charts, nil
}

// ListChartsByRepositoryAndCluster returns a list of charts filtered by helm repository and cluster.
func (i *HelmChartIndexer) ListChartsByRepositoryAndCluster(ctx context.Context, clusterRef types.NamespacedName, repoRef ObjectReference, kind string) ([]Chart, error) {
	// select name, version and layer from the charts
	// if a kind is provided also filter by kind
	sqlStatement := `
SELECT name, version, layer FROM helm_charts
WHERE repo_name = $1 AND repo_namespace = $2
AND cluster_name = $3 AND cluster_namespace = $4`

	if kind != "" {
		sqlStatement += " AND kind = $5"
	}

	rows, err := i.CacheDB.QueryContext(ctx, sqlStatement, repoRef.Name, repoRef.Namespace, clusterRef.Name, clusterRef.Namespace, kind)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var charts []Chart
	for rows.Next() {
		var chart Chart
		// layer can be null
		var layer sql.NullString
		if err := rows.Scan(&chart.Name, &chart.Version, &layer); err != nil {
			return nil, fmt.Errorf("failed to scan database: %w", err)
		}
		if layer.Valid {
			chart.Layer = layer.String
		}
		charts = append(charts, chart)
	}

	return charts, nil
}

// GetLatestVersion returns the latest version of a chart in a repo and cluster.
func (i *HelmChartIndexer) GetLatestVersion(ctx context.Context, clusterRef, repoRef types.NamespacedName, name string) (string, error) {
	sqlStatement := `
SELECT version FROM helm_charts
WHERE name = $1
AND repo_name = $2 AND repo_namespace = $3
AND cluster_name = $4 AND cluster_namespace = $5`

	rows, err := i.CacheDB.QueryContext(ctx, sqlStatement, name, repoRef.Name, repoRef.Namespace, clusterRef.Name, clusterRef.Namespace)
	if err != nil {
		return "", fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return "", fmt.Errorf("failed to scan database: %w", err)
		}
		versions = append(versions, version)
	}

	if len(versions) == 0 {
		return "", nil
	}

	sorted, err := ReverseSemVerSort(versions)
	if err != nil {
		return "", fmt.Errorf("retrieving latest version %s: %w", name, err)
	}

	return sorted[0], nil
}

// GetLayer returns the layer of a chart in a repo and cluster.
func (i *HelmChartIndexer) GetLayer(ctx context.Context, clusterRef, repoRef types.NamespacedName, name, version string) (string, error) {
	sqlStatement := `
SELECT layer FROM helm_charts
WHERE name = $1 AND version = $2
AND repo_name = $3 AND repo_namespace = $4
AND cluster_name = $5 AND cluster_namespace = $6`

	rows, err := i.CacheDB.QueryContext(ctx, sqlStatement, name, version, repoRef.Name, repoRef.Namespace, clusterRef.Name, clusterRef.Namespace)
	if err != nil {
		return "", fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return "", nil
	} else {
		var layer sql.NullString
		if err := rows.Scan(&layer); err != nil {
			return "", fmt.Errorf("failed to scan database: %w", err)
		}
		if layer.Valid {
			return layer.String, nil
		}
		return "", nil
	}
}

func (i *HelmChartIndexer) Delete(ctx context.Context, repoRef ObjectReference, clusterRef types.NamespacedName) error {
	sqlStatement := `
DELETE FROM helm_charts
WHERE repo_name = $1 AND repo_namespace = $2
AND cluster_name = $3 AND cluster_namespace = $4`

	_, err := i.CacheDB.ExecContext(ctx, sqlStatement, repoRef.Name, repoRef.Namespace, clusterRef.Name, clusterRef.Namespace)
	return err
}

// DeleteAllChartsForCluster deletes all charts for a cluster
func (i *HelmChartIndexer) DeleteAllChartsForCluster(ctx context.Context, clusterRef types.NamespacedName) error {
	sqlStatement := `
DELETE FROM helm_charts
WHERE cluster_name = $1 AND cluster_namespace = $2`

	_, err := i.CacheDB.ExecContext(ctx, sqlStatement, clusterRef.Name, clusterRef.Namespace)
	return err
}

func applySchema(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS helm_charts (
	name text,
	version text,
	kind text,
	valuesYaml blob,
	layer text,
	repo_kind text,
	repo_api_version text,
	repo_name text,
	repo_namespace text,
	cluster_name text,
	cluster_namespace text);`)
	return err
}

func createDB(cacheLocation string) (*sql.DB, error) {
	dbFileLocation := filepath.Join(cacheLocation, dbFile)
	// make sure the directory exists
	if err := os.MkdirAll(cacheLocation, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	db, err := sql.Open("sqlite3", dbFileLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %q, %w", cacheLocation, err)
	}
	// From the readme: https://github.com/mattn/go-sqlite3
	db.SetMaxOpenConns(1)

	if err := applySchema(db); err != nil {
		return db, err
	}

	return db, nil
}

func ReverseSemVerSort(versions []string) ([]string, error) {
	vs := make([]*semver.Version, len(versions))

	for i, r := range versions {
		v, err := semver.NewVersion(r)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", r, err)
		}

		vs[i] = v
	}

	sort.Sort(sort.Reverse(semver.Collection(vs)))

	result := make([]string, len(versions))
	for i := range vs {
		result[i] = vs[i].String()
	}

	return result, nil
}
