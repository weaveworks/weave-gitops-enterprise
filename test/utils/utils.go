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

// CreateFakeClient create fake client for testing with default wge schemes
func CreateFakeClient(t *testing.T, clusterState ...runtime.Object) client.Client {
	t.Helper()
	scheme, err := kube.CreateScheme()
	if err != nil {
		t.Fatalf("error creating fake client: %v", err)
	}

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(clusterState...).
		Build()
}

// CreateLogger create logger
func CreateLogger() logger.Logger {
	return logger.NewCLILogger(os.Stdout)
}
