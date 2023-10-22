package utils

import (
	"os"
	"testing"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// CreateFakeClient create test fake client
func CreateFakeClient(t *testing.T, objects ...runtime.Object) (client.WithWatch, error) {
	scheme, err := kube.CreateScheme()
	if err != nil {
		return nil, err
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
	return fakeClient, nil
}

// CreateLogger create logger
func CreateLogger() logger.Logger {
	return logger.NewCLILogger(os.Stdout)
}
