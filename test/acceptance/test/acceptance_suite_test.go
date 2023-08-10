package acceptance

import (
	"fmt"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var theT *testing.T

func GomegaFail(message string, callerSkip ...int) {
	randID := randString(16)
	logger.Error("Spec has failed, capturing failure...")
	logger.Tracef("Dumping artifacts to %s with prefix %s", artifactsBaseDir, randID)

	takeScreenShot(randID) //Save the screenshot of failure
	dumpingDOM(randID)
	dumpBrowserLogs(randID)
	dumpResources(randID)
	dumpClusterInfo(randID)
	dumpConfigRepo(randID)
	dumpTenantInfo(randID)

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
	setupTestEnvironment()                                         // Read OS environment variables and initialize the test environment
	initializeLogger("acceptance-tests.log")                       // Initilaize the global logger and tee Ginkgowriter
	installWeaveGitopsControllers()                                // Install weave gitops core and enterprise controllers
	initializeWebdriver(testUiUrl)                                 // Initilize web driver for whole test suite run

	ginkgo.By(fmt.Sprintf("Login as a %s user", userCredentials.UserType), func() {
		loginUser() // Login to the weaveworks enterprise dashboard

		if userCredentials.UserType == OidcUserLogin {
			cliOidcLogin() // CLI OIDC Login
		}
	})

	checkClusterService(wgeEndpointUrl) // Cluster service should be running before running any test for enterprise
})

var _ = ginkgo.AfterSuite(func() {
	//Tear down the suite level setup
	ginkgo.By(fmt.Sprintf("Logout as a %s user", userCredentials.UserType), func() {
		if webDriver != nil {
			gomega.Expect(webDriver.Navigate(testUiUrl)).To(gomega.Succeed()) // Make sure the UI should not has any popups and modal dialogs
			logoutUser()                                                      // Logout to the weaveworks enterprise
			gomega.Expect(webDriver.Destroy()).To(gomega.Succeed())
		}
	})

	deleteRepo(gitProviderEnv) // Delete the config repository to keep the org clean

	if _, err := logFile.Stat(); err == nil {
		logFile.Close()
	}
})
