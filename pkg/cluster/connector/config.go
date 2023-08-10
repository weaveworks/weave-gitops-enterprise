package connector

import (
	"fmt"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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
