package app_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-logr/logr"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/app"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	wego_server "github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/services/applicationv2"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/authfakes"
	"github.com/weaveworks/weave-gitops/pkg/services/servicesfakes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var validEntitlement = `eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJsaWNlbmNlZFVudGlsIjoxNzg5MzgxMDE1LCJpYXQiOjE2MzE2MTQ2MTUsImlzcyI6InNhbGVzQHdlYXZlLndvcmtzIiwibmJmIjoxNjMxNjE0NjE1LCJzdWIiOiJ0ZWFtLXBlc3RvQHdlYXZlLndvcmtzIn0.klRpQQgbCtshC3PuuD4DdI3i-7Z0uSGQot23YpsETphFq4i3KK4NmgfnDg_WA3Pik-C2cJgG8WWYkWnemWQJAw`

func TestWeaveGitOpsHandlers(t *testing.T) {
	ctx := context.Background()
	defer ctx.Done()

	c := createFakeClient(t, createSecret(validEntitlement))
	db, err := utils.Open("", "sqlite", "", "", "")
	if err != nil {
		t.Fatalf("expected no errors but got %v", err)
	}
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		capiv1.AddToScheme,
		sourcev1beta1.AddToScheme,
	}
	err = schemeBuilder.AddToScheme(scheme)
	if err != nil {
		t.Fatalf("expected no errors but got %v", err)
	}

	dc := discovery.NewDiscoveryClient(fakeclientset.NewSimpleClientset().Discovery().RESTClient())

	if err != nil {
		t.Fatalf("expected no errors but got %v", err)
	}
	go func(ctx context.Context) {
		appsConfig := fakeAppsConfig(c)
		err := app.RunInProcessGateway(ctx, "0.0.0.0:8001",
			app.WithCAPIClustersNamespace("default"),
			app.WithEntitlementSecretKey(client.ObjectKey{Name: "name", Namespace: "namespace"}),
			app.WithKubernetesClient(c),
			app.WithDiscoveryClient(dc),
			app.WithDatabase(db),
			app.WithApplicationsConfig(appsConfig),
			app.WithApplicationsOptions(wego_server.WithClientGetter(NewFakeClientGetter(c))),
			app.WithTemplateLibrary(&templates.CRDLibrary{
				Log:       logr.Discard(),
				Client:    c,
				Namespace: "default",
			}),
			app.WithGitProvider(git.NewGitProviderService(logr.Discard())),
		)
		t.Logf("%v", err)
	}(ctx)

	time.Sleep(1 * time.Second)
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
	appFactory := &servicesfakes.FakeFactory{}
	kubeClient := &kubefakes.FakeKube{}
	k8s := fake.NewClientBuilder().WithScheme(kube.CreateScheme()).Build()
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
		Factory:        appFactory,
		FetcherFactory: NewFakeFetcherFactory(applicationv2.NewFetcher(k8s)),
		Logger:         logr.Discard(),
		JwtClient:      jwtClient,
		ClusterConfig:  wego_server.ClusterConfig{},
	}
}

func createFakeClient(t *testing.T, clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
	}
	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		t.Fatal(err)
	}

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
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{"entitlement": []byte(s)},
	}
}

// FIXME: expose this in wego core

type FakeFetcherFactory struct {
	fake applicationv2.Fetcher
}

func NewFakeFetcherFactory(fake applicationv2.Fetcher) wego_server.FetcherFactory {
	return &FakeFetcherFactory{
		fake: fake,
	}
}

func (f *FakeFetcherFactory) Create(client client.Client) applicationv2.Fetcher {
	return f.fake
}

type FakeClientGetter struct {
	client client.Client
}

func NewFakeClientGetter(client client.Client) wego_server.ClientGetter {
	return &FakeClientGetter{
		client: client,
	}
}

func (g *FakeClientGetter) Client(ctx context.Context) (client.Client, error) {
	return g.client, nil
}
