package app_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"net/http"
	"net/http/cookiejar"
	"os"
	"testing"
	"time"

	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/app"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	wego_server "github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/services/applicationv2"
	"github.com/weaveworks/weave-gitops/pkg/services/applicationv2/applicationv2fakes"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/authfakes"
	"github.com/weaveworks/weave-gitops/pkg/services/servicesfakes"
	"golang.org/x/crypto/bcrypt"
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
	os.Setenv("WEAVE_GITOPS_AUTH_ENABLED", "true")
	defer os.Unsetenv("WEAVE_GITOPS_AUTH_ENABLED")

	ctx := context.Background()
	defer ctx.Done()

	password := "my-secret-password"
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	hashedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "admin-password-hash",
			Namespace: "wego-system",
		},
		Data: map[string][]byte{
			"password": hashed,
		},
	}

	c := createFakeClient(t, createSecret(validEntitlement), hashedSecret)
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
			app.WithApplicationsOptions(wego_server.WithClientGetter(kubefakes.NewFakeClientGetter(c))),
			app.WithTemplateLibrary(&templates.CRDLibrary{
				Log:          logr.Discard(),
				ClientGetter: kubefakes.NewFakeClientGetter(c),
				Namespace:    "default",
			}),
			app.WithGitProvider(git.NewGitProviderService(logr.Discard())),
			app.WithClientGetter(kubefakes.NewFakeClientGetter(c)),
			app.WithOIDCConfig(app.OIDCAuthenticationOptions{TokenDuration: time.Hour}),
		)
		t.Logf("%v", err)
	}(ctx)

	jar, err := cookiejar.New(&cookiejar.Options{})
	assert.NoError(t, err)
	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	time.Sleep(1 * time.Second)

	// Check this route is public
	res, err := client.Get("https://localhost:8001/gitops/api/agent.yaml?token=derp")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	// login
	res1, err := client.Post("https://localhost:8001/oauth2/sign_in", "application/json", bytes.NewReader([]byte(`{"password":"my-secret-password"}`)))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res1.StatusCode)

	res, err = client.Get("https://localhost:8001/v1/applications")
	if err != nil {
		t.Fatalf("expected no errors but got: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status code to be %d but got %d instead", http.StatusOK, res.StatusCode)
	}
	res, err = client.Get("https://localhost:8001/v1/pineapples")
	if err != nil {
		t.Fatalf("expected no errors but got: %v", err)
	}
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code to be %d but got %d instead", http.StatusNotFound, res.StatusCode)
	}

}

func fakeAppsConfig(c client.Client) *wego_server.ApplicationsConfig {
	appFactory := &servicesfakes.FakeFactory{}
	k8s := fake.NewClientBuilder().WithScheme(kube.CreateScheme()).Build()
	jwtClient := &authfakes.FakeJWTClient{
		VerifyJWTStub: func(s string) (*auth.Claims, error) {
			return &auth.Claims{
				ProviderToken: "provider-token",
			}, nil
		},
	}
	return &wego_server.ApplicationsConfig{
		Factory:        appFactory,
		FetcherFactory: applicationv2fakes.NewFakeFetcherFactory(applicationv2.NewFetcher(k8s)),
		Logger:         logr.Discard(),
		JwtClient:      jwtClient,
		ClusterConfig:  kube.ClusterConfig{},
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

func TestNoIssuerURL(t *testing.T) {
	os.Setenv("WEAVE_GITOPS_AUTH_ENABLED", "true")
	defer os.Unsetenv("WEAVE_GITOPS_AUTH_ENABLED")

	tempDir, err := os.MkdirTemp("", "*")
	require.NoError(t, err)
	cmd := app.NewAPIServerCommand(logr.Discard(), tempDir)
	cmd.SetArgs([]string{
		"ui", "run",
		"--oidc-client-id=client-id",
	})

	err = cmd.Execute()
	assert.ErrorIs(t, err, cmderrors.ErrNoIssuerURL)
}

func TestNoClientID(t *testing.T) {
	os.Setenv("WEAVE_GITOPS_AUTH_ENABLED", "true")
	defer os.Unsetenv("WEAVE_GITOPS_AUTH_ENABLED")

	tempDir, err := os.MkdirTemp("", "*")
	require.NoError(t, err)
	cmd := app.NewAPIServerCommand(logr.Discard(), tempDir)
	cmd.SetArgs([]string{
		"ui", "run",
		"--oidc-issuer-url=http://weave.works",
	})

	err = cmd.Execute()
	assert.ErrorIs(t, err, cmderrors.ErrNoClientID)
}

func TestNoClientSecret(t *testing.T) {
	os.Setenv("WEAVE_GITOPS_AUTH_ENABLED", "true")
	defer os.Unsetenv("WEAVE_GITOPS_AUTH_ENABLED")

	tempDir, err := os.MkdirTemp("", "*")
	require.NoError(t, err)
	cmd := app.NewAPIServerCommand(logr.Discard(), tempDir)
	cmd.SetArgs([]string{
		"ui", "run",
		"--oidc-issuer-url=http://weave.works",
		"--oidc-client-id=client-id",
	})

	err = cmd.Execute()
	assert.ErrorIs(t, err, cmderrors.ErrNoClientSecret)
}

func TestNoRedirectURL(t *testing.T) {
	os.Setenv("WEAVE_GITOPS_AUTH_ENABLED", "true")
	defer os.Unsetenv("WEAVE_GITOPS_AUTH_ENABLED")

	tempDir, err := os.MkdirTemp("", "*")
	require.NoError(t, err)
	defer os.Remove(tempDir)
	cmd := app.NewAPIServerCommand(logr.Discard(), tempDir)
	cmd.SetArgs([]string{
		"ui", "run",
		"--oidc-issuer-url=http://weave.works",
		"--oidc-client-id=client-id",
		"--oidc-client-secret=client-secret",
	})

	err = cmd.Execute()
	assert.ErrorIs(t, err, cmderrors.ErrNoRedirectURL)
}
