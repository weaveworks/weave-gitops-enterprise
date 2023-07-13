package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"

	"gorm.io/gorm"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectCategory string

const (
	CategoryAutomation ObjectCategory = "automation"
	CategorySource     ObjectCategory = "source"
	CategoryEvent      ObjectCategory = "event"
)

type Object struct {
	gorm.Model
	ID                  string          `gorm:"primaryKey;autoIncrement:false"`
	Cluster             string          `json:"cluster" gorm:"type:text"`
	Namespace           string          `json:"namespace" gorm:"type:text"`
	APIGroup            string          `json:"apiGroup" gorm:"type:text"`
	APIVersion          string          `json:"apiVersion" gorm:"type:text"`
	Kind                string          `json:"kind" gorm:"type:text"`
	Name                string          `json:"name" gorm:"type:text"`
	Status              string          `json:"status" gorm:"type:text"`
	Message             string          `json:"message" gorm:"type:text"`
	Category            ObjectCategory  `json:"category" gorm:"type:text"`
	KubernetesDeletedAt time.Time       `json:"kubernetesDeletedAt"`
	Unstructured        json.RawMessage `json:"unstructured" gorm:"type:blob"`
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
	if o.APIVersion == "" {
		return fmt.Errorf("missing api version field")
	}
	if o.Kind == "" {
		return fmt.Errorf("missing kind field")
	}

	if o.Category == "" {
		return errors.New("category is required")
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

// https://pkg.go.dev/github.com/ttys3/bleve/mapping#Classifier
// Type returns a collection identifier to help with indexing
func (o Object) Type() string {
	return "object"
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
	RetentionPolicy() configuration.RetentionPolicy
}

func IsExpired(policy configuration.RetentionPolicy, obj Object) bool {
	currentTime := time.Now()
	retention := time.Duration(policy)
	expirationTime := currentTime.Add(-retention)

	ts := obj.KubernetesDeletedAt

	if ts.IsZero() {
		return false
	}

	if ts.Before(expirationTime) {
		return true
	}

	return false
}
