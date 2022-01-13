package acceptance

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/agouti"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

var (
	DOCKER_IO_USER       string
	DOCKER_IO_PASSWORD   string
	GITHUB_USER          string
	GITHUB_PASSWORD      string
	GIT_PROVIDER         string
	GITHUB_ORG           string
	GITHUB_TOKEN         string
	GITLAB_TOKEN         string
	CLUSTER_REPOSITORY   string
	GIT_REPOSITORY_URL   string
	SELENIUM_SERVICE_URL string
	GITOPS_BIN_PATH      string
	CAPI_ENDPOINT_URL    string
	DEFAULT_UI_URL       string

	webDriver    *agouti.Page
	screenshotNo = 1
)

const (
	WGE_WINDOW_NAME          string = "weave-gitops-enterprise"
	GITOPS_DEFAULT_NAMESPACE string = "wego-system"
	WINDOW_SIZE_X            int    = 1800
	WINDOW_SIZE_Y            int    = 2500
	ARTEFACTS_BASE_DIR       string = "/tmp/workspace/test/"
	SCREENSHOTS_DIR          string = ARTEFACTS_BASE_DIR + "screenshots/"
	CLUSTER_INFO_DIR         string = ARTEFACTS_BASE_DIR + "cluster-info/"
	JUNIT_TEST_REPORT_FILE   string = ARTEFACTS_BASE_DIR + "acceptance-test-results.xml"

	ASSERTION_DEFAULT_TIME_OUT   time.Duration = 15 * time.Second
	ASSERTION_1SECOND_TIME_OUT   time.Duration = 1 * time.Second
	ASSERTION_10SECONDS_TIME_OUT time.Duration = 10 * time.Second
	ASSERTION_30SECONDS_TIME_OUT time.Duration = 30 * time.Second
	ASSERTION_1MINUTE_TIME_OUT   time.Duration = 1 * time.Minute
	ASSERTION_2MINUTE_TIME_OUT   time.Duration = 2 * time.Minute
	ASSERTION_3MINUTE_TIME_OUT   time.Duration = 3 * time.Minute
	ASSERTION_5MINUTE_TIME_OUT   time.Duration = 5 * time.Minute
	ASSERTION_6MINUTE_TIME_OUT   time.Duration = 6 * time.Minute

	POLL_INTERVAL_1SECONDS        time.Duration = 1 * time.Second
	POLL_INTERVAL_5SECONDS        time.Duration = 5 * time.Second
	POLL_INTERVAL_15SECONDS       time.Duration = 15 * time.Second
	POLL_INTERVAL_100MILLISECONDS time.Duration = 100 * time.Millisecond
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// Describes all the UI acceptance tests
func DescribeSpecsUi(gitopsTestRunner GitopsTestRunner) {
	DescribeClusters(gitopsTestRunner)
	DescribeTemplates(gitopsTestRunner)
	DescribeApplications(gitopsTestRunner)
}

// Describes all the CLI acceptance tests
func DescribeSpecsCli(gitopsTestRunner GitopsTestRunner) {
	DescribeCliHelp()
	DescribeCliGet(gitopsTestRunner)
	DescribeCliAddDelete(gitopsTestRunner)
	DescribeCliUpgrade(gitopsTestRunner)
}

func GetWebDriver() *agouti.Page {
	return webDriver
}

func SetDefaultUIURL(url string) {
	DEFAULT_UI_URL = url
}

func SetSeleniumServiceUrl(url string) {
	SELENIUM_SERVICE_URL = url
}

func TakeScreenShot(name string) string {
	if webDriver != nil {
		filepath := path.Join(SCREENSHOTS_DIR, name+".png")
		_ = webDriver.Screenshot(filepath)
		return filepath
	}
	return ""
}

func RandString(length int) string {
	return stringWithCharset(length, charset)
}

func SetupTestEnvironment() {
	SELENIUM_SERVICE_URL = "http://localhost:4444/wd/hub"
	DEFAULT_UI_URL = getEnv("TEST_UI_URL", "http://localhost:8000")
	CAPI_ENDPOINT_URL = getEnv("TEST_CAPI_ENDPOINT_URL", "http://localhost:8000")
	GITOPS_BIN_PATH = getEnv("GITOPS_BIN_PATH", "/usr/local/bin/gitops")

	GITHUB_USER = getEnv("GITHUB_USER", "")
	GITHUB_PASSWORD = getEnv("GITHUB_PASSWORD", "")
	GIT_PROVIDER = getEnv("GIT_PROVIDER", "")
	GITHUB_ORG = getEnv("GITHUB_ORG", "")
	GITHUB_TOKEN = getEnv("GITHUB_TOKEN", "")
	GITLAB_TOKEN = getEnv("GITLAB_TOKEN", "")
	CLUSTER_REPOSITORY = getEnv("CLUSTER_REPOSITORY", "")
	GIT_REPOSITORY_URL = "https://" + path.Join("github.com", GITHUB_ORG, CLUSTER_REPOSITORY)

	DOCKER_IO_USER = getEnv("DOCKER_IO_USER", "")
	DOCKER_IO_PASSWORD = getEnv("DOCKER_IO_PASSWORD", "")
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func takeNextScreenshot() {
	TakeScreenShot(fmt.Sprintf("test-%v", screenshotNo))
	screenshotNo += 1
}

func initializeWebdriver(wgeURL string) {
	var err error
	if webDriver == nil {
		switch runtime.GOOS {
		case "darwin":
			chromeDriver := agouti.ChromeDriver(agouti.ChromeOptions("args", []string{"--disable-gpu", "--no-sandbox"}))
			err = chromeDriver.Start()
			Expect(err).NotTo(HaveOccurred())
			webDriver, err = chromeDriver.NewPage()
			Expect(err).NotTo(HaveOccurred())
		case "linux":
			webDriver, err = agouti.NewPage(SELENIUM_SERVICE_URL, agouti.Debug, agouti.Desired(agouti.Capabilities{
				"chromeOptions": map[string]interface{}{"args": []string{"--disable-gpu", "--no-sandbox"}, "w3c": false}}))
			Expect(err).NotTo(HaveOccurred())
		}

		err = webDriver.Size(WINDOW_SIZE_X, WINDOW_SIZE_Y)
		Expect(err).NotTo(HaveOccurred())
	} else {
		// Clear localstorage, cookie etc
		Expect(webDriver.Reset()).To(Succeed())
	}

	By("When I navigate to WGE UI Page", func() {
		Expect(webDriver.Navigate(wgeURL)).To(Succeed())
	})

	By(fmt.Sprintf("And I set the default WGE window name to: %s", WGE_WINDOW_NAME), func() {
		pages.SetWindowName(webDriver, WGE_WINDOW_NAME)
		weaveGitopsWindowName := pages.GetWindowName(webDriver)
		Expect(weaveGitopsWindowName).To(Equal(WGE_WINDOW_NAME))

	})
}

// Run a command, passing through stdout/stderr to the OS standard streams
func runCommandPassThrough(env []string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	if len(env) > 0 {
		cmd.Env = env
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCommandPassThroughWithoutOutput(env []string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	if len(env) > 0 {
		cmd.Env = env
	}
	return cmd.Run()
}

func runCommandAndReturnStringOutput(commandToRun string) (stdOut string, stdErr string) {
	command := exec.Command("sh", "-c", commandToRun)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session, ASSERTION_2MINUTE_TIME_OUT).Should(gexec.Exit())

	return strings.Trim(string(session.Wait().Out.Contents()), "\n"), strings.Trim(string(session.Wait().Err.Contents()), "\n")
}

func showItems(itemType string) error {
	if itemType != "" {
		return runCommandPassThrough([]string{}, "kubectl", "get", itemType, "--all-namespaces", "-o", "wide")
	}
	return runCommandPassThrough([]string{}, "kubectl", "get", "all", "--all-namespaces", "-o", "wide")
}

func dumpClusterInfo(testName string) error {
	return runCommandPassThrough([]string{}, "../../utils/scripts/dump-cluster-info.sh", testName, CLUSTER_INFO_DIR)
}

// utility functions
func deleteFile(name []string) error {
	for _, name := range name {
		log.Printf("Deleting: %s", name)
		err := os.RemoveAll(name)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteDirectory(name []string) error {
	return deleteFile(name)
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// WaitUntil runs checkDone until a timeout is reached
func waitUntil(out io.Writer, poll, timeout time.Duration, checkDone func() error) error {
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(poll) {
		err := checkDone()
		if err == nil {
			return nil
		}
		fmt.Fprintf(out, "error occurred %s, retrying in %s\n", err, poll.String())
	}
	return fmt.Errorf("timeout reached %s", timeout.String())
}
