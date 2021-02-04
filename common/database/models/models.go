package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/types"
)

// Event table, ref: https://godoc.org/k8s.io/api/core/v1#Event
type Event struct {
	UID          types.UID `gorm:"primaryKey"`
	CreatedAt    datatypes.Date
	RegisteredAt datatypes.Date
	Name         string
	Namespace    string
	Labels       string
	Annotations  string
	ClusterName  string
	Cluster      Cluster `gorm:"foreignKey:Name"`
	Reason       string
	Message      string
	Type         string
	RawEvent     datatypes.JSON
}

// Cluster table
type Cluster struct {
	gorm.Model
	Name              string
	Namespace         string
	Labels            string
	Annotations       string
	ControlPlaneNodes int
	WorkerNodes       int
	CNIInfo           string
	CSIInfo           string
	CRIInfo           string
	Version           string
	IngressInfo       string
}

// GitRepository table
type GitRepository struct {
	gorm.Model
	URL               string `gorm:"primaryKey"`
	Namespace         string
	SecretRef         string
	Interval          time.Time
	Timeout           time.Time
	Branch            string `gorm:"primaryKey"`
	Ignore            *string
	Suspend           bool
	GitImplementation string
	RawGitRepo        datatypes.JSON
}

// Workspace table
type Workspace struct {
	gorm.Model
	Name                string
	Namespace           string
	Namespaces          string
	MemberRole          string
	GitProviderHostname string
	GitProvider         GitProvider `gorm:"foreignKey:Hostname"`
	GitRepositoryID     string
	// Setting as primary key the ID. Git repos can share the same URL
	// but have different branches
	GitRepository GitRepository `gorm:"foreignKey:ID"`
	RawWorkspace  datatypes.JSON
}

// GitProvider table
type GitProvider struct {
	gorm.Model
	Hostname        string `gorm:"primaryKey"`
	Type            string
	SecretName      string
	SecretNamespace string
}
