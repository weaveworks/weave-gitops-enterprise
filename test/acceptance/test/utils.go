package acceptance

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/agouti"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/capi"
	"github.com/weaveworks/wks/common/database/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/types"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var webDriver *agouti.Page
var gitProvider string
var seleniumServiceUrl string
var defaultUIURL = "http://localhost:8090"
var defaultMccpBinPath = "/usr/local/bin/mccp"
var defaultCapiEndpointURL = "http://localhost:8090"

func GetWebDriver() *agouti.Page {
	return webDriver
}

func SetWebDriver(wb *agouti.Page) {
	webDriver = wb
}

func GetMCCBinPath() string {
	if os.Getenv("MCCP_BIN_PATH") != "" {
		return os.Getenv("MCCP_BIN_PATH")
	}
	return defaultMccpBinPath
}

func GetWkpUrl() string {
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
	seleniumServiceUrl = url
}

var GITHUB_ORG string
var CLUSTER_REPOSITORY string

const ARTEFACTS_BASE_DIR string = "/tmp/workspace/test/"
const SCREENSHOTS_DIR string = ARTEFACTS_BASE_DIR + "screenshots/"
const JUNIT_TEST_REPORT_FILE string = ARTEFACTS_BASE_DIR + "wkp_junit.xml"

const ASSERTION_DEFAULT_TIME_OUT time.Duration = 15 * time.Second
const ASSERTION_10SECONDS_TIME_OUT time.Duration = 10 * time.Second
const ASSERTION_1SECOND_TIME_OUT time.Duration = 1 * time.Second
const ASSERTION_1MINUTE_TIME_OUT time.Duration = 1 * time.Minute
const ASSERTION_2MINUTE_TIME_OUT time.Duration = 2 * time.Minute
const ASSERTION_5MINUTE_TIME_OUT time.Duration = 5 * time.Minute
const ASSERTION_6MINUTE_TIME_OUT time.Duration = 6 * time.Minute

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
		webDriver.Screenshot(filepath)
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
func DescribeSpecsMccpUi(mccpTestRunner MCCPTestRunner) {
	DescribeMCCPClusters(mccpTestRunner)
	DescribeMCCPTemplates(mccpTestRunner)
}

// Describes all the CLI acceptance tests
func DescribeSpecsMccpCli(mccpTestRunner MCCPTestRunner) {
	DescribeMccpCliHelp()
	DescribeMccpCliList(mccpTestRunner)
	DescribeMccpCliRender(mccpTestRunner)
}

// Interface that can be implemented either with:
// - "Real" commands like "exec(kubectl...)"
// - "Mock" commands like db.Create(cluster_info...)

type MCCPTestRunner interface {
	ResetDatabase() error
	FireAlert(name, severity, message string, fireFor time.Duration) error
	KubectlApply(env []string, tokenURL string) error
	KubectlDelete(env []string, tokenURL string) error
	KubectlDeleteAllAgents(env []string) error
	TimeTravelToLastSeen() error
	TimeTravelToAlertsResolved() error
	AddWorkspace(env []string, clusterName string) error
	CreateApplyCapitemplates(templateCount int, templateFile string) []string
	DeleteApplyCapiTemplates(templateFiles []string)
}

// "DB" backend that creates/delete rows

type DatabaseMCCPTestRunner struct {
	DB     *gorm.DB
	Client goclient.Client
}

func (b DatabaseMCCPTestRunner) TimeTravelToLastSeen() error {
	oneMinuteAgo := time.Now().UTC().Add(time.Minute * -2)
	b.DB.Exec("update cluster_info set updated_at = ?", oneMinuteAgo)
	return nil
}

func (b DatabaseMCCPTestRunner) TimeTravelToAlertsResolved() error {
	b.DB.Where("1 = 1").Delete(&models.Alert{})
	return nil
}

func (b DatabaseMCCPTestRunner) ResetDatabase() error {
	b.DB.Where("1 = 1").Delete(&models.Cluster{})
	return nil
}

func (b DatabaseMCCPTestRunner) KubectlApply(env []string, tokenURL string) error {
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

func (b DatabaseMCCPTestRunner) KubectlDelete(env []string, tokenURL string) error {
	//
	// No more cluster_infos will be created anyway..
	// FIXME: maybe we add a polling loop that keeps creating cluster_info while its connected
	//
	return nil
}

func (b DatabaseMCCPTestRunner) KubectlDeleteAllAgents(env []string) error {
	// No more cluster_infos will be created anyway..
	return nil
}

func (b DatabaseMCCPTestRunner) FireAlert(name, severity, message string, fireFor time.Duration) error {
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

func (b DatabaseMCCPTestRunner) AddWorkspace(env []string, clusterName string) error {
	var firstCluster models.Cluster
	b.DB.Where("Name = ?", clusterName).First(&firstCluster)

	b.DB.Create(&models.Workspace{
		ClusterToken: firstCluster.Token,
		Name:         "mccp-devs-workspace",
		Namespace:    "wkp-workspace",
	})

	return nil
}

func (b DatabaseMCCPTestRunner) CreateApplyCapitemplates(templateCount int, templateFile string) []string {
	templateFiles, err := generateTestCapiTemplates(templateCount, templateFile)
	Expect(err).To(BeNil(), "Failed to generate CAPITemplate template test files")
	By("Apply/Insall CAPITemplate templates", func() {
		for _, fileName := range templateFiles {
			template, err := capi.ParseFile(fileName)
			Expect(err).To(BeNil(), "Failed to parse CAPITemplate template files")
			err = b.Client.Create(context.Background(), template)
			Expect(err).To(BeNil(), "Failed to create CAPITemplate template files")
		}
	})

	return templateFiles
}

func (b DatabaseMCCPTestRunner) DeleteApplyCapiTemplates(templateFiles []string) {
	By("Delete CAPITemplate templates", func() {
		for _, fileName := range templateFiles {
			template, err := capi.ParseFile(fileName)
			Expect(err).To(BeNil(), "Failed to parse CAPITemplate template files")
			err = b.Client.Delete(context.Background(), template)
			Expect(err).To(BeNil(), "Failed to delete CAPITemplate template files")
		}
	})
}

// "Real" backend that call kubectl and posts to alertmanagement

type RealMCCPTestRunner struct{}

func (b RealMCCPTestRunner) TimeTravelToLastSeen() error {
	return nil
}

func (b RealMCCPTestRunner) TimeTravelToAlertsResolved() error {
	return nil
}

func (b RealMCCPTestRunner) ResetDatabase() error {
	return runCommandPassThrough([]string{}, "../../utils/scripts/mccp-setup-helpers.sh", "reset")
}

func (b RealMCCPTestRunner) KubectlApply(env []string, tokenURL string) error {
	err := runCommandPassThrough(env, "kubectl", "apply", "-f", tokenURL)
	fmt.Println("Leaf cluster pods after apply")
	if err := runCommandPassThrough(env, "kubectl", "get", "pods", "-A"); err != nil {
		fmt.Printf("Error getting leaf cluster pods after apply: %v\n", err)
	}
	return err
}

func (b RealMCCPTestRunner) KubectlDelete(env []string, tokenURL string) error {
	return runCommandPassThrough(env, "kubectl", "delete", "-f", tokenURL)
}

func (b RealMCCPTestRunner) KubectlDeleteAllAgents(env []string) error {
	return runCommandPassThrough(env, "kubectl", "delete", "-n", "wkp-agent", "deploy", "wkp-agent")
}

func (b RealMCCPTestRunner) FireAlert(name, severity, message string, fireFor time.Duration) error {
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
	req, err := http.NewRequest("POST", GetWkpUrl()+"/alertmanager/api/v2/alerts", &populated)
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
		return fmt.Errorf("Alertmanager didn't like the alert: %v", resp.StatusCode)
	}

	return nil
}

func (b RealMCCPTestRunner) AddWorkspace(env []string, clusterName string) error {
	return runCommandPassThrough(env, "kubectl", "apply", "-f", "../../utils/data/mccp-workspace.yaml")
}

// This function will crete the test capiTemplate files and do the kubectl apply for capiserver availability
func (b RealMCCPTestRunner) CreateApplyCapitemplates(templateCount int, templateFile string) []string {
	templateFiles, err := generateTestCapiTemplates(templateCount, templateFile)
	Expect(err).To(BeNil(), "Failed to generate CAPITemplate template test files")

	By("Apply/Insall CAPITemplate templates", func() {
		for _, fileName := range templateFiles {
			err = runCommandPassThrough([]string{}, "kubectl", "apply", "-f", fileName)
			Expect(err).To(BeNil(), "Failed to apply/install CAPITemplate template files")
		}
	})

	return templateFiles
}

// This function deletes the test capiTemplate files and do the kubectl delete to clean the cluster
func (b RealMCCPTestRunner) DeleteApplyCapiTemplates(templateFiles []string) {
	By("Delete CAPITemplate templates", func() {

		for _, fileName := range templateFiles {
			err := b.KubectlDelete([]string{}, fileName)
			Expect(err).To(BeNil(), "Failed to delete CAPITemplate template")
		}
	})

	err := deleteFile(templateFiles)
	Expect(err).To(BeNil(), "Failed to delete CAPITemplate template test files")
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

// ***************** Repository helper functions ************************

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

func deleteRepo(repoName string) {
	log.Printf("Delete application repo: %s", path.Join(GITHUB_ORG, repoName))
	_ = runCommandPassThrough([]string{}, "hub", "delete", "-y", path.Join(GITHUB_ORG, repoName))
}

func initAndCreateEmptyRepo(repoName string, IsPrivateRepo bool) string {
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
	Eventually(session).Should(gexec.Exit())

	Expect(WaitUntil(os.Stdout, time.Second, 20*time.Second, func() error {
		cmd := fmt.Sprintf(`hub api repos/%s/%s`, GITHUB_ORG, repoName)
		command := exec.Command("sh", "-c", cmd)
		return command.Run()
	})).ShouldNot(HaveOccurred())

	return repoAbsolutePath
}

func gitAddCommitPush(repoAbsolutePath string, fileToAdd string) {
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

func pullBranch(repoAbsolutePath string, branch string) {
	command := exec.Command("sh", "-c", fmt.Sprintf(`
                            cd %s &&
                            git pull origin %s`, repoAbsolutePath, branch))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
}

func listPullRequest(repoAbsolutePath string) []string {
	command := exec.Command("sh", "-c", fmt.Sprintf(`
                            cd %s &&
                            hub pr list --limit 1 --base main --format='%%t|%%H|%%U'`, repoAbsolutePath))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())

	return strings.Split(string(session.Wait().Out.Contents()), "|")
}

func getRepoVisibility(org string, repo string) string {
	command := exec.Command("sh", "-c", fmt.Sprintf("hub api --flat repos/%s/%s|grep -i private|cut -f2", org, repo))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	visibilityStr := strings.TrimSpace(string(session.Wait().Out.Contents()))
	log.Printf("Repo visibility private=%s", visibilityStr)
	return visibilityStr
}
