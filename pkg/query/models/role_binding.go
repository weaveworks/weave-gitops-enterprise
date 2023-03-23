package models

import (
	"fmt"

	"gorm.io/gorm"
)

type RoleBinding struct {
	gorm.Model
	ID          string    `gorm:"primaryKey;autoIncrement:false"`
	Cluster     string    `gorm:"type:text"`
	Namespace   string    `gorm:"type:text"`
	Kind        string    `gorm:"type:text"`
	Name        string    `gorm:"type:text"`
	RoleRefName string    `gorm:"type:text"`
	RoleRefKind string    `gorm:"type:text"`
	Subjects    []Subject `gorm:"foreignKey:RoleBindingID"`
}

// Subject is a subject that is contained in a role binding.
// This is a "second-class"  object and is only here for database schema reasons.
type Subject struct {
	gorm.Model
	Kind          string `gorm:"type:text"`
	Name          string `gorm:"type:text"`
	Namespace     string `gorm:"type:text"`
	APIGroup      string `gorm:"type:text"`
	RoleBindingID string
}

func (o *RoleBinding) GetID() string {
	return fmt.Sprintf("%s/%s/%s/%s", o.Cluster, o.Namespace, o.Kind, o.Name)
}

func (o RoleBinding) Validate() error {
	if o.Cluster == "" {
		return fmt.Errorf("missing cluster field")
	}
	if o.Name == "" {
		return fmt.Errorf("missing name field")
	}

	if o.Kind == "" {
		return fmt.Errorf("missing kind field")
	}

	return nil
}
