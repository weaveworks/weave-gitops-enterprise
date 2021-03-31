package test

import (
	gcontext "context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	broker "github.com/weaveworks/wks/cmd/gitops-repo-broker/server"
	"github.com/weaveworks/wks/common/database/models"
	"github.com/weaveworks/wks/common/database/utils"
	acceptancetest "github.com/weaveworks/wks/test/acceptance/test"
	"github.com/weaveworks/wks/test/acceptance/test/pages"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

//
// Test suite
//

const uiURL = "http://localhost:4046"
const brokerURL = "http://localhost:8000"
const seleniumURL = "http://localhost:4444/wd/hub"

var db *gorm.DB
var dbURI string

func resetDb(db *gorm.DB) {
	db.Where("1 = 1").Delete(&models.Cluster{})
	db.Where("1 = 1").Delete(&models.Alert{})
	db.Where("1 = 1").Delete(&models.ClusterInfo{})
}

func createAlert(db *gorm.DB, token, name, severity, message string, fireFor time.Duration) {
	labels := fmt.Sprintf(`{ "alertname": "%s", "severity": "%s" }`, name, severity)
	annotations := fmt.Sprintf(`{ "message": "%s" }`, message)
	db.Create(&models.Alert{
		ClusterToken: token,
		UpdatedAt:    time.Now().UTC(),
		Labels:       datatypes.JSON(labels),
		Annotations:  datatypes.JSON(annotations),
		Severity:     severity,
		StartsAt:     time.Now().UTC().Add(fireFor * -1),
		EndsAt:       time.Now().UTC().Add(fireFor),
	})
}

func AssertClusterOrder(clustersPage *pages.ClustersPage, clusterNames []string) {
	for i, v := range clusterNames {
		Eventually(clustersPage.ClustersList.Find(fmt.Sprintf("tr:nth-child(%d) td:nth-child(1)", i+1))).Should(MatchText(v))
	}
}

func createCluster(db *gorm.DB, name, status string) {
	db.Create(&models.Cluster{Name: name, Token: name})
	if status == "Ready" {
		db.Create(&models.ClusterInfo{
			UID:          types.UID(name),
			ClusterToken: name,
			UpdatedAt:    time.Now().UTC(),
		})
	} else if status == "Not Connected" {
		// do nothing
	} else if status == "Last seen" {
		// do nothing
		db.Create(&models.ClusterInfo{
			UID:          types.UID(name),
			ClusterToken: name,
			UpdatedAt:    time.Now().UTC().Add(time.Minute * -2),
		})
	} else if status == "Alerting" {
		db.Create(&models.ClusterInfo{
			UID:          types.UID(name),
			ClusterToken: name,
			UpdatedAt:    time.Now().UTC(),
		})
		createAlert(db, name, "ExampleAlert", "warning", "oh no", time.Second*30)
	} else if status == "Critical" {
		db.Create(&models.ClusterInfo{
			UID:          types.UID(name),
			ClusterToken: name,
			UpdatedAt:    time.Now().UTC(),
		})
		createAlert(db, name, "ExampleAlert", "critical", "oh no", time.Second*30)
	}
}

var _ = Describe("Integration suite", func() {

	var webDriver *agouti.Page

	BeforeEach(func() {
		var err error
		webDriver, err = agouti.NewPage(seleniumURL, agouti.Debug, agouti.Desired(agouti.Capabilities{
			"chromeOptions": map[string][]string{
				"args": {
					"--disable-gpu",
					"--no-sandbox",
				}}}))
		Expect(err).NotTo(HaveOccurred())
		Expect(webDriver.Navigate(uiURL + "/clusters")).To(Succeed())
	})

	AfterEach(func() {
		Expect(webDriver.Destroy()).To(Succeed())
	})

	Describe("Cluster page", func() {
		BeforeEach(func() {
			resetDb(db)
			db.Create(&models.Cluster{Name: "ewq"})
		})
		Describe("How the page should look", func() {
			It("should say Clusters in the title", func() {
				Eventually(webDriver).Should(HaveTitle("WKP · Clusters"))
			})

			It("should have a single cluster named ewq", func() {
				Eventually(webDriver.All("table tbody tr")).Should(HaveCount(1))
				Eventually(webDriver.First("table tbody tr td")).Should(HaveText("ewq"))
			})
		})
	})

	Describe("Sorting clusters!", func() {
		BeforeEach(func() {
			// Create some stuff in the db
			resetDb(db)
			createCluster(db, "cluster-1-ready", "Ready")
			createCluster(db, "cluster-2-critical", "Critical")
			createCluster(db, "cluster-3-alerting", "Alerting")
			createCluster(db, "cluster-4-not-connected", "Not Connected")
			createCluster(db, "cluster-5-last-seen", "Last seen")
		})

		Describe("How clicking on the headers should sort things", func() {
			It("Should have some items in the table", func() {
				clustersPage := pages.GetClustersPage(webDriver)
				Eventually(clustersPage.ClustersList.All("tr")).Should(HaveCount(5))
			})

			It("Should sort the cluster by status initially", func() {
				AssertClusterOrder(pages.GetClustersPage(webDriver), []string{
					"cluster-3-alerting",
					"cluster-2-critical",
					"cluster-5-last-seen",
					"cluster-4-not-connected",
					"cluster-1-ready",
				})
			})

			It("should reverse the order when I click on the status header", func() {
				pages.GetClustersPage(webDriver).HeaderStatus.Click()
				AssertClusterOrder(pages.GetClustersPage(webDriver), []string{
					"cluster-1-ready",
					"cluster-4-not-connected",
					"cluster-5-last-seen",
					"cluster-2-critical",
					"cluster-3-alerting",
				})
			})

			It("It should sort by name asc when you click on the name header", func() {
				pages.GetClustersPage(webDriver).HeaderName.Click()
				AssertClusterOrder(pages.GetClustersPage(webDriver), []string{
					"cluster-1-ready",
					"cluster-2-critical",
					"cluster-3-alerting",
					"cluster-4-not-connected",
					"cluster-5-last-seen",
				})
			})

			It("It should sort by name desc when you click on the name header again", func() {
				pages.GetClustersPage(webDriver).HeaderName.Click()
				AssertClusterOrder(pages.GetClustersPage(webDriver), []string{
					"cluster-1-ready",
					"cluster-2-critical",
					"cluster-3-alerting",
					"cluster-4-not-connected",
					"cluster-5-last-seen",
				})
				pages.GetClustersPage(webDriver).HeaderName.Click()
				AssertClusterOrder(pages.GetClustersPage(webDriver), []string{
					"cluster-5-last-seen",
					"cluster-4-not-connected",
					"cluster-3-alerting",
					"cluster-2-critical",
					"cluster-1-ready",
				})
			})
		})
	})
})

//
// Helpers
//

func getLocalPath(localPath string) string {
	testDir, _ := os.Getwd()
	path, _ := filepath.Abs(fmt.Sprintf("%s/../../../%s", testDir, localPath))
	return path
}

func ListenAndServe(ctx gcontext.Context, srv *http.Server) error {
	listenContext, listenCancel := gcontext.WithCancel(ctx)
	var listenError error
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			listenError = err
		}
		listenCancel()
	}()
	defer srv.Shutdown(gcontext.Background())

	<-listenContext.Done()

	return listenError
}

func RunBroker(ctx gcontext.Context, dbURI string) error {
	srv, err := broker.NewServer(ctx, broker.ParamSet{
		DbURI:       dbURI,
		DbType:      "sqlite",
		PrivKeyFile: dbURI,
	})
	if err != nil {
		return err
	}
	return ListenAndServe(ctx, srv)
}

func RunUIServer(ctx gcontext.Context, brokerURL string) {
	cmd := exec.CommandContext(ctx, "node", "server.js")
	cmd.Dir = getLocalPath("ui")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(
		os.Environ(),
		[]string{
			"NODE_ENV=production",
			"API_SERVER=" + brokerURL,
		}...,
	)
	err := cmd.Start()

	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()
	if err != nil {
		log.Println("waiting on cmd:", err)
	}
}

func GetDB(t *testing.T) (*gorm.DB, string) {
	f, err := ioutil.TempFile("", "mccpdb")
	log.Infof("db at %v", f.Name())
	dbURI := f.Name()
	require.NoError(t, err)
	db, err := utils.Open(dbURI, "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)
	return db, dbURI
}

func waitFor200(ctx gcontext.Context, url string, timeout time.Duration) error {
	log.Infof("Waiting for 200 from %v for %v", url, timeout)
	waitCtx, cancel := gcontext.WithTimeout(ctx, timeout)
	defer cancel()

	return wait.PollUntil(time.Second*1, func() (bool, error) {
		client := http.Client{
			Timeout: 5 * time.Second,
		}
		resp, err := client.Get(url)
		if err != nil {
			return false, nil
		}
		return resp.StatusCode == http.StatusOK, nil
	}, waitCtx.Done())
}

func gomegaFail(message string, callerSkip ...int) {
	fmt.Println("gomegaFail:")
	fmt.Println(message)
	webDriver := acceptancetest.GetWebDriver()
	if webDriver != nil {
		filepath := acceptancetest.TakeScreenShot(acceptancetest.String(16)) //Save the screenshot of failure
		fmt.Printf("\033[1;34mFailure screenshot is saved in file %s\033[0m \n", filepath)
	}
	// Pass this down to the default handler for onward processing
	Fail(message, callerSkip...)
}

//
// "main"
//

func TestMccpUI(t *testing.T) {
	db, dbURI = GetDB(t)

	var wg sync.WaitGroup
	ctx, cancel := gcontext.WithCancel(gcontext.Background())

	// Increment the WaitGroup synchronously in the main method, to avoid
	// racing with the goroutine starting.
	wg.Add(1)
	go func() {
		err := RunBroker(ctx, dbURI)
		assert.NoError(t, err)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		RunUIServer(ctx, brokerURL)
		wg.Done()
	}()

	// Test ui is proxying through to broker
	err := waitFor200(ctx, uiURL+"/gitops/api/clusters", time.Second*30)
	require.NoError(t, err)

	//
	// Test env stuff
	//
	RegisterFailHandler(Fail)
	// Screenshot on fail
	RegisterFailHandler(gomegaFail)
	// Screenshots
	_ = os.RemoveAll(acceptancetest.ARTEFACTS_BASE_DIR)
	_ = os.MkdirAll(acceptancetest.SCREENSHOTS_DIR, 0700)
	// WKP-UI can be a bit slow
	SetDefaultEventuallyTimeout(acceptancetest.ASSERTION_DEFAULT_TIME_OUT)

	// Load up the acceptance suite suite
	mccpRunner := acceptancetest.DatabaseMCCPTestRunner{DB: db}
	acceptancetest.DescribeMCCPAcceptance(mccpRunner)
	acceptancetest.SetSeleniumServiceUrl(seleniumURL)
	acceptancetest.SetWkpUrl(uiURL)

	AfterSuite(func() {
		webDriver := acceptancetest.GetWebDriver()
		//Tear down the suite level setup
		if webDriver != nil {
			Expect(webDriver.Destroy()).To(Succeed())
		}

		// Clean up ui-server and broker
		cancel()
		// Wait for the child goroutine to finish, which will only occur when
		// the child process has stopped and the call to cmd.Wait has returned.
		// This prevents main() exiting prematurely.
		wg.Wait()
	})

	// JUnit style test report
	junitReporter := reporters.NewJUnitReporter(acceptancetest.JUNIT_TEST_REPORT_FILE)
	// Run it!
	RunSpecsWithDefaultAndCustomReporters(t, "WKP Integration Suite", []Reporter{junitReporter})
}
