package connector

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ConfigForContext will return the kube config given a context name and set of path options if exists
func ConfigForContext(ctx context.Context, pathOpts *clientcmd.PathOptions, contextName string) (*rest.Config, error) {
	logger := log.FromContext(ctx)
	config, err := pathOpts.GetStartingConfig()
	if err != nil {
		return nil, err
	}

	configContext := config.Contexts[contextName]
	if configContext == nil {
		return nil, fmt.Errorf("failed to get context %s", contextName)
	}

	overrides := clientcmd.ConfigOverrides{
		Context: *configContext,
	}
	clientConfig := clientcmd.NewDefaultClientConfig(*config, &overrides)
	logger.Info("Config for context retrieved", "context", contextName)
	return clientConfig.ClientConfig()
}

// kubeConfigWithToken takes a rest.Config and generates a KubeConfig with the
// named context and configured user credentials from the provided token.
func kubeConfigWithToken(ctx context.Context, config *rest.Config, context string, token []byte) (*clientcmdapi.Config, error) {
	logger := log.FromContext(ctx)
	clusterName := context + "-cluster"
	username := clusterName + "-user"

	cfg := clientcmdapi.NewConfig()
	cfg.Kind = ""       // legacy field
	cfg.APIVersion = "" // legacy field
	cfg.Clusters[context] = &clientcmdapi.Cluster{
		Server:                   config.Host,
		CertificateAuthorityData: config.CAData,
		InsecureSkipTLSVerify:    config.Insecure,
	}
	cfg.AuthInfos[username] = &clientcmdapi.AuthInfo{
		Token: string(token),
	}
	cfg.Contexts[context] = &clientcmdapi.Context{
		Cluster:  context,
		AuthInfo: username,
	}
	cfg.CurrentContext = context
	logger.Info("kubeconfig with token generated successfully")

	return cfg, nil
}

// Load kubeconfig from provided path or load in-cluster config or using default recommended locations
func loadConfig(kubeconfigPath string) (*rest.Config, error) {

	// If a kubeconfig flag is specified with the config location, use that
	if len(kubeconfigPath) > 0 {
		loader := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
		deferedClientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, nil)
		config, err := deferedClientConfig.ClientConfig()
		if err != nil {
			return nil, err
		}
		return config, nil
	}

	// If the recommended kubeconfig env variable is not specified,
	// try the in-cluster config.
	kubeconfigPath = os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
	if len(kubeconfigPath) == 0 {
		config, err := rest.InClusterConfig()
		if err == nil && config != nil {
			return config, nil
		}
	}

	// If kubeconfig env variable is set, or there is no in-cluster config,
	// try the default recommended locations.
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if _, ok := os.LookupEnv("HOME"); !ok {
		u, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("could not get current user: %w", err)
		}
		loadingRules.Precedence = append(loadingRules.Precedence, filepath.Join(u.HomeDir, clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName))
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, nil).ClientConfig()
	if err != nil {
		return nil, err
	}
	return config, nil

}
