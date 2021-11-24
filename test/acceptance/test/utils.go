package acceptance

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/agouti"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/capi"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/types"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var DOCKER_IO_USER string
var DOCKER_IO_PASSWORD string
var GITHUB_USER string
var GITHUB_PASSWORD string
var GIT_PROVIDER string
var GITHUB_ORG string
var GITHUB_TOKEN string
var CLUSTER_REPOSITORY string
var GIT_REPOSITORY_URL string
var SELENIUM_SERVICE_URL string

var webDriver *agouti.Page
var defaultUIURL = "http://localhost:8090"
var defaultPctlBinPath = "/usr/local/bin/pctl"
var defaultGitopsBinPath = "/usr/local/bin/gitops"
var defaultCapiEndpointURL = "http://localhost:8090"

const GITOPS_DEFAULT_NAMESPACE = "wego-system"

func GetWebDriver() *agouti.Page {
	return webDriver
}

func SetWebDriver(wb *agouti.Page) {
	webDriver = wb
}

func GetPctlBinPath() string {
	if os.Getenv("PCTL_BIN_PATH") != "" {
		return os.Getenv("PCTL_BIN_PATH")
	}
	return defaultPctlBinPath
}

func GetGitopsBinPath() string {
	if os.Getenv("GITOPS_BIN_PATH") != "" {
		return os.Getenv("GITOPS_BIN_PATH")
	}
	return defaultGitopsBinPath
}

func GetWGEUrl() string {
	if os.Getenv("TEST_UI_URL") != "" {
		return os.Getenv("TEST_UI_URL")
	}
	return defaultUIURL
}

func GetCapiEndpointUrl() string {
	if os.Getenv("TEST_CAPI_ENDPOINT_URL") != "" {
		return os.Getenv("TEST_CAPI_ENDPOINT_URL")
	}
	return defaultCapiEndpointURL
}

func SetDefaultUIURL(url string) {
	defaultUIURL = url
}

func SetSeleniumServiceUrl(url string) {
	SELENIUM_SERVICE_URL = url
}

const WINDOW_SIZE_X int = 1800
const WINDOW_SIZE_Y int = 2500
const ARTEFACTS_BASE_DIR string = "/tmp/workspace/test/"
const SCREENSHOTS_DIR string = ARTEFACTS_BASE_DIR + "screenshots/"
const CLUSTER_INFO_DIR string = ARTEFACTS_BASE_DIR + "cluster-info/"
const JUNIT_TEST_REPORT_FILE string = ARTEFACTS_BASE_DIR + "acceptance-test-results.xml"

const ASSERTION_DEFAULT_TIME_OUT time.Duration = 15 * time.Second
const ASSERTION_1SECOND_TIME_OUT time.Duration = 1 * time.Second
const ASSERTION_10SECONDS_TIME_OUT time.Duration = 10 * time.Second
const ASSERTION_30SECONDS_TIME_OUT time.Duration = 30 * time.Second
const ASSERTION_1MINUTE_TIME_OUT time.Duration = 1 * time.Minute
const ASSERTION_2MINUTE_TIME_OUT time.Duration = 2 * time.Minute
const ASSERTION_3MINUTE_TIME_OUT time.Duration = 3 * time.Minute
const ASSERTION_5MINUTE_TIME_OUT time.Duration = 5 * time.Minute
const ASSERTION_6MINUTE_TIME_OUT time.Duration = 6 * time.Minute

const POLL_INTERVAL_15SECONDS time.Duration = 15 * time.Second
const POLL_INTERVAL_5SECONDS time.Duration = 5 * time.Second

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

func RandString(length int) string {
	return StringWithCharset(length, charset)
}

// WaitUntil runs checkDone until a timeout is reached
func WaitUntil(out io.Writer, poll, timeout time.Duration, checkDone func() error) error {
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(poll) {
		err := checkDone()
		if err == nil {
			return nil
		}
		fmt.Fprintf(out, "error occurred %s, retrying in %s\n", err, poll.String())
	}
	return fmt.Errorf("timeout reached %s", timeout.String())
}

func TakeScreenShot(name string) string {
	if webDriver != nil {
		filepath := path.Join(SCREENSHOTS_DIR, name+".png")
		_ = webDriver.Screenshot(filepath)
		return filepath
	}
	return ""
}

var n = 1

func TakeNextScreenshot() {
	TakeScreenShot(fmt.Sprintf("test-%v", n))
	n += 1
}

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

// Interface that can be implemented either with:
// - "Real" commands like "exec(kubectl...)"
// - "Mock" commands like db.Create(cluster_info...)

type GitopsTestRunner interface {
	ResetDatabase() error
	VerifyWegoPodsRunning()
	FireAlert(name, severity, message string, fireFor time.Duration) error
	KubectlApply(env []string, tokenURL string) error
	KubectlDelete(env []string, tokenURL string) error
	KubectlDeleteAllAgents(env []string) error
	TimeTravelToLastSeen() error
	TimeTravelToAlertsResolved() error
	AddWorkspace(env []string, clusterName string) error
	CreateApplyCapitemplates(templateCount int, templateFile string) []string
	DeleteApplyCapiTemplates(templateFiles []string)
	CreateIPCredentials(infrastructureProvider string)
	DeleteIPCredentials(infrastructureProvider string)
	CheckClusterService(capiEndpointURL string)
	RestartDeploymentPods(env []string, appName string, namespace string) error

	// Git repository helper functions
	DeleteRepo(repoName string)
	InitAndCreateEmptyRepo(repoName string, IsPrivateRepo bool) string
	GitAddCommitPush(repoAbsolutePath string, fileToAdd string)
	CreateGitRepoBranch(repoAbsolutePath string, branchName string) string
	PullBranch(repoAbsolutePath string, branch string)
	ListPullRequest(repoAbsolutePath string) []string
	MergePullRequest(repoAbsolutePath string, prBranch string)
	GetRepoVisibility(org string, repo string) string
}

func InitializeWebdriver(wgeURL string) {
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
	}

	By("When I navigate to MCCP UI Page", func() {
		Expect(webDriver.Navigate(wgeURL)).To(Succeed())
	})
}

// "DB" backend that creates/delete rows

type DatabaseGitopsTestRunner struct {
	DB     *gorm.DB
	Client goclient.Client
}

func (b DatabaseGitopsTestRunner) TimeTravelToLastSeen() error {
	oneMinuteAgo := time.Now().UTC().Add(time.Minute * -2)
	b.DB.Exec("update cluster_info set updated_at = ?", oneMinuteAgo)
	return nil
}

func (b DatabaseGitopsTestRunner) TimeTravelToAlertsResolved() error {
	b.DB.Where("1 = 1").Delete(&models.Alert{})
	return nil
}

func (b DatabaseGitopsTestRunner) ResetDatabase() error {
	b.DB.Where("1 = 1").Delete(&models.Cluster{})
	return nil
}

func (b DatabaseGitopsTestRunner) VerifyWegoPodsRunning() {

}

func (b DatabaseGitopsTestRunner) KubectlApply(env []string, tokenURL string) error {
	u, err := url.Parse(tokenURL)
	if err != nil {
		return err
	}
	token := u.Query()["token"][0]

	b.DB.Create(&models.ClusterInfo{
		UID:          types.UID(RandString(10)),
		ClusterToken: token,
		UpdatedAt:    time.Now().UTC(),
	})
	b.DB.Create(&models.GitCommit{
		ClusterToken: token,
		Sha:          "abcdef123456",
		AuthorName:   "Alice",
		AuthorEmail:  "alice@acme.org",
		AuthorDate:   time.Now().UTC().Add(time.Hour * -1),
		Message:      "Fixed it",
	})
	b.DB.Create(&models.FluxInfo{
		ClusterToken: token,
		Name:         "flux",
		Namespace:    "wkp-flux",
		RepoURL:      "git@github.com:wkp/my-cluster",
		RepoBranch:   "main",
	})
	return nil
}

func (b DatabaseGitopsTestRunner) KubectlDelete(env []string, tokenURL string) error {
	//
	// No more cluster_infos will be created anyway..
	// FIXME: maybe we add a polling loop that keeps creating cluster_info while its connected
	//
	return nil
}

func (b DatabaseGitopsTestRunner) KubectlDeleteAllAgents(env []string) error {
	// No more cluster_infos will be created anyway..
	return nil
}

func (b DatabaseGitopsTestRunner) FireAlert(name, severity, message string, fireFor time.Duration) error {
	var firstCluster models.Cluster
	b.DB.Last(&firstCluster)

	//
	// FIXME: we shouldn't need this. The UI should stop showing the alerts after 30s anyway
	// But its not filtering on endsAt right now.
	//
	go func() {
		time.Sleep(fireFor)
		b.DB.Where("1 = 1").Delete(&models.Alert{})
	}()

	labels := fmt.Sprintf(`{ "alertname": "%s", "severity": "%s" }`, name, severity)
	annotations := fmt.Sprintf(`{ "message": "%s" }`, message)
	b.DB.Create(&models.Alert{
		ClusterToken: firstCluster.Token,
		UpdatedAt:    time.Now().UTC(),
		Labels:       datatypes.JSON(labels),
		Annotations:  datatypes.JSON(annotations),
		Severity:     severity,
		StartsAt:     time.Now().UTC().Add(fireFor * -1),
		EndsAt:       time.Now().UTC().Add(fireFor),
	})

	return nil
}

func (b DatabaseGitopsTestRunner) AddWorkspace(env []string, clusterName string) error {
	var firstCluster models.Cluster
	b.DB.Where("Name = ?", clusterName).First(&firstCluster)

	b.DB.Create(&models.Workspace{
		ClusterToken: firstCluster.Token,
		Name:         "mccp-devs-workspace",
		Namespace:    "wkp-workspace",
	})

	return nil
}

func (b DatabaseGitopsTestRunner) CreateApplyCapitemplates(templateCount int, templateFile string) []string {
	templateFiles, err := generateTestCapiTemplates(templateCount, templateFile)
	Expect(err).To(BeNil(), "Failed to generate CAPITemplate template test files")
	By("Apply/Install CAPITemplate templates", func() {
		for _, fileName := range templateFiles {
			template, err := capi.ParseFile(fileName)
			Expect(err).To(BeNil(), "Failed to parse CAPITemplate template files")
			err = b.Client.Create(context.Background(), template)
			Expect(err).To(BeNil(), "Failed to create CAPITemplate template files")
		}
	})

	return templateFiles
}

func (b DatabaseGitopsTestRunner) DeleteApplyCapiTemplates(templateFiles []string) {
	By("Delete CAPITemplate templates", func() {
		for _, fileName := range templateFiles {
			template, err := capi.ParseFile(fileName)
			Expect(err).To(BeNil(), "Failed to parse CAPITemplate template files")
			err = b.Client.Delete(context.Background(), template)
			Expect(err).To(BeNil(), "Failed to delete CAPITemplate template files")
		}
	})
}

func (b DatabaseGitopsTestRunner) CheckClusterService(capiEndpointURL string) {

}

func (b DatabaseGitopsTestRunner) RestartDeploymentPods(env []string, appName string, namespace string) error {
	return nil
}

func (b DatabaseGitopsTestRunner) CreateIPCredentials(infrastructureProvider string) {

}

func (b DatabaseGitopsTestRunner) DeleteIPCredentials(infrastructureProvider string) {

}

func (b DatabaseGitopsTestRunner) DeleteRepo(repoName string) {

}

func (b DatabaseGitopsTestRunner) InitAndCreateEmptyRepo(repoName string, IsPrivateRepo bool) string {
	return ""
}

func (b DatabaseGitopsTestRunner) GitAddCommitPush(repoAbsolutePath string, fileToAdd string) {

}

func (b DatabaseGitopsTestRunner) CreateGitRepoBranch(repoAbsolutePath string, branchName string) string {
	return ""
}

func (b DatabaseGitopsTestRunner) PullBranch(repoAbsolutePath string, branch string) {

}

func (b DatabaseGitopsTestRunner) ListPullRequest(repoAbsolutePath string) []string {
	return []string{}
}

func (b DatabaseGitopsTestRunner) MergePullRequest(repoAbsolutePath string, prBranch string) {

}

func (b DatabaseGitopsTestRunner) GetRepoVisibility(org string, repo string) string {
	return ""
}

// "Real" backend that call kubectl and posts to alertmanagement

type RealGitopsTestRunner struct{}

func (b RealGitopsTestRunner) TimeTravelToLastSeen() error {
	return nil
}

func (b RealGitopsTestRunner) TimeTravelToAlertsResolved() error {
	return nil
}

func (b RealGitopsTestRunner) ResetDatabase() error {
	return runCommandPassThrough([]string{}, "../../utils/scripts/wego-enterprise.sh", "reset_mccp")
}

func (b RealGitopsTestRunner) VerifyWegoPodsRunning() {
	VerifyEnterpriseControllers("my-mccp", "", GITOPS_DEFAULT_NAMESPACE)
}

func (b RealGitopsTestRunner) KubectlApply(env []string, tokenURL string) error {
	err := runCommandPassThrough(env, "kubectl", "apply", "-f", tokenURL)
	fmt.Println("Cluster pods after apply")
	if err := runCommandPassThrough(env, "kubectl", "get", "pods", "-A"); err != nil {
		fmt.Printf("Error getting cluster pods after apply: %v\n", err)
	}
	return err
}

func (b RealGitopsTestRunner) KubectlDelete(env []string, tokenURL string) error {
	return runCommandPassThrough(env, "kubectl", "delete", "-f", tokenURL)
}

func (b RealGitopsTestRunner) KubectlDeleteAllAgents(env []string) error {
	return runCommandPassThrough(env, "kubectl", "delete", "-n", "wkp-agent", "deploy", "wkp-agent")
}

func (b RealGitopsTestRunner) FireAlert(name, severity, message string, fireFor time.Duration) error {
	const alertTemplate = `
    [
      {
        "labels": {
          "alertname": "{{ .Name }}",
          "severity": "{{ .Severity }}"
        },
        "annotations": {
          "message": "{{ .Message }}"
        },
        "startsAt": "{{ .StartsAt }}",
        "endsAt": "{{ .EndsAt }}"
      }
    ]
    `

	t, err := template.New("alert").Parse(alertTemplate)
	if err != nil {
		return err
	}
	var populated bytes.Buffer
	err = t.Execute(&populated, struct {
		Name     string
		Severity string
		Message  string
		StartsAt string
		EndsAt   string
	}{
		name,
		severity,
		message,
		time.Now().UTC().Add(fireFor * -1).Format(time.RFC3339),
		time.Now().UTC().Add(fireFor).Format(time.RFC3339),
	})

	if err != nil {
		return err
	}

	fmt.Print(populated.String())
	req, err := http.NewRequest("POST", GetWGEUrl()+"/alertmanager/api/v2/alerts", &populated)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("alertmanager didn't like the alert: %v", resp.StatusCode)
	}

	return nil
}

func (b RealGitopsTestRunner) AddWorkspace(env []string, clusterName string) error {
	return runCommandPassThrough(env, "kubectl", "apply", "-f", "../../utils/data/mccp-workspace.yaml")
}

// This function will crete the test capiTemplate files and do the kubectl apply for capiserver availability
func (b RealGitopsTestRunner) CreateApplyCapitemplates(templateCount int, templateFile string) []string {
	templateFiles, err := generateTestCapiTemplates(templateCount, templateFile)
	Expect(err).To(BeNil(), "Failed to generate CAPITemplate template test files")

	By("Apply/Install CAPITemplate templates", func() {
		for _, fileName := range templateFiles {
			err = runCommandPassThrough([]string{}, "kubectl", "apply", "-f", fileName)
			Expect(err).To(BeNil(), "Failed to apply/install CAPITemplate template files")
		}
	})

	return templateFiles
}

// This function deletes the test capiTemplate files and do the kubectl delete to clean the cluster
func (b RealGitopsTestRunner) DeleteApplyCapiTemplates(templateFiles []string) {
	By("Delete CAPITemplate templates", func() {

		for _, fileName := range templateFiles {
			err := b.KubectlDelete([]string{}, fileName)
			Expect(err).To(BeNil(), "Failed to delete CAPITemplate template")
		}
	})

	err := deleteFile(templateFiles)
	Expect(err).To(BeNil(), "Failed to delete CAPITemplate template test files")
}

func (b RealGitopsTestRunner) CheckClusterService(capiEndpointURL string) {
	output := func() string {
		command := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", capiEndpointURL+"/v1/templates")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		return string(session.Wait(ASSERTION_30SECONDS_TIME_OUT).Out.Contents())
	}
	Eventually(output, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp("200"), "Cluster Service is not healthy")
}

func (b RealGitopsTestRunner) RestartDeploymentPods(env []string, appName string, namespace string) error {
	// Restart the deployment pods
	err := runCommandPassThrough(env, "kubectl", "rollout", "restart", "deployment", appName, "-n", namespace)
	if err == nil {
		// Wait for all the deployments replicas to rolled out successfully
		err = runCommandPassThrough(env, "kubectl", "rollout", "status", "deployment", appName, "-n", namespace)
	}
	return err
}

func (b RealGitopsTestRunner) CreateIPCredentials(infrastructureProvider string) {
	if infrastructureProvider == "AWS" {
		By("Install AWSClusterStaticIdentity CRD", func() {
			err := runCommandPassThrough([]string{}, "kubectl", "apply", "-f", "../../utils/data/infrastructure.cluster.x-k8s.io_awsclusterstaticidentities.yaml")
			Expect(err).To(BeNil(), "Failed to install AWSClusterStaticIdentity CRD")
			err = runCommandPassThrough([]string{}, "kubectl", "wait", "--for=condition=established", "--timeout=90s", "crd/awsclusterstaticidentities.infrastructure.cluster.x-k8s.io")
			Expect(err).To(BeNil(), "Failed to verify AWSClusterStaticIdentity CRD")
		})

		By("Install AWSClusterRoleIdentity CRD", func() {
			err := runCommandPassThrough([]string{}, "kubectl", "apply", "-f", "../../utils/data/infrastructure.cluster.x-k8s.io_awsclusterroleidentities.yaml")
			Expect(err).To(BeNil(), "Failed to install AWSClusterRoleIdentity CRD")
			err = runCommandPassThrough([]string{}, "kubectl", "wait", "--for=condition=established", "--timeout=90s", "crd/awsclusterroleidentities.infrastructure.cluster.x-k8s.io")
			Expect(err).To(BeNil(), "Failed to verify AWSClusterRoleIdentity CRD")
		})

		By("Create AWS Secret, AWSClusterStaticIdentity and AWSClusterRoleIdentity)", func() {
			err := runCommandPassThrough([]string{}, "kubectl", "apply", "-f", "../../utils/data/aws_cluster_credentials.yaml")
			Expect(err).To(BeNil(), "Failed to create AWS credentials")
		})

	} else if infrastructureProvider == "AZURE" {
		By("Install AzureClusterIdentity CRD", func() {
			err := runCommandPassThrough([]string{}, "kubectl", "apply", "-f", "../../utils/data/infrastructure.cluster.x-k8s.io_azureclusteridentities.yaml")
			Expect(err).To(BeNil(), "Failed to install AzureClusterIdentity CRD")
			err = runCommandPassThrough([]string{}, "kubectl", "wait", "--for=condition=established", "--timeout=90s", "crd/azureclusteridentities.infrastructure.cluster.x-k8s.io")
			Expect(err).To(BeNil(), "Failed to verify AzureClusterIdentity CRD")
		})

		By("Create Azure Secret and AzureClusterIdentity)", func() {
			err := runCommandPassThrough([]string{}, "kubectl", "apply", "-f", "../../utils/data/azure_cluster_credentials.yaml")
			Expect(err).To(BeNil(), "Failed to create Azure credentials")
		})
	}

}

func (b RealGitopsTestRunner) DeleteIPCredentials(infrastructureProvider string) {
	if infrastructureProvider == "AWS" {
		_ = runCommandPassThrough([]string{}, "kubectl", "delete", "-f", "../../utils/data/aws_cluster_credentials.yaml")
		_ = runCommandPassThrough([]string{}, "kubectl", "delete", "-f", "../../utils/data/infrastructure.cluster.x-k8s.io_awsclusterroleidentities.yaml")
		_ = runCommandPassThrough([]string{}, "kubectl", "delete", "-f", "../../utils/data/infrastructure.cluster.x-k8s.io_awsclusterstaticidentities.yaml")

	} else if infrastructureProvider == "AZURE" {
		_ = runCommandPassThrough([]string{}, "kubectl", "delete", "-f", "../../utils/data/azure_cluster_credentials.yaml")
		_ = runCommandPassThrough([]string{}, "kubectl", "delete", "-f", "../../utils/data/infrastructure.cluster.x-k8s.io_azureclusteridentities.yaml")
	}
}

func (b RealGitopsTestRunner) DeleteRepo(repoName string) {
	log.Printf("Delete application repo: %s", path.Join(GITHUB_ORG, repoName))
	_ = runCommandPassThrough([]string{}, "hub", "delete", "-y", path.Join(GITHUB_ORG, repoName))

	output := func() string {
		command := exec.Command("sh", "-c", fmt.Sprintf(`git ls-remote https://github.com/%s/%s`, GITHUB_ORG, GITHUB_ORG))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		return string(session.Wait().Err.Contents())
	}
	Eventually(output, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp("Repository not found"))
}

func (b RealGitopsTestRunner) InitAndCreateEmptyRepo(repoName string, IsPrivateRepo bool) string {
	log.Printf("Init and create repo: %s, %v\n", repoName, IsPrivateRepo)
	repoAbsolutePath := path.Join("/tmp/", repoName)
	privateRepo := ""
	if IsPrivateRepo {
		privateRepo = "-p"
	}
	command := exec.Command("sh", "-c", fmt.Sprintf(`
                            mkdir %s &&
                            cd %s &&
                            git init &&
                            git checkout -b main &&
                            hub create %s %s`, repoAbsolutePath, repoAbsolutePath, path.Join(GITHUB_ORG, repoName), privateRepo))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit(), "err %v, out, %v", string(session.Err.Contents()), string(session.Out.Contents()))

	log.Printf("Waiting for repo to be created %v/%v", GITHUB_ORG, repoName)
	Expect(WaitUntil(os.Stdout, POLL_INTERVAL_5SECONDS, ASSERTION_1MINUTE_TIME_OUT, func() error {
		cmd := fmt.Sprintf(`hub api repos/%s/%s`, GITHUB_ORG, repoName)
		command := exec.Command("sh", "-c", cmd)
		return command.Run()
	})).ShouldNot(HaveOccurred())

	return repoAbsolutePath
}

func (b RealGitopsTestRunner) GitAddCommitPush(repoAbsolutePath string, fileToAdd string) {
	command := exec.Command("sh", "-c", fmt.Sprintf(`
                            cp -r -f %s %s &&
                            cd %s &&
                            git add . &&
                            git commit -m 'add workload manifest' &&
                            git push -u origin main`, fileToAdd, repoAbsolutePath, repoAbsolutePath))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	fmt.Println(string(session.Wait().Err.Contents()))
}

func GitUpdateCommitPush(repoAbsolutePath string, commitMessage string) {
	log.Infof("Pushing changes made to file(s) in repo: %s", repoAbsolutePath)
	if commitMessage == "" {
		commitMessage = "edit repo file"
	}

	_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("cd %s && git add -u && git add -A && git commit -m '%s' && git pull --rebase && git push origin HEAD", repoAbsolutePath, commitMessage))
}

func GitSetUpstream(repoAbsolutePath string, upstreamBranch string) {
	log.Infof("Setting tracking/upstream remote branch %s", upstreamBranch)
	_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("cd %s && git branch -u origin/%s", repoAbsolutePath, upstreamBranch))
}

func GetGitRepositoryURL(repoAbsolutePath string) string {
	repoURL, _ := runCommandAndReturnStringOutput(fmt.Sprintf(`cd %s && git config --get remote.origin.url`, repoAbsolutePath))
	return strings.Trim(repoURL, "\n")
}

func (b RealGitopsTestRunner) CreateGitRepoBranch(repoAbsolutePath string, branchName string) string {
	command := exec.Command("sh", "-c", fmt.Sprintf("cd %s && git checkout -b %s && git push --set-upstream origin %s", repoAbsolutePath, branchName, branchName))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	return string(session.Wait().Out.Contents())
}

func (b RealGitopsTestRunner) PullBranch(repoAbsolutePath string, branch string) {
	command := exec.Command("sh", "-c", fmt.Sprintf(`
                            cd %s &&
                            git pull origin %s --rebase`, repoAbsolutePath, branch))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
}

func (b RealGitopsTestRunner) ListPullRequest(repoAbsolutePath string) []string {
	command := exec.Command("sh", "-c", fmt.Sprintf(`
                            cd %s &&
                            hub pr list --limit 1 --base main --format='%%t|%%H|%%U%%n'`, repoAbsolutePath))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())

	return strings.Split(string(session.Wait().Out.Contents()), "|")
}

func (b RealGitopsTestRunner) GetRepoVisibility(org string, repo string) string {
	command := exec.Command("sh", "-c", fmt.Sprintf("hub api --flat repos/%s/%s|grep -i private|cut -f2", org, repo))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	visibilityStr := strings.TrimSpace(string(session.Wait().Out.Contents()))
	log.Printf("Repo visibility private=%s", visibilityStr)
	return visibilityStr
}

func (b RealGitopsTestRunner) MergePullRequest(repoAbsolutePath string, prBranch string) {
	command := exec.Command("sh", "-c", fmt.Sprintf(`
                            cd %s &&
							git fetch
							git checkout main &&
							git pull --no-ff &&
                            git merge --no-ff --no-edit origin/%s &&
							git push origin main`, repoAbsolutePath, prBranch))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
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

	return string(session.Wait().Out.Contents()), string(session.Wait().Err.Contents())
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

// showItems displays the current set of a specified object type in tabular format
func showItems(itemType string) error {
	if itemType != "" {
		return runCommandPassThrough([]string{}, "kubectl", "get", itemType, "--all-namespaces", "-o", "wide")
	}
	return runCommandPassThrough([]string{}, "kubectl", "get", "all", "--all-namespaces", "-o", "wide")
}

func dumpClusterInfo(namespaces, testName string) error {
	return runCommandPassThrough([]string{}, "../../utils/scripts/dump-cluster-info.sh", namespaces, testName, CLUSTER_INFO_DIR)
}

// This function generates multiple capitemplate files from a single capitemplate to be used as test data
func generateTestCapiTemplates(templateCount int, templateFile string) (templateFiles []string, err error) {
	// Read input capitemplate
	contents, err := ioutil.ReadFile(fmt.Sprintf("../../utils/data/%s", templateFile))

	if err != nil {
		return templateFiles, err
	}

	// Prepare  data to insert into the template.
	type TemplateInput struct {
		Count int
	}

	// Create a new template and parse the letter into it.
	t := template.Must(template.New("capi-template").Parse(string(contents)))

	// Execute the template for each count.
	for i := 0; i < templateCount; i++ {
		input := TemplateInput{i}

		fileName := fmt.Sprintf("%s%d", templateFile, i)

		f, err := os.Create(path.Join("/tmp", fileName))
		if err != nil {
			return templateFiles, err
		}
		templateFiles = append(templateFiles, f.Name())

		err = t.Execute(f, input)
		if err != nil {
			log.Println("executing template:", err)
		}

		f.Close()
	}

	return templateFiles, nil
}

// Utility function to delete all the files passed in a list
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

//Utility function delete directory
func deleteDirectory(name []string) error {
	return deleteFile(name)
}

// Utility function to check if file exists
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

//Utility function to compute public IP of cluster workload node
func ClusterWorkloadNonePublicIP(clusterKind string) string {
	var expernal_ip string
	if clusterKind == "EKS" || clusterKind == "GKE" {
		node_name, _ := runCommandAndReturnStringOutput(`kubectl get node --selector='!node-role.kubernetes.io/master' -o name | head -n 1`)
		worker_name := strings.Trim(strings.Split(node_name, "/")[1], "\n")
		expernal_ip, _ = runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl get nodes -o jsonpath="{.items[?(@.metadata.name=='%s')].status.addresses[?(@.type=='ExternalIP')].address}"`, worker_name))
	} else {
		switch runtime.GOOS {
		case "darwin":
			expernal_ip, _ = runCommandAndReturnStringOutput(`ifconfig en0 | grep -i MASK | awk '{print $2}' | cut -f2 -d:`)
		case "linux":
			expernal_ip, _ = runCommandAndReturnStringOutput(`ifconfig eth0 | grep -i MASK | awk '{print $2}' | cut -f2 -d:`)
		}
	}
	return strings.Trim(expernal_ip, "\n")
}

func CreateCluster(clusterType string, clusterName string, configFile string) {
	if clusterType == "kind" {
		err := runCommandPassThrough([]string{}, "kind", "create", "cluster", "--name", clusterName, "--image=kindest/node:v1.20.7", "--config", "../../utils/data/"+configFile)
		Expect(err).ShouldNot(HaveOccurred())
	} else {
		Fail(fmt.Sprintf("%s cluster type is not supported for test WGE upgrade", clusterType))
	}
}

func createTestFile(fileName string, fileContents string) string {
	testFilePath := filepath.Join(os.TempDir(), fileName)

	command := exec.Command("sh", "-c", fmt.Sprintf(`
							cd /tmp &&
                            echo "%s" > %s`, fileContents, testFilePath))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())

	return testFilePath
}

func deleteClusters(clusterType string, clusters []string) {
	for _, cluster := range clusters {
		if clusterType == "kind" {
			log.Printf("Deleting cluster: %s", cluster)
			err := runCommandPassThrough([]string{}, "kind", "delete", "cluster", "--name", cluster)
			Expect(err).ShouldNot(HaveOccurred())
		} else {
			err := runCommandPassThrough([]string{}, "kubectl", "get", "cluster", cluster)
			if err == nil {
				log.Printf("Deleting cluster: %s", cluster)
				err := runCommandPassThrough([]string{}, "kubectl", "delete", "cluster", cluster)
				Expect(err).ShouldNot(HaveOccurred())
				err = runCommandPassThrough([]string{}, "kubectl", "get", "cluster", cluster)
				Expect(err).Should(HaveOccurred(), fmt.Sprintf("Failed to delete cluster %s", cluster))
			}
		}
	}
}

func installInfrastructureProvider(name string) {
	if name == "docker" {
		command := exec.Command("clusterctl", "init", "--infrastructure", "docker")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, ASSERTION_2MINUTE_TIME_OUT).Should(gexec.Exit())

		Eventually(string(session.Wait().Err.Contents())).Should(MatchRegexp(`(Installing Provider="infrastructure-docker"|installing provider "infrastructure-docker")`))
	} else {
		Fail(fmt.Sprintf("%s infrastructure Provider is not supported for test run", name))
	}
}

// gitops system helper functions
func waitForResource(resourceType string, resourceName string, namespace string, timeout time.Duration) error {
	pollInterval := 5
	if timeout < 5*time.Second {
		timeout = 5 * time.Second
	}

	timeoutInSeconds := int(timeout.Seconds())
	for i := pollInterval; i < timeoutInSeconds; i += pollInterval {
		log.Infof("Waiting for %s in namespace: %s... : %d second(s) passed of %d seconds timeout", resourceType+"/"+resourceName, namespace, i, timeoutInSeconds)
		err := runCommandPassThroughWithoutOutput([]string{}, "sh", "-c", fmt.Sprintf("kubectl get %s %s -n %s", resourceType, resourceName, namespace))
		if err == nil {
			log.Infof("%s are available in cluster", resourceType+"/"+resourceName)
			command := exec.Command("sh", "-c", fmt.Sprintf("kubectl get %s %s -n %s", resourceType, resourceName, namespace))
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
			noResourcesFoundMessage := fmt.Sprintf("No resources found in %s namespace", namespace)
			if strings.Contains(string(session.Wait().Out.Contents()), noResourcesFoundMessage) {
				log.Infof("Got message => {" + noResourcesFoundMessage + "} Continue looking for resource(s)")
				continue
			}
			return nil
		}
		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
	return fmt.Errorf("error: Failed to find the resource %s of type %s, timeout reached", resourceName, resourceType)
}

func VerifyCoreControllers(namespace string) {
	Expect(waitForResource("deploy", "helm-controller", namespace, ASSERTION_2MINUTE_TIME_OUT))
	Expect(waitForResource("deploy", "kustomize-controller", namespace, ASSERTION_2MINUTE_TIME_OUT))
	Expect(waitForResource("deploy", "notification-controller", namespace, ASSERTION_2MINUTE_TIME_OUT))
	Expect(waitForResource("deploy", "source-controller", namespace, ASSERTION_2MINUTE_TIME_OUT))
	Expect(waitForResource("deploy", "image-automation-controller", namespace, ASSERTION_2MINUTE_TIME_OUT))
	Expect(waitForResource("deploy", "image-reflector-controller", namespace, ASSERTION_2MINUTE_TIME_OUT))
	Expect(waitForResource("pods", "", namespace, ASSERTION_2MINUTE_TIME_OUT))

	By("And I wait for the gitops core controllers to be ready", func() {
		command := exec.Command("sh", "-c", fmt.Sprintf("kubectl wait --for=condition=Ready --timeout=%s -n %s --all pod --selector='app!=wego-app'", "180s", namespace))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, ASSERTION_3MINUTE_TIME_OUT).Should(gexec.Exit())
	})
}

func VerifyEnterpriseControllers(releaseName string, mccpPrefix, namespace string) {
	// SOMETIMES (?) (with helm install ./local-path), the mccpPrefix is skipped
	Expect(waitForResource("deploy", releaseName+"-"+mccpPrefix+"gitops-repo-broker", namespace, ASSERTION_2MINUTE_TIME_OUT))
	Expect(waitForResource("deploy", releaseName+"-"+mccpPrefix+"event-writer", namespace, ASSERTION_2MINUTE_TIME_OUT))
	Expect(waitForResource("deploy", releaseName+"-"+mccpPrefix+"cluster-service", namespace, ASSERTION_2MINUTE_TIME_OUT))
	Expect(waitForResource("deploy", releaseName+"-nginx-ingress-controller", namespace, ASSERTION_2MINUTE_TIME_OUT))
	// FIXME
	// const maxDeploymentLength = 63
	// Expect(waitForResource("deploy", (releaseName + "-nginx-ingress-controller-default-backend")[:maxDeploymentLength], namespace, ASSERTION_2MINUTE_TIME_OUT))
	Expect(waitForResource("deploy", releaseName+"-wkp-ui-server", namespace, ASSERTION_2MINUTE_TIME_OUT))
	Expect(waitForResource("pods", "", namespace, ASSERTION_2MINUTE_TIME_OUT))

	By("And I wait for the gitops enterprise controllers to be ready", func() {
		command := exec.Command("sh", "-c", fmt.Sprintf("kubectl wait --for=condition=Ready --timeout=%s -n %s --all pod --selector='app!=wego-app'", "180s", namespace))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, ASSERTION_3MINUTE_TIME_OUT).Should(gexec.Exit())
	})
}

func verifyWegoAddCommand(appName string, namespace string) {
	command := exec.Command("sh", "-c", fmt.Sprintf(" kubectl wait --for=condition=Ready --timeout=60s -n %s GitRepositories --all", namespace))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session, ASSERTION_5MINUTE_TIME_OUT).Should(gexec.Exit())
	Expect(waitForResource("GitRepositories", appName, namespace, ASSERTION_5MINUTE_TIME_OUT)).To(Succeed())
}

func InstallAndVerifyGitops(gitopsNamespace string, manifestRepoURL string) {
	By("And I run 'gitops install' command with namespace "+gitopsNamespace, func() {
		command := exec.Command("sh", "-c", fmt.Sprintf("%s install --app-config-url %s --namespace=%s", GetGitopsBinPath(), manifestRepoURL, gitopsNamespace))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, ASSERTION_2MINUTE_TIME_OUT).Should(gexec.Exit())
		Expect(string(session.Err.Contents())).Should(BeEmpty())
		VerifyCoreControllers(gitopsNamespace)
	})
}

func InstallAndVerifyPctl(gitopsNamespace string) {
	By("And I run 'pctl install' command with flux-namespace "+gitopsNamespace, func() {
		command := exec.Command("sh", "-c", fmt.Sprintf("%s install --flux-namespace=%s", GetPctlBinPath(), gitopsNamespace))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, ASSERTION_2MINUTE_TIME_OUT).Should(gexec.Exit())
		Expect(string(session.Err.Contents())).Should(BeEmpty())

		By("And I wait for the pctl controller to be ready", func() {
			command := exec.Command("sh", "-c", "kubectl wait --for=condition=Ready --timeout=120s -n profiles-system --all pod")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session, ASSERTION_2MINUTE_TIME_OUT).Should(gexec.Exit())
		})
	})
}

func RemoveGitopsCapiClusters(appName string, clusternames []string, nameSpace string) {
	SusspendGitopsApplication(appName, nameSpace)

	deleteClusters("capi", clusternames)

	DeleteGitopsApplication(appName, nameSpace)
	DeleteGitopsDeploySecret(nameSpace)
}

func SusspendGitopsApplication(appName string, nameSpace string) {
	command := fmt.Sprintf("suspend app %s", appName)
	By(fmt.Sprintf("And I run gitops suspend app command '%s'", command), func() {
		command := exec.Command("sh", "-c", fmt.Sprintf("%s %s", GetGitopsBinPath(), command))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit())
	})
}

func ListGitopsApplication(appName string, nameSpace string) string {
	var session *gexec.Session
	var err error

	cmd := fmt.Sprintf("get app %s", appName)
	command := exec.Command("sh", "-c", fmt.Sprintf("%s %s", GetGitopsBinPath(), cmd))
	session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())

	return string(session.Out.Contents())
}

func DeleteGitopsApplication(appName string, nameSpace string) {
	command := "delete app " + appName
	By(fmt.Sprintf("And I run gitops delete app command '%s'", command), func() {
		command := exec.Command("sh", "-c", fmt.Sprintf("%s %s", GetGitopsBinPath(), command))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit())

		appDeleted := func() bool {
			status := ListGitopsApplication(appName, nameSpace)
			return status == ""
		}
		Eventually(appDeleted, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(BeTrue(), fmt.Sprintf("%s application failed to delete", appName))
	})
}

func DeleteGitopsDeploySecret(nameSpace string) {
	command := fmt.Sprintf(`kubectl get secrets -n %[1]v  | grep Opaque | grep wego- | cut -d' ' -f1 | xargs kubectl delete secrets -n %[1]v`, nameSpace)
	By("And I delete deploy key secret", func() {
		command := exec.Command("sh", "-c", command)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit())
	})
}
