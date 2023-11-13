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

type Object struct {
	gorm.Model
	ID                  string                       `gorm:"primaryKey;autoIncrement:false"`
	Cluster             string                       `json:"cluster" gorm:"type:text"`
	Namespace           string                       `json:"namespace" gorm:"type:text"`
	APIGroup            string                       `json:"apiGroup" gorm:"type:text"`
	APIVersion          string                       `json:"apiVersion" gorm:"type:text"`
	Kind                string                       `json:"kind" gorm:"type:text"`
	Name                string                       `json:"name" gorm:"type:text"`
	Status              string                       `json:"status" gorm:"type:text"`
	Message             string                       `json:"message" gorm:"type:text"`
	Category            configuration.ObjectCategory `json:"category" gorm:"type:text"`
	KubernetesDeletedAt time.Time                    `json:"kubernetesDeletedAt"`
	Unstructured        json.RawMessage              `json:"unstructured" gorm:"type:blob"`
	Tenant              string                       `json:"tenant" gorm:"type:text"`
	Labels              map[string]string            `json:"labels" gorm:"-"`
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
	Object() NormalizedObject
	TransactionType() TransactionType
	RetentionPolicy() configuration.RetentionPolicy
}

type NormalizedObject interface {
	client.Object
	// GetStatus returns the status of the object, as determined by the ObjectKind StatusFunc
	GetStatus() (configuration.ObjectStatus, error)
	// GetMessage returns the message of the object, as determined by the ObjectKind MessageFunc
	GetMessage() (string, error)
	// GetCategory returns the category of the object, as determined by the ObjectKind Category
	GetCategory() (configuration.ObjectCategory, error)
	// GetLabels returns the labels for the object
	GetRelevantLabels() map[string]string
	// Raw returns the underlying client.Object
	Raw() client.Object
}

type defaultNormalizedObject struct {
	client.Object
	config configuration.ObjectKind
}

func (n defaultNormalizedObject) GetStatus() (configuration.ObjectStatus, error) {
	return n.config.StatusFunc(n.Object), nil
}

func (n defaultNormalizedObject) GetMessage() (string, error) {
	return n.config.MessageFunc(n.Object), nil
}

// GetRelevantLabels returns the object labels that have been configured to be selected.
func (n defaultNormalizedObject) GetRelevantLabels() map[string]string {
	labels := map[string]string{}
	objectLabels := n.GetLabels()
	for _, labelKey := range n.config.Labels {
		if objectLabels[labelKey] != "" {
			labels[labelKey] = objectLabels[labelKey]
		}
	}
	return labels
}

func (n defaultNormalizedObject) GetCategory() (configuration.ObjectCategory, error) {
	if n.config.Category == "" {
		return "", fmt.Errorf("category not found for object kind %q", n.config.Gvk.Kind)
	}
	return n.config.Category, nil
}

func (n defaultNormalizedObject) Raw() client.Object {
	return n.Object
}

func NewNormalizedObject(obj client.Object, config configuration.ObjectKind) NormalizedObject {
	return defaultNormalizedObject{
		Object: obj,
		config: config,
	}
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
