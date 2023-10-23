package steps

import (
	"testing"

	"github.com/weaveworks/weave-gitops-enterprise/test/utils"
	"k8s.io/apimachinery/pkg/runtime"
)

const testNamespace = "test-ns"

func makeTestConfig(t *testing.T, config Config, objects ...runtime.Object) Config {
	fakeClient := utils.CreateFakeClient(t, objects...)
	cliLogger := utils.CreateLogger()
	return Config{
		KubernetesClient:   fakeClient,
		Logger:             cliLogger,
		Namespace:          config.Namespace,
		WGEVersion:         config.WGEVersion,
		DomainType:         config.DomainType,
		Username:           config.UserDomain,
		Password:           config.Password,
		UserDomain:         config.UserDomain,
		PrivateKeyPath:     config.PrivateKeyPath,
		PrivateKeyPassword: config.PrivateKeyPassword,
	}
}
