package models

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Object struct {
	gorm.Model
	ID         string `gorm:"primaryKey;autoIncrement:false"`
	Cluster    string `gorm:"type:text"`
	Namespace  string `gorm:"type:text"`
	APIGroup   string `gorm:"type:text"`
	APIVersion string `gorm:"type:text"`
	Kind       string `gorm:"type:text"`
	Name       string `gorm:"type:text"`
	Status     string `gorm:"type:text"`
	Message    string `gorm:"type:text"`
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
	if o.APIGroup == "" {
		return fmt.Errorf("missing api group field")
	}
	if o.APIVersion == "" {
		return fmt.Errorf("missing api version field")
	}
	if o.Kind == "" {
		return fmt.Errorf("missing kind field")
	}
	return nil
}

func (o *Object) GetID() string {
	return fmt.Sprintf("%s/%s/%s/%s", o.Cluster, o.Namespace, o.GroupVersionKind(), o.Name)
}

func (o *Object) String() string {
	return o.GetID()
}

func (o Object) GroupVersionKind() string {
	s := []string{o.APIGroup, o.APIVersion, o.Kind}

	if o.APIVersion == "" {
		s = []string{o.APIGroup, o.Kind}
	}

	return strings.Join(s, "/")
}

type TransactionType string

const (
	TransactionTypeUpsert    TransactionType = "upsert"
	TransactionTypeDelete    TransactionType = "delete"
	TransactionTypeDeleteAll TransactionType = "deleteAll"
)

//counterfeiter:generate . ObjectTransaction
type ObjectTransaction interface {
	ClusterName() string
	Object() client.Object
	TransactionType() TransactionType
}
