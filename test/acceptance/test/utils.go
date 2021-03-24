package acceptance

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/sclevine/agouti"
	"github.com/weaveworks/wks/common/database/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/types"
)

var webDriver *agouti.Page
var gitProvider string
var seleniumServiceUrl string
var wkpUrl = "http://localhost:8090"

func GetWebDriver() *agouti.Page {
	return webDriver
}

func SetWebDriver(wb *agouti.Page) {
	webDriver = wb
}

func SetWkpUrl(url string) {
	wkpUrl = url
}

func SetSeleniumServiceUrl(url string) {
	seleniumServiceUrl = url
}

const ARTEFACTS_BASE_DIR string = "/tmp/workspace/test/"
const SCREENSHOTS_DIR string = ARTEFACTS_BASE_DIR + "screenshots/"
const JUNIT_TEST_REPORT_FILE string = ARTEFACTS_BASE_DIR + "wkp_junit.xml"

const ASSERTION_DEFAULT_TIME_OUT time.Duration = 15 * time.Second // 15 seconds
const ASSERTION_1MINUTE_TIME_OUT time.Duration = 1 * time.Minute  // 1 Minute
const ASSERTION_5MINUTES_TIME_OUT time.Duration = 5 * time.Minute // 5 Minutes

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

var n = 1

func TakeNextScreenshot() {
	TakeScreenShot(fmt.Sprintf("test-%v", n))
	n += 1
}

// Interface that can be implemented either with:
// - "Real" commands like "exec(kubectl...)"
// - "Mock" commands like db.Create(cluster_info...)

type MCCPTestRunner interface {
	ResetDatabase() error
	FireAlert(name, severity, message string, fireFor time.Duration) error
	KubectlApply(tokenURL string) error
	KubectlDelete(tokenURL string) error
	KubectlDeleteAllAgents() error
	TimeTravelToLastSeen() error
	TimeTravelToAlertsResolved() error
}

// "DB" backend that creates/delete rows

type DatabaseMCCPTestRunner struct {
	DB *gorm.DB
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

func (b DatabaseMCCPTestRunner) KubectlApply(tokenURL string) error {
	bits := strings.Split(tokenURL, "=")
	token := bits[len(bits)-1]
	b.DB.Create(&models.ClusterInfo{
		UID:          types.UID(String(10)),
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

func (b DatabaseMCCPTestRunner) KubectlDelete(tokenURL string) error {
	//
	// No more cluster_infos will be created anyway..
	// FIXME: maybe we add a polling loop that keeps creating cluster_info while its connected
	//
	return nil
}

func (b DatabaseMCCPTestRunner) KubectlDeleteAllAgents() error {
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

// "Real" backend that call kubectl and posts to alertmanagement

type RealMCCPTestRunner struct{}

func (b RealMCCPTestRunner) TimeTravelToLastSeen() error {
	return nil
}

func (b RealMCCPTestRunner) TimeTravelToAlertsResolved() error {
	return nil
}

func (b RealMCCPTestRunner) ResetDatabase() error {
	return runCommandPassThrough("../../utils/scripts/mccp-setup-helpers.sh", "reset")
}

func (b RealMCCPTestRunner) KubectlApply(tokenURL string) error {
	return runCommandPassThrough("kubectl", "apply", "-f", tokenURL)
}

func (b RealMCCPTestRunner) KubectlDelete(tokenURL string) error {
	return runCommandPassThrough("kubectl", "delete", "-f", tokenURL)
}

func (b RealMCCPTestRunner) KubectlDeleteAllAgents() error {
	return runCommandPassThrough("kubectl", "delete", "-n", "wkp-agent", "deploy", "wkp-agent")
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
	req, err := http.NewRequest("POST", wkpUrl+"/alertmanager/api/v2/alerts", &populated)
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

// Run a command, passing through stdout/stderr to the OS standard streams
func runCommandPassThrough(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
