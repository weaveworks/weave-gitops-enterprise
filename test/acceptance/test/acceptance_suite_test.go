package acceptance

import (
	"fmt"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var theT *testing.T

func GomegaFail(message string, callerSkip ...int) {
	randID := RandString(16)
	logger.Error("Spec has failed, capturing failure...")
	logger.Tracef("Dumping artifacts to %s with prefix %s", artifacts_base_dir, randID)

	TakeScreenShot(randID) //Save the screenshot of failure
	DumpingDOM(randID)
	DumpBrowserLogs(randID)
	DumpResources(randID)
	DumpClusterInfo(randID)
	DumpConfigRepo(randID)

	//Pass this down to the default handler for onward processing
	ginkgo.Fail(message, callerSkip...)
}

func TestAcceptance(t *testing.T) {

	theT = t //Save the testing instance for later use

	gomega.RegisterFailHandler(ginkgo.Fail)

	//Intercept the assertiona Failure
	gomega.RegisterFailHandler(GomegaFail)

	ginkgo.RunSpecs(t, "Weave GitOps Enterprise Acceptance Tests")
}

var _ = ginkgo.BeforeSuite(func() {
	gomega.SetDefaultEventuallyTimeout(ASSERTION_DEFAULT_TIME_OUT) // Things are slow when running on Kind
	SetupTestEnvironment()                                         // Read OS environment variables and initialize the test environment
	InitializeLogger("acceptance-tests.log")                       // Initilaize the global logger and tee Ginkgowriter
	InstallWeaveGitopsControllers()                                // Install weave gitops core and enterprise controllers
	InitializeWebdriver(test_ui_url)                               // Initilize web driver for whole test suite run

	ginkgo.By(fmt.Sprintf("Login as a %s user", userCredentials.UserType), func() {
		loginUser() // Login to the weaveworks enterprise dashboard

		if userCredentials.UserType == OidcUserLogin {
			cliOidcLogin() // CLI OIDC Login
		}
	})

	CheckClusterService(wge_endpoint_url) // Cluster service should be running before running any test for enterprise
})

var _ = ginkgo.AfterSuite(func() {
	//Tear down the suite level setup
	ginkgo.By(fmt.Sprintf("Logout as a %s user", userCredentials.UserType), func() {
		gomega.Expect(webDriver.Navigate(test_ui_url)).To(gomega.Succeed()) // Make sure the UI should not has any popups and modal dialogs
		logoutUser()                                                        // Logout to the weaveworks enterprise
	})

	deleteRepo(gitProviderEnv) // Delete the config repository to keep the org clean
	if webDriver != nil {
		gomega.Expect(webDriver.Destroy()).To(gomega.Succeed())
	}

	if _, err := logFile.Stat(); err == nil {
		logFile.Close()
	}
})
