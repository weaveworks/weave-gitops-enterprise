package client

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetClient(kubeconfigFile string) (kubernetes.Interface, error) {
	var config *rest.Config
	var err error
	if kubeconfigFile != "" {
		// Out-of-cluster config
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigFile)
		if err != nil {
			log.Errorf("Unable to connect to Kubernetes API using out-of-cluster config: %v.", err)
			return nil, err
		}
	} else {
		// In-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Errorf("Unable to connect to Kubernetes API using in-cluster config: %v.", err)
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Errorf("Unable to create a client to %s: %v.", config.Host, err)
		return nil, err
	}

	log.Infof("Kubernetes host: %s", config.Host)
	return clientset, nil
}
