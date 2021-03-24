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
	"gorm.io/gorm"
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

var _ = Describe("Integration suite", func() {

	var page *agouti.Page

	BeforeEach(func() {
		db.Where("1 = 1").Delete(&models.Cluster{})
		c := models.Cluster{Name: "ewq"}
		db.Create(&c)
	})

	BeforeEach(func() {
		var err error
		page, err = agouti.NewPage(seleniumURL, agouti.Debug, agouti.Desired(agouti.Capabilities{
			"chromeOptions": map[string][]string{
				"args": {
					"--disable-gpu",
					"--no-sandbox",
				}}}))
		Expect(err).NotTo(HaveOccurred())
		Expect(page.Navigate(uiURL + "/clusters")).To(Succeed())
	})

	Describe("Cluster page", func() {
		AfterEach(func() {
			Expect(page.Destroy()).To(Succeed())
		})

		It("should say Clusters in the title", func() {
			Eventually(page).Should(HaveTitle("WKP · Clusters"))
		})

		It("should have a single cluster named ewq", func() {
			Eventually(page.All("table tbody tr")).Should(HaveCount(1))
			Eventually(page.First("table tbody tr td")).Should(HaveText("ewq"))
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
	defer func() {
		cancel()
		// Wait for the child goroutine to finish, which will only occur when
		// the child process has stopped and the call to cmd.Wait has returned.
		// This prevents main() exiting prematurely.
		wg.Wait()
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
	})

	// JUnit style test report
	junitReporter := reporters.NewJUnitReporter(acceptancetest.JUNIT_TEST_REPORT_FILE)
	// Run it!
	RunSpecsWithDefaultAndCustomReporters(t, "WKP Integration Suite", []Reporter{junitReporter})
}
