package acceptance

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
)

var theT *testing.T

func GomegaFail(message string, callerSkip ...int) {
	randID := RandString(16)
	if webDriver != nil {
		filepath := TakeScreenShot(randID) //Save the screenshot of failure
		fmt.Printf("\033[1;34mFailure screenshot is saved in file %s\033[0m \n", filepath)
	}

	// Show management cluster pods etc.
	_ = ShowItems("")
	_ = DumpClusterInfo(randID)

	//Pass this down to the default handler for onward processing
	ginkgo.Fail(message, callerSkip...)
}

func TestAcceptance(t *testing.T) {

	theT = t //Save the testing instance for later use

	//Cleanup the workspace dir, it helps when running locally
	err := os.RemoveAll(ARTEFACTS_BASE_DIR)
	assert.NoError(t, err)
	err = os.MkdirAll(SCREENSHOTS_DIR, 0700)
	assert.NoError(t, err)

	RegisterFailHandler(Fail)

	//Intercept the assertiona Failure
	gomega.RegisterFailHandler(GomegaFail)

	if os.Getenv("WGE_ACCEPTANCE") == "true" {

		// Runs the UI tests
		DescribeSpecsUi(RealGitopsTestRunner{})
		// Runs the CLI tests
		DescribeSpecsCli(RealGitopsTestRunner{})
	}

	//JUnit style test report
	junitReporter := reporters.NewJUnitReporter(JUNIT_TEST_REPORT_FILE)
	RunSpecsWithDefaultAndCustomReporters(t, "WGE Acceptance Suite", []Reporter{junitReporter})

}

var _ = BeforeSuite(func() {
	//Set the suite level defaults

	SetDefaultEventuallyTimeout(ASSERTION_DEFAULT_TIME_OUT) //Things are slow on WKP UI

	SELENIUM_SERVICE_URL = "http://localhost:4444/wd/hub"
	GITHUB_USER = os.Getenv("GITHUB_USER")
	GITHUB_PASSWORD = os.Getenv("GITHUB_PASSWORD")
	GIT_PROVIDER = os.Getenv("GIT_PROVIDER")
	GITHUB_ORG = os.Getenv("GITHUB_ORG")
	GITHUB_TOKEN = os.Getenv("GITHUB_TOKEN")
	CLUSTER_REPOSITORY = os.Getenv("CLUSTER_REPOSITORY")
	GIT_REPOSITORY_URL = "https://" + path.Join("github.com", GITHUB_ORG, CLUSTER_REPOSITORY)

	DOCKER_IO_USER = os.Getenv("DOCKER_IO_USER")
	DOCKER_IO_PASSWORD = os.Getenv("DOCKER_IO_PASSWORD")
})

var _ = AfterSuite(func() {
	//Tear down the suite level setup
	if webDriver != nil {
		Expect(webDriver.Destroy()).To(Succeed())
	}
})
