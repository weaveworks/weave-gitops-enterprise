package acceptance

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"reflect"
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
	userCredentials      UserCredentials
	git_repository_url   string
	selenium_service_url string
	gitops_bin_path      string
	capi_provider        string
	capi_endpoint_url    string
	test_ui_url          string
	artifacts_base_dir   string

	webDriver *agouti.Page
)

const (
	WGE_WINDOW_NAME                string = "weave-gitops-enterprise"
	GITOPS_DEFAULT_NAMESPACE       string = "flux-system"
	CLUSTER_SERVICE_DEPLOYMENT_APP string = "my-mccp-cluster-service"
	SCREENSHOTS_DIR_NAME           string = "screenshots"

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
	DescribePolicies(gitopsTestRunner)
	DescribeViolations(gitopsTestRunner)
}

// Describes all the CLI acceptance tests
func DescribeSpecsCli(gitopsTestRunner GitopsTestRunner) {
	// FIXME: CLI acceptances are disabled due to authentication not being supported
	// DescribeCliHelp()
	// DescribeCliGet(gitopsTestRunner)
	// DescribeCliAddDelete(gitopsTestRunner)
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

func SetupTestEnvironment() {
	selenium_service_url = "http://localhost:4444/wd/hub"
	test_ui_url = fmt.Sprintf(`https://%s:%s`, GetEnv("MANAGEMENT_CLUSTER_CNAME", "localhost"), GetEnv("UI_NODEPORT", "30080"))
	capi_endpoint_url = fmt.Sprintf(`https://%s:%s`, GetEnv("MANAGEMENT_CLUSTER_CNAME", "localhost"), GetEnv("UI_NODEPORT", "30080"))
	gitops_bin_path = GetEnv("GITOPS_BIN_PATH", "/usr/local/bin/gitops")
	capi_provider = GetEnv("CAPI_PROVIDER", "capd")
	artifacts_base_dir = GetEnv("ARTIFACTS_BASE_DIR", "/tmp/gitops-test/")

	gitProviderEnv = initGitProviderData()
	git_repository_url = "https://" + path.Join(gitProviderEnv.Hostname, gitProviderEnv.Org, gitProviderEnv.Repo)

	userCredentials = initUserCredentials()

	//Cleanup the workspace dir, it helps when running locally
	err := os.RemoveAll(artifacts_base_dir)
	Expect(err).ShouldNot(HaveOccurred())
	err = os.MkdirAll(path.Join(artifacts_base_dir, SCREENSHOTS_DIR_NAME), 0700)
	Expect(err).ShouldNot(HaveOccurred())
}

func InstallWeaveGitopsControllers() {
	// gitops binary must exists, it is required to install weave gitops controllers
	Expect(fileExists(gitops_bin_path)).To(BeTrue(), fmt.Sprintf("%s can not be found.", gitops_bin_path))
	// TODO: check flux bin is available too.

	if controllerStatus(CLUSTER_SERVICE_DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE) == nil {
		repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
		initAndCreateEmptyRepo(gitProviderEnv, true)
		bootstrapAndVerifyFlux(gitProviderEnv, GITOPS_DEFAULT_NAMESPACE, getGitRepositoryURL(repoAbsolutePath))
		logger.Info("No need to install Weave gitops enterprise controllers, managemnt cluster is already configured and setup.")

	} else {
		logger.Info("Installing Weave gitops controllers on to management cluster along with respective configurations and setting such as config repo creation etc.")

		// Config repo must exist first before installing gitops controller
		initAndCreateEmptyRepo(gitProviderEnv, true)

		logger.Info("Starting weave-gitops-enterprise installation...")
		// wego-enterprise.sh script install core and enterprise controller and setup the management cluster along with required resources, secrets and entitlements etc.
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

func InitializeWebdriver(wgeURL string) {
	var err error
	if webDriver == nil {
		switch runtime.GOOS {
		case "darwin":
			a := make(map[string]bool)
			a["enableNetwork"] = true
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

		err = webDriver.Size(1800, 2500)
		Expect(err).NotTo(HaveOccurred(), "Failed to resize browser window")

	} else {
		logger.Info("Clearing cookies")
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

func ShowItems(itemType string) {
	logger.Info("Dumping cluster objects/resources...")
	if itemType != "" {
		_ = runCommandPassThrough("kubectl", "get", itemType, "--all-namespaces", "-o", "wide")
	}
	_ = runCommandPassThrough("kubectl", "get", "all", "--all-namespaces", "-o", "wide")

	logger.Info(fmt.Sprintf("Dumping %s congigmap", CLUSTER_SERVICE_DEPLOYMENT_APP))
	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("kubectl get configmap %s -n flux-system -o yaml", CLUSTER_SERVICE_DEPLOYMENT_APP))

	logger.Info("Dumping cluster crds...")
	_ = runCommandPassThrough("kubectl", "get", "crds", "-o", "wide")
}

func DumpClusterInfo(testName string) {
	logger.Info("Dumping cluster-info...")

	logsPath := "/tmp/dumped-cluster-logs"
	archiveLogsPath := path.Join(artifacts_base_dir, "cluster-info")
	archivedPath := path.Join(archiveLogsPath, testName+".tar.gz")

	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`rm -rf %s && mkdir -p %s`, logsPath, archiveLogsPath))
	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl cluster-info dump --all-namespaces --output-directory %s`, logsPath))
	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`cd %s && tar -czf %s .`, logsPath, archivedPath))
}

func DumpConfigRepo(testName string) {
	repoPath := "/tmp/config-repo"
	archiveRepoPath := path.Join(artifacts_base_dir, "config-repo")
	archivedPath := path.Join(archiveRepoPath, testName+".tar.gz")

	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`rm -rf %s && mkdir -p %s`, repoPath, archiveRepoPath))
	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`git clone git@%s:%s/%s.git %s`, gitProviderEnv.Hostname, gitProviderEnv.Org, gitProviderEnv.Repo, repoPath))
	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`cd %s && tar -czf %s .`, repoPath, archivedPath))
}

func DumpBrowserLogs(console bool, network bool) {
	fetchLogs := false
	for _, label := range CurrentSpecReport().LeafNodeLabels {
		if label == "browser-logs" {
			fetchLogs = true
		}
	}

	if !fetchLogs {
		return
	}

	if console {
		logger.Info("Dumping browser console logs...")
		consoleLog, _ := webDriver.ReadAllLogs("browser")
		for _, l := range consoleLog {
			logger.Trace(l)
		}
	}

	if network {
		logger.Info("Dumping network logs...")
		var networkLog interface{}
		Expect(webDriver.RunScript(`return window.performance.getEntries();`, map[string]interface{}{}, &networkLog)).ShouldNot(HaveOccurred())

		switch reflect.TypeOf(networkLog).Kind() {
		case reflect.Slice:
			s := reflect.ValueOf(networkLog)
			for i := 0; i < s.Len(); i++ {
				logger.Trace(s.Index(i))
			}
		}
	}
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

func createDirectory(dirPath string) error {
	return os.MkdirAll(dirPath, os.ModePerm)
}

func copyFile(sourceFile, destination string) error {

	if info, err := os.Stat(destination); err == nil {
		if info.IsDir() {
			_, src := path.Split(sourceFile)
			destination = path.Join(destination, src)
		}

		input, err := ioutil.ReadFile(sourceFile)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(destination, input, 0644)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}
