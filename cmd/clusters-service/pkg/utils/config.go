package utils

import "github.com/spf13/viper"

type Config struct {
	ClusterName                       string `mapstructure:"CLUSTER_NAME"`
	GitProviderType                   string `mapstructure:"GIT_PROVIDER_TYPE"`
	GitProviderHostname               string `mapstructure:"GIT_PROVIDER_HOSTNAME"`
	CAPIClustersNamespace             string `mapstructure:"CAPI_CLUSTERS_NAMESPACE"`
	CAPITemplatesNamespace            string `mapstructure:"CAPI_TEMPLATES_NAMESPACE"`
	InjectPruneAnnotation             string `mapstructure:"INJECT_PRUNE_ANNOTATION"`
	CAPITemplatesRepositoryUrl        string `mapstructure:"CAPI_TEMPLATES_REPOSITORY_URL"`
	CAPIRepositoryPath                string `mapstructure:"CAPI_REPOSITORY_PATH"`
	CAPITemplatesRepositoryApiUrl     string `mapstructure:"CAPI_TEMPLATES_REPOSITORY_API_URL"`
	CAPITemplatesRepositoryBaseBranch string `mapstructure:"CAPI_TEMPLATES_REPOSITORY_BASE_BRANCH"`
	CheckpointDisable                 int    `mapstructure:"CHECKPOINT_DISABLE"`
	WeaveGitopsAuthEnabled            string `mapstructure:"WEAVE_GITOPS_AUTH_ENABLED"`
	OIDCIssuerUrl                     string `mapstructure:"OIDC_ISSUER_URL"`
	OIDCRedirectUrl                   string `mapstructure:"OIDC_REDIRECT_URL"`
	OIDCCookieDuration                string `mapstructure:"OIDC_COOKIE_DURATION"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
