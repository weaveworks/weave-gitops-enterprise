package steps

import (
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultAdminUsername = "wego-admin"
	defaultAdminPassword = "password"
)

// inputs names
const (
	UserName     = "username"
	Password     = "password"
	WGEVersion   = "wgeVersion"
	UserDomain   = "userDomain"
	DiscoveryURL = "discoveryURL"
	ClientID     = "clientID"
	ClientSecret = "clientSecret"
)

// input/output types
const (
	successMsg           = "successMsg"
	failureMsg           = "failureMsg"
	multiSelectionChoice = "multiSelect"
	stringInput          = "string"
	passwordInput        = "password"
	confirmInput         = "confirm"
	typeSecret           = "secret"
	typeFile             = "file"
	typePortforward      = "portforward"
)

// ConfigBuilder contains all the different configuration options that a user can introduce
type ConfigBuilder struct {
	logger                  logger.Logger
	kubeconfig              string
	username                string
	password                string
	wGEVersion              string
	domainType              string
	domain                  string
	privateKeyPath          string
	privateKeyPassword      string
	PromptedForDiscoveryURL bool
}

func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{}
}

func (c *ConfigBuilder) WithLogWriter(logger logger.Logger) *ConfigBuilder {
	c.logger = logger
	return c
}

func (c *ConfigBuilder) WithUsername(username string) *ConfigBuilder {
	c.username = username
	return c
}

func (c *ConfigBuilder) WithPassword(password string) *ConfigBuilder {
	c.password = password
	return c
}

func (c *ConfigBuilder) WithKubeconfig(kubeconfig string) *ConfigBuilder {
	c.kubeconfig = kubeconfig
	return c
}

func (c *ConfigBuilder) WithVersion(version string) *ConfigBuilder {
	c.wGEVersion = version
	return c
}

func (c *ConfigBuilder) WithDomainType(domainType string) *ConfigBuilder {
	c.domainType = domainType
	return c

}

func (c *ConfigBuilder) WithDomain(domain string) *ConfigBuilder {
	c.domain = domain
	return c

}

func (c *ConfigBuilder) WithPrivateKey(privateKeyPath string, privateKeyPassword string) *ConfigBuilder {
	c.privateKeyPath = privateKeyPath
	c.privateKeyPassword = privateKeyPassword
	return c
}

func (c *ConfigBuilder) WithPromptedForDiscoveryURL(prompted bool) *ConfigBuilder {
	c.PromptedForDiscoveryURL = prompted
	return c
}

// Config is the configuration struct to user for WGE installation. It includes
// configuration values as well as other required structs like clients
type Config struct {
	KubernetesClient k8s_client.Client
	Logger           logger.Logger

	ExistsWgeVersion string // existing wge version in the cluster
	WGEVersion       string // user want this version in the cluster

	Username string // cluster user username
	Password string // cluster user password

	DomainType string
	UserDomain string

	PrivateKeyPath     string
	PrivateKeyPassword string

	PromptedForDiscoveryURL bool
}

// Builds creates a valid config so boostrap could be executed. It uses values introduced
// and checks the requirements for the environmnet.
func (cb *ConfigBuilder) Build() (Config, error) {
	l := cb.logger

	kubernetesClient, err := utils.GetKubernetesClient(cb.kubeconfig)
	if err != nil {
		return Config{}, fmt.Errorf("failed to get kubernetes client. error: %s", err)
	}
	l.Actionf("created client to cluster")

	// TODO this should be part of a verify wge step, potentially select wge version
	//installedVersion, err := utils.GetHelmReleaseVersion(kubernetesClient, WgeHelmReleaseName, WGEDefaultNamespace)
	//if err != nil {
	//	if !errors.IsNotFound(err) {
	//		return Config{}, fmt.Errorf("unexpected error finding weave gitops helm release: %w", err)
	//	}
	//	l.Successf("weave gitops not found in the cluster")
	//} else if installedVersion != "" {
	//	return Config{}, fmt.Errorf("cannot config bootstrap: weave gitops `%s` exists in the cluster", installedVersion)
	//}

	// validate ssh keys
	if cb.privateKeyPath != "" {
		_, err = os.ReadFile(cb.privateKeyPath)
		if err != nil {
			return Config{}, fmt.Errorf("cannot read ssh key: %v", err)
		}
	}

	//TODO we should do validations in case invalid values and throw an error early
	return Config{
		KubernetesClient:        kubernetesClient,
		WGEVersion:              cb.wGEVersion,
		Username:                cb.username,
		Password:                cb.password,
		Logger:                  cb.logger,
		DomainType:              cb.domainType,
		UserDomain:              cb.domain,
		PrivateKeyPath:          cb.privateKeyPath,
		PrivateKeyPassword:      cb.privateKeyPassword,
		PromptedForDiscoveryURL: cb.PromptedForDiscoveryURL,
	}, nil

}

type fileContent struct {
	Name      string
	Content   string
	CommitMsg string
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

// OIDCConfig store the OIDC config
type OIDCConfig struct {
	IssuerURL    string `json:"issuerURL"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	RedirectURL  string `json:"redirectURL"`
}
