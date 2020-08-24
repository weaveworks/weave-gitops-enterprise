package acceptance

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
)

var theT *testing.T
var webDriver *agouti.Page

var gitProvider string
var seleniumServiceUrl string
var wkpDashboardUrl string

const ASSERTION_DEFAULT_TIME_OUT time.Duration = 15 * time.Second //15 seconds

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

func TakeScreenShot(driver *agouti.Page) {
	if webDriver != nil {

		filepath := "./screenshots/" + String(16) + ".png"
		driver.Screenshot(filepath)
		fmt.Printf("\033[1;34mFailure screenshot is saved in file %s\033[0m \n", filepath)
	}
}

func GomegaFail(message string, callerSkip ...int) {

	TakeScreenShot(webDriver) //Save the screenshot of failure

	//Pass this down to the default handler for onward processing
	ginkgo.Fail(message, callerSkip...)
}

func TestAcceptance(t *testing.T) {
	theT = t //Save the testing instance for later use
	_ = os.RemoveAll("./screenshots")
	_ = os.Mkdir("./screenshots", 0700)
	RegisterFailHandler(Fail)
	gomega.RegisterFailHandler(GomegaFail)
	RunSpecs(t, "WKP Acceptance Suite")
}

var _ = BeforeSuite(func() {
	//Set the suite level defaults

	SetDefaultEventuallyTimeout(ASSERTION_DEFAULT_TIME_OUT) //Things are slow on WKP UI

	gitProvider = "github" //TODO - Read from config.yaml
	seleniumServiceUrl = "http://localhost:4444/wd/hub"
	wkpDashboardUrl = "http://localhost:8090/"

})

var _ = AfterSuite(func() {
	//Tear down the suite level setup
	if webDriver != nil {
		Expect(webDriver.Destroy()).To(Succeed())
	}
})
