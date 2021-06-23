package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
)

var theT *testing.T

func GomegaFail(message string, callerSkip ...int) {
	if webDriver != nil {
		filepath := TakeScreenShot(String(16)) //Save the screenshot of failure
		fmt.Printf("\033[1;34mFailure screenshot is saved in file %s\033[0m \n", filepath)
	}

	//Show pods
	showItems("pods")
	//Pass this down to the default handler for onward processing
	ginkgo.Fail(message, callerSkip...)
}

// showItems displays the current set of a specified object type in tabular format
func showItems(itemType string) error {
	return runCommandPassThrough([]string{}, "kubectl", "get", itemType, "--all-namespaces", "-o", "wide")
}

func TestAcceptance(t *testing.T) {

	theT = t //Save the testing instance for later use

	//Cleanup the workspace dir, it helps when running locally
	_ = os.RemoveAll(ARTEFACTS_BASE_DIR)
	_ = os.MkdirAll(SCREENSHOTS_DIR, 0700)

	RegisterFailHandler(Fail)

	//Intercept the assertiona Failure
	gomega.RegisterFailHandler(GomegaFail)

	defaultSuite := true
	if os.Getenv("MCCP_ACCEPTANCE") == "true" {
		DescribeMCCPClusters(RealMCCPTestRunner{})
		DescribeMCCPTemplates(RealMCCPTestRunner{})
		defaultSuite = false
	}

	if os.Getenv("WKP_ACCEPTANCE") == "true" || defaultSuite {
		DescribeWKPUIAcceptance()
		DescribeWorkspacesAcceptance()
	}

	//JUnit style test report
	junitReporter := reporters.NewJUnitReporter(JUNIT_TEST_REPORT_FILE)
	RunSpecsWithDefaultAndCustomReporters(t, "WKP Acceptance Suite", []Reporter{junitReporter})

}

var _ = BeforeSuite(func() {
	//Set the suite level defaults

	SetDefaultEventuallyTimeout(ASSERTION_DEFAULT_TIME_OUT) //Things are slow on WKP UI

	gitProvider = "github" //TODO - Read from config.yaml
	seleniumServiceUrl = "http://localhost:4444/wd/hub"
})

var _ = AfterSuite(func() {
	//Tear down the suite level setup
	if webDriver != nil {
		Expect(webDriver.Destroy()).To(Succeed())
	}
})
