package commands

import (
	"github.com/weaveworks/weave-gitops/pkg/logger"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"
)

// Bootstrapper is the main struct for WGE installation setup
type Bootstrapper struct {
	Username         string
	Password         string
	KubernetesClient k8s_client.Client
	WGEVersion       string
	UserDomain       string
	Logger           logger.Logger
}

// ValuesFile store the wge values
type valuesFile struct {
	Config             ValuesWGEConfig        `json:"config,omitempty"`
	Ingress            map[string]interface{} `json:"ingress,omitempty"`
	TLS                map[string]interface{} `json:"tls,omitempty"`
	PolicyAgent        map[string]interface{} `json:"policy-agent,omitempty"`
	PipelineController map[string]interface{} `json:"pipeline-controller,omitempty"`
	GitOpsSets         map[string]interface{} `json:"gitopssets-controller,omitempty"`
	EnablePipelines    bool                   `json:"enablePipelines,omitempty"`
	EnableTerraformUI  bool                   `json:"enableTerraformUI,omitempty"`
	Global             global                 `json:"global,omitempty"`
	ClusterController  clusterController      `json:"cluster-controller,omitempty"`
}

// ValuesWGEConfig store the wge values config field
type ValuesWGEConfig struct {
	CAPI map[string]interface{} `json:"capi,omitempty"`
	OIDC map[string]interface{} `json:"oidc,omitempty"`
}

// ClusterController store the wge values cluster controller field
type clusterController struct {
	Enabled           bool                     `json:"enabled,omitempty"`
	FullNameOverride  string                   `json:"fullnameOverride,omitempty"`
	ControllerManager clusterControllerManager `json:"controllerManager,omitempty"`
}

// ClusterController store the wge values clustercontrollermanager  field
type clusterControllerManager struct {
	Manager clusterControllerManagerManager `json:"manager,omitempty"`
}

// ClusterControllerManagerManager store the wge values clustercontrollermanager manager  field
type clusterControllerManagerManager struct {
	Image clusterControllerImage `json:"image,omitempty"`
}

// ClusterControllerManagerManager store the wge values clustercontrollermanager image  field
type clusterControllerImage struct {
	Repository string `json:"repository,omitempty"`
	Tag        string `json:"tag,omitempty"`
}

// Global store the global variables
type global struct {
	CapiEnabled bool `json:"capiEnabled,omitempty"`
}

// HelmChartResponse store the chart versions response
type helmChartResponse struct {
	ApiVersion string
	Entries    map[string][]chartEntry
	Generated  string
}

// ChartEntry store the HelmChartResponse entries
type chartEntry struct {
	ApiVersion string
	Name       string
	Version    string
}
