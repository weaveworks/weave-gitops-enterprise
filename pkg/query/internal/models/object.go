package models

type Object struct {
	Id        string
	Cluster   string
	Namespace string
	Kind      string
	Name      string
	Status    string
	Message   string
	Operation string
}
