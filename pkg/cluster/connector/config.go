package connector

import (
	"context"
	"fmt"

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
