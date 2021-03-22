package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"gorm.io/datatypes"

	"github.com/gorilla/mux"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/cmd/gitops-repo-broker/internal/common"
	"github.com/weaveworks/wks/common/database/models"
	"gorm.io/gorm"
)

var (
	ErrNilDB          = errors.New("The database has not been initialised.")
	ErrInvalidPayload = errors.New("Invalid payload")
	ErrRowUnpack      = errors.New("Error constructing response")
)

// Signature for json.MarshalIndent accepted as a method parameter for unit tests
type MarshalIndent func(v interface{}, prefix, indent string) ([]byte, error)

// Signature for json.Unmarshal accepted as a method parameter for unit test
type Unmarshal func(data []byte, v interface{}) error

// Signature for api.Generate accepted as a method parameter for unit test
type GenerateToken func() (string, error)

// DB helpers (FIXME: maybe move somewhere)

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
}

const sqliteClusterInfoTimeDifference = "strftime('%s', 'now') - strftime('%s', ci.updated_at)"
const postgresClusterInfoTimeDifference = "EXTRACT(EPOCH FROM (NOW() - ci.updated_at))"

func getClusters(db *gorm.DB, extraQuery string, extraValues ...interface{}) ([]ClusterView, error) {
	var resultOrder []string
	// The `WHERE 1=1` clause is a null condition that allows us to append more clauses to this query by concatenating them
	// without worrying about including a WHERE clause beforehand.
	queryString := `
	SELECT
		c.id AS ID,
		c.name AS Name, 
		c.token AS Token,
		c.ingress_url AS IngressURL, 
		ci.type AS Type, 
		ni.name AS NodeName, 
		ci.updated_at AS UpdatedAt,
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
		CASE
			WHEN %[1]s IS NULL OR 
				 %[1]s > 1800 THEN 'notConnected'
			WHEN %[1]s BETWEEN 60 AND 1800 THEN 'lastSeen'
			WHEN (select count(*) from alerts a where a.cluster_token = c.token and severity = 'critical') > 0 THEN 'critical'
			WHEN (select count(*) from alerts a where a.cluster_token = c.token and severity != 'none' and severity is not null) > 0 THEN 'alerting'
		ELSE 'ready'
		END AS ClusterStatus
	FROM 
		clusters c 
		LEFT JOIN cluster_info ci ON c.token = ci.cluster_token 
		LEFT JOIN node_info ni ON c.token = ni.cluster_token
		LEFT JOIN flux_info fi ON c.token = fi.cluster_token
		LEFT JOIN git_commits gc ON c.token = gc.cluster_token
		LEFT JOIN workspaces ws ON c.token = ws.cluster_token
	WHERE
		1 = 1
`

	var rows []ClusterListRow
	if db.Dialector.Name() == "postgres" {
		DB, err := db.DB()
		if err != nil {
			return []ClusterView{}, err
		}
		queryString = fmt.Sprintf(queryString, postgresClusterInfoTimeDifference)
		result, err := DB.Query(queryString + extraQuery)
		if err != nil {
			return []ClusterView{}, err
		}
		for result.Next() {
			rows = append(rows, clusterListRowScan(result))
		}
	} else {
		queryString = fmt.Sprintf(queryString, sqliteClusterInfoTimeDifference)
		if err := db.Raw(queryString + extraQuery).Scan(&rows).Error; err != nil {
			return nil, ErrNilDB
		}
	}

	clusters := map[string]*ClusterView{}
	for _, r := range rows {
		fmt.Println("cluster status in row: ", r.ClusterStatus)
		resultOrder = insertUnique(resultOrder, r.Name.String)
		if cl, ok := clusters[r.Name.String]; !ok {
			// Add new cluster with node to map
			c := ClusterView{
				ID:         r.ID,
				Name:       r.Name.String,
				Token:      r.Token.String,
				Type:       r.Type.String,
				IngressURL: r.IngressURL.String,
				UpdatedAt:  r.UpdatedAt.Time,
				Status:     r.ClusterStatus,
			}
			unpackClusterRow(&c, r)
			clusters[r.Name.String] = &c
		} else {
			unpackClusterRow(cl, r)
		}
	}

	clusterList := []ClusterView{}
	for _, c := range resultOrder {
		clusterItem, ok := clusters[c]
		if !ok {
			return nil, ErrRowUnpack
		}
		clusterList = append(clusterList, *clusterItem)
	}

	return clusterList, nil
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
		&row.NodeName,
		&row.UpdatedAt,
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
		&row.ClusterStatus)
	if err != nil {
		log.Debug("error while scanning sql row: ", err)
	}
	return row
}

func unpackClusterRow(c *ClusterView, r ClusterListRow) {
	// Do not add nodes if they don't exist yet
	if r.NodeName.Valid && !nodeExists(*c, NodeView{
		Name:           r.NodeName.String,
		IsControlPlane: r.IsControlPlane.Bool,
		KubeletVersion: r.KubeletVersion.String,
	}) {
		c.Nodes = append(c.Nodes, NodeView{
			Name:           r.NodeName.String,
			IsControlPlane: r.IsControlPlane.Bool,
			KubeletVersion: r.KubeletVersion.String,
		})
	}

	// Append flux info for the cluster
	var fluxLogInfo datatypes.JSON
	if r.FluxLogInfo != nil {
		fluxLogInfo = *r.FluxLogInfo
	}
	if r.FluxName.Valid && !fluxInfoExists(*c, FluxInfoView{
		r.FluxName.String,
		r.FluxNamespace.String,
		r.FluxRepoBranch.String,
		r.FluxRepoURL.String,
		fluxLogInfo,
	}) {
		c.FluxInfo = append(c.FluxInfo, FluxInfoView{
			Name:       r.FluxName.String,
			Namespace:  r.FluxNamespace.String,
			RepoURL:    r.FluxRepoURL.String,
			RepoBranch: r.FluxRepoBranch.String,
			LogInfo:    fluxLogInfo,
		})
	}

	if r.GitCommitSha.Valid && !gitCommitExists(*c, GitCommitView{
		Sha:         r.GitCommitSha.String,
		AuthorName:  r.GitCommitAuthorName.String,
		AuthorEmail: r.GitCommitAuthorEmail.String,
		AuthorDate:  r.GitCommitAuthorDate,
		Message:     r.GitCommitMessage.String,
	}) {
		c.GitCommits = append(c.GitCommits, GitCommitView{
			Sha:         r.GitCommitSha.String,
			AuthorName:  r.GitCommitAuthorName.String,
			AuthorEmail: r.GitCommitAuthorEmail.String,
			AuthorDate:  r.GitCommitAuthorDate,
			Message:     r.GitCommitMessage.String,
		})
	}

	var wsView WorkspaceView
	if r.WorkspaceName.Valid && r.WorkspaceNamespace.Valid {
		wsView = WorkspaceView{
			Name:      r.WorkspaceName.String,
			Namespace: r.WorkspaceNamespace.String,
		}
	}

	if r.WorkspaceName.Valid && !workspaceExists(*c, wsView) {
		c.Workspaces = append(c.Workspaces, wsView)
	}
}

func getCluster(db *gorm.DB, id uint) (*ClusterView, error) {
	clusters, err := getClusters(db, fmt.Sprintf(" AND c.id = %d", id))
	if err != nil {
		return nil, err
	}
	if len(clusters) == 0 {
		return nil, nil
	}
	return &clusters[0], nil
}

// Handlers

func FindCluster(db *gorm.DB, marshalIndentFn MarshalIndent) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		cluster, err := getClusterFromRequest(r, db)
		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}
		if cluster == nil {
			common.WriteError(w, fmt.Errorf("cluster not found"), http.StatusNotFound)
			return
		}

		respondWithJSON(w, http.StatusOK, cluster, marshalIndentFn)
	}
}

func ListClusters(db *gorm.DB, marshalIndentFn MarshalIndent) func(w http.ResponseWriter, r *http.Request) {
	extraQuery := ""
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		// Read sort-by-column from the url string if provided, otherwise default to sorting by cluster name
		sortColumn := "Name"
		sortByParam, ok := r.URL.Query()["sortBy"]
		if ok {
			log.Debugf("sorting by column: %s\n", sortByParam)
			sortColumn = sortByParam[0]
		}

		// Read sort order from the url string if provided, otherwise sort desc
		sortOrder := "ASC"
		sortOrderParam, ok := r.URL.Query()["order"]
		if ok {
			log.Debugf("sorting by order: %s\n", sortOrderParam)
			if sortOrderParam[0] == "ASC" || sortOrderParam[0] == "DESC" {
				sortOrder = sortOrderParam[0]
			}
		}

		extraQuery = fmt.Sprintf(" ORDER BY %s %s", sortColumn, sortOrder)

		clusters, err := getClusters(db, extraQuery)
		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}

		res := ClustersResponse{
			Clusters: clusters,
		}
		data, err := marshalIndentFn(res, "", " ")
		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(data))
	}
}

func RegisterCluster(db *gorm.DB, validate *validator.Validate, unmarshalFn Unmarshal, marshalFn MarshalIndent, generateTokenFn GenerateToken) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			common.WriteError(w, err, http.StatusBadRequest)
			return
		}

		crr := &ClusterRegistrationRequest{}
		if err := unmarshalFn(reqBody, crr); err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}

		err = validate.Struct(crr)
		if err != nil {
			log.Errorf("Failed to validate payload: %v", err)
			common.WriteError(w, ErrInvalidPayload, http.StatusBadRequest)
			return
		}

		t, err := generateTokenFn()
		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}

		c := &models.Cluster{
			Name:       crr.Name,
			IngressURL: crr.IngressURL,
			Token:      t,
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(c).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}

		res := ClusterRegistrationResponse{
			ID:         c.ID,
			Name:       c.Name,
			IngressURL: c.IngressURL,
			Token:      c.Token,
		}
		output, err := marshalFn(res, "", " ")
		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", output)
	}
}

func UpdateCluster(db *gorm.DB, unmarshalFn Unmarshal, marshalFn MarshalIndent) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			cluster, err := getClusterFromRequest(r, tx)
			if err != nil {
				return err
			}
			if cluster == nil {
				common.WriteError(w, fmt.Errorf("cluster not found"), http.StatusNotFound)
				return nil
			}

			reqBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				common.WriteError(w, err, http.StatusBadRequest)
				return nil
			}

			clusterUpdates := &models.Cluster{}
			if err := unmarshalFn(reqBody, clusterUpdates); err != nil {
				return err
			}

			c := models.Cluster{}
			tx.First(&c, cluster.ID)
			c.IngressURL = clusterUpdates.IngressURL
			c.Name = clusterUpdates.Name
			// This particular variant of `Updates` (with a struct), does non-zero checking
			// and so don't update the Name or IngressURL if they are "zero"
			// https://gorm.io/docs/update.html#Update-Selected-Fields
			result := tx.Model(c).Updates(c)
			if result.Error != nil {
				common.WriteError(w, result.Error, http.StatusBadRequest)
				return nil
			}

			clusterView, err := getCluster(tx, c.ID)
			if err != nil {
				return err
			}
			respondWithJSON(w, http.StatusOK, clusterView, marshalFn)
			return nil
		})

		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
		}
	}
}

func ListAlerts(db *gorm.DB, marshalIndentFn MarshalIndent) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		var rows []AlertsClusterRow
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Raw(`
			SELECT
				a.id,
				a.annotations,
				a.ends_at,
				a.fingerprint,
				a.inhibited_by,
				a.silenced_by,
				a.severity,
				a.state,
				a.starts_at,
				a.updated_at,
				a.generator_url,
				a.labels,
				c.id as ClusterID,
				c.name as ClusterName,
				c.ingress_url as ClusterIngressURL
			FROM
				alerts a,
				clusters c
			WHERE
				a.cluster_token = c.token and
				a.severity != 'none' and a.severity is not null
			ORDER BY
				a.severity
			`).Scan(&rows).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			log.Errorf("Failed to query for alerts: %v", err)
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		res, err := toAlertResponse(rows)
		if err != nil {
			log.Errorf("Failed to generate response for alerts: %v", err)
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		respondWithJSON(w, http.StatusOK, res, marshalIndentFn)
	}
}

func UnregisterCluster(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		id, err := getClusterIDFromRequest(r)
		if err != nil {
			log.Errorf("Parameter 'id' is not a uint: %v", err)
			common.WriteError(w, err, http.StatusBadRequest)
			return
		}

		cluster, err := getCluster(db, uint(id))
		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			log.Errorf("Failed to load cluster(%d): %v", id, err)
			return
		}

		if cluster == nil {
			w.WriteHeader(http.StatusNotFound)
			log.Errorf("Cluster(%d) was not found", id)
			return
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			dependentObjectsToDelete := []interface{}{
				&models.Event{},
				&models.NodeInfo{},
				&models.ClusterInfo{},
				&models.Alert{},
				&models.FluxInfo{},
				&models.GitCommit{},
				&models.Workspace{},
			}

			for _, o := range dependentObjectsToDelete {
				if err := tx.Where("cluster_token = ?", cluster.Token).Delete(o).Error; err != nil {
					return fmt.Errorf("failed to delete %T records when unregistering Cluster %q: %w", o, cluster.Token, err)
				}
			}

			if err := tx.Delete(&models.Cluster{}, id).Error; err != nil {
				return fmt.Errorf("failed to delete Cluster(%d) record when unregistering Cluster %q: %w", id, cluster.Token, err)
			}

			return nil
		})
		if err != nil {
			log.Errorf("Failed to unregister Cluster(%d): %v", id, err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// types

type ClusterRegistrationRequest struct {
	Name       string `json:"name" validate:"required"`
	IngressURL string `json:"ingressUrl" validate:"omitempty,url"`
}

type ClusterRegistrationResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	IngressURL string `json:"ingressUrl"`
	Token      string `json:"token"`
}

type NodeView struct {
	Name           string `json:"name"`
	IsControlPlane bool   `json:"isControlPlane"`
	KubeletVersion string `json:"kubeletVersion"`
}

type ClusterView struct {
	ID         uint            `json:"id"`
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Token      string          `json:"token"`
	IngressURL string          `json:"ingressUrl"`
	Nodes      []NodeView      `json:"nodes,omitempty"`
	Status     string          `json:"status"`
	UpdatedAt  time.Time       `json:"updatedAt"`
	FluxInfo   []FluxInfoView  `json:"fluxInfo,omitempty"`
	GitCommits []GitCommitView `json:"gitCommits,omitempty"`
	Workspaces []WorkspaceView `json:"workspaces,omitempty"`
}

type FluxInfoView struct {
	Name       string         `json:"name"`
	Namespace  string         `json:"namespace"`
	RepoURL    string         `json:"repoUrl"`
	RepoBranch string         `json:"repoBranch"`
	LogInfo    datatypes.JSON `json:"logInfo"`
}

type ClustersResponse struct {
	Clusters []ClusterView `json:"clusters"`
}

type AlertView struct {
	ID          uint                   `json:"id"`
	Fingerprint string                 `json:"fingerprint"`
	State       string                 `json:"state"`
	Severity    string                 `json:"severity"`
	InhibitedBy string                 `json:"inhibited_by"`
	SilencedBy  string                 `json:"silenced_by"`
	Annotations map[string]interface{} `json:"annotations"`
	Labels      map[string]interface{} `json:"labels"`
	StartsAt    time.Time              `json:"starts_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	EndsAt      time.Time              `json:"ends_at"`
	Cluster     ClusterView            `json:"cluster"`
}

type AlertsResponse struct {
	Alerts []AlertView `json:"alerts"`
}

type AlertsClusterRow struct {
	ID                uint
	Fingerprint       string
	State             string
	Severity          string
	InhibitedBy       string
	SilencedBy        string
	Annotations       datatypes.JSON
	Labels            datatypes.JSON
	StartsAt          time.Time
	UpdatedAt         time.Time
	EndsAt            time.Time
	ClusterID         uint
	ClusterName       string
	ClusterIngressURL string
}

type GitCommitView struct {
	Sha         string       `json:"sha"`
	AuthorName  string       `json:"author_name"`
	AuthorEmail string       `json:"author_email"`
	AuthorDate  sql.NullTime `json:"author_date"`
	Message     string       `json:"message"`
}

type WorkspaceView struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// helpers
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}, marshalIndentFn MarshalIndent) {
	response, err := marshalIndentFn(payload, "", " ")
	if err != nil {
		common.WriteError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func getClusterFromRequest(r *http.Request, db *gorm.DB) (*ClusterView, error) {
	id, err := getClusterIDFromRequest(r)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse id: %v", err)
	}

	return getCluster(db, uint(id))
}

func getClusterIDFromRequest(r *http.Request) (uint64, error) {
	vars := mux.Vars(r)
	idString := vars["id"]
	return strconv.ParseUint(idString, 10, 64)
}

func nodeExists(c ClusterView, n NodeView) bool {
	for _, existingNode := range c.Nodes {
		if existingNode.Name == n.Name {
			return true
		}
	}
	return false
}

func fluxInfoExists(c ClusterView, fi FluxInfoView) bool {
	for _, existingFluxInfo := range c.FluxInfo {
		if existingFluxInfo.Name == fi.Name && existingFluxInfo.Namespace == fi.Namespace {
			return true
		}
	}
	return false
}

func toAlertResponse(rows []AlertsClusterRow) (*AlertsResponse, error) {

	res := AlertsResponse{
		Alerts: []AlertView{},
	}

	for _, r := range rows {
		var labels map[string]interface{}
		var annotations map[string]interface{}
		if err := json.Unmarshal(r.Labels, &labels); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(r.Annotations, &annotations); err != nil {
			return nil, err
		}
		res.Alerts = append(res.Alerts, AlertView{
			ID:          r.ID,
			Fingerprint: r.Fingerprint,
			State:       r.State,
			Severity:    r.Severity,
			InhibitedBy: r.InhibitedBy,
			SilencedBy:  r.SilencedBy,
			Annotations: annotations,
			Labels:      labels,
			StartsAt:    r.StartsAt,
			UpdatedAt:   r.UpdatedAt,
			EndsAt:      r.EndsAt,
			Cluster: ClusterView{
				ID:         r.ClusterID,
				Name:       r.ClusterName,
				IngressURL: r.ClusterIngressURL,
			},
		})
	}

	return &res, nil
}

func gitCommitExists(c ClusterView, gc GitCommitView) bool {
	for _, existingGitCommitInfo := range c.GitCommits {
		if existingGitCommitInfo.Sha == gc.Sha {
			return true
		}
	}
	return false
}

func workspaceExists(c ClusterView, ws WorkspaceView) bool {
	for _, existingWorkspace := range c.Workspaces {
		if existingWorkspace.Name == ws.Name && existingWorkspace.Namespace == ws.Namespace {
			return true
		}
	}
	return false
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
