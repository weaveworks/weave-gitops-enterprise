package app_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"testing"
	"time"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/core/logger"
	core_core "github.com/weaveworks/weave-gitops/core/server"
	"github.com/weaveworks/weave-gitops/pkg/featureflags"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	wego_server "github.com/weaveworks/weave-gitops/pkg/server"
	server_auth "github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/authfakes"
	"github.com/weaveworks/weave-gitops/pkg/services/servicesfakes"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	pipectrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/capi/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/app"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/mgmtfetcher"
	mgmtfetcherfake "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/mgmtfetcher/fake"
	"github.com/weaveworks/weave-gitops-enterprise/internal/grpctesting"
	pipepb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
)

var validEntitlement = `eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJsaWNlbmNlZFVudGlsIjoxNzg5MzgxMDE1LCJpYXQiOjE2MzE2MTQ2MTUsImlzcyI6InNhbGVzQHdlYXZlLndvcmtzIiwibmJmIjoxNjMxNjE0NjE1LCJzdWIiOiJ0ZWFtLXBlc3RvQHdlYXZlLndvcmtzIn0.klRpQQgbCtshC3PuuD4DdI3i-7Z0uSGQot23YpsETphFq4i3KK4NmgfnDg_WA3Pik-C2cJgG8WWYkWnemWQJAw`

func TestWeaveGitOpsHandlers(t *testing.T) {
	ctx := context.Background()
	defer ctx.Done()
	password := "my-secret-password"
	runtimeNamespace := "flux-system"
	c := createK8sClient(t, password, runtimeNamespace)

	port := "8001"

	client := runServer(t, ctx, c, runtimeNamespace, "0.0.0.0:"+port)

	// login
	res1, err := client.Post(fmt.Sprintf("https://localhost:%s/oauth2/sign_in", port), "application/json", bytes.NewReader([]byte(`{"username":"testsuite","password":"my-secret-password"}`)))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res1.StatusCode)

	res, err := client.Post(fmt.Sprintf("https://localhost:%s/v1/objects", port), "application/json", bytes.NewBuffer([]byte(`{"kind": "Kustomization"}`)))
	if err != nil {
		t.Fatalf("expected no errors but got: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		buf := make([]byte, 1024)
		_, _ = res.Body.Read(buf)
		t.Fatalf("expected status code to be %d but got %d instead: %v", http.StatusOK, res.StatusCode, string(buf))
	}
	res, err = client.Get(fmt.Sprintf("https://localhost:%s/v1/pineapples", port))
	if err != nil {
		t.Fatalf("expected no errors but got: %v", err)
	}
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code to be %d but got %d instead", http.StatusNotFound, res.StatusCode)
	}

}

func TestPipelinesServer(t *testing.T) {
	ctx := context.Background()
	defer ctx.Done()

	password := "my-secret-password"
	runtimeNamespace := "flux-system"
	c := createK8sClient(t, password, runtimeNamespace)

	port := "8002"

	ff := featureflags.Get("WEAVE_GITOPS_FEATURE_PIPELINES")
	t.Cleanup(func() {
		featureflags.Set("WEAVE_GITOPS_FEATURE_PIPELINES", ff)
	})
	featureflags.Set("WEAVE_GITOPS_FEATURE_PIPELINES", "true")

	client := runServer(t, ctx, c, runtimeNamespace, "0.0.0.0:"+port)

	p := &pipectrl.Pipeline{}
	p.Name = "my-pipeline"
	p.Namespace = "flux-system"

	assert.NoError(t, c.Create(ctx, p))

	res1, err := client.Post(fmt.Sprintf("https://localhost:%s/oauth2/sign_in", port), "application/json", bytes.NewReader([]byte(`{"username":"testsuite","password":"my-secret-password"}`)))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res1.StatusCode)

	res, err := client.Get(fmt.Sprintf("https://localhost:%s/v1/pipelines", port))
	assert.NoError(t, err)

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, res.StatusCode, http.StatusOK, string(body))

	msg := pipepb.ListPipelinesResponse{}

	assert.NoError(t, json.Unmarshal(body, &msg))

	assert.Len(t, msg.Pipelines, 1)
	assert.Equal(t, msg.Pipelines[0].Name, "my-pipeline")

}

func createK8sClient(t *testing.T, pw string, ns string, objects ...runtime.Object) client.Client {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	assert.NoError(t, err)

	runtimeNamespace := ns

	hashedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-user-auth",
			Namespace: runtimeNamespace,
		},
		Data: map[string][]byte{
			"username": []byte("testsuite"),
			"password": hashed,
		},
	}

	objs := []runtime.Object{createSecret(validEntitlement), hashedSecret}
	objs = append(objs, objects...)

	return createFakeClient(t, objs...)
}

func runServer(t *testing.T, ctx context.Context, k client.Client, ns string, addr string) *http.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		capiv1.AddToScheme,
		sourcev1.AddToScheme,
		pipectrl.AddToScheme,
	}

	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		t.Fatalf("expected no errors but got %v", err)
	}

	dc := discovery.NewDiscoveryClient(fakeclientset.NewSimpleClientset().Discovery().RESTClient())
	if err != nil {
		t.Fatalf("expected no errors but got %v", err)
	}

	log, err := logger.New("debug", false)
	if err != nil {
		t.Fatalf("expected no errors but got %v", err)
	}

	go func(ctx context.Context) {
		coreConfig := fakeCoreConfig(t, log)
		appsConfig := fakeAppsConfig(k, log)
		clientSet := fakeclientset.NewSimpleClientset(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "flux-system"}})
		mgmtFetcher := mgmtfetcher.NewManagementCrossNamespacesFetcher(&mgmtfetcherfake.FakeNamespaceCache{
			Namespaces: []*corev1.Namespace{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "flux-system",
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Namespace",
					},
				},
			},
		}, kubefakes.NewFakeClientGetter(k), &mgmtfetcherfake.FakeAuthClientGetter{})

		err = app.RunInProcessGateway(ctx, addr,
			app.WithCAPIClustersNamespace("default"),
			app.WithEntitlementSecretKey(client.ObjectKey{Name: "name", Namespace: "namespace"}),
			app.WithKubernetesClient(k),
			app.WithDiscoveryClient(dc),
			app.WithCoreConfig(coreConfig),
			app.WithApplicationsConfig(appsConfig),
			app.WithApplicationsOptions(wego_server.WithClientGetter(kubefakes.NewFakeClientGetter(k))),
			app.WithRuntimeNamespace(ns),
			app.WithGitProvider(git.NewGitProviderService(log)),
			app.WithClientGetter(kubefakes.NewFakeClientGetter(k)),
			app.WithAuthConfig(
				map[server_auth.AuthMethod]bool{server_auth.UserAccount: true},
				app.OIDCAuthenticationOptions{TokenDuration: time.Hour},
			),
			app.WithKubernetesClientSet(clientSet),
			app.WithClustersManager(grpctesting.MakeClustersManager(k)),
			app.WithManagemetFetcher(mgmtFetcher),
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

	return client
}

func fakeCoreConfig(t *testing.T, log logr.Logger) core_core.CoreServerConfig {

	clustersManager := &clustersmngrfakes.FakeClustersManager{}

	// A fake to support kustomizations, sorry, this is pretty frgaile and will likely break.
	clientsPool := &clustersmngrfakes.FakeClientsPool{}
	clientsPool.ClientsReturns(map[string]client.Client{})

	client := clustersmngr.NewClient(clientsPool, map[string][]corev1.Namespace{})
	clustersManager.GetImpersonatedClientReturns(client, nil)
	clustersManager.GetServerClientReturns(client, nil)

	coreConfig, err := core_core.NewCoreConfig(log, &rest.Config{}, "test", clustersManager)
	if err != nil {
		t.Fatal(err)
	}

	return coreConfig
}

func fakeAppsConfig(c client.Client, log logr.Logger) *wego_server.ApplicationsConfig {
	appFactory := &servicesfakes.FakeFactory{}
	jwtClient := &authfakes.FakeJWTClient{
		VerifyJWTStub: func(s string) (*auth.Claims, error) {
			return &auth.Claims{
				ProviderToken: "provider-token",
			}, nil
		},
	}
	return &wego_server.ApplicationsConfig{
		Factory:       appFactory,
		Logger:        log,
		JwtClient:     jwtClient,
		ClusterConfig: kube.ClusterConfig{},
	}
}

func createFakeClient(t *testing.T, clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		pipectrl.AddToScheme,
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
	tempDir, err := os.MkdirTemp("", "*")
	require.NoError(t, err)
	cmd := app.NewAPIServerCommand(logr.Discard(), tempDir)
	cmd.SetArgs([]string{
		"ui", "run",
		"--oidc-client-id=client-id",
	})

	err = cmd.Execute()
	assert.ErrorIs(t, err, app.ErrNoIssuerURL)
}

func TestNoClientID(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "*")
	require.NoError(t, err)
	cmd := app.NewAPIServerCommand(logr.Discard(), tempDir)
	cmd.SetArgs([]string{
		"ui", "run",
		"--oidc-issuer-url=http://weave.works",
	})

	err = cmd.Execute()
	assert.ErrorIs(t, err, app.ErrNoClientID)
}

func TestNoClientSecret(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "*")
	require.NoError(t, err)
	cmd := app.NewAPIServerCommand(logr.Discard(), tempDir)
	cmd.SetArgs([]string{
		"ui", "run",
		"--oidc-issuer-url=http://weave.works",
		"--oidc-client-id=client-id",
	})

	err = cmd.Execute()
	assert.ErrorIs(t, err, app.ErrNoClientSecret)
}

func TestNoRedirectURL(t *testing.T) {
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
	assert.ErrorIs(t, err, app.ErrNoRedirectURL)
}
