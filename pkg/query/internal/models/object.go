package models

import "sigs.k8s.io/controller-runtime/pkg/client"

type Object struct {
	Cluster   string
	Namespace string
	Kind      string
	Name      string
	Status    string
	Message   string
	Operation string
}

//counterfeiter:generate . ObjectRecord
type ObjectRecord interface {
	ClusterName() string
	Object() client.Object
}
