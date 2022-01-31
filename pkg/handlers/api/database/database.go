package database

import (
	"database/sql"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/handlers/api/views"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrNilDB     = errors.New("The database has not been initialised.")
	ErrRowUnpack = errors.New("Error constructing response")
)

const sqliteClusterInfoTimeDifference = "strftime('%s', 'now') - strftime('%s', ci.updated_at)"
const postgresClusterInfoTimeDifference = "EXTRACT(EPOCH FROM (NOW() - ci.updated_at))"
const clustersQueryString = `
WITH unique_capi_clusters AS (
	SELECT
		MIN(id) as id,
		name,
		namespace
	FROM
		capi_clusters
	GROUP BY
		name,
		namespace
), clusters_with_status AS (
	SELECT
		c.id AS ID,
		c.name AS Name,
		c.token AS Token,
		c.ingress_url AS IngressURL,
		c.capi_name AS capi_name,
		c.capi_namespace AS capi_namespace,
		ci.type AS Type,
		ci.updated_at AS UpdatedAt,
		CASE
			WHEN %[1]s BETWEEN 60 AND 1800 THEN 'lastSeen'
			WHEN (select count(*) from alerts a where a.cluster_token = c.token and severity = 'critical') > 0 THEN 'critical'
			WHEN (select count(*) from alerts a where a.cluster_token = c.token and severity != 'none' and severity is not null) > 0 THEN 'alerting'
			WHEN  %[1]s <= 1800 THEN 'ready'
			WHEN (select count(*) from pull_requests pr inner join cluster_pull_requests cpr on pr.id = cpr.pull_request_id where cpr.cluster_id = c.id and pr.type = 'delete') > 0 and ci.updated_at is null THEN 'pullRequestCreated'
			WHEN cc.id is not null THEN 'clusterFound'
			WHEN (select count(*) from pull_requests pr inner join cluster_pull_requests cpr on pr.id = cpr.pull_request_id where cpr.cluster_id = c.id) > 0 and ci.updated_at is null THEN 'pullRequestCreated'
			WHEN %[1]s IS NULL OR %[1]s > 1800 THEN 'notConnected'
		END AS ClusterStatus
	FROM
		clusters c
		LEFT JOIN cluster_info ci ON c.token = ci.cluster_token
		LEFT JOIN unique_capi_clusters cc ON c.capi_name = cc.name and c.capi_namespace = cc.namespace
), paginated_ordered_clusters AS (
	SELECT
		c.*
	FROM
		clusters_with_status AS c
		LEFT JOIN cluster_statuses cs ON c.ClusterStatus = cs.Status
	%[2]s
	%[3]s
	%[4]s
)
SELECT
	c.ID,
	c.Name,
	c.Token,
	c.IngressURL,
	c.Type,
	c.UpdatedAt,
	c.ClusterStatus,
	ni.name AS NodeName,
	ni.is_control_plane AS IsControlPlane,
	ni.kubelet_version AS KubeletVersion,
	fi.name AS FluxName,
	fi.namespace AS FluxNamespace,
	fi.repo_url AS FluxRepoURL,
	fi.repo_branch AS FluxRepoBranch,
	fi.Syncs AS FluxLogInfo,
	(select count(*) from alerts a where a.cluster_token = c.token and severity = 'critical') as CriticalAlertsCount,
	(select count(*) from alerts a where a.cluster_token = c.token and severity != 'none' and severity is not null) AS AlertsCount,
	gc.author_name AS GitCommitAuthorName,
	gc.author_email AS GitCommitAuthorEmail,
	gc.author_date AS GitCommitAuthorDate,
	gc.message AS GitCommitMessage,
	gc.sha AS GitCommitSha,
	ws.name AS WorkspaceName,
	ws.namespace AS WorkspaceNamespace,
	cc.name as CapiName,
	cc.namespace as CapiNamespace,
	cc.object as CapiCluster,
	pr.url as PullRequestURL,
	pr.type as PullRequestType
FROM
	paginated_ordered_clusters AS c
	LEFT JOIN node_info ni ON c.token = ni.cluster_token
	LEFT JOIN cluster_pull_requests cpr ON cpr.cluster_id = c.ID
	LEFT JOIN pull_requests pr ON pr.ID = cpr.pull_request_id
	LEFT JOIN flux_info fi ON c.token = fi.cluster_token
	LEFT JOIN git_commits gc ON c.token = gc.cluster_token
	LEFT JOIN workspaces ws ON c.token = ws.cluster_token
	LEFT JOIN capi_clusters cc ON c.capi_name = cc.name and c.capi_namespace = cc.namespace
	LEFT JOIN cluster_statuses cs ON c.ClusterStatus = cs.Status
%[3]s
`

type ClusterListRow struct {
	ID             uint
	Name           sql.NullString
	Token          sql.NullString
	IngressURL     sql.NullString
	Type           sql.NullString
	NodeName       sql.NullString
	UpdatedAt      sql.NullTime
	IsControlPlane sql.NullBool
	KubeletVersion sql.NullString

	// for alerts
	CriticalAlertsCount uint
	AlertsCount         uint

	// for flux info
	FluxName       sql.NullString
	FluxNamespace  sql.NullString
	FluxRepoURL    sql.NullString
	FluxRepoBranch sql.NullString
	FluxLogInfo    *datatypes.JSON

	// Git commit
	GitCommitAuthorName  sql.NullString
	GitCommitAuthorEmail sql.NullString
	GitCommitAuthorDate  sql.NullTime
	GitCommitMessage     sql.NullString
	GitCommitSha         sql.NullString

	// Workspace
	WorkspaceName      sql.NullString
	WorkspaceNamespace sql.NullString

	ClusterStatus string

	CapiName        sql.NullString
	CapiNamespace   sql.NullString
	CapiCluster     *datatypes.JSON
	PullRequestURL  sql.NullString
	PullRequestType sql.NullString
}

type Pagination struct {
	Page int
	Size int
}

type GetClustersRequest struct {
	SortColumn string
	SortOrder  string
	ID         uint
	Pagination
}

type GetClustersResponse struct {
	Clusters []views.ClusterView
	Total    int64
}

func GetClusters(db *gorm.DB, req GetClustersRequest) (*GetClustersResponse, error) {
	if db == nil {
		return nil, ErrNilDB
	}

	// Sort by cluster name by default
	if req.SortColumn == "" {
		req.SortColumn = "Name"
	}
	// Sort in ascending order by default
	if req.SortOrder == "" {
		req.SortOrder = "ASC"
	}

	whereClause := buildWhereClause(req)

	orderByClause := buildOrderByClause(req)

	limitOffsetClause := buildLimitOffsetClause(req)

	var resultOrder []string

	queryString := clustersQueryString

	var rows []ClusterListRow
	if db.Dialector.Name() == "postgres" {
		DB, err := db.DB()
		if err != nil {
			return nil, err
		}
		queryString = fmt.Sprintf(queryString, postgresClusterInfoTimeDifference, whereClause, orderByClause, limitOffsetClause)
		log.Debugf("raw query: %s\n", queryString)
		result, err := DB.Query(queryString)
		if err != nil {
			return nil, err
		}
		for result.Next() {
			rows = append(rows, clusterListRowScan(result))
		}
	} else {
		queryString = fmt.Sprintf(queryString, sqliteClusterInfoTimeDifference, whereClause, orderByClause, limitOffsetClause)
		if err := db.Raw(queryString).Scan(&rows).Error; err != nil {
			log.Debugf("Failed to execute query %q: %v", queryString, err)
			return nil, ErrNilDB
		}
	}

	clusters := map[string]*views.ClusterView{}
	for _, r := range rows {
		resultOrder = insertUnique(resultOrder, r.Name.String)
		if cl, ok := clusters[r.Name.String]; !ok {
			// Add new cluster with node to map
			var capiCluster datatypes.JSON
			if r.CapiCluster != nil {
				capiCluster = *r.CapiCluster
			}
			c := views.ClusterView{
				ID:            r.ID,
				Name:          r.Name.String,
				Token:         r.Token.String,
				Type:          r.Type.String,
				IngressURL:    r.IngressURL.String,
				UpdatedAt:     r.UpdatedAt.Time,
				Status:        r.ClusterStatus,
				CAPIName:      r.CapiName.String,
				CAPINamespace: r.CapiNamespace.String,
				CAPICluster:   capiCluster,
			}
			unpackClusterRow(&c, r)
			clusters[r.Name.String] = &c
		} else {
			unpackClusterRow(cl, r)
		}
	}

	clusterList := []views.ClusterView{}
	for _, c := range resultOrder {
		clusterItem, ok := clusters[c]
		if !ok {
			return nil, ErrRowUnpack
		}
		clusterList = append(clusterList, *clusterItem)
	}

	response := &GetClustersResponse{}
	var total int64
	if err := db.Model(&models.Cluster{}).Count(&total).Error; err == nil {
		response.Total = total
	} else {
		log.Errorf("Unable to query for total number of records: %v", err)
	}
	response.Clusters = clusterList

	return response, nil
}

func GetCluster(db *gorm.DB, id uint) (*views.ClusterView, error) {
	response, err := GetClusters(db, GetClustersRequest{
		ID: id,
	})
	if err != nil {
		return nil, err
	}

	if len(response.Clusters) > 0 {
		return &response.Clusters[0], nil
	}
	return nil, nil
}

func clusterListRowScan(sqlResult *sql.Rows) ClusterListRow {
	var row ClusterListRow
	cols, _ := sqlResult.Columns()
	log.Debugf("sqlResult: %+v\n", cols)
	err := sqlResult.Scan(
		&row.ID,
		&row.Name,
		&row.Token,
		&row.IngressURL,
		&row.Type,
		&row.UpdatedAt,
		&row.ClusterStatus,
		&row.NodeName,
		&row.IsControlPlane,
		&row.KubeletVersion,
		&row.FluxName,
		&row.FluxNamespace,
		&row.FluxRepoURL,
		&row.FluxRepoBranch,
		&row.FluxLogInfo,
		&row.CriticalAlertsCount,
		&row.AlertsCount,
		&row.GitCommitAuthorName,
		&row.GitCommitAuthorEmail,
		&row.GitCommitAuthorDate,
		&row.GitCommitMessage,
		&row.GitCommitSha,
		&row.WorkspaceName,
		&row.WorkspaceNamespace,

		&row.CapiName,
		&row.CapiNamespace,
		&row.CapiCluster,
		&row.PullRequestURL,
		&row.PullRequestType,
	)
	if err != nil {
		log.Debug("error while scanning sql row: ", err)
	}
	return row
}

func AlertListRowScan(sqlResult *sql.Rows) views.AlertsClusterRow {
	var row views.AlertsClusterRow
	cols, _ := sqlResult.Columns()
	log.Debugf("sqlResult: %+v\n", cols)
	err := sqlResult.Scan(
		&row.ID,
		&row.Annotations,
		&row.EndsAt,
		&row.Fingerprint,
		&row.InhibitedBy,
		&row.SilencedBy,
		&row.Severity,
		&row.State,
		&row.StartsAt,
		&row.UpdatedAt,
		&row.Labels,
		&row.ClusterID,
		&row.ClusterName,
		&row.ClusterIngressURL)
	if err != nil {
		log.Debug("error while scanning sql row: ", err)
	}
	return row
}

// Utility function to keep order of SQL result while unpacking result rows
func insertUnique(resultOrder []string, name string) []string {
	for _, existingName := range resultOrder {
		if existingName == name {
			return resultOrder
		}
	}
	resultOrder = append(resultOrder, name)
	return resultOrder
}

func unpackClusterRow(c *views.ClusterView, r ClusterListRow) {
	// Do not add nodes if they don't exist yet
	if r.NodeName.Valid && !nodeExists(*c, views.NodeView{
		Name:           r.NodeName.String,
		IsControlPlane: r.IsControlPlane.Bool,
		KubeletVersion: r.KubeletVersion.String,
	}) {
		c.Nodes = append(c.Nodes, views.NodeView{
			Name:           r.NodeName.String,
			IsControlPlane: r.IsControlPlane.Bool,
			KubeletVersion: r.KubeletVersion.String,
		})
	}

	if r.PullRequestURL.Valid {
		if c.PullRequest == nil || r.PullRequestType.String == "delete" {
			c.PullRequest = &views.PullRequestView{URL: r.PullRequestURL.String, Type: r.PullRequestType.String}
		}
	}

	// Append flux info for the cluster
	var fluxLogInfo datatypes.JSON
	if r.FluxLogInfo != nil {
		fluxLogInfo = *r.FluxLogInfo
	}
	if r.FluxName.Valid && !fluxInfoExists(*c, views.FluxInfoView{
		Name:       r.FluxName.String,
		Namespace:  r.FluxNamespace.String,
		RepoURL:    r.FluxRepoURL.String,
		RepoBranch: r.FluxRepoBranch.String,
		LogInfo:    fluxLogInfo,
	}) {
		c.FluxInfo = append(c.FluxInfo, views.FluxInfoView{
			Name:       r.FluxName.String,
			Namespace:  r.FluxNamespace.String,
			RepoURL:    r.FluxRepoURL.String,
			RepoBranch: r.FluxRepoBranch.String,
			LogInfo:    fluxLogInfo,
		})
	}

	if r.GitCommitSha.Valid && !gitCommitExists(*c, views.GitCommitView{
		Sha:         r.GitCommitSha.String,
		AuthorName:  r.GitCommitAuthorName.String,
		AuthorEmail: r.GitCommitAuthorEmail.String,
		AuthorDate:  r.GitCommitAuthorDate,
		Message:     r.GitCommitMessage.String,
	}) {
		c.GitCommits = append(c.GitCommits, views.GitCommitView{
			Sha:         r.GitCommitSha.String,
			AuthorName:  r.GitCommitAuthorName.String,
			AuthorEmail: r.GitCommitAuthorEmail.String,
			AuthorDate:  r.GitCommitAuthorDate,
			Message:     r.GitCommitMessage.String,
		})
	}

	var wsView views.WorkspaceView
	if r.WorkspaceName.Valid && r.WorkspaceNamespace.Valid {
		wsView = views.WorkspaceView{
			Name:      r.WorkspaceName.String,
			Namespace: r.WorkspaceNamespace.String,
		}
	}

	if r.WorkspaceName.Valid && !workspaceExists(*c, wsView) {
		c.Workspaces = append(c.Workspaces, wsView)
	}
}

func nodeExists(c views.ClusterView, n views.NodeView) bool {
	for _, existingNode := range c.Nodes {
		if existingNode.Name == n.Name {
			return true
		}
	}
	return false
}

func fluxInfoExists(c views.ClusterView, fi views.FluxInfoView) bool {
	for _, existingFluxInfo := range c.FluxInfo {
		if existingFluxInfo.Name == fi.Name && existingFluxInfo.Namespace == fi.Namespace {
			return true
		}
	}
	return false
}

func gitCommitExists(c views.ClusterView, gc views.GitCommitView) bool {
	for _, existingGitCommitInfo := range c.GitCommits {
		if existingGitCommitInfo.Sha == gc.Sha {
			return true
		}
	}
	return false
}

func workspaceExists(c views.ClusterView, ws views.WorkspaceView) bool {
	for _, existingWorkspace := range c.Workspaces {
		if existingWorkspace.Name == ws.Name && existingWorkspace.Namespace == ws.Namespace {
			return true
		}
	}
	return false
}

func buildWhereClause(req GetClustersRequest) string {
	if req.ID == 0 {
		return ""
	}

	return fmt.Sprintf("WHERE c.ID = %d", req.ID)
}

func buildOrderByClause(req GetClustersRequest) string {
	if req.ID != 0 {
		return ""
	}

	orderByClause := fmt.Sprintf(`ORDER BY c.%s %s, c.Name ASC`, req.SortColumn, req.SortOrder)
	if req.SortColumn == "ClusterStatus" {
		orderByClause = fmt.Sprintf(`ORDER BY cs.ID %s, c.Name ASC`, req.SortOrder)
	}

	return orderByClause
}

func buildLimitOffsetClause(req GetClustersRequest) string {
	if req.ID != 0 {
		return ""
	}

	var limit, page int
	if req.Size <= 10 {
		limit = 10
	} else if req.Size >= 100 {
		limit = 100
	} else {
		limit = req.Size
	}

	if req.Page <= 1 {
		page = 1
	} else {
		page = req.Page
	}

	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, (page-1)*limit)
}
