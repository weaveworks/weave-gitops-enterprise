package payload

type FluxInfo struct {
	Token       string               `json:"token"`
	Deployments []FluxDeploymentInfo `json:"fluxinfo"`
}

type FluxDeploymentInfo struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Args      []string `json:"args"`
	Image     string   `json:"image"`
}
