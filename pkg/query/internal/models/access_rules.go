package models

type AccessRule struct {
	Cluster         string
	Principal       string
	Namespace       string
	AccessibleKinds []string
}
