package connector

import (
	"fmt"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ConnectCluster will ensure that the cluster referenced by newContext can be
// accessed from the cluster pointed to by hubContext.
//
// func ConnectCluster(ctx context.Context, kubeconfigName, newContext, hubContext string) error {
// }

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
