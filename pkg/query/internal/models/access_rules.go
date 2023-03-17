package models

import (
	"fmt"

	"gorm.io/gorm"
)

type AccessRule struct {
	gorm.Model
	ID              string   `gorm:"primaryKey;autoIncrement:false"`
	Cluster         string   `gorm:"type:text"`
	Principal       string   `gorm:"type:text"`
	Namespace       string   `gorm:"type:text"`
	AccessibleKinds []string `gorm:"type:text"`
}

func (a AccessRule) Validate() error {
	if a.Cluster == "" {
		return fmt.Errorf("cluster is empty")
	}
	if a.Principal == "" {
		return fmt.Errorf("principal is empty")
	}
	if a.Namespace == "" {
		return fmt.Errorf("namespace is empty")
	}
	if len(a.AccessibleKinds) == 0 {
		return fmt.Errorf("accessible kinds is empty")
	}
	return nil
}

func (a AccessRule) GetID() string {
	return fmt.Sprintf("%s/%s/%s", a.Cluster, a.Namespace, a.Principal)
}
