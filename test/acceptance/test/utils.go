package acceptance

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/agouti"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

type customFormatter struct {
	log.TextFormatter
}

var (
	logger               *logrus.Logger
	logFile              *os.File
	gitProviderEnv       GitProviderEnv
	git_repository_url   string
	selenium_service_url string
	gitops_bin_path      string
	capi_endpoint_url    string
	test_ui_url          string
	artifacts_base_dir   string

	webDriver *agouti.Page
)

const (
	WGE_WINDOW_NAME                string = "weave-gitops-enterprise"
	GITOPS_DEFAULT_NAMESPACE       string = "wego-system"
	CLUSTER_SERVICE_DEPLOYMENT_APP string = "my-mccp-cluster-service"
	SCREENSHOTS_DIR_NAME           string = "screenshots"
	WINDOW_SIZE_X                  int    = 1800
	WINDOW_SIZE_Y                  int    = 2500

	ASSERTION_DEFAULT_TIME_OUT   time.Duration = 15 * time.Second
	ASSERTION_1SECOND_TIME_OUT   time.Duration = 1 * time.Second
	ASSERTION_10SECONDS_TIME_OUT time.Duration = 10 * time.Second
	ASSERTION_30SECONDS_TIME_OUT time.Duration = 30 * time.Second
	ASSERTION_1MINUTE_TIME_OUT   time.Duration = 1 * time.Minute
	ASSERTION_2MINUTE_TIME_OUT   time.Duration = 2 * time.Minute
	ASSERTION_3MINUTE_TIME_OUT   time.Duration = 3 * time.Minute
	ASSERTION_5MINUTE_TIME_OUT   time.Duration = 5 * time.Minute
	ASSERTION_6MINUTE_TIME_OUT   time.Duration = 6 * time.Minute
	ASSERTION_15MINUTE_TIME_OUT  time.Duration = 15 * time.Minute

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
	test_ui_url = url
}

func SetSeleniumServiceUrl(url string) {
	selenium_service_url = url
}

func TakeScreenShot(name string) string {
	if webDriver != nil {
		filepath := path.Join(artifacts_base_dir, SCREENSHOTS_DIR_NAME, name+".png")
		_ = webDriver.Screenshot(filepath)
		return filepath
	}
	return ""
}

func RandString(length int) string {
	return stringWithCharset(length, charset)
}

func getCheckoutRepoPath() string {
	currDir, err := os.Getwd()
	Expect(err).ShouldNot(HaveOccurred())

	re := regexp.MustCompile(`^(.*/weave-gitops-enterprise)`)
	repoDir := re.FindStringSubmatch(currDir)
	Expect(len(repoDir)).Should(Equal(2))

	return repoDir[1]
}

func enterpriseChartVersion() string {
	version := GetEnv("ENTERPRISE_CHART_VERSION", "")
	if version == "" {
		version, _ = runCommandAndReturnStringOutput(`git describe --always --abbrev=7 | sed 's/^[^0-9]*//'`)
	}
	return version
}

func SetupTestEnvironment() {
	selenium_service_url = "http://localhost:4444/wd/hub"
	test_ui_url = fmt.Sprintf(`https://%s:%s`, GetEnv("MANAGEMENT_CLUSTER_CNAME", "localhost"), GetEnv("UI_NODEPORT", "30080"))
	capi_endpoint_url = fmt.Sprintf(`https://%s:%s`, GetEnv("MANAGEMENT_CLUSTER_CNAME", "localhost"), GetEnv("UI_NODEPORT", "30080"))
	gitops_bin_path = GetEnv("GITOPS_BIN_PATH", "/usr/local/bin/gitops")
	artifacts_base_dir = GetEnv("ARTIFACTS_BASE_DIR", "/tmp/gitops-test/")

	gitProviderEnv = initGitProviderData()
	git_repository_url = "https://" + path.Join(gitProviderEnv.Hostname, gitProviderEnv.Org, gitProviderEnv.Repo)

	//Cleanup the workspace dir, it helps when running locally
	err := os.RemoveAll(artifacts_base_dir)
	Expect(err).ShouldNot(HaveOccurred())
	err = os.MkdirAll(path.Join(artifacts_base_dir, SCREENSHOTS_DIR_NAME), 0700)
	Expect(err).ShouldNot(HaveOccurred())
}

func InstallWeaveGitopsControllers() {
	if controllerStatus(CLUSTER_SERVICE_DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE) == nil {
		logger.Info("No need to install Weave gitops controllers, managemnt cluster is already configured and setup.")

	} else {
		logger.Info("Installing Weave gitops controllers on to management cluster along with respective configurations and setting such as config repo creation etc.")

		// Config repo must exist first before installing gitops controller
		initAndCreateEmptyRepo(gitProviderEnv, true)

		logger.Info("Starting weave-gitops-enterprise installation...")
		//wego-enterprise.sh script install core and enterprise controller and setup the management cluster along with required resources, secrets and entitlements etc.
		checkoutRepoPath := getCheckoutRepoPath()
		setupScriptPath := path.Join(checkoutRepoPath, "test", "utils", "scripts", "wego-enterprise.sh")
		_, _ = runCommandAndReturnStringOutput(fmt.Sprintf(`%s setup %s`, setupScriptPath, checkoutRepoPath), ASSERTION_15MINUTE_TIME_OUT)
	}
}

func GetEnv(key, fallback string) string {
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

func initializeWebdriver(wgeURL string) {
	var err error
	if webDriver == nil {
		switch runtime.GOOS {
		case "darwin":
			chromeDriver := agouti.ChromeDriver(
				agouti.ChromeOptions("w3c", false),
				agouti.ChromeOptions("args", []string{"--disable-gpu", "--no-sandbox", "--disable-blink-features=AutomationControlled", "--ignore-ssl-errors=yes", "--ignore-certificate-errors"}),
				agouti.ChromeOptions("excludeSwitches", []string{"enable-automation"}))

			err = chromeDriver.Start()
			Expect(err).NotTo(HaveOccurred())
			webDriver, err = chromeDriver.NewPage()
			Expect(err).NotTo(HaveOccurred())
		case "linux":
			webDriver, err = agouti.NewPage(selenium_service_url, agouti.Debug, agouti.Desired(agouti.Capabilities{
				"acceptInsecureCerts": true,
				"chromeOptions": map[string]interface{}{
					"args":            []string{"--disable-gpu", "--no-sandbox", "--disable-blink-features=AutomationControlled"},
					"w3c":             false,
					"excludeSwitches": []string{"enable-automation"},
				}}))
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

func (f *customFormatter) Format(entry *log.Entry) ([]byte, error) {
	// ansi color codes are required for the colored output otherwise console output would have no lose colors for the log levels
	var levelColor int
	switch entry.Level {
	case log.DebugLevel:
		levelColor = 31 // gray
	case log.InfoLevel:
		levelColor = 36 // cyan
	case log.WarnLevel:
		levelColor = 33 // orange
	case log.ErrorLevel, log.FatalLevel, log.PanicLevel:
		levelColor = 31 // red
	default:
		return []byte(fmt.Sprintf("\t%s\n", entry.Message)), nil
	}
	return []byte(fmt.Sprintf("\x1b[%dm%s\x1b[0m: %s \x1b[38;5;243m%s\x1b[0m\n", levelColor, strings.ToUpper(entry.Level.String()), entry.Message, entry.Time.Format(f.TimestampFormat))), nil
}

func InitializeLogger(logFileName string) {
	logger = &logrus.Logger{
		Out:   os.Stdout,
		Level: logrus.TraceLevel,
		Formatter: &customFormatter{log.TextFormatter{
			FullTimestamp:          true,
			TimestampFormat:        "01/02/06 15:04:05.000",
			ForceColors:            true,
			DisableLevelTruncation: true,
		},
		},
	}

	file_name := path.Join(artifacts_base_dir, logFileName)
	logFile, err := os.OpenFile(file_name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err == nil {
		GinkgoWriter.TeeTo(logFile)
		logger.SetOutput(io.MultiWriter(logFile, os.Stdout))
	} else {
		logger.Warnf("Failed to create log file: '%s', Error: %d", file_name, err)
	}
}

// Run a command, passing through stdout/stderr to the OS standard streams
func runCommandPassThrough(name string, arg ...string) error {
	return runCommandPassThroughWithEnv([]string{}, name, arg...)
}

func runCommandPassThroughWithoutOutput(name string, arg ...string) error {
	return runCommandPassThroughWithoutOutputWithEnv([]string{}, name, arg...)
}

func runCommandPassThroughWithEnv(env []string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	if len(env) > 0 {
		cmd.Env = env
	}
	cmd.Stdout = logger.WriterLevel(logrus.TraceLevel)
	cmd.Stderr = logger.WriterLevel(logrus.TraceLevel)
	return cmd.Run()
}

func runCommandPassThroughWithoutOutputWithEnv(env []string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	if len(env) > 0 {
		cmd.Env = env
	}
	return cmd.Run()
}

func runCommandAndReturnStringOutput(commandToRun string, timeout ...time.Duration) (stdOut string, stdErr string) {
	assert_timeout := ASSERTION_DEFAULT_TIME_OUT
	if len(timeout) > 0 {
		assert_timeout = timeout[0]
	}

	command := exec.Command("sh", "-c", commandToRun)
	session, err := gexec.Start(command, logger.WriterLevel(logrus.TraceLevel), logger.WriterLevel(logrus.TraceLevel))
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session, assert_timeout).Should(gexec.Exit())

	return strings.Trim(string(session.Wait().Out.Contents()), "\n"), strings.Trim(string(session.Wait().Err.Contents()), "\n")
}

func ShowItems(itemType string) error {
	if itemType != "" {
		return runCommandPassThrough("kubectl", "get", itemType, "--all-namespaces", "-o", "wide")
	}
	err := runCommandPassThrough("kubectl", "get", "all", "--all-namespaces", "-o", "wide")
	if err != nil {
		return fmt.Errorf("failed to get all resources %s", err)
	}
	return runCommandPassThrough("kubectl", "get", "crds", "-o", "wide")
}

func DumpClusterInfo(testName string) error {
	scriptPath := path.Join(getCheckoutRepoPath(), "test", "utils", "scripts", "dump-cluster-info.sh")
	return runCommandPassThrough(scriptPath, testName, path.Join(artifacts_base_dir, "cluster-info"))
}

func getDownloadedKubeconfigPath(clusterName string) string {
	return path.Join(os.Getenv("HOME"), "Downloads", fmt.Sprintf("%s.kubeconfig", clusterName))
}

// utility functions
func deleteFile(name []string) error {
	for _, name := range name {
		logger.Tracef("Deleting: %s", name)
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
