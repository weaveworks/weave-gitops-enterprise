package payload

type WorkspaceInfo struct {
	Token      string      `json:"token"`
	Workspaces []Workspace `json:"workspace"`
}

type Workspace struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
