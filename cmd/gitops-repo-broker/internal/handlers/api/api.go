package api

import (
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
	Name           string
	Token          string
	IngressURL     string
	Type           string
	NodeName       string
	IsControlPlane bool
	KubeletVersion string

	// for alerts
	CriticalAlertsCount uint
	AlertsCount         uint
	UpdatedAt           time.Time

	// for flux info
	FluxName       string
	FluxNamespace  string
	FluxRepoURL    string
	FluxRepoBranch string

	// Git commit
	GitCommitAuthorName  string
	GitCommitAuthorEmail string
	GitCommitAuthorDate  time.Time
	GitCommitMessage     string
	GitCommitSha         string
}

func getClusters(db *gorm.DB, extraQuery string, extraValues ...interface{}) ([]ClusterView, error) {

	var rows []ClusterListRow
	if err := db.Raw(`
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
				(select count(*) from alerts a where a.token = c.token and severity = 'critical') as CriticalAlertsCount,
				(select count(*) from alerts a where a.token = c.token and severity != 'none' and severity is not null) AS AlertsCount,
				gc.author_name AS GitCommitAuthorName,
				gc.author_email AS GitCommitAuthorEmail,
				gc.author_date AS GitCommitAuthorDate,
				gc.message AS GitCommitMessage,
				gc.sha AS GitCommitSha
			FROM 
				clusters c 
				LEFT JOIN cluster_info ci ON c.token = ci.token 
				LEFT JOIN node_info ni ON c.token = ni.token
				LEFT JOIN flux_info fi ON c.token = fi.cluster_token
				LEFT JOIN git_commits gc ON c.token = gc.cluster_token
			WHERE c.deleted_at IS NULL
		`+extraQuery, extraValues).Scan(&rows).Error; err != nil {
		return nil, ErrNilDB
	}

	clusters := map[string]*ClusterView{}
	for _, r := range rows {
		if cl, ok := clusters[r.Name]; !ok {
			// Add new cluster with node to map
			c := ClusterView{
				ID:         r.ID,
				Name:       r.Name,
				Token:      r.Token,
				Type:       r.Type,
				IngressURL: r.IngressURL,
				UpdatedAt:  r.UpdatedAt,
				Status:     getClusterStatus(r),
			}
			unpackClusterRow(&c, r)
			clusters[r.Name] = &c
		} else {
			unpackClusterRow(cl, r)
		}
	}

	clusterList := []ClusterView{}
	for _, c := range clusters {
		clusterList = append(clusterList, *c)
	}

	return clusterList, nil
}

func unpackClusterRow(c *ClusterView, r ClusterListRow) {
	// Do not add nodes if they don't exist yet
	if r.NodeName != "" && !nodeExists(*c, NodeView{
		Name:           r.NodeName,
		IsControlPlane: r.IsControlPlane,
		KubeletVersion: r.KubeletVersion,
	}) {
		c.Nodes = append(c.Nodes, NodeView{
			Name:           r.NodeName,
			IsControlPlane: r.IsControlPlane,
			KubeletVersion: r.KubeletVersion,
		})
	}

	// Append flux info for the cluster
	if r.FluxName != "" && !fluxInfoExists(*c, FluxInfoView{
		r.FluxName,
		r.FluxNamespace,
		r.FluxRepoBranch,
		r.FluxRepoURL,
	}) {
		c.FluxInfo = append(c.FluxInfo, FluxInfoView{
			Name:       r.FluxName,
			Namespace:  r.FluxNamespace,
			RepoURL:    r.FluxRepoURL,
			RepoBranch: r.FluxRepoBranch,
		})
	}

	if r.GitCommitSha != "" && !gitCommitExists(*c, GitCommitView{
		Sha:         r.GitCommitSha,
		AuthorName:  r.GitCommitAuthorName,
		AuthorEmail: r.GitCommitAuthorEmail,
		AuthorDate:  r.GitCommitAuthorDate,
		Message:     r.GitCommitMessage,
	}) {
		c.GitCommits = append(c.GitCommits, GitCommitView{
			Sha:         r.GitCommitSha,
			AuthorName:  r.GitCommitAuthorName,
			AuthorEmail: r.GitCommitAuthorEmail,
			AuthorDate:  r.GitCommitAuthorDate,
			Message:     r.GitCommitMessage,
		})
	}
}

func getCluster(db *gorm.DB, id uint) (*ClusterView, error) {
	clusters, err := getClusters(db, " AND c.id = ?", id)
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
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		clusters, err := getClusters(db, "")
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
				a.token = c.token and
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
	GitCommits []GitCommitView `json:"git_commits,omitempty"`
}

type FluxInfoView struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	RepoURL    string `json:"repoUrl"`
	RepoBranch string `json:"repoBranch"`
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
	Sha         string    `json:"sha"`
	AuthorName  string    `json:"author_name"`
	AuthorEmail string    `json:"author_email"`
	AuthorDate  time.Time `json:"author_date"`
	Message     string    `json:"message"`
}

// helpers

func getClusterStatus(c ClusterListRow) string {
	// Connection status first, if its gone away alerts might be stale.
	timeNow := time.Now()
	diff := timeNow.Sub(c.UpdatedAt)
	intDiff := int64(diff / time.Minute)

	if intDiff > 1 && intDiff < 30 {
		return "lastSeen"
	}
	if intDiff > 30 {
		return "notConnected"
	}

	if c.CriticalAlertsCount > 0 {
		return "critical"
	}
	if c.AlertsCount > 0 {
		return "alerting"
	}

	return "ready"
}

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
	vars := mux.Vars(r)
	idString := vars["id"]
	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse id %v: %v", idString, err)
	}

	return getCluster(db, uint(id))
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
