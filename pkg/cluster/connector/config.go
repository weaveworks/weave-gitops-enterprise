package connector

import (
	"fmt"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// ConfigForContext will return the kube config given a context name and set of path options if exists
func ConfigForContext(pathOpts *clientcmd.PathOptions, contextName string) (*rest.Config, error) {
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

	return clientConfig.ClientConfig()
}

// kubeConfigWithToken takes a rest.Config and generates a KubeConfig with the
// named context and configured user credentials from the provided token.
func kubeConfigWithToken(config *rest.Config, context string, token []byte) (*clientcmdapi.Config, error) {
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

	return cfg, nil
}
