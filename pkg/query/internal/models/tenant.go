package models

import (
	"fmt"

	"gorm.io/gorm"
)

type Tenant struct {
	gorm.Model

	ID          string `gorm:"primaryKey;autoIncrement:false"`
	ClusterName string `json:"clusterName" gorm:"type:text"`
	Name        string `json:"name" gorm:"type:text"`
	Namespace   string `json:"namespace" gorm:"type:text"`
}

// Validate validates the fields of a Tenant
func (t *Tenant) Validate() error {
	if t.ClusterName == "" {
		return fmt.Errorf("missing cluster name field")
	}

	if t.Name == "" {
		return fmt.Errorf("missing name field")
	}

	if t.Namespace == "" {
		return fmt.Errorf("missing namespace field")
	}

	return nil
}

func (t *Tenant) GetID() string {
	return fmt.Sprintf("%s/%s", t.ClusterName, t.Name)
}

func (t *Tenant) GetClusterNamespacePair() string {
	return fmt.Sprintf("%s/%s", t.ClusterName, t.Namespace)
}
