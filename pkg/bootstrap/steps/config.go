package steps

import (
	"fmt"
	"io"
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
	confirmYes           = "y"
	confirmNo            = "n"
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
	inPrivateKeyPath     = "privateKeyPath"
	inPrivateKeyPassword = "privateKeyPassword"
	inDiscoveryURL       = "discoveryURL"
	inClientID           = "clientID"
	inClientSecret       = "clientSecret"
	inOidcInstalled      = "oidcInstalled"
	inExistingOIDC       = "existingOIDC"
	inRepoURL            = "repoURL"
	inBranch             = "branch"
	inRepoPath           = "repoPath"
	inGitUserName        = "username"
	inGitPassword        = "gitPassowrd"
	inBootstrapFlux      = "bootstrapFlux"
	inComponentsExtra    = "componentsExtra"
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
	logger             logger.Logger
	kubeconfig         string
	password           string
	wgeVersion         string
	privateKeyPath     string
	privateKeyPassword string
	// privateKeyPasswordChanged indicates true when the value privateKeyPassword
	// comes from the user input. false otherwise.
	privateKeyPasswordChanged bool
	silent                    bool
	export                    bool
	gitUsername               string
	gitToken                  string
	repoURL                   string
	repoBranch                string
	repoPath                  string
	authType                  string
	installOIDC               string
	discoveryURL              string
	clientID                  string
	clientSecret              string
	PromptedForDiscoveryURL   bool
	bootstrapFlux             bool
	componentsExtra           []string
	outWriter                 io.Writer
	inReader                  io.Reader
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

func (c *ConfigBuilder) WithGitAuthentication(privateKeyPath, privateKeyPassword string, privateKeyPasswordChanged bool,
	gitUsername, gitToken string) *ConfigBuilder {
	c.privateKeyPath = privateKeyPath
	c.privateKeyPassword = privateKeyPassword
	c.privateKeyPasswordChanged = privateKeyPasswordChanged
	c.gitUsername = gitUsername
	c.gitToken = gitToken

	return c
}

func (c *ConfigBuilder) WithGitRepository(repoURL, branch, repoPath string) *ConfigBuilder {
	c.repoURL = repoURL
	c.repoBranch = branch
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

func (c *ConfigBuilder) WithSilent(silent bool) *ConfigBuilder {
	c.silent = silent
	return c
}

func (c *ConfigBuilder) WithBootstrapFluxFlag(bootstrapFlux bool) *ConfigBuilder {
	c.bootstrapFlux = bootstrapFlux
	return c
}

func (c *ConfigBuilder) WithComponentsExtra(componentsExtra []string) *ConfigBuilder {
	c.componentsExtra = componentsExtra
	return c
}

func (c *ConfigBuilder) WithExport(export bool) *ConfigBuilder {
	c.export = export
	return c
}

func (c *ConfigBuilder) WithInReader(inReader io.Reader) *ConfigBuilder {
	c.inReader = inReader
	return c
}

func (c *ConfigBuilder) WithOutWriter(outWriter io.Writer) *ConfigBuilder {
	c.outWriter = outWriter
	return c
}

// Config is the configuration struct to user for WGE installation. It includes
// configuration values as well as other required structs like clients
type Config struct {
	KubernetesClient k8s_client.Client
	// TODO move me to a better package
	GitClient utils.GitClient
	// TODO move me to a better package
	FluxClient utils.FluxClient
	Logger     logger.Logger
	// InReader holds the stream to read input from
	InReader io.Reader
	// OutWriter holds the output to write to
	OutWriter  io.Writer
	FluxConfig FluxConfig
	WgeConfig  WgeConfig

	ClusterUserAuth ClusterUserAuthConfig
	ModesConfig     ModesConfig

	// TODO refactor me to git ssh auth config type
	PrivateKeyPath            string
	PrivateKeyPassword        string
	PrivateKeyPasswordChanged bool

	// TODO refactor me to git https auth config type
	GitUsername string
	GitToken    string

	// GitRepository contains the configuration for the git repo
	GitRepository GitRepositoryConfig

	AuthType                string
	InstallOIDC             string
	DiscoveryURL            string
	IssuerURL               string
	ClientID                string
	ClientSecret            string
	RedirectURL             string
	PromptedForDiscoveryURL bool

	BootstrapFlux   bool
	ComponentsExtra ComponentsExtraConfig
}

// Builds creates a valid config so boostrap could be executed. It uses values introduced
// and checks the requirements for the environments.
func (cb *ConfigBuilder) Build() (Config, error) {
	l := cb.logger

	if cb.inReader == nil {
		return Config{}, fmt.Errorf("input cannot be nil")
	}

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

	clusterUserAuthConfig, err := NewClusterUserAuthConfig(cb.password, kubeHttp.Client)
	if err != nil {
		return Config{}, fmt.Errorf("error creating cluster user auth configuration: %v", err)
	}

	fluxConfig, err := NewFluxConfig(cb.logger, kubeHttp.Client)
	if err != nil {
		return Config{}, fmt.Errorf("error creating flux configuration: %v", err)
	}

	gitRepositoryConfig, err := NewGitRepositoryConfig(cb.repoURL, cb.repoBranch, cb.repoPath, fluxConfig)
	if err != nil {
		return Config{}, fmt.Errorf("error creating git repository configuration: %v", err)
	}

	wgeConfig, err := NewWgeConfig(cb.wgeVersion, kubeHttp.Client, fluxConfig.IsInstalled)
	if err != nil {
		return Config{}, fmt.Errorf("cannot create WGE configuration: %v", err)
	}

	componentsExtraConfig, err := NewInstallExtraComponentsConfig(cb.componentsExtra, kubeHttp.Client, fluxConfig.IsInstalled)
	if err != nil {
		return Config{}, fmt.Errorf("cannot create components extra configuration: %v", err)
	}

	//TODO we should do validations in case invalid values and throw an error early
	return Config{
		KubernetesClient: kubeHttp.Client,
		GitClient:        &utils.GoGitClient{},
		FluxClient:       &utils.CmdFluxClient{},
		InReader:         cb.inReader,
		OutWriter:        cb.outWriter,
		WgeConfig:        wgeConfig,
		ClusterUserAuth:  clusterUserAuthConfig,
		GitRepository:    gitRepositoryConfig,
		Logger:           cb.logger,
		ModesConfig: ModesConfig{
			Silent: cb.silent,
			Export: cb.export,
		},
		PrivateKeyPath:            cb.privateKeyPath,
		PrivateKeyPassword:        cb.privateKeyPassword,
		PrivateKeyPasswordChanged: cb.privateKeyPasswordChanged,
		GitUsername:               cb.gitUsername,
		GitToken:                  cb.gitToken,
		AuthType:                  cb.authType,
		InstallOIDC:               cb.installOIDC,
		DiscoveryURL:              cb.discoveryURL,
		ClientID:                  cb.clientID,
		ClientSecret:              cb.clientSecret,
		PromptedForDiscoveryURL:   cb.PromptedForDiscoveryURL,
		ComponentsExtra:           componentsExtraConfig,
		FluxConfig:                fluxConfig,
		BootstrapFlux:             cb.bootstrapFlux,
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
	Service            map[string]interface{} `json:"service,omitempty"`
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
