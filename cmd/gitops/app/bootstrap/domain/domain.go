package domain

// ValuesFile store the wge values
type ValuesFile struct {
	Config             ValuesWGEConfig        `json:"config,omitempty"`
	Ingress            map[string]interface{} `json:"ingress,omitempty"`
	TLS                map[string]interface{} `json:"tls,omitempty"`
	PolicyAgent        map[string]interface{} `json:"policy-agent,omitempty"`
	PipelineController map[string]interface{} `json:"pipeline-controller,omitempty"`
	EnablePipelines    bool                   `json:"enablePipelines,omitempty"`
	EnableTerraformUI  bool                   `json:"enableTerraformUI,omitempty"`
}

// ValuesWGEConfig store the wge values config field
type ValuesWGEConfig struct {
	CAPI map[string]interface{} `json:"capi,omitempty"`
	OIDC map[string]interface{} `json:"oidc,omitempty"`
}

// HelmChartResponse store the chart versions response
type HelmChartResponse struct {
	ApiVersion string
	Entries    map[string][]ChartEntry
	Generated  string
}

// ChartEntry store the HelmChartResponse entries
type ChartEntry struct {
	ApiVersion string
	Name       string
	Version    string
}

const (
	POLICY_AGENT_VALUES_NAME = "policy-agent"
)
