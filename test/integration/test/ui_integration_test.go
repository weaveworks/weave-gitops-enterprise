package test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gitopsv1 "github.com/weaveworks/cluster-controller/api/v1alpha1"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	core_core "github.com/weaveworks/weave-gitops/core/server"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	wego_server "github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/authfakes"
	"github.com/weaveworks/weave-gitops/pkg/services/servicesfakes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/capi/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/app"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/clusters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
	acceptancetest "github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test"
)

//
// Test suite
//

const capiServerPort = "8000"
const uiURL = "http://localhost:5000"
const seleniumURL = "http://localhost:4444/wd/hub"

const entitlement = `eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJsaWNlbmNlZFVudGlsIjoxNzg5MzgxMDE1LCJpYXQiOjE2MzE2MTQ2MTUsImlzcyI6InNhbGVzQHdlYXZlLndvcmtzIiwibmJmIjoxNjMxNjE0NjE1LCJzdWIiOiJ0ZWFtLXBlc3RvQHdlYXZlLndvcmtzIn0.klRpQQgbCtshC3PuuD4DdI3i-7Z0uSGQot23YpsETphFq4i3KK4NmgfnDg_WA3Pik-C2cJgG8WWYkWnemWQJAw`

// "super secret password"
const passwordValue = "super secret password"
const passwordHash = "$2a$10$m1A6CstgCbYrkNyPA8rF1egTaY2Qg8lpbv3zvyrpp.rcWWhJBIStu"

func AssertRowCellContains(element *agouti.Selection, text string) {
	Eventually(element).Should(BeFound())
	Eventually(element, acceptancetest.ASSERTION_1SECOND_TIME_OUT).Should(HaveText(text))
}

var intWebDriver *agouti.Page

//
// Helpers
//

func getLocalPath(localPath string) string {
	testDir, _ := os.Getwd()
	path, _ := filepath.Abs(fmt.Sprintf("%s/../../../%s", testDir, localPath))
	return path
}

func ListenAndServe(ctx context.Context, srv *http.Server) error {
	listenContext, listenCancel := context.WithCancel(ctx)
	var listenError error
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			listenError = err
		}
		listenCancel()
	}()
	defer func() {
		_ = srv.Shutdown(context.Background())
	}()

	<-listenContext.Done()

	return listenError
}

func fakeCoreConfig(t *testing.T, log logr.Logger) core_core.CoreServerConfig {
	t.Helper()

	clientsFactory := &clustersmngrfakes.FakeClientsFactory{}
	clientsPool := &clustersmngrfakes.FakeClientsPool{}
	client := clustersmngr.NewClient(clientsPool, map[string][]corev1.Namespace{})
	clientsFactory.GetServerClientReturns(client, nil)

	return core_core.NewCoreConfig(log, &rest.Config{}, "test", clientsFactory)
}

func RunCAPIServer(t *testing.T, ctx context.Context, cl client.Client, discoveryClient discovery.DiscoveryInterface) error {
	templatesLibrary := &templates.CRDLibrary{
		Log:           logr.Discard(),
		ClientGetter:  kubefakes.NewFakeClientGetter(cl),
		CAPINamespace: "default",
	}

	clustersLibrary := &clusters.CRDLibrary{
		Log:          logr.Discard(),
		ClientGetter: kubefakes.NewFakeClientGetter(cl),
		Namespace:    "default",
	}

	jwtClient := &authfakes.FakeJWTClient{
		VerifyJWTStub: func(s string) (*auth.Claims, error) {
			return &auth.Claims{
				ProviderToken: "provider-token",
			}, nil
		},
	}

	fakeAppsConfig := &wego_server.ApplicationsConfig{
		Factory:       &servicesfakes.FakeFactory{},
		JwtClient:     jwtClient,
		Logger:        logr.Discard(),
		ClusterConfig: kube.ClusterConfig{},
	}

	fakeCoreConfig := fakeCoreConfig(t, logr.Discard())

	viper.SetDefault("capi-clusters-namespace", "default")

	return app.RunInProcessGateway(ctx, "0.0.0.0:"+capiServerPort,
		app.WithCAPIClustersNamespace("default"),
		app.WithEntitlementSecretKey(client.ObjectKey{Name: "entitlement", Namespace: "default"}),
		app.WithTemplateLibrary(templatesLibrary),
		app.WithClustersLibrary(clustersLibrary),
		app.WithKubernetesClient(cl),
		app.WithDiscoveryClient(discoveryClient),
		app.WithApplicationsConfig(fakeAppsConfig),
		app.WithApplicationsOptions(wego_server.WithClientGetter(kubefakes.NewFakeClientGetter(cl))),
		app.WithGitProvider(git.NewGitProviderService(logr.Discard())),
		app.WithClientGetter(kubefakes.NewFakeClientGetter(cl)),
		app.WithCoreConfig(fakeCoreConfig),
		app.WithOIDCConfig(
			app.OIDCAuthenticationOptions{
				TokenDuration: time.Hour,
			},
		),
	)
}

func RunUIServer(ctx context.Context) {
	// is configured to proxy to
	// - 8000 for clusters-service
	cmd := exec.CommandContext(ctx, "node", "server.js")
	cmd.Dir = getLocalPath("ui-cra")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(
		os.Environ(),
		[]string{
			"PROXY_HOST=https://localhost:" + capiServerPort,
		}...,
	)

	err := cmd.Start()

	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()
	if err != nil {
		log.Println("waiting on cmd:", err)
	}
}

func waitFor200(ctx context.Context, url string, timeout time.Duration) error {
	log.Infof("Waiting for 200 from %v for %v", url, timeout)
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return wait.PollUntil(time.Second*1, func() (bool, error) {
		client := http.Client{
			Timeout: 5 * time.Second,
		}
		resp, err := client.Get(url)
		if err != nil {
			return false, nil
		}
		return resp.StatusCode == http.StatusOK, nil
	}, waitCtx.Done())
}

func gomegaFail(message string, callerSkip ...int) {
	fmt.Println("gomegaFail:")
	fmt.Println(message)
	webDriver := acceptancetest.GetWebDriver()
	if webDriver != nil {
		filepath := acceptancetest.TakeScreenShot(acceptancetest.RandString(16)) //Save the screenshot of failure
		fmt.Printf("\033[1;34mFailure screenshot is saved in file %s\033[0m \n", filepath)
	}
	// Pass this down to the default handler for onward processing
	Fail(message, callerSkip...)
}

//
// "main"
//

func TestMccpUI(t *testing.T) {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		appsv1.AddToScheme,
		capiv1.AddToScheme,
		corev1.AddToScheme,
		gitopsv1.AddToScheme,
	}
	err := schemeBuilder.AddToScheme(scheme)
	assert.NoError(t, err)

	// Add entitlement secret
	entitlementSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "entitlement",
			Namespace: "default",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{"entitlement": []byte(entitlement)},
	}

	authSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-user-auth",
			Namespace: "flux-system",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"username": []byte("wego-admin"),
			"password": []byte(passwordHash),
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(entitlementSecret, authSecret).
		Build()

	discoveryClient := discovery.NewDiscoveryClient(fakeclientset.NewSimpleClientset().Discovery().RESTClient())

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// Increment the WaitGroup synchronously in the main method, to avoid
	// racing with the goroutine starting.
	wg.Add(1)
	go func() {
		RunUIServer(ctx)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		err := RunCAPIServer(t, ctx, cl, discoveryClient)
		require.NoError(t, err)
		wg.Done()
	}()

	// Test ui is proxying through to cluster-service
	err = waitFor200(ctx, uiURL+"/v1/featureflags", time.Second*30)
	require.NoError(t, err)

	//
	// Test env stuff
	//
	RegisterFailHandler(Fail)
	// Screenshot on fail
	RegisterFailHandler(gomegaFail)
	// Screenshots
	ARTIFACTS_BASE_DIR := acceptancetest.GetEnv("ARTIFACTS_BASE_DIR", "/tmp/gitops-test/")
	_ = os.RemoveAll(ARTIFACTS_BASE_DIR)
	_ = os.MkdirAll(path.Join(ARTIFACTS_BASE_DIR, acceptancetest.SCREENSHOTS_DIR_NAME), 0700)
	// WKP-UI can be a bit slow
	SetDefaultEventuallyTimeout(acceptancetest.ASSERTION_5MINUTE_TIME_OUT)

	// Load up the acceptance suite suite
	mccpRunner := acceptancetest.DatabaseGitopsTestRunner{Client: cl}

	acceptancetest.SetSeleniumServiceUrl(seleniumURL)
	acceptancetest.SetDefaultUIURL(uiURL)
	acceptancetest.DescribeSpecsUi(mccpRunner)

	BeforeSuite(func() {
		acceptancetest.InitializeLogger("ui-integration-tests.log") // Initilaize the global logger and tee Ginkgowriter
		acceptancetest.InitializeWebdriver(uiURL)                   // Initilize web driver for whole test suite run
		By("Logging in", func() {
			acceptancetest.LoginUserFlow(acceptancetest.UserCredentials{
				UserType:     acceptancetest.ClusterUserLogin,
				UserName:     acceptancetest.AdminUserName,
				UserPassword: passwordValue,
			})
		})
	})

	AfterSuite(func() {
		webDriver := acceptancetest.GetWebDriver()
		//Tear down the suite level setup
		if webDriver != nil {
			Expect(webDriver.Destroy()).To(Succeed())
		}

		if intWebDriver != nil {
			Expect(intWebDriver.Destroy()).To(Succeed())
		}
		// Clean up ui-server
		cancel()
		// Wait for the child goroutine to finish, which will only occur when
		// the child process has stopped and the call to cmd.Wait has returned.
		// This prevents main() exiting prematurely.
		wg.Wait()
	})

	// Run it!
	RunSpecs(t, "Weave GitOps Enterprise Integration UI Tests")

}
