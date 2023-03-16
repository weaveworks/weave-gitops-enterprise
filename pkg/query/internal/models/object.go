package models

import (
	"fmt"

	"gorm.io/gorm"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Object struct {
	gorm.Model
	Cluster   string
	Namespace string
	Kind      string
	Name      string
	Status    string
	Message   string
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
