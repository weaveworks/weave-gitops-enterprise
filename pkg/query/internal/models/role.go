package models

import (
	"fmt"
	"strings"

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
	APIGroups     string `gorm:"type:text"`
	Resources     string `gorm:"type:text"`
	Verbs         string `gorm:"type:text"`
	RoleID        string
	ResourceNames string `gorm:"type:text"`
}

func (o Role) GetID() string {
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

	if o.PolicyRules == nil || len(o.PolicyRules) == 0 {
		return fmt.Errorf("missing policy rules")
	}

	return nil
}

// We join the arrays into a string because the database may not support arrays.
// These helpers are used to ensure we always use the same separator.
var separator = ","

func JoinRuleData(data []string) string {
	return strings.Join(data, separator)
}

func SplitRuleData(data string) []string {
	if len(data) == 0 {
		return nil
	}
	return strings.Split(data, separator)
}
