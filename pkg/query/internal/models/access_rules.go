package models

type AccessRule struct {
	Cluster         string
	Namespace       string
	AccessibleKinds []string
	Subjects        []Subject
}
