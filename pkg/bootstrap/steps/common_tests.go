package steps

import (
	"testing"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"k8s.io/apimachinery/pkg/runtime"
)

func makeTestConfig(t *testing.T, config Config, objects ...runtime.Object) (Config, error) {
	fakeClient, err := utils.CreateFakeClient(t, objects...)
	if err != nil {
		return Config{}, err
	}
	cliLogger := utils.CreateLogger()
	config = Config{
		KubernetesClient:   fakeClient,
		Logger:             cliLogger,
		WGEVersion:         config.WGEVersion,
		DomainType:         config.DomainType,
		Username:           config.UserDomain,
		Password:           config.Password,
		UserDomain:         config.UserDomain,
		PrivateKeyPath:     config.PrivateKeyPath,
		PrivateKeyPassword: config.PrivateKeyPassword,
	}
	return config, err
}
