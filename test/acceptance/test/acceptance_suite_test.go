package acceptance

import (
	"fmt"
	"testing"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
)

var theT *testing.T

func GomegaFail(message string, callerSkip ...int) {
	randID := RandString(16)
	if webDriver != nil {
		filepath := TakeScreenShot(randID) //Save the screenshot of failure
		logger.Errorf("Failure screenshot is saved in file %s", filepath)
	}

	// Show management cluster pods etc.
	ShowItems("")
	DumpClusterInfo(randID)
	DumpConfigRepo(randID)

	//Pass this down to the default handler for onward processing
	ginkgo.Fail(message, callerSkip...)
}

func TestAcceptance(t *testing.T) {

	theT = t //Save the testing instance for later use

	RegisterFailHandler(Fail)

	//Intercept the assertiona Failure
	gomega.RegisterFailHandler(GomegaFail)

	// Runs the UI tests
	DescribeSpecsUi(RealGitopsTestRunner{})
	// Runs the CLI tests
	// FIXME: CLI acceptances are disabled due to authentication not being supported
	// DescribeSpecsCli(RealGitopsTestRunner{})

	RunSpecs(t, "Weave GitOps Enterprise Acceptance Tests")

}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(ASSERTION_DEFAULT_TIME_OUT) // Things are slow when running on Kind
	SetupTestEnvironment()                                  // Read OS environment variables and initialize the test environment
	InitializeLogger("acceptance-tests.log")                // Initilaize the global logger and tee Ginkgowriter
	InstallWeaveGitopsControllers()                         // Install weave gitops core and enterprise controllers
	InitializeWebdriver(test_ui_url)                        // Initilize web driver for whole test suite run

	By(fmt.Sprintf("Login as a %s user", userCredentials.UserType), func() {
		loginUser() // Login to the weaveworks enterprise
	})

	CheckClusterService(capi_endpoint_url) // Cluster service should be running before running any test for enterprise
})

var _ = AfterSuite(func() {
	//Tear down the suite level setup
	By(fmt.Sprintf("Logout as a %s user", userCredentials.UserType), func() {
		Expect(webDriver.Navigate(test_ui_url)).To(Succeed()) // Make sure the UI should not has any popups and modal dialogs
		logoutUser()                                          // Logout to the weaveworks enterprise
	})

	deleteRepo(gitProviderEnv) // Delete the config repository to keep the org clean
	if webDriver != nil {
		Expect(webDriver.Destroy()).To(Succeed())
	}

	if _, err := logFile.Stat(); err == nil {
		logFile.Close()
	}
})
