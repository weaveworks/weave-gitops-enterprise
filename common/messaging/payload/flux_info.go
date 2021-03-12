package payload

type FluxInfo struct {
	Token       string               `json:"token"`
	Deployments []FluxDeploymentInfo `json:"fluxinfo"`
}

type FluxLogInfo struct {
	Timestamp string `json:"ts"`
	URL       string `json:"url"`
	Branch    string `json:"branch"`
	Head      string `json:"head"`
	Event     string `json:"event"`
}

type FluxDeploymentInfo struct {
	Name      string        `json:"name"`
	Namespace string        `json:"namespace"`
	Args      []string      `json:"args"`
	Image     string        `json:"image"`
	Syncs     []FluxLogInfo `json:"fluxSyncs"`
}
