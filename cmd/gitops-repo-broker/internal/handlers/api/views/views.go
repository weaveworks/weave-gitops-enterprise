package views

import (
	"database/sql"
	"time"

	"gorm.io/datatypes"
)

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

type ClusterUpdateRequest struct {
	Name       string `json:"name" validate:"required"`
	IngressURL string `json:"ingressUrl" validate:"omitempty,url"`
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
