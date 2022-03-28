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
	SetDefaultEventuallyTimeout(ASSERTION_DEFAULT_TIME_OUT) // Things are slow when running on Kind
	SetupTestEnvironment()                                  // Read OS environment variables and initialize the test environment
	InitializeLogger("acceptance-tests.log")                // Initilaize the global logger and tee Ginkgowriter
	InstallWeaveGitopsControllers()                         // Install weave gitops core and enterprise controllers
	InitializeWebdriver(test_ui_url)                        // Initilize web driver for whole test suite run

	if EnableUserLogin {
		loginUserName := AdminUserName
		if login_user_type == OidcUserLogin {
			loginUserName = gitProviderEnv.Username
		}

		By(fmt.Sprintf("Login %s user as: %s", login_user_type, loginUserName), func() {
			loginUser(login_user_type) // Login to the weaveworks enterprise
		})
	}

	CheckClusterService(capi_endpoint_url) // Cluster service should be running before running any test for enterprise
})

var _ = AfterSuite(func() {
	//Tear down the suite level setup
	if EnableUserLogin {
		By(fmt.Sprintf("Logout %s user as: %s", login_user_type, loginUserName), func() {
			logoutUser(login_user_type) // Login to the weaveworks enterprise
		})
	}

	deleteRepo(gitProviderEnv) // Delete the config repository to keep the org clean
	if webDriver != nil {
		Expect(webDriver.Destroy()).To(Succeed())
	}

	if _, err := logFile.Stat(); err == nil {
		logFile.Close()
	}
})
