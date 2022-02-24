package acceptance

import (
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
	_ = ShowItems("")
	_ = DumpClusterInfo(randID)

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
	DescribeSpecsCli(RealGitopsTestRunner{})

	RunSpecs(t, "Weave GitOps Enterprise Acceptance Tests")

}

var _ = BeforeSuite(func() {
	//Set the suite level defaults

	SetDefaultEventuallyTimeout(ASSERTION_DEFAULT_TIME_OUT) //Things are slow on WKP UI
	SetupTestEnvironment()                                  // Read OS environment variables and initialize the test environment
	InitializeLogger("acceptance-tests.log")                // Initilaize the global logger and tee Ginkgowriter
	InstallWeaveGitopsControllers()                         // Install weave gitops core and enterprise controllers
})

var _ = AfterSuite(func() {
	//Tear down the suite level setup

	deleteRepo(gitProviderEnv) // Delete the config repository to keep the org clean
	if webDriver != nil {
		Expect(webDriver.Destroy()).To(Succeed())
	}

	if _, err := logFile.Stat(); err == nil {
		logFile.Close()
	}
})
