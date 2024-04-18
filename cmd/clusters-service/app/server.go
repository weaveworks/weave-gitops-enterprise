package app

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/alexedwards/scs/v2"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	flaggerv1beta1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
	flux_logger "github.com/fluxcd/pkg/runtime/logger"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	grpc_runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	clusterreflectorv1alpha1 "github.com/weaveworks/cluster-reflector-controller/api/v1alpha1"
	gitopssetsv1alpha1 "github.com/weaveworks/gitopssets-controller/api/v1alpha1"
	pipelinev1alpha1 "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	pacv2beta1 "github.com/weaveworks/policy-agent/api/v2beta1"
	pacv2beta2 "github.com/weaveworks/policy-agent/api/v2beta2"
	pd "github.com/weaveworks/progressive-delivery/pkg/server"
	capiv1 "github.com/weaveworks/templates-controller/apis/capi/v1alpha2"
	gapiv1 "github.com/weaveworks/templates-controller/apis/gitops/v1alpha2"
	tfctrl "github.com/weaveworks/tf-controller/api/v1alpha1"
	csgit "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/mgmtfetcher"
	capi_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/server"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/version"
	"github.com/weaveworks/weave-gitops-enterprise/common/entitlement"
	gitauth "github.com/weaveworks/weave-gitops-enterprise/pkg/api/gitauth"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/fetcher"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/namespaces"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/estimation"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/git"
	gitauth_server "github.com/weaveworks/weave-gitops-enterprise/pkg/gitauth/server"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/indexer"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/monitoring"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/monitoring/metrics"
	pipelines "github.com/weaveworks/weave-gitops-enterprise/pkg/pipelines/server"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/preview"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	queryserver "github.com/weaveworks/weave-gitops-enterprise/pkg/query/server"
	tfserver "github.com/weaveworks/weave-gitops-enterprise/pkg/terraform"
	wge_version "github.com/weaveworks/weave-gitops-enterprise/pkg/version"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	core_fetcher "github.com/weaveworks/weave-gitops/core/clustersmngr/fetcher"
	"github.com/weaveworks/weave-gitops/core/logger"
	"github.com/weaveworks/weave-gitops/core/nsaccess"
	core_core "github.com/weaveworks/weave-gitops/core/server"
	core_core_proto "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/featureflags"
	"github.com/weaveworks/weave-gitops/pkg/health"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	core "github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"github.com/weaveworks/weave-gitops/pkg/telemetry"
	"google.golang.org/protobuf/reflect/protoreflect"
	authv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	k8scache "k8s.io/client-go/tools/cache"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	defaultConfigFilename = "config"

	// Allowed login requests per second
	loginRequestRateLimit = 20

	// resync for informers to guarantee that no event was missed
	sharedFactoryResync = 20 * time.Minute
)

var (
	ErrNoIssuerURL    = errors.New("the OIDC issuer URL flag (--oidc-issuer-url) has not been set")
	ErrNoClientID     = errors.New("the OIDC client ID flag (--oidc-client-id) has not been set")
	ErrNoClientSecret = errors.New("the OIDC client secret flag (--oidc-client-secret) has not been set")
	ErrNoRedirectURL  = errors.New("the OIDC redirect URL flag (--oidc-redirect-url) has not been set")
)

func EnterprisePublicRoutes() []string {
	return core.PublicRoutes
}

// Options contains all the options for the `ui run` command.
type Params struct {
	HelmRepoNamespace                 string                    `mapstructure:"helm-repo-namespace"`
	HelmRepoName                      string                    `mapstructure:"helm-repo-name"`
	ProfileCacheLocation              string                    `mapstructure:"profile-cache-location"`
	HtmlRootPath                      string                    `mapstructure:"html-root-path"`
	RoutePrefix                       string                    `mapstructure:"route-prefix"`
	OIDC                              OIDCAuthenticationOptions `mapstructure:",squash"`
	GitProviderType                   string                    `mapstructure:"git-provider-type"`
	GitProviderHostname               string                    `mapstructure:"git-provider-hostname"`
	CAPIClustersNamespace             string                    `mapstructure:"capi-clusters-namespace"`
	CAPITemplatesNamespace            string                    `mapstructure:"capi-templates-namespace"`
	InjectPruneAnnotation             string                    `mapstructure:"inject-prune-annotation"`
	AddBasesKustomization             string                    `mapstructure:"add-bases-kustomization"`
	CAPIEnabled                       bool                      `mapstructure:"capi-enabled"`
	CAPITemplatesRepositoryUrl        string                    `mapstructure:"capi-templates-repository-url"`
	CAPIRepositoryPath                string                    `mapstructure:"capi-repository-path"`
	CAPIRepositoryClustersPath        string                    `mapstructure:"capi-repository-clusters-path"`
	CAPITemplatesRepositoryApiUrl     string                    `mapstructure:"capi-templates-repository-api-url"`
	CAPITemplatesRepositoryBaseBranch string                    `mapstructure:"capi-templates-repository-base-branch"`
	RuntimeNamespace                  string                    `mapstructure:"runtime-namespace"`
	GitProviderToken                  string                    `mapstructure:"git-provider-token"`
	AuthMethods                       []string                  `mapstructure:"auth-methods"`
	TLSCert                           string                    `mapstructure:"tls-cert"`
	TLSKey                            string                    `mapstructure:"tls-key"`
	NoTLS                             bool                      `mapstructure:"no-tls"`
	Cluster                           string                    `mapstructure:"cluster-name"`
	UseK8sCachedClients               bool                      `mapstructure:"use-k8s-cached-clients"`
	UIConfig                          string                    `mapstructure:"ui-config"`
	PipelineControllerAddress         string                    `mapstructure:"pipeline-controller-address"`
	CostEstimationFilters             string                    `mapstructure:"cost-estimation-filters"`
	CostEstimationAPIRegion           string                    `mapstructure:"cost-estimation-api-region"`
	CostEstimationFilename            string                    `mapstructure:"cost-estimation-csv-file"`
	GitProviderCSRFCookieDomain       string                    `mapstructure:"git-provider-csrf-cookie-domain"`
	GitProviderCSRFCookiePath         string                    `mapstructure:"git-provider-csrf-cookie-path"`
	GitProviderCSRFCookieDuration     time.Duration             `mapstructure:"git-provider-csrf-cookie-duration"`
	CollectorServiceAccountName       string                    `mapstructure:"collector-serviceaccount-name"`
	CollectorServiceAccountNamespace  string                    `mapstructure:"collector-serviceaccount-namespace"`
	MonitoringEnabled                 bool                      `mapstructure:"monitoring-enabled"`
	MonitoringBindAddress             string                    `mapstructure:"monitoring-bind-address"`
	MetricsEnabled                    bool                      `mapstructure:"monitoring-metrics-enabled"`
	ProfilingEnabled                  bool                      `mapstructure:"monitoring-profiling-enabled"`
	ExplorerCleanerDisabled           bool                      `mapstructure:"explorer-cleaner-disabled"`
	NoAuthUser                        string                    `mapstructure:"insecure-no-authentication-user"`
	ExplorerEnabledFor                []string                  `mapstructure:"explorer-enabled-for"`
}

type OIDCAuthenticationOptions struct {
	IssuerURL      string        `mapstructure:"oidc-issuer-url"`
	ClientID       string        `mapstructure:"oidc-client-id"`
	ClientSecret   string        `mapstructure:"oidc-client-secret"`
	RedirectURL    string        `mapstructure:"oidc-redirect-url"`
	TokenDuration  time.Duration `mapstructure:"oidc-token-duration"`
	ClaimUsername  string        `mapstructure:"oidc-claim-username"`
	ClaimGroups    string        `mapstructure:"oidc-claim-groups"`
	CustomScopes   []string      `mapstructure:"custom-oidc-scopes"`
	UsernamePrefix string        `mapstructure:"oidc-username-prefix"`
	GroupsPrefix   string        `mapstructure:"oidc-groups-prefix"`
}

func NewAPIServerCommand() *cobra.Command {
	p := &Params{}
	var logOptions flux_logger.Options

	cmd := &cobra.Command{
		Use:          "capi-server",
		Version:      fmt.Sprintf("Version: %s, Image Tag: %s", version.Version, wge_version.ImageTag),
		Long:         "The capi-server servers and handles REST operations for CAPI templates.",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := initializeConfig(cmd)
			if err != nil {
				return fmt.Errorf("error initializing viper env, %w", err)
			}
			err = viper.Unmarshal(p)
			if err != nil {
				return fmt.Errorf("error unmarshalling flags and env into config struct %w", err)
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return checkParams(*p)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return StartServer(context.Background(), *p, logOptions)
		},
	}

	cmdFlags := cmd.Flags()

	// Have to declare a flag for viper to correctly read and then bind environment variables too
	// FIXME: why? We don't actually use the flags in helm templates etc.
	//
	cmdFlags.String("helm-repo-namespace", os.Getenv("RUNTIME_NAMESPACE"), "the namespace of the Helm Repository resource to scan for profiles")
	cmdFlags.String("helm-repo-name", "weaveworks-charts", "the name of the Helm Repository resource to scan for profiles")
	cmdFlags.String("profile-cache-location", "/tmp/helm-cache", "the location where the cache Profile data lives")
	cmdFlags.String("route-prefix", "", "Mount the UI and API endpoint under a path prefix, e.g. /weave-gitops-enterprise")
	cmdFlags.String("html-root-path", "/html", "Where to serve static assets from")
	cmdFlags.String("git-provider-type", "", "")
	cmdFlags.String("git-provider-hostname", "", "")
	cmdFlags.Bool("capi-enabled", true, "")
	cmdFlags.String("capi-clusters-namespace", corev1.NamespaceAll, "where to look for GitOps cluster resources, defaults to looking in all namespaces")
	cmdFlags.String("capi-templates-namespace", "", "where to look for CAPI template resources, required")
	cmdFlags.String("inject-prune-annotation", "", "")
	cmdFlags.String("add-bases-kustomization", "enabled", "Add a kustomization to point to ./bases when creating leaf clusters")
	cmdFlags.String("capi-templates-repository-url", "", "")
	cmdFlags.String("capi-repository-path", "", "")
	cmdFlags.String("capi-repository-clusters-path", "./clusters", "")
	cmdFlags.String("capi-templates-repository-api-url", "", "")
	cmdFlags.String("capi-templates-repository-base-branch", "", "")
	cmdFlags.String("runtime-namespace", "flux-system", "Namespace hosting Gitops configuration objects (e.g. cluster-user-auth secrets)")
	cmdFlags.String("git-provider-token", "", "")
	cmdFlags.String("tls-cert-file", "", "filename for the TLS certficate, in-memory generated if omitted")
	cmdFlags.String("tls-private-key", "", "filename for the TLS key, in-memory generated if omitted")
	cmdFlags.Bool("no-tls", false, "do not attempt to read TLS certificates")
	cmdFlags.String("cluster-name", "management", "name of the management cluster")

	cmdFlags.StringSlice("auth-methods", []string{"oidc", "token-passthrough", "user-account"}, "Which auth methods to use, valid values are 'oidc', 'token-pass-through' and 'user-account'")
	cmdFlags.String("oidc-issuer-url", "", "The URL of the OpenID Connect issuer")
	cmdFlags.String("oidc-client-id", "", "The client ID for the OpenID Connect client")
	cmdFlags.String("oidc-client-secret", "", "The client secret to use with OpenID Connect issuer")
	cmdFlags.String("oidc-redirect-url", "", "The OAuth2 redirect URL")
	cmdFlags.Duration("oidc-token-duration", time.Hour, "The duration of the ID token. It should be set in the format: number + time unit (s,m,h) e.g., 20m")
	cmdFlags.String("oidc-claim-username", "", "JWT claim to use as the user name. By default email, which is expected to be a unique identifier of the end user. Admins can choose other claims, such as sub or name, depending on their provider")
	cmdFlags.String("oidc-claim-groups", "", "JWT claim to use as the user's group. If the claim is present it must be an array of strings")
	cmdFlags.StringSlice("custom-oidc-scopes", auth.DefaultScopes, "Customise the requested scopes for then OIDC authentication flow - openid will always be requested")
	cmdFlags.String("oidc-username-prefix", "", "If provided, all usernames will be prefixed with this value to prevent conflicts with other authentication strategies")
	cmdFlags.String("oidc-groups-prefix", "", "If provided, all groups will be prefixed with this value to prevent conflicts with other authentication strategies")

	cmdFlags.String("insecure-no-authentication-user", "", "A kubernetes user to impersonate for all requests, no authentication will be performed")

	cmdFlags.Bool("use-k8s-cached-clients", true, "Enables the use of cached clients")
	cmdFlags.String("ui-config", "", "UI configuration, JSON encoded")
	cmdFlags.String("pipeline-controller-address", pipelines.DefaultPipelineControllerAddress, "Pipeline controller address")

	cmdFlags.String("cost-estimation-filters", "", "Cost estimation filters")
	cmdFlags.String("cost-estimation-api-region", "", "API region for cost estimation queries")
	cmdFlags.String("cost-estimation-csv-file", "", "Filename to parse as Cost Estimation data")
	// Used to configure the cookie holding the CSRF token that gets created during the OAuth flow
	cmdFlags.String("git-provider-csrf-cookie-domain", "", "The domain of the CSRF cookie")
	cmdFlags.String("git-provider-csrf-cookie-path", "", "The path of the CSRF cookie")
	cmdFlags.Duration("git-provider-csrf-cookie-duration", 5*time.Minute, "The duration of the CSRF cookie before it expires")
	// Explorer configuration
	cmdFlags.String("collector-serviceaccount-name", "", "name of the serviceaccount that collector impersonates to watch leaf clusters.")
	cmdFlags.String("collector-serviceaccount-namespace", "", "namespace of the serviceaccount that collector impersonates to watch leaf clusters.")
	cmdFlags.Bool("explorer-cleaner-disabled", false, "Enables the Explorer object cleaner that manages retaining objects")
	cmdFlags.StringSlice("explorer-enabled-for", []string{}, "List of components that the Explorer is enabled for")

	// Monitoring
	cmdFlags.Bool("monitoring-enabled", false, "creates monitoring server")
	cmdFlags.String("monitoring-bind-address", "", "monitoring server binding address")
	cmdFlags.Bool("monitoring-metrics-enabled", false, "exposes metrics endpoint in monitoring server. requires monitoring enabled.")
	cmdFlags.Bool("monitoring-profiling-enabled", false, "exposes profiling endpoint in monitoring server. requires monitoring enabled.")

	cmdFlags.VisitAll(func(fl *flag.Flag) {
		if strings.HasPrefix(fl.Name, "cost-estimation") {
			cobra.CheckErr(cmdFlags.MarkHidden(fl.Name))
		}
	})

	logOptions.BindFlags(cmdFlags)

	return cmd
}

func checkParams(params Params) error {
	issuerURL := params.OIDC.IssuerURL
	clientID := params.OIDC.ClientID
	clientSecret := params.OIDC.ClientSecret
	redirectURL := params.OIDC.RedirectURL

	authMethods, err := auth.ParseAuthMethodArray(params.AuthMethods)
	if err != nil {
		return fmt.Errorf("could not parse auth methods while checking params: %w", err)
	}

	if !authMethods[auth.OIDC] {
		return nil
	}

	if issuerURL != "" || clientID != "" || clientSecret != "" || redirectURL != "" {
		if issuerURL == "" {
			return ErrNoIssuerURL
		}

		if clientID == "" {
			return ErrNoClientID
		}

		if clientSecret == "" {
			return ErrNoClientSecret
		}

		if redirectURL == "" {
			return ErrNoRedirectURL
		}
	}

	return nil
}

func initializeConfig(cmd *cobra.Command) error {
	// Set the base name of the config file, without the file extension.
	viper.SetConfigName(defaultConfigFilename)

	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// Align flag and env var names
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)

	// Read all env var values into viper
	viper.AutomaticEnv()

	// Read all flag values into viper
	// So they can be read from `viper.Get`, (sometimes user by weave-gitops (core))
	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		return err
	}

	return nil
}

func StartServer(ctx context.Context, p Params, logOptions flux_logger.Options) error {
	log := flux_logger.NewLogger(logOptions)

	log.Info("Starting server", "log-options", logOptions)

	featureflags.SetFromEnv(os.Environ())

	if p.CAPITemplatesNamespace == "" {
		return errors.New("CAPI templates namespace not set")
	}

	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		gapiv1.AddToScheme,
		sourcev1.AddToScheme,
		gitopsv1alpha1.AddToScheme,
		authv1.AddToScheme,
	}

	if p.CAPIEnabled {
		schemeBuilder = append(schemeBuilder, capiv1.AddToScheme)
	}

	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		return err
	}
	kubeClientConfig, err := config.GetConfig()
	if err != nil {
		return err
	}
	kubernetesClientSet, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}
	kubeClient, err := client.New(kubeClientConfig, client.Options{Scheme: scheme})
	if err != nil {
		return err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(kubeClientConfig)
	if err != nil {
		return err
	}

	appsConfig, err := gitauth_server.DefaultApplicationsConfig(log)
	if err != nil {
		return fmt.Errorf("could not create wego default config: %w", err)
	}

	chartsCache, err := helm.NewChartIndexer(p.ProfileCacheLocation, p.Cluster)
	if err != nil {
		return fmt.Errorf("could not create charts cache: %w", err)
	}

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	userPrefixes := kube.UserPrefixes{
		UsernamePrefix: p.OIDC.UsernamePrefix,
		GroupsPrefix:   p.OIDC.GroupsPrefix,
	}
	configGetter := kube.NewImpersonatingConfigGetter(kubeClientConfig, false, userPrefixes)
	clientGetter := kube.NewDefaultClientGetter(configGetter, "",
		capiv1.AddToScheme,
		pacv2beta1.AddToScheme,
		pacv2beta2.AddToScheme,
		gitopsv1alpha1.AddToScheme,
		gitopssetsv1alpha1.AddToScheme,
		clusterv1.AddToScheme,
		gapiv1.AddToScheme,
		pipelinev1alpha1.AddToScheme,
		esv1beta1.AddToScheme,
	)

	rest, clusterName, err := kube.RestConfig()
	if err != nil {
		return fmt.Errorf("could not retrieve cluster rest config: %w", err)
	}

	clustersManagerScheme, err := kube.CreateScheme()
	if err != nil {
		return fmt.Errorf("could not create scheme: %w", err)
	}

	authMethods, err := auth.ParseAuthMethodArray(p.AuthMethods)
	if err != nil {
		return fmt.Errorf("could not parse auth methods: %w", err)
	}

	builder := runtime.NewSchemeBuilder(
		capiv1.AddToScheme,
		pacv2beta1.AddToScheme,
		pacv2beta2.AddToScheme,
		esv1beta1.AddToScheme,
		flaggerv1beta1.AddToScheme,
		pipelinev1alpha1.AddToScheme,
		tfctrl.AddToScheme,
		gitopsv1alpha1.AddToScheme,
		gitopssetsv1alpha1.AddToScheme,
		clusterv1.AddToScheme,
		gapiv1.AddToScheme,
	)
	if err := builder.AddToScheme(clustersManagerScheme); err != nil {
		return err
	}

	mgmtCluster, err := cluster.NewSingleCluster(p.Cluster, rest, clustersManagerScheme, userPrefixes, cluster.DefaultKubeConfigOptions...)
	if err != nil {
		return fmt.Errorf("could not create mgmt cluster: %w", err)
	}

	if featureflags.Get("WEAVE_GITOPS_FEATURE_TELEMETRY") != "false" {
		err := telemetry.InitTelemetry(ctx, mgmtCluster)
		if err != nil {
			// If there's an error turning on telemetry, that's not a
			// thing that should interrupt anything else
			log.Info("Couldn't enable telemetry", "error", err)
		}
	}

	if p.UseK8sCachedClients {
		log.Info("Using cached clients")
		mgmtCluster = cluster.NewDelegatingCacheCluster(mgmtCluster, rest, clustersManagerScheme)
	} else {
		log.Info("Using un-cached clients")
	}

	gcf := fetcher.NewGitopsClusterFetcher(log, mgmtCluster, p.CAPIClustersNamespace, clustersManagerScheme, p.UseK8sCachedClients, userPrefixes, cluster.DefaultKubeConfigOptions...)
	scf := core_fetcher.NewSingleClusterFetcher(mgmtCluster)
	fetchers := []clustersmngr.ClusterFetcher{scf, gcf}
	if featureflags.Get("WEAVE_GITOPS_FEATURE_RUN_UI") == "true" {
		sessionFetcher := fetcher.NewRunSessionFetcher(log, mgmtCluster, clustersManagerScheme, p.UseK8sCachedClients, userPrefixes, cluster.DefaultKubeConfigOptions...)
		fetchers = append(fetchers, sessionFetcher)
	}

	clustersManager := clustersmngr.NewClustersManager(
		fetchers,
		nsaccess.NewChecker(nsaccess.DefautltWegoAppRules),
		log,
	)

	controllerContext := ctrl.SetupSignalHandler()

	indexer := indexer.NewClusterHelmIndexerTracker(chartsCache, p.Cluster, indexer.NewIndexer)
	go func() {
		err := indexer.Start(controllerContext, clustersManager, log)
		if err != nil {
			log.Error(err, "failed to start indexer")
			os.Exit(1)
		}
	}()

	clustersManager.Start(ctx)

	var estimator estimation.Estimator
	if featureflags.Get("WEAVE_GITOPS_FEATURE_COST_ESTIMATION") != "" {
		log.Info("Cost estimation feature flag is enabled")
		est, err := makeCostEstimator(ctx, log, p)
		if err != nil {
			return err
		}
		estimator = est
	}

	healthChecker := health.NewHealthChecker()
	coreCfg, err := core_core.NewCoreConfig(
		log, rest, clusterName, clustersManager, healthChecker,
	)
	if err != nil {
		return fmt.Errorf("could not create core config: %w", err)
	}

	err = coreCfg.PrimaryKinds.Add("GitOpsSet", gitopssetsv1alpha1.GroupVersion.WithKind("GitOpsSet"))
	if err != nil {
		return fmt.Errorf("failed to add GitOpsSet primary kind: %w", err)
	}

	err = coreCfg.PrimaryKinds.Add("AutomatedClusterDiscovery", clusterreflectorv1alpha1.GroupVersion.WithKind("AutomatedClusterDiscovery"))
	if err != nil {
		return fmt.Errorf("could not add AutomatedClusterDiscovery to primary kinds: %w", err)
	}

	sessionManager := scs.New()
	// TODO: Make this configurable
	sessionManager.Lifetime = 24 * time.Hour

	return RunInProcessGateway(ctx, "0.0.0.0:8000",
		WithLog(log),
		WithProfileHelmRepository(types.NamespacedName{Name: p.HelmRepoName, Namespace: p.HelmRepoNamespace}),
		WithKubernetesClient(kubeClient),
		WithDiscoveryClient(discoveryClient),
		WithGitProvider(csgit.NewGitProviderService(log)),
		WithApplicationsConfig(appsConfig),
		WithCoreConfig(coreCfg),
		WithGrpcRuntimeOptions(
			[]grpc_runtime.ServeMuxOption{
				grpc_runtime.WithIncomingHeaderMatcher(CustomIncomingHeaderMatcher),
				grpc_runtime.WithForwardResponseOption(IssueGitProviderCSRFCookie(p.GitProviderCSRFCookieDomain, p.GitProviderCSRFCookiePath, p.GitProviderCSRFCookieDuration)),
				middleware.WithGrpcErrorLogging(log),
			},
		),
		WithCAPIClustersNamespace(p.CAPIClustersNamespace),
		WithHtmlRootPath(p.HtmlRootPath),
		WithClientGetter(clientGetter),
		WithAuthConfig(authMethods, p.OIDC, p.NoAuthUser, sessionManager),
		WithTLSConfig(p.TLSCert, p.TLSKey, p.NoTLS),
		WithCAPIEnabled(p.CAPIEnabled),
		WithRuntimeNamespace(p.RuntimeNamespace),
		WithClustersManager(clustersManager),
		WithChartsCache(chartsCache),
		WithKubernetesClientSet(kubernetesClientSet),
		WithManagementCluster(p.Cluster),
		WithTemplateCostEstimator(estimator),
		WithUIConfig(p.UIConfig),
		WithPipelineControllerAddress(p.PipelineControllerAddress),
		WithCollectorServiceAccount(p.CollectorServiceAccountName, p.CollectorServiceAccountNamespace),
		WithMonitoring(p.MonitoringEnabled, p.MonitoringBindAddress, p.MetricsEnabled, p.ProfilingEnabled, log),
		WithExplorerCleanerDisabled(p.ExplorerCleanerDisabled),
		WithExplorerEnabledFor(p.ExplorerEnabledFor),
		WithRoutePrefix(p.RoutePrefix),
	)
}

// RunInProcessGateway starts the invoke in process http gateway.
func RunInProcessGateway(ctx context.Context, addr string, setters ...Option) error {
	args := defaultOptions()
	for _, setter := range setters {
		setter(args)
	}
	if args.KubernetesClient == nil {
		return errors.New("kubernetes client is not set")
	}
	if args.KubernetesClientSet == nil {
		return errors.New("kubernetes client set is not set")
	}
	if args.DiscoveryClient == nil {
		return errors.New("kubernetes discovery client is not set")
	}
	if args.GitProvider == nil {
		return errors.New("git provider is not set")
	}
	if args.ApplicationsConfig == nil {
		return errors.New("applications config is not set")
	}
	if args.ClientGetter == nil {
		return errors.New("kubernetes client getter is not set")
	}
	if args.CoreServerConfig.ClustersManager == nil {
		return errors.New("clusters manager is not set")
	}
	if args.SessionManager == nil {
		return errors.New("session manager is not set")
	}
	// TokenDuration at least should be set
	if args.OIDC.TokenDuration == 0 {
		return errors.New("OIDC configuration is not set")
	}

	estimator := estimation.NilEstimator()
	if args.Estimator != nil {
		estimator = args.Estimator
	}

	grpcMux := grpc_runtime.NewServeMux(args.GrpcRuntimeOptions...)

	factory := informers.NewSharedInformerFactory(args.KubernetesClientSet, sharedFactoryResync)
	namespacesCache, err := namespaces.NewNamespacesInformerCache(factory)
	if err != nil {
		return fmt.Errorf("failed to create informer cache for namespaces: %w", err)
	}

	userPrefixes := kube.UserPrefixes{UsernamePrefix: args.OIDC.UsernamePrefix, GroupsPrefix: args.OIDC.GroupsPrefix}
	authClientGetter, err := mgmtfetcher.NewUserConfigAuth(args.CoreServerConfig.RestCfg, args.Cluster, userPrefixes)
	if err != nil {
		return fmt.Errorf("failed to set up auth client getter")
	}
	if args.ManagementFetcher == nil {
		args.ManagementFetcher = mgmtfetcher.NewManagementCrossNamespacesFetcher(namespacesCache, args.ClientGetter, authClientGetter)
	}

	// Add weave-gitops enterprise handlers
	clusterServer := server.NewClusterServer(
		server.ServerOpts{
			Logger:                args.Log,
			ClustersManager:       args.CoreServerConfig.ClustersManager,
			GitProvider:           args.GitProvider,
			ClientGetter:          args.ClientGetter,
			DiscoveryClient:       args.DiscoveryClient,
			ClustersNamespace:     args.CAPIClustersNamespace,
			ProfileHelmRepository: args.ProfileHelmRepository,
			CAPIEnabled:           args.CAPIEnabled,
			ChartJobs:             helm.NewJobs(),
			ChartsCache:           args.ChartsCache,
			ValuesFetcher:         helm.NewValuesFetcher(),
			RestConfig:            args.CoreServerConfig.RestCfg,
			ManagementFetcher:     args.ManagementFetcher,
			Cluster:               args.Cluster,
			Estimator:             estimator,
			UIConfig:              args.UIConfig,
		},
	)
	if err := capi_proto.RegisterClustersServiceHandlerServer(ctx, grpcMux, clusterServer); err != nil {
		return fmt.Errorf("failed to register clusters service handler server: %w", err)
	}

	//Add weave-gitops core handlers
	wegoApplicationServer := gitauth_server.NewApplicationsServer(args.ApplicationsConfig, args.ApplicationsOptions...)
	if err := gitauth.RegisterGitAuthHandlerServer(ctx, grpcMux, wegoApplicationServer); err != nil {
		return fmt.Errorf("failed to register application handler server: %w", err)
	}

	// Add logging middleware
	grpcHttpHandler := middleware.WithLogging(args.Log, grpcMux)

	appsServer, err := core_core.NewCoreServer(args.CoreServerConfig)
	if err != nil {
		return fmt.Errorf("unable to create new kube client: %w", err)
	}

	if err = core_core_proto.RegisterCoreHandlerServer(ctx, grpcMux, appsServer); err != nil {
		return fmt.Errorf("could not register new app server: %w", err)
	}

	// Add progressive-delivery handlers
	if err := pd.Hydrate(ctx, grpcMux, pd.ServerOpts{
		ClustersManager: args.CoreServerConfig.ClustersManager,
		Logger:          args.Log,
	}); err != nil {
		return fmt.Errorf("failed to register progressive delivery handler server: %w", err)
	}

	if featureflags.Get("WEAVE_GITOPS_FEATURE_EXPLORER") != "" {
		_, err := queryserver.Hydrate(ctx, grpcMux, queryserver.ServerOpts{
			Logger:              args.Log,
			DiscoveryClient:     args.DiscoveryClient,
			ClustersManager:     args.ClustersManager,
			SkipCollection:      false,
			ObjectKinds:         configuration.SupportedObjectKinds,
			ServiceAccount:      args.CollectorServiceAccount,
			EnableObjectCleaner: !args.ExplorerCleanerDisabled,
			EnabledFor:          args.ExplorerEnabledFor,
		})
		if err != nil {
			return fmt.Errorf("hydrating query server: %w", err)
		}
	}

	if featureflags.Get("WEAVE_GITOPS_FEATURE_PIPELINES") != "" {
		if err := pipelines.Hydrate(ctx, grpcMux, pipelines.ServerOpts{
			Logger:                    args.Log,
			ClustersManager:           args.ClustersManager,
			ManagementFetcher:         args.ManagementFetcher,
			Cluster:                   args.Cluster,
			PipelineControllerAddress: args.PipelineControllerAddress,
			GitProvider:               args.GitProvider,
		}); err != nil {
			return fmt.Errorf("hydrating pipelines server: %w", err)
		}
	}

	if featureflags.Get("WEAVE_GITOPS_FEATURE_TERRAFORM_UI") != "" {
		if err := tfserver.Hydrate(ctx, grpcMux, tfserver.ServerOpts{
			Logger:         args.Log,
			ClientsFactory: args.ClustersManager,
			Scheme:         args.KubernetesClient.Scheme(),
		}); err != nil {
			return fmt.Errorf("hydrating terraform server: %w", err)
		}
	}

	if err := preview.Hydrate(ctx, grpcMux, preview.ServerOpts{
		Logger:          args.Log,
		ProviderCreator: git.NewFactory(args.Log),
	}); err != nil {
		return fmt.Errorf("hydrating preview server")
	}

	// UI
	args.Log.Info("Attaching FileServer", "HtmlRootPath", args.HtmlRootPath)

	assetFS := os.DirFS(args.HtmlRootPath)
	assertFSHandler := http.FileServer(http.FS(assetFS))
	redirectHandler := core.IndexHTMLHandler(assetFS, args.Log, args.RoutePrefix)
	assetHandler := core.AssetHandler(assertFSHandler, redirectHandler)

	mux := http.NewServeMux()

	_, err = url.Parse(args.OIDC.IssuerURL)
	if err != nil {
		return fmt.Errorf("invalid issuer URL: %w", err)
	}

	_, err = url.Parse(args.OIDC.RedirectURL)
	if err != nil {
		return fmt.Errorf("invalid redirect URL: %w", err)
	}

	tsv, err := auth.NewHMACTokenSignerVerifier(args.OIDC.TokenDuration)
	if err != nil {
		return fmt.Errorf("could not create HMAC token signer: %w", err)
	}

	authMethods := args.AuthMethods
	if args.NoAuthUser != "" {
		args.Log.V(logger.LogLevelWarn).Info("Anonymous mode enabled", "noAuthUser", args.NoAuthUser)
		authMethods = map[auth.AuthMethod]bool{auth.Anonymous: true}
	}

	if len(authMethods) == 0 {
		return errors.New("no authentication methods set")
	}

	// FIXME: Slightly awkward bit of logging..
	authMethodsStrings := []string{}
	for authMethod, enabled := range authMethods {
		if enabled {
			authMethodsStrings = append(authMethodsStrings, authMethod.String())
		}
	}
	args.Log.Info("setting enabled auth methods", "enabled", authMethodsStrings)

	if len(args.OIDC.CustomScopes) != 0 {
		args.Log.Info("setting custom OIDC scopes", "scopes", args.OIDC.CustomScopes)
	}

	authServerConfig, err := auth.NewAuthServerConfig(
		args.Log,
		auth.OIDCConfig{
			IssuerURL:     args.OIDC.IssuerURL,
			ClientID:      args.OIDC.ClientID,
			ClientSecret:  args.OIDC.ClientSecret,
			RedirectURL:   args.OIDC.RedirectURL,
			TokenDuration: args.OIDC.TokenDuration,
			ClaimsConfig: &auth.ClaimsConfig{
				Username: args.OIDC.ClaimUsername,
				Groups:   args.OIDC.ClaimGroups,
			},
			UsernamePrefix: args.OIDC.UsernamePrefix,
			GroupsPrefix:   args.OIDC.GroupsPrefix,
			Scopes:         args.OIDC.CustomScopes,
		},
		args.KubernetesClient,
		tsv,
		args.RuntimeNamespace,
		authMethods,
		args.NoAuthUser,
		args.SessionManager,
	)
	if err != nil {
		return fmt.Errorf("could not create auth server: %w", err)
	}
	srv, err := auth.NewAuthServer(
		ctx,
		authServerConfig,
	)
	if err != nil {
		return fmt.Errorf("could not create auth server: %w", err)
	}

	args.Log.Info("Registering callback route")
	if err := auth.RegisterAuthServer(mux, "/oauth2", srv, loginRequestRateLimit); err != nil {
		return fmt.Errorf("failed to register auth routes: %w", err)
	}

	// Secure `/v1` and `/gitops/api` API routes
	grpcHttpHandler = auth.WithAPIAuth(grpcHttpHandler, srv, EnterprisePublicRoutes(), args.SessionManager)

	// monitoring server
	var monitoringServer *http.Server
	if args.MonitoringOptions.Enabled {
		monitoringServer, err = monitoring.NewServer(args.MonitoringOptions)
		if err != nil {
			return fmt.Errorf("cannot create monitoring server: %w", err)
		}
		args.Log.Info("monitoring server started")
	}

	commonMiddleware := func(mux http.Handler) http.Handler {
		wrapperHandler := middleware.WithProviderToken(args.ApplicationsConfig.JwtClient, mux, args.Log)

		if args.MonitoringOptions.MetricsOptions.Enabled {
			wrapperHandler = metrics.WithHttpMetrics(wrapperHandler)
		}

		return entitlement.EntitlementHandler(
			ctx,
			args.Log,
			args.KubernetesClient,
			args.EntitlementSecretKey,
			entitlement.CheckEntitlementHandler(args.Log, wrapperHandler, EnterprisePublicRoutes()),
		)
	}

	mux.Handle("/v1/", commonMiddleware(grpcHttpHandler))

	staticAssetsWithGz := gziphandler.GzipHandler(assetHandler)

	mux.Handle("/", staticAssetsWithGz)

	if args.RoutePrefix != "" {
		mux = core.WithRoutePrefix(mux, args.RoutePrefix)
	}

	handler := http.Handler(mux)
	handler = args.SessionManager.LoadAndSave(handler)

	s := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	factoryStopCh := make(chan struct{})
	factory.Start(factoryStopCh)
	k8scache.WaitForCacheSync(factoryStopCh,
		namespacesCache.CacheSync(),
	)

	go func() {
		<-ctx.Done()
		args.Log.Info("Shutting down the http gateway server")
		close(factoryStopCh)
		if err := s.Shutdown(context.Background()); err != nil {
			args.Log.Error(err, "Failed to shutdown http gateway server")
		}
		if args.MonitoringOptions.Enabled && monitoringServer != nil {
			if err := monitoringServer.Shutdown(ctx); err != nil {
				args.Log.Error(err, "Failed to shutdown management server")
			}
		}
	}()

	args.Log.Info("Starting to listen and serve", "address", addr)

	if err := ListenAndServe(s, args.NoTLS, args.TLSCert, args.TLSKey, args.Log); err != http.ErrServerClosed {
		args.Log.Error(err, "Failed to listen and serve")
		return err
	}
	return nil
}

func TLSConfig(hosts []string) (*tls.Config, error) {
	certPEMBlock, keyPEMBlock, err := generateKeyPair(hosts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS keys %w", err)
	}

	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to generate X509 key pair %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	return tlsConfig, nil
}

// Adapted from https://go.dev/src/crypto/tls/generate_cert.go
func generateKeyPair(hosts []string) ([]byte, []byte, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failing to generate new ecdsa key: %w", err)
	}

	// A CA is supposed to choose unique serial numbers, that is, unique for the CA.
	maxSerialNumber := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, maxSerialNumber)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate a random serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Weaveworks"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	certPEMBlock := &bytes.Buffer{}

	err = pem.Encode(certPEMBlock, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode cert pem: %w", err)
	}

	keyPEMBlock := &bytes.Buffer{}

	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to marshal ECDSA private key: %v", err)
	}

	err = pem.Encode(keyPEMBlock, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode key pem: %w", err)
	}

	return certPEMBlock.Bytes(), keyPEMBlock.Bytes(), nil
}

func ListenAndServe(srv *http.Server, noTLS bool, tlsCert, tlsKey string, log logr.Logger) error {
	if noTLS {
		log.Info("TLS connections disabled")
		return srv.ListenAndServe()
	}

	if tlsCert == "" && tlsKey == "" {
		log.Info("TLS cert and key not specified, generating and using in-memory keys")

		tlsConfig, err := TLSConfig([]string{"localhost", "0.0.0.0", "127.0.0.1", "weave.gitops.enterprise.com"})
		if err != nil {
			return fmt.Errorf("failed to generate a TLSConfig: %w", err)
		}

		srv.TLSConfig = tlsConfig
		// if TLSCert and TLSKey are both empty (""), ListenAndServeTLS will ignore
		// and happily use the TLSConfig supplied above
		return srv.ListenAndServeTLS("", "")
	}

	if tlsCert == "" || tlsKey == "" {
		return cmderrors.ErrNoTLSCertOrKey
	}

	log.Info("Using TLS", "cert", tlsCert, "key", tlsKey)

	return srv.ListenAndServeTLS(tlsCert, tlsKey)
}

// CustomIncomingHeaderMatcher allows the Accept header to be passed to the gRPC handlers.
// The Accept header is used by the gRPC handlers to determine whether a response other
// than `application/json` is requested.
func CustomIncomingHeaderMatcher(key string) (string, bool) {
	switch key {
	case "Accept":
		return key, true
	default:
		return grpc_runtime.DefaultHeaderMatcher(key)
	}
}

func defaultOptions() *Options {
	return &Options{
		Log: logr.Discard(),
	}
}

func makeCostEstimator(ctx context.Context, log logr.Logger, p Params) (estimation.Estimator, error) {
	var pricer estimation.Pricer
	if p.CostEstimationFilename != "" {
		log.Info("configuring cost estimation from CSV", "filename", p.CostEstimationFilename)
		pr, err := estimation.NewCSVPricerFromFile(log, p.CostEstimationFilename)
		if err != nil {
			return nil, err
		}
		pricer = pr
	} else {
		if p.CostEstimationFilters == "" {
			return nil, errors.New("cost estimation filters cannot be empty")
		}
		cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(p.CostEstimationAPIRegion))
		if err != nil {
			log.Error(err, "unable to load AWS SDK config, cost estimation will not be available")
		} else {
			svc := pricing.NewFromConfig(cfg)
			pricer = estimation.NewAWSPricer(log, svc)
		}
	}
	log.Info("Setting default cost estimation filters", "filters", p.CostEstimationFilters)
	filters, err := estimation.ParseFilterQueryString(p.CostEstimationFilters)
	if err != nil {
		return nil, fmt.Errorf("could not parse cost estimation filters: %w", err)
	}
	log.Info("Parsed default cost estimation filters", "filters", filters)

	return estimation.NewAWSClusterEstimator(pricer, filters), nil
}

// IssueGitProviderCSRFCookie gets executed before sending the HTTP response and checks if any gRPC handlers have
// previously set any custom metadata with the key `x-git-provider-csrf`. If so, it will read its value and issue
// an HTTP cookie with that value. The cookie will be read during the OAuth flow, when the user authenticates to a
// git provider in order to receive a short-lived token.
func IssueGitProviderCSRFCookie(domain string, path string, duration time.Duration) func(ctx context.Context, w http.ResponseWriter, p protoreflect.ProtoMessage) error {
	return func(ctx context.Context, w http.ResponseWriter, p protoreflect.ProtoMessage) error {
		md, ok := grpc_runtime.ServerMetadataFromContext(ctx)
		if !ok {
			return nil
		}

		if vals := md.HeaderMD.Get(gitauth_server.GitProviderCSRFHeaderName); len(vals) > 0 {
			state := vals[0]
			md.HeaderMD.Delete(gitauth_server.GitProviderCSRFHeaderName)
			w.Header().Del(grpc_runtime.MetadataHeaderPrefix + gitauth_server.GitProviderCSRFHeaderName)
			cookie := &http.Cookie{
				Name:     gitauth_server.GitProviderCSRFCookieName,
				Value:    state,
				Domain:   domain,
				Path:     path,
				Secure:   true,
				HttpOnly: true,
				Expires:  time.Now().UTC().Add(duration),
			}
			http.SetCookie(w, cookie)
		}

		return nil
	}
}
