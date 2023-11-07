package steps

import (
	"errors"
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"
)

// auth types
const (
	AuthOIDC = "oidc"
)

const (
	defaultAdminPassword = "password"
)

// git schemes
const (
	httpsScheme = "https"
	sshScheme   = "ssh"
)

// inputs names
const (
	inPassword           = "password"
	inWGEVersion         = "wgeVersion"
	inUserDomain         = "userDomain"
	inPrivateKeyPath     = "privateKeyPath"
	inPrivateKeyPassword = "privateKeyPassword"
	inExistingCreds      = "existingCreds"
	inDomainType         = "domainType"
	inDiscoveryURL       = "discoveryURL"
	inClientID           = "clientID"
	inClientSecret       = "clientSecret"
	inOidcInstalled      = "oidcInstalled"
	inExistingOIDC       = "existingOIDC"
	inRepoURL            = "repoURL"
	inBranch             = "branch"
	inRepoPath           = "repoPath"
	inGitUserName        = "username"
	inGitToken           = "gitToken"
	inBootstrapFlux      = "bootstrapFlux"
)

// input/output types
const (
	multiSelectionChoice = "multiSelect"
	stringInput          = "string"
	passwordInput        = "password"
	confirmInput         = "confirm"
	typeSecret           = "secret"
	typeFile             = "file"
)

// ConfigBuilder contains all the different configuration options that a user can introduce
type ConfigBuilder struct {
	logger                  logger.Logger
	kubeconfig              string
	password                string
	wgeVersion              string
	domainType              string
	domain                  string
	privateKeyPath          string
	privateKeyPassword      string
	gitUsername             string
	gitToken                string
	repoURL                 string
	branch                  string
	repoPath                string
	authType                string
	installOIDC             string
	discoveryURL            string
	clientID                string
	clientSecret            string
	PromptedForDiscoveryURL bool
}

func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{}
}

func (c *ConfigBuilder) WithLogWriter(logger logger.Logger) *ConfigBuilder {
	c.logger = logger
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
	c.wgeVersion = version
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

func (c *ConfigBuilder) WithGitAuthentication(privateKeyPath, privateKeyPassword, gitUsername, gitToken string) *ConfigBuilder {
	c.privateKeyPath = privateKeyPath
	c.privateKeyPassword = privateKeyPassword
	c.gitUsername = gitUsername
	c.gitToken = gitToken

	return c
}

func (c *ConfigBuilder) WithFluxGitRepository(repoURL, branch, repoPath string) *ConfigBuilder {
	c.repoURL = repoURL
	c.branch = branch
	c.repoPath = repoPath
	return c
}

func (c *ConfigBuilder) WithOIDCConfig(discoveryURL string, clientID string, clientSecret string, prompted bool) *ConfigBuilder {
	c.authType = AuthOIDC
	c.discoveryURL = discoveryURL
	c.clientID = clientID
	c.clientSecret = clientSecret
	if discoveryURL != "" && clientID != "" && clientSecret != "" {
		prompted = false
	}
	c.PromptedForDiscoveryURL = prompted
	c.installOIDC = "y" // todo: change to parameter
	return c
}

// Config is the configuration struct to user for WGE installation. It includes
// configuration values as well as other required structs like clients
type Config struct {
	KubernetesClient k8s_client.Client
	Logger           logger.Logger

	WGEVersion string // user want this version in the cluster

	Password string // cluster user password

	DomainType string
	UserDomain string

	GitScheme string

	FluxInstallated    bool
	PrivateKeyPath     string
	PrivateKeyPassword string

	GitUsername string
	GitToken    string

	RepoURL  string
	Branch   string
	RepoPath string

	AuthType                string
	InstallOIDC             string
	DiscoveryURL            string
	IssuerURL               string
	ClientID                string
	ClientSecret            string
	RedirectURL             string
	PromptedForDiscoveryURL bool
}

// Builds creates a valid config so boostrap could be executed. It uses values introduced
// and checks the requirements for the environments.
func (cb *ConfigBuilder) Build() (Config, error) {
	l := cb.logger
	l.Actionf("creating client to cluster")
	kubeHttp, err := utils.GetKubernetesHttp(cb.kubeconfig)
	if err != nil {
		return Config{}, fmt.Errorf("failed to get kubernetes client. error: %s", err)
	}
	l.Successf("created client to cluster: %s", kubeHttp.ClusterName)

	// validate ssh keys
	if cb.privateKeyPath != "" {
		_, err = os.ReadFile(cb.privateKeyPath)
		if err != nil {
			return Config{}, fmt.Errorf("cannot read ssh key: %v", err)
		}
	}

	if cb.password != "" && len(cb.password) < 6 {
		return Config{}, errors.New("password minimum characters should be >= 6")
	}

	// parse repo scheme
	var scheme string
	if cb.repoURL != "" {
		scheme, err = parseRepoScheme(cb.repoURL)
		if err != nil {
			return Config{}, err
		}
	}

	//TODO we should do validations in case invalid values and throw an error early
	return Config{
		KubernetesClient:        kubeHttp.Client,
		WGEVersion:              cb.wgeVersion,
		Password:                cb.password,
		Logger:                  cb.logger,
		DomainType:              cb.domainType,
		UserDomain:              cb.domain,
		GitScheme:               scheme,
		Branch:                  cb.branch,
		RepoPath:                cb.repoPath,
		RepoURL:                 cb.repoURL,
		PrivateKeyPath:          cb.privateKeyPath,
		PrivateKeyPassword:      cb.privateKeyPassword,
		GitUsername:             cb.gitUsername,
		GitToken:                cb.gitToken,
		AuthType:                cb.authType,
		InstallOIDC:             cb.installOIDC,
		DiscoveryURL:            cb.discoveryURL,
		ClientID:                cb.clientID,
		ClientSecret:            cb.clientSecret,
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
