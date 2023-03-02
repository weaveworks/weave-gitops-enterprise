package models

type AccessRule struct {
	Cluster         string
	Role            string
	Namespace       string
	AccessibleKinds []string
}
