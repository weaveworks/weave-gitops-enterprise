package models

import (
	"fmt"

	"gorm.io/gorm"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Object struct {
	gorm.Model
	Cluster   string `gorm:"type:text"`
	Namespace string `gorm:"type:text"`
	Kind      string `gorm:"type:text"`
	Name      string `gorm:"type:text"`
	Status    string `gorm:"type:text"`
	Message   string `gorm:"type:text"`
}

func (o Object) Validate() error {
	if o.Cluster == "" {
		return fmt.Errorf("missing cluster field")
	}
	if o.Name == "" {
		return fmt.Errorf("missing name field")
	}
	if o.Namespace == "" {
		return fmt.Errorf("missing namespace field")
	}
	if o.Kind == "" {
		return fmt.Errorf("missing kind field")
	}
	return nil
}

//counterfeiter:generate . ObjectRecord
type ObjectRecord interface {
	ClusterName() string
	Object() client.Object
}
