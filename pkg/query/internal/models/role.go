package models

import (
	"fmt"

	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	ID          string       `gorm:"primaryKey;autoIncrement:false"`
	Cluster     string       `gorm:"type:text"`
	Namespace   string       `gorm:"type:text"`
	Kind        string       `gorm:"type:text"`
	Name        string       `gorm:"type:text"`
	PolicyRules []PolicyRule `gorm:"foreignKey:RoleID"`
}

// PolicyRule is a rule that applies to a role.
// This is a "second-class"  object and is only here for database schema reasons..
type PolicyRule struct {
	gorm.Model
	APIGroups string `gorm:"type:text"`
	Resources string `gorm:"type:text"`
	Verbs     string `gorm:"type:text"`
	RoleID    string
}

func (o *Role) GetID() string {
	return fmt.Sprintf("%s/%s/%s/%s", o.Cluster, o.Namespace, o.Kind, o.Name)
}

func (o Role) Validate() error {
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
