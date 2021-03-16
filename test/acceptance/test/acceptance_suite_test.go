package acceptance

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
)

var theT *testing.T
var webDriver *agouti.Page

var gitProvider string
var seleniumServiceUrl string

const wkpUrl = "http://localhost:8090"

const ARTEFACTS_BASE_DIR string = "/tmp/workspace/test/"
const SCREENSHOTS_DIR string = ARTEFACTS_BASE_DIR + "screenshots/"
const JUNIT_TEST_REPORT_FILE string = ARTEFACTS_BASE_DIR + "wkp_junit.xml"

const ASSERTION_DEFAULT_TIME_OUT time.Duration = 15 * time.Second //15 seconds
const ASSERTION_1MINUTE_TIME_OUT time.Duration = 1 * time.Minute  //5 Minutes
const ASSERTION_2MINUTES_TIME_OUT time.Duration = 5 * time.Minute //2 Minutes
const ASSERTION_5MINUTES_TIME_OUT time.Duration = 5 * time.Minute //5 Minutes
const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}

func TakeScreenShot(name string) string {
	if webDriver != nil {
		filepath := path.Join(SCREENSHOTS_DIR, name+".png")
		webDriver.Screenshot(filepath)
		return filepath
	}
	return ""
}

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

// Run a command, passing through stdout/stderr to the OS standard streams
func runCommandPassThrough(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// showItems displays the current set of a specified object type in tabular format
func showItems(itemType string) error {
	return runCommandPassThrough("kubectl", "get", itemType, "--all-namespaces", "-o", "wide")
}

func TestAcceptance(t *testing.T) {

	theT = t //Save the testing instance for later use

	//Cleanup the workspace dir, it helps when running locally
	_ = os.RemoveAll(ARTEFACTS_BASE_DIR)
	_ = os.MkdirAll(SCREENSHOTS_DIR, 0700)

	RegisterFailHandler(Fail)

	//Intercept the assertiona Failure
	gomega.RegisterFailHandler(GomegaFail)

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
