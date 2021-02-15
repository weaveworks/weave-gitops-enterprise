package test

import (
	gcontext "context"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	broker "github.com/weaveworks/wks/cmd/gitops-repo-broker/server"
	"github.com/weaveworks/wks/common/database/models"
	"github.com/weaveworks/wks/common/database/utils"
	"gorm.io/gorm"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/wait"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

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
	dbURI := f.Name()
	require.NoError(t, err)
	db, err := utils.Open(dbURI)
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)
	return db, dbURI
}

func AddCluster(t *testing.T, db *gorm.DB) {
	c := models.ClusterInfo{Name: "ewq"}
	result := db.Create(&c)
	require.NoError(t, result.Error)
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

func TestMccpUI(t *testing.T) {
	uiURL := "http://localhost:4046"
	brokerURL := "http://localhost:8000"
	// Where it runs locally / circle
	seleniumURL := "http://localhost:4444/wd/hub"

	var wg sync.WaitGroup
	ctx, cancel := gcontext.WithCancel(gcontext.Background())

	db, dbURI := GetDB(t)

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

	RegisterFailHandler(Fail)
	var page *agouti.Page

	AddCluster(t, db)

	Describe("Clusters page", func() {
		BeforeEach(func() {
			var err error
			page, err = agouti.NewPage(seleniumURL, agouti.Debug, agouti.Desired(agouti.Capabilities{
				"chromeOptions": map[string][]string{
					"args": {
						"--disable-gpu",
						"--no-sandbox",
					}}}))
			require.NoError(t, err)
			Expect(page.Navigate(uiURL + "/clusters")).To(Succeed())
		})

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

	RunSpecs(t, "Integration Suite")
}
