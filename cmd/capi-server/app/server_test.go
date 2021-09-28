package app_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/app"
	"github.com/weaveworks/weave-gitops/pkg/apputils/apputilsfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	wego_server "github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/authfakes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var validEntitlement = `eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJsaWNlbmNlZFVudGlsIjoxNzg5MzgxMDE1LCJpYXQiOjE2MzE2MTQ2MTUsImlzcyI6InNhbGVzQHdlYXZlLndvcmtzIiwibmJmIjoxNjMxNjE0NjE1LCJzdWIiOiJ0ZWFtLXBlc3RvQHdlYXZlLndvcmtzIn0.klRpQQgbCtshC3PuuD4DdI3i-7Z0uSGQot23YpsETphFq4i3KK4NmgfnDg_WA3Pik-C2cJgG8WWYkWnemWQJAw`

func TestWeaveGitOpsHandlers(t *testing.T) {
	ctx := context.Background()
	defer ctx.Done()

	c := createFakeClient(createSecret(validEntitlement))
	go func(ctx context.Context) {
		appsConfig := fakeAppsConfig(c)
		err := app.RunInProcessGateway(ctx, "0.0.0.0:8001", nil, nil, c, nil, nil, "default", appsConfig, client.ObjectKey{Name: "name", Namespace: "namespace"}, logr.Discard())
		t.Logf("%v", err)

	}(ctx)

	time.Sleep(100 * time.Millisecond)
	res, err := http.Get("http://localhost:8001/v1/applications")
	if err != nil {
		t.Fatalf("expected no errors but got: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status code to be %d but got %d instead", http.StatusOK, res.StatusCode)
	}
	res, err = http.Get("http://localhost:8001/v1/pineapples")
	if err != nil {
		t.Fatalf("expected no errors but got: %v", err)
	}
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code to be %d but got %d instead", http.StatusNotFound, res.StatusCode)
	}
}

func fakeAppsConfig(c client.Client) *wego_server.ApplicationsConfig {
	appFactory := &apputilsfakes.FakeAppFactory{}
	kubeClient := &kubefakes.FakeKube{}
	jwtClient := &authfakes.FakeJWTClient{
		VerifyJWTStub: func(s string) (*auth.Claims, error) {
			return &auth.Claims{
				ProviderToken: "provider-token",
			}, nil
		},
	}
	appFactory.GetKubeServiceStub = func() (kube.Kube, error) {
		return kubeClient, nil
	}
	return &wego_server.ApplicationsConfig{
		AppFactory: appFactory,
		KubeClient: c,
		Logger:     logr.Discard(),
		JwtClient:  jwtClient,
	}
}

func createFakeClient(clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
	}
	schemeBuilder.AddToScheme(scheme)

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(clusterState...).
		Build()

	return c
}

func createSecret(s string) *corev1.Secret {
	// When reading a secret, only Data contains any data, StringData is empty
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "name",
			Namespace: "namespace",
		},
		Type: "Opaque",
		Data: map[string][]byte{"entitlement": []byte(s)},
	}
}
