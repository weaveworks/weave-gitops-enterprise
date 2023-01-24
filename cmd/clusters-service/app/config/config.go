package config

import "time"

// Values onto the values.yaml file
// `yaml` tags map onto the keys in the values.yaml file
// `mapstructure` tags map onto the viper config keys (cli-flag, env-var, etc)
type Values struct {
	Config ValuesConfig `yaml:"config" mapstructure:",squash"`
	Global ValuesGlobal `yaml:"global" mapstructure:",squash"`
	TLS    ValuesTLS    `yaml:"tls" mapstructure:",squash"`
}

type ValuesGlobal struct {
	UseK8sCachedClients bool `yaml:"useK8sCachedClients" mapstructure:"use-k8s-cached-clients"`
}

type ValuesTLS struct {
	Enabled    *bool  `yaml:"enabled" mapstructure:"tls-enabled"`
	SecretName string `yaml:"secretName" mapstructure:"tls-secret-name"`
}

type ValuesConfig struct {
	LogLevel string        `yaml:"logLevel"`
	Cluster  ConfigCluster `yaml:"cluster" mapstructure:",squash"`
	Git      ConfigGit     `yaml:"git" mapstructure:",squash"`
	CAPI     ConfigCAPI    `yaml:"capi" mapstructure:",squash"`
	// Checkpoint     ConfigCheckpoint     `yaml:"checkpoint" mapstructure:",squash"`
	OIDC           ConfigOIDC           `yaml:"oidc" mapstructure:",squash"`
	Auth           ConfigAuth           `yaml:"auth" mapstructure:",squash"`
	UI             ConfigUI             `yaml:"ui" mapstructure:",squash"`
	CostEstimation ConfigCostEstimation `yaml:"costEstimation" mapstructure:",squash"`

	// This that we don't want to expose to usually expose to the user
	Internal ConfigInternal `yaml:"internal" mapstructure:",squash"`
}

type ConfigCAPI struct {
	Enabled   bool `yaml:"enabled" mapstructure:"capi-enabled"`
	Templates struct {
		Namespace             string `yaml:"namespace" mapstructure:"capi-templates-namespace"`
		InjectPruneAnnotation string `yaml:"injectPruneAnnotation" mapstructure:"inject-prune-annotation"`
		AddBasesKustomization string `yaml:"addBasesKustomization" mapstructure:"add-bases-kustomization"`
	} `mapstructure:"templates"`
	Clusters struct {
		Namespace string `yaml:"namespace" mapstructure:"namespace"`
	} `mapstructure:"clusters"`
	RepositoryURL          string `yaml:"repositoryURL" mapstructure:"capi-templates-repository-url"`
	RepositoryApiURL       string `yaml:"repositoryApiURL" mapstructure:"capi-templates-repository-api-url"`
	RepositoryPath         string `yaml:"repositoryPath" mapstructure:"capi-repository-path"`
	RepositoryClustersPath string `yaml:"repositoryClustersPath" mapstructure:"capi-repository-clusters-path"`
	BaseBranch             string `yaml:"baseBranch" mapstructure:"capi-templates-repository-base-branch"`
}

type ConfigOIDC struct {
	Enabled                 bool          `yaml:"enabled" mapstructure:"oidc-enabled"`
	IssuerURL               string        `yaml:"issuerURL" mapstructure:"oidc-issuer-url"`
	RedirectURL             string        `yaml:"redirectURL" mapstructure:"oidc-redirect-url"`
	TokenDuration           time.Duration `yaml:"cookieDuration" mapstructure:"oidc-token-duration"`
	ClientID                string        `yaml:"clientID" mapstructure:"oidc-client-id"`
	ClientSecret            string        `yaml:"clientSecret" mapstructure:"oidc-client-secret"`
	ClaimUsername           string        `yaml:"claimUsername" mapstructure:"oidc-claim-username"`
	ClaimGroups             string        `yaml:"claimGroups" mapstructure:"oidc-claim-groups"`
	ClientCredentialsSecret string        `yaml:"clientCredentialsSecret" mapstructure:"oidc-client-credentials-secret"`
	CustomScopes            []string      `yaml:"customScopes" mapstructure:"oidc-custom-scopes"`
}

type ConfigAuth struct {
	UserAccount struct {
		Enabled bool `yaml:"enabled" mapstructure:"user-account-enabled"`
	} `yaml:"userAccount" mapstructure:",squash"`
	TokenPassthrough struct {
		Enabled bool `yaml:"enabled" mapstructure:"token-passthrough-enabled"`
	} `yaml:"tokenPassthrough" mapstructure:",squash"`
}

type ConfigUI struct {
	LogoURL string `yaml:"logoURL" mapstructure:"ui-logo-url"`
	Footer  struct {
		BackgroundColor string `yaml:"backgroundColor" mapstructure:"ui-footer-background-color"`
		Color           string `yaml:"color" mapstructure:"ui-footer-color"`
		Content         string `yaml:"content" mapstructure:"ui-footer-content"`
		HideVersion     bool   `yaml:"hideVersion" mapstructure:"ui-footer-hide-version"`
	} `yaml:"footer" mapstructure:",squash"`
}

type ConfigCluster struct {
	Name string `mapstructure:"cluster-name"`
}

type ConfigGit struct {
	Type     string `yaml:"type" mapstructure:"git-type"`
	Hostname string `yaml:"hostname" mapstructure:"git-hostname"`
}

// type ConfigCheckpoint struct {
// 	Enabled bool `mapstructure:"checkpoint-enabled"`
// }

type ConfigCostEstimation struct {
	EstimationFilter string `yaml:"estimationFilter" mapstructure:"cost-estimation-filters"`
	APIRegion        string `yaml:"apiRegion" mapstructure:"cost-estimation-api-region"`
	CSVFile          string `yaml:"csvFile" mapstructure:"cost-estimation-csv-file"`
}

// For completeness we include all the fields in the config struct
type ConfigInternal struct {
	DevMode                   bool              `yaml:"devMode" mapstructure:"dev-mode"`
	EntitlementSecret         ConfigEntitlement `yaml:"entitlementSecret" mapstructure:",squash"`
	HelmRepo                  ConfigHelmRepo    `yaml:"helmRepo" mapstructure:",squash"`
	HtmlRootPath              string            `yaml:"htmlRootPath" mapstructure:"html-root-path"`
	PipelineControllerAddress string            `yaml:"pipelineControllerAddress" mapstructure:"pipeline-controller-address"`
	ProfileCacheLocation      string            `yaml:"profileCacheLocation" mapstructure:"profile-cache-location"`
	RuntimeNamespace          string            `yaml:"runtimeNamespace" mapstructure:"runtime-namespace"`
	TLSCert                   string            `yaml:"tlsCert" mapstructure:"tls-cert"`
	TLSKey                    string            `yaml:"tlsKey" mapstructure:"tls-key"`

	// Is this used at all? Maybe read by some part of OSS?
	GitProviderToken string `mapstructure:"git-provider-token"`

	// legacy, conflicts with tls.enabled, prefer using that
	NoTLS *bool `yaml:"noTLS" mapstructure:"no-tls"`
}

type ConfigEntitlement struct {
	Name      string `yaml:"name" mapstructure:"entitlement-secret"`
	Namespace string `yaml:"namespace" mapstructure:"entitlement-secret-namespace"`
}

type ConfigHelmRepo struct {
	Name      string `yaml:"name" mapstructure:"helm-repo-name"`
	Namespace string `yaml:"namespace" mapstructure:"helm-repo-namespace"`
}
