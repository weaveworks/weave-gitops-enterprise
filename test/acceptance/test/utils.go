package acceptance

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/agouti"
	"github.com/sclevine/agouti/api"
	"github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

type CustomFormatter struct {
	logrus.TextFormatter
}

var (
	logger             *logrus.Logger
	logFile            *os.File
	gitProviderEnv     GitProviderEnv
	userCredentials    UserCredentials
	mgmtClusterKind    string
	seleniumServiceUrl string
	gitopsBinPath      string
	wgeEndpointUrl     string
	testUiUrl          string
	artifactsBaseDir   string
	testScriptsPath    string
	testDataPath       string

	webDriver        *agouti.Page
	enterpriseWindow *api.Window
)

const (
	KindMgmtCluster = "kind"
	EKSMgmtCluster  = "eks"
	GKEMgmtCluster  = "gke"
)

const (
	WGE_WINDOW_NAME                string = "weave-gitops-enterprise"
	GITOPS_DEFAULT_NAMESPACE       string = "flux-system"
	CLUSTER_SERVICE_DEPLOYMENT_APP string = "my-mccp-cluster-service"
	SCREENSHOTS_DIR_NAME           string = "screenshots"

	ASSERTION_DEFAULT_TIME_OUT   time.Duration = 15 * time.Second
	ASSERTION_1SECOND_TIME_OUT   time.Duration = 1 * time.Second
	ASSERTION_5SECONDS_TIME_OUT  time.Duration = 5 * time.Second
	ASSERTION_10SECONDS_TIME_OUT time.Duration = 10 * time.Second
	ASSERTION_15SECONDS_TIME_OUT time.Duration = 15 * time.Second
	ASSERTION_30SECONDS_TIME_OUT time.Duration = 30 * time.Second
	ASSERTION_1MINUTE_TIME_OUT   time.Duration = 1 * time.Minute
	ASSERTION_2MINUTE_TIME_OUT   time.Duration = 2 * time.Minute
	ASSERTION_3MINUTE_TIME_OUT   time.Duration = 3 * time.Minute
	ASSERTION_5MINUTE_TIME_OUT   time.Duration = 5 * time.Minute
	ASSERTION_6MINUTE_TIME_OUT   time.Duration = 6 * time.Minute
	ASSERTION_10MINUTE_TIME_OUT  time.Duration = 10 * time.Minute
	ASSERTION_15MINUTE_TIME_OUT  time.Duration = 15 * time.Minute

	POLL_INTERVAL_1SECONDS        time.Duration = 1 * time.Second
	POLL_INTERVAL_3SECONDS        time.Duration = 3 * time.Second
	POLL_INTERVAL_5SECONDS        time.Duration = 5 * time.Second
	POLL_INTERVAL_15SECONDS       time.Duration = 15 * time.Second
	POLL_INTERVAL_100MILLISECONDS time.Duration = 100 * time.Millisecond
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func randString(length int) string {
	return stringWithCharset(length, charset)
}

func getCheckoutRepoPath() string {
	currDir, err := os.Getwd()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	re := regexp.MustCompile(`^(.*/weave-gitops-enterprise)`)
	repoDir := re.FindStringSubmatch(currDir)
	gomega.Expect(len(repoDir)).Should(gomega.Equal(2))

	return repoDir[1]
}

func setupTestEnvironment() {
	mgmtClusterKind = GetEnv("MANAGEMENT_CLUSTER_KIND", "kind")
	seleniumServiceUrl = "http://localhost:4444/wd/hub"
	testUiUrl = fmt.Sprintf(`http://%s:%s`, GetEnv("MANAGEMENT_CLUSTER_CNAME", "localhost"), GetEnv("UI_NODEPORT", "30080"))
	wgeEndpointUrl = fmt.Sprintf(`http://%s:%s`, GetEnv("MANAGEMENT_CLUSTER_CNAME", "localhost"), GetEnv("UI_NODEPORT", "30080"))
	gitopsBinPath = GetEnv("GITOPS_BIN_PATH", "/usr/local/bin/gitops")
	artifactsBaseDir = GetEnv("ARTIFACTS_BASE_DIR", "/tmp/gitops-test/")
	testScriptsPath = path.Join(getCheckoutRepoPath(), "test", "utils", "scripts")
	testDataPath = path.Join(getCheckoutRepoPath(), "test", "utils", "data")

	gitProviderEnv = initGitProviderData()

	userCredentials = initUserCredentials()

	//Cleanup the workspace dir, it helps when running locally
	err := os.RemoveAll(artifactsBaseDir)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = os.MkdirAll(path.Join(artifactsBaseDir, SCREENSHOTS_DIR_NAME), 0700)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func installWeaveGitopsControllers() {
	// gitops binary must exists, it is required to install weave gitops controllers
	gomega.Expect(fileExists(gitopsBinPath)).To(gomega.BeTrue(), fmt.Sprintf("%s can not be found.", gitopsBinPath))
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

func initializeWebdriver(wgeURL string) {
	var err error
	if webDriver == nil {
		switch runtime.GOOS {
		case "darwin":
			a := make(map[string]bool)
			a["enableNetwork"] = true
			chromeDriver := agouti.ChromeDriver(
				agouti.ChromeOptions("binary", "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"),
				agouti.ChromeOptions("w3c", false),
				agouti.ChromeOptions("args", []string{"--disable-gpu", "--no-sandbox", "--disable-blink-features=AutomationControlled", "--ignore-ssl-errors=yes", "--ignore-certificate-errors"}),
				agouti.ChromeOptions("excludeSwitches", []string{"enable-automation"}),
				agouti.ChromeOptions("useAutomationExtension", false))
			err = chromeDriver.Start()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			webDriver, err = chromeDriver.NewPage()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

		case "linux":
			webDriver, err = agouti.NewPage(seleniumServiceUrl, agouti.Debug, agouti.Desired(agouti.Capabilities{
				"acceptInsecureCerts": true,
				"chromeOptions": map[string]interface{}{
					"args":                   []string{"--disable-gpu", "--no-sandbox", "--disable-blink-features=AutomationControlled"},
					"w3c":                    false,
					"excludeSwitches":        []string{"enable-automation"},
					"useAutomationExtension": false,
				}}))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}

		err = webDriver.Size(1800, 2500)
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Failed to resize browser window")

	} else {
		logger.Info("Clearing cookies")
		// Clear localstorage, cookie etc
		gomega.Expect(webDriver.Reset()).To(gomega.Succeed())
	}

	ginkgo.By("When I navigate to WGE UI Page", func() {
		gomega.Expect(webDriver.Navigate(wgeURL)).To(gomega.Succeed())
	})

	ginkgo.By(fmt.Sprintf("And I set the default WGE window name to: %s", WGE_WINDOW_NAME), func() {
		pages.SetWindowName(webDriver, WGE_WINDOW_NAME)
		gomega.Expect(pages.GetWindowName(webDriver)).To(gomega.Equal(WGE_WINDOW_NAME))
		enterpriseWindow, err = webDriver.Session().GetWindow()
		gomega.Expect(err).To(gomega.BeNil(), "Failed to get wevegitops enterprise window")
	})
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// ansi color codes are required for the colored output otherwise console output would have no lose colors for the log levels
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = 31 // gray
	case logrus.InfoLevel:
		levelColor = 36 // cyan
	case logrus.WarnLevel:
		levelColor = 33 // orange
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = 31 // red
	default:
		return []byte(fmt.Sprintf("\t%s\n", entry.Message)), nil
	}
	return []byte(fmt.Sprintf("\x1b[%dm%s\x1b[0m: %s \x1b[38;5;243m%s\x1b[0m\n", levelColor, strings.ToUpper(entry.Level.String()), entry.Message, entry.Time.Format(f.TimestampFormat))), nil
}

func initializeLogger(logFileName string) {
	logger = &logrus.Logger{
		Out:   os.Stdout,
		Level: logrus.TraceLevel,
		Formatter: &CustomFormatter{logrus.TextFormatter{
			FullTimestamp:          true,
			TimestampFormat:        "01/02/06 15:04:05.000",
			ForceColors:            true,
			DisableLevelTruncation: true,
		},
		},
	}

	file_name := path.Join(artifactsBaseDir, logFileName)
	logFile, err := os.OpenFile(file_name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err == nil {
		ginkgo.GinkgoWriter.TeeTo(logFile)
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
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Error starting Cmd: "+commandToRun)
	gomega.Eventually(session, assert_timeout).Should(gexec.Exit())

	return strings.Trim(string(session.Wait().Out.Contents()), "\n"), strings.Trim(string(session.Wait().Err.Contents()), "\n")
}

func currentSpecType(specLabel string) bool {
	for _, ctx := range ginkgo.CurrentSpecReport().ContainerHierarchyLabels {
		for _, label := range ctx {
			if label == specLabel {
				return true
			}
		}
	}
	return false
}

func takeScreenShot(name string) {
	if currentSpecType("cli") {
		return
	}

	logger.Info("Saving screenshot ...")
	if webDriver != nil {
		filepath := path.Join(artifactsBaseDir, SCREENSHOTS_DIR_NAME, name+".png")
		_ = webDriver.Screenshot(filepath)
	}
}

func dumpingDOM(name string) {
	if currentSpecType("cli") {
		return
	}

	logger.Info("Dumping DOM ... ")
	if webDriver != nil {
		filepath := path.Join(artifactsBaseDir, SCREENSHOTS_DIR_NAME, name+".html")
		var htmlDocument interface{}
		_ = webDriver.RunScript(`return document.documentElement.innerHTML;`, map[string]interface{}{}, &htmlDocument)
		stringDocument := fmt.Sprintf("%v", htmlDocument)
		if htmlDocument != nil {
			_ = os.WriteFile(filepath, []byte(stringDocument), 0644)
		}
	}
}

func dumpResources(testName string) {
	resourcesPath := "/tmp/resource-info"
	archiveResourcePath := path.Join(artifactsBaseDir, "resource-info")
	archivedPath := path.Join(archiveResourcePath, testName+".tar.gz")
	logger.Info("Dumping cluster objects/resources ...")

	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`rm -rf %[1]v && mkdir -p %[1]v && mkdir -p %[2]v`, resourcesPath, archiveResourcePath))

	summaryResourceTypes := []string{"all", "crds"}
	for _, res := range summaryResourceTypes {
		_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("kubectl get %s --all-namespaces -o wide > %s", res, path.Join(resourcesPath, res+".txt")))
	}

	resourceTypes := []string{"gitopsclusters", "gitrepositories", "kustomizations"}
	for _, res := range resourceTypes {
		_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("kubectl get %s --all-namespaces -o yaml > %s", res, path.Join(resourcesPath, res+".txt")))
	}

	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("kubectl get configmap %s -n flux-system -o yaml > %s", CLUSTER_SERVICE_DEPLOYMENT_APP, path.Join(resourcesPath, CLUSTER_SERVICE_DEPLOYMENT_APP+"-configmap.txt")))
	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`cd %s && tar -czf %s .`, resourcesPath, archivedPath))
}

func dumpClusterInfo(testName string) {
	logsPath := "/tmp/dumped-cluster-logs"
	archiveLogsPath := path.Join(artifactsBaseDir, "cluster-info")
	archivedPath := path.Join(archiveLogsPath, testName+".tar.gz")
	logger.Info("Dumping cluster-info ...")

	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`rm -rf %s && mkdir -p %s`, logsPath, archiveLogsPath))
	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl cluster-info dump --all-namespaces --output-directory %s`, logsPath))
	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`cd %s && tar -czf %s .`, logsPath, archivedPath))
}

func dumpConfigRepo(testName string) {
	repoPath := "/tmp/config-repo"
	archiveRepoPath := path.Join(artifactsBaseDir, "config-repo")
	archivedPath := path.Join(archiveRepoPath, testName+".tar.gz")
	logger.Info("Dumping git-repo ...")

	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`rm -rf %s && mkdir -p %s`, repoPath, archiveRepoPath))
	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`git clone git@%s:%s/%s.git %s`, gitProviderEnv.Hostname, gitProviderEnv.Org, gitProviderEnv.Repo, repoPath))
	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`cd %s && tar -czf %s .`, repoPath, archivedPath))
}

func dumpTenantInfo(testName string) {
	archiveTenantPath := path.Join(artifactsBaseDir, "tenants")
	logger.Info("Dumping tenant info ...")

	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`cp /tmp/rendered-tenant.yaml %s`, archiveTenantPath))
	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`cp /tmp/generated-tenant.yaml %s`, archiveTenantPath))
}

func dumpBrowserLogs(testName string) {
	if webDriver == nil {
		return
	}
	if currentSpecType("cli") {
		return
	}

	writeSlicetoFile := func(fileName string, dataLog interface{}) {
		f, _ := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		defer f.Close()
		dataWriter := bufio.NewWriter(f)

		val := reflect.ValueOf(dataLog)
		switch reflect.TypeOf(dataLog).Kind() {
		case reflect.Slice:
			for i := 0; i < val.Len(); i++ {
				_, _ = dataWriter.WriteString(fmt.Sprintf("%v", val.Index(i)) + "\n")
			}
		}
		dataWriter.Flush()
	}

	browserLogsPath := "/tmp/browser-logs"
	archiveLogsPath := path.Join(artifactsBaseDir, "browser-logs")
	archivedPath := path.Join(archiveLogsPath, testName+".tar.gz")
	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`rm -rf %[1]v && mkdir -p %[1]v && mkdir -p %[2]v`, browserLogsPath, archiveLogsPath))

	logger.Info("Dumping browser console logs ...")
	consoleLog, _ := webDriver.ReadAllLogs("browser")
	writeSlicetoFile(path.Join(browserLogsPath, "console.txt"), consoleLog)

	logger.Info("Dumping browser network logs ...")
	var networkLog interface{}
	gomega.Expect(webDriver.RunScript(`return window.performance.getEntries();`, map[string]interface{}{}, &networkLog)).ShouldNot(gomega.HaveOccurred())
	writeSlicetoFile(path.Join(browserLogsPath, "network.txt"), networkLog)

	_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`cd %s && tar -czf %s .`, browserLogsPath, archivedPath))
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

		input, err := os.ReadFile(sourceFile)
		if err != nil {
			return err
		}

		err = os.WriteFile(destination, input, 0644)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}

func StringToLines(s string) (lines []string, err error) {
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	err = scanner.Err()
	return
}
