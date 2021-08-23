package models

import (
	"time"

	"gorm.io/datatypes"
	"k8s.io/apimachinery/pkg/types"
)

// Embedded struct similar to gorm.Model but excluding DeletedAt field since we do hard deletes.
type Model struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Event table, ref: https://godoc.org/k8s.io/api/core/v1#Event
type Event struct {
	UID          types.UID `gorm:"primaryKey"`
	ClusterToken string
	CreatedAt    datatypes.Date
	RegisteredAt datatypes.Date
	Name         string
	Namespace    string
	Labels       string
	Annotations  string
	ClusterName  string
	Reason       string
	Message      string
	Type         string
	RawEvent     datatypes.JSON
}

// ClusterInfo table
type ClusterInfo struct {
	UID          types.UID `gorm:"primaryKey"`
	ClusterToken string
	Type         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (ClusterInfo) TableName() string {
	return "cluster_info"
}

// NodeInfo table
type NodeInfo struct {
	ID             uint `gorm:"primaryKey"`
	UID            types.UID
	ClusterToken   string
	ClusterInfoUID types.UID
	Name           string
	IsControlPlane bool
	KubeletVersion string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (NodeInfo) TableName() string {
	return "node_info"
}

// Alert Table
type Alert struct {
	ID           uint `gorm:"primaryKey"`
	ClusterToken string
	Annotations  datatypes.JSON
	EndsAt       time.Time
	Fingerprint  string
	InhibitedBy  string
	SilencedBy   string
	Severity     string
	State        string
	StartsAt     time.Time
	UpdatedAt    time.Time
	GeneratorURL string
	Labels       datatypes.JSON
	RawAlert     datatypes.JSON
}

// Cluster table. Stores cluster configuration.
type Cluster struct {
	Model
	Token         string `gorm:"uniqueIndex"`
	Name          string `gorm:"uniqueIndex"`
	IngressURL    string
	CAPIName      string         `gorm:"column:capi_name"`
	CAPINamespace string         `gorm:"column:capi_namespace"`
	PullRequests  []*PullRequest `gorm:"many2many:cluster_pull_requests;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// ClusterStatus table. Used to sort cluster status by importance.
type ClusterStatus struct {
	ID     uint   `gorm:"primarykey"`
	Status string `gorm:"column:status"`
}

// FluxInfo table
type FluxInfo struct {
	ClusterToken string `gorm:"primaryKey"`
	Name         string `gorm:"primaryKey"`
	Namespace    string `gorm:"primaryKey"`
	Args         string
	Image        string
	RepoURL      string
	RepoBranch   string
	Syncs        datatypes.JSON
}

func (FluxInfo) TableName() string {
	return "flux_info"
}

type GitCommit struct {
	ClusterToken   string `gorm:"primaryKey"`
	Sha            string `gorm:"primaryKey"`
	AuthorName     string
	AuthorEmail    string
	AuthorDate     time.Time
	CommitterName  string
	CommitterEmail string
	CommitterDate  time.Time
	Message        string
}

type Workspace struct {
	ClusterToken string `gorm:"primaryKey"`
	Name         string `gorm:"primaryKey"`
	Namespace    string `gorm:"primaryKey"`
}

type PullRequest struct {
	Model
	URL  string
	Type string // maybe create/delete/modify
}

type CAPICluster struct {
	Model
	ClusterToken string
	Name         string
	Namespace    string
	CAPIVersion  string `gorm:"column:capi_version"`
	Object       datatypes.JSON
}

func (CAPICluster) TableName() string {
	return "capi_clusters"
}
