package acceptance

import (
	"fmt"
	"testing"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
)

var theT *testing.T

func GomegaFail(message string, callerSkip ...int) {
	randID := RandString(16)
	if webDriver != nil {
		filepath := TakeScreenShot(randID) //Save the screenshot of failure
		fmt.Printf("\033[1;34mFailure screenshot is saved in file %s\033[0m \n", filepath)
	}

	// Show management cluster pods etc.
	_ = showItems("")
	_ = dumpClusterInfo(randID)

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
})

var _ = AfterSuite(func() {
	//Tear down the suite level setup
	if webDriver != nil {
		Expect(webDriver.Destroy()).To(Succeed())
	}
})
