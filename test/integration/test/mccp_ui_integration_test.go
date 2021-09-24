package test

import (
	gcontext "context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/app"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/pkg/templates"
	broker "github.com/weaveworks/weave-gitops-enterprise/cmd/gitops-repo-broker/server"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/models"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
	acceptancetest "github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
	"github.com/weaveworks/weave-gitops/pkg/apputils/apputilsfakes"
	wego_server "github.com/weaveworks/weave-gitops/pkg/server"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

//
// Test suite
//

const capiServerPort = "8000"
const brokerPort = "8090"
const uiURL = "http://localhost:5000"
const seleniumURL = "http://localhost:4444/wd/hub"

var db *gorm.DB
var dbURI string

const entitlement = `eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJsaWNlbmNlZFVudGlsIjoxNzg5MzgxMDE1LCJpYXQiOjE2MzE2MTQ2MTUsImlzcyI6InNhbGVzQHdlYXZlLndvcmtzIiwibmJmIjoxNjMxNjE0NjE1LCJzdWIiOiJ0ZWFtLXBlc3RvQHdlYXZlLndvcmtzIn0.klRpQQgbCtshC3PuuD4DdI3i-7Z0uSGQot23YpsETphFq4i3KK4NmgfnDg_WA3Pik-C2cJgG8WWYkWnemWQJAw`

func resetDb(db *gorm.DB) {
	// https://gorm.io/docs/delete.html#Block-Global-Delete
	db.Where("1 = 1").Delete(&models.Cluster{})
	db.Where("1 = 1").Delete(&models.Alert{})
	db.Where("1 = 1").Delete(&models.ClusterInfo{})
}

func createAlert(db *gorm.DB, token, name, severity, message string, fireFor time.Duration, startsAt time.Time) {
	labels := fmt.Sprintf(`{ "alertname": "%s", "severity": "%s" }`, name, severity)
	annotations := fmt.Sprintf(`{ "message": "%s" }`, message)
	db.Create(&models.Alert{
		ClusterToken: token,
		Labels:       datatypes.JSON(labels),
		Annotations:  datatypes.JSON(annotations),
		Severity:     severity,
		StartsAt:     startsAt,
		EndsAt:       time.Now().UTC().Add(fireFor),
	})
}

func AssertClusterOrder(clustersPage *pages.ClustersPage, clusterNames []string) {
	getClusterNames := func() []string {
		names := []string{}
		elements := clustersPage.ClustersList.All("tr.summary")
		count, err := elements.Count()
		Expect(err).NotTo(HaveOccurred())
		for i := 0; i < count; i++ {
			el := elements.At(i).Find("td:nth-child(2)")
			name, err := el.Text()
			Expect(err).NotTo(HaveOccurred())
			names = append(names, name)
		}
		return names
	}

	Eventually(getClusterNames, acceptancetest.ASSERTION_10SECONDS_TIME_OUT).Should(Equal(clusterNames))
}

func AssertAlertsOrder(clustersPage *pages.ClustersPage, alertNames []string) {
	getAlertNames := func() []string {
		names := []string{}
		for _, a := range pages.AlertsFiringInAlertsWidget(clustersPage) {
			name, _ := a.Message.Text()
			names = append(names, name)
		}
		return names
	}

	Eventually(getAlertNames, acceptancetest.ASSERTION_10SECONDS_TIME_OUT).Should(Equal(alertNames))
	Consistently(getAlertNames, acceptancetest.ASSERTION_10SECONDS_TIME_OUT).Should(Equal(alertNames))
}

func createCluster(db *gorm.DB, name string, clusterType string, status string) {
	db.Create(&models.Cluster{Name: name, Token: name})
	if status == "Ready" {
		db.Create(&models.ClusterInfo{
			UID:          types.UID(name),
			ClusterToken: name,
			Type:         clusterType,
			UpdatedAt:    time.Now().UTC(),
		})
	} else if status == "Not Connected" {
		// do nothing
	} else if status == "Last seen" {
		// do nothing
		db.Create(&models.ClusterInfo{
			UID:          types.UID(name),
			ClusterToken: name,
			Type:         clusterType,
			UpdatedAt:    time.Now().UTC().Add(time.Minute * -2),
		})
	} else if status == "Alerting" {
		db.Create(&models.ClusterInfo{
			UID:          types.UID(name),
			ClusterToken: name,
			Type:         clusterType,
			UpdatedAt:    time.Now().UTC(),
		})
		createAlert(db, name, "ExampleAlert", "warning", "oh no", time.Second*30, time.Now().UTC())
	} else if status == "Critical" {
		db.Create(&models.ClusterInfo{
			UID:          types.UID(name),
			ClusterToken: name,
			Type:         clusterType,
			UpdatedAt:    time.Now().UTC(),
		})
		createAlert(db, name, "ExampleAlert", "critical", "oh no", time.Second*30, time.Now().UTC())
	}
}

func AssertTooltipContains(page *pages.ClustersPage, element *agouti.Selection, text string) {
	Eventually(element).Should(BeFound())
	Expect(element.MouseToElement()).Should(Succeed())
	Eventually(page.Tooltip).Should(BeFound())
	Eventually(page.Tooltip, acceptancetest.ASSERTION_1SECOND_TIME_OUT).Should(MatchText(text))
}

func AssertClusterIconIsDisplayed(icon *agouti.Selection, shouldBeFound bool) {
	if shouldBeFound {
		Eventually(icon, acceptancetest.ASSERTION_1MINUTE_TIME_OUT).Should(BeFound())
	} else {
		Eventually(icon, acceptancetest.ASSERTION_1MINUTE_TIME_OUT).ShouldNot(BeFound())
	}
}

func createNodeInfo(db *gorm.DB, clusterName, name, version string, isControlPlane bool) {
	var cluster models.Cluster
	var clusterInfo models.ClusterInfo
	db.Where("Name = ?", clusterName).First(&cluster)
	db.Where("cluster_token = ?", cluster.Token).First(&clusterInfo)

	db.Create(&models.NodeInfo{
		ClusterToken:   cluster.Token,
		ClusterInfoUID: types.UID(clusterName),
		Name:           name,
		IsControlPlane: isControlPlane,
		KubeletVersion: version,
	})
}

func AssertRowCellContains(element *agouti.Selection, text string) {
	Eventually(element).Should(BeFound())
	Eventually(element, acceptancetest.ASSERTION_1SECOND_TIME_OUT).Should(HaveText(text))
}

func createFluxInfo(db *gorm.DB, clusterName, name, namespace, repoURL, repoBranch string) {
	image := "docker.io/fluxcd/flux:v0.8.1"

	var cluster models.Cluster
	db.Where("Name = ?", clusterName).First(&cluster)

	db.Create(&models.FluxInfo{
		ClusterToken: cluster.Token,
		Name:         name,
		Namespace:    namespace,
		Args:         "",
		Image:        image,
		RepoURL:      repoURL,
		RepoBranch:   repoBranch,
	})
}

var intWebDriver *agouti.Page

var _ = Describe("Integration suite", func() {

	var page *pages.ClustersPage

	BeforeEach(func() {
		var err error
		if intWebDriver == nil {
			intWebDriver, err = agouti.NewPage(seleniumURL, agouti.Debug, agouti.Desired(agouti.Capabilities{
				"chromeOptions": map[string][]string{
					"args": {
						"--disable-gpu",
						"--no-sandbox",
					}}}))
			Expect(err).NotTo(HaveOccurred())
		}

		// reload fresh page each time
		Expect(intWebDriver.Navigate(uiURL)).To(Succeed())
		page = pages.GetClustersPage(intWebDriver)
		resetDb(db)
	})

	Describe("Tooltips!", func() {
		Describe("The column header tooltips", func() {
			It("should show a tooltip containing 'name' on mouse over", func() {
				AssertTooltipContains(page, page.HeaderName, "Name")
			})
			It("should show a tooltip containing 'version' on mouse over", func() {
				AssertTooltipContains(page, page.HeaderNodeVersion, "version")
			})
			It("should show a tooltip containing 'status' on mouse over", func() {
				AssertTooltipContains(page, page.HeaderStatus, "status")
			})
			It("should show a tooltip containing 'git' on mouse over", func() {
				AssertTooltipContains(page, page.HeaderGitActivity, "git")
			})
			It("should show a tooltip containing 'workspaces' on mouse over", func() {
				AssertTooltipContains(page, page.HeaderWorkspaces, "Workspaces")
			})
		})

		Describe("Cluster row tooltips", func() {
			var cluster *pages.ClusterInformation

			BeforeEach(func() {
				name := "ewq"
				createCluster(db, name, "", "Last seen")
				db.Create(&models.NodeInfo{
					ClusterInfoUID: types.UID(name),
					ClusterToken:   name,
					Name:           "cp-1",
					IsControlPlane: true,
					KubeletVersion: "v1.19",
				})
				db.Create(&models.Workspace{
					ClusterToken: name,
					Name:         "app-dev",
					Namespace:    "wkp-workspaces",
				})
				cluster = pages.FindClusterInList(page, name)
			})

			It("should show a tooltip containing with cp/version on mouse over", func() {
				AssertTooltipContains(page, cluster.NodesVersions, "1 Control plane nodes v1.19")
			})
			It("should show a tooltip containing app-dev on mouse over", func() {
				AssertTooltipContains(page, cluster.TeamWorkspaces, "app-dev")
			})
			It("should show a tooltip on status column cluster w/ last seen", func() {
				AssertTooltipContains(page, cluster.Status, "Last seen")
			})
		})
	})

	Describe("Cluster type icons!", func() {
		var kindCluster *pages.ClusterInformation
		var gkeCluster *pages.ClusterInformation
		var awsCluster *pages.ClusterInformation
		var eiCluster *pages.ClusterInformation
		var unknownCluster *pages.ClusterInformation

		BeforeEach(func() {
			createCluster(db, "kind", "kind", "Last seen")
			createCluster(db, "gke", "gke", "Last seen")
			createCluster(db, "aws", "aws", "Last seen")
			createCluster(db, "existingInfra", "existingInfra", "Last seen")
			createCluster(db, "unknown", "unknown", "Last seen")
		})

		It("should show the corresponding cluster type icon if exists", func() {
			kindCluster = pages.FindClusterInList(page, "kind")
			gkeCluster = pages.FindClusterInList(page, "gke")
			awsCluster = pages.FindClusterInList(page, "aws")
			eiCluster = pages.FindClusterInList(page, "existingInfra")
			unknownCluster = pages.FindClusterInList(page, "unknown")

			AssertClusterIconIsDisplayed(kindCluster.Icon, true)
			AssertClusterIconIsDisplayed(gkeCluster.Icon, true)
			AssertClusterIconIsDisplayed(awsCluster.Icon, true)
			AssertClusterIconIsDisplayed(eiCluster.Icon, true)
			AssertClusterIconIsDisplayed(unknownCluster.Icon, false)
		})
	})

	Describe("Sorting clusters!", func() {
		BeforeEach(func() {
			// Create some stuff in the db
			createCluster(db, "cluster-1-ready", "", "Ready")
			createCluster(db, "cluster-2-critical", "", "Critical")
			createCluster(db, "cluster-3-alerting", "", "Alerting")
			createCluster(db, "cluster-4-not-connected", "", "Not Connected")
			createCluster(db, "cluster-5-last-seen", "", "Last seen")
		})

		Describe("How clicking on the headers should sort things", func() {
			It("Should have some items in the table", func() {
				Eventually(page.ClustersList.All("tr.summary")).Should(HaveCount(5))
			})

			It("Should sort the cluster by status initially", func() {
				AssertClusterOrder(page, []string{
					"cluster-2-critical",
					"cluster-3-alerting",
					"cluster-5-last-seen",
					"cluster-1-ready",
					"cluster-4-not-connected",
				})
			})

			It("should reverse the order when I click on the status header", func() {
				Expect(page.HeaderStatus.Click()).Should(Succeed())
				AssertClusterOrder(page, []string{
					"cluster-4-not-connected",
					"cluster-1-ready",
					"cluster-5-last-seen",
					"cluster-3-alerting",
					"cluster-2-critical",
				})
			})

			It("It should sort by name asc when you click on the name header", func() {
				Expect(page.HeaderName.Click()).Should(Succeed())
				AssertClusterOrder(page, []string{
					"cluster-1-ready",
					"cluster-2-critical",
					"cluster-3-alerting",
					"cluster-4-not-connected",
					"cluster-5-last-seen",
				})
			})

			It("It should sort by name desc when you click on the name header again", func() {
				Expect(page.HeaderName.Click()).Should(Succeed())
				AssertClusterOrder(page, []string{
					"cluster-1-ready",
					"cluster-2-critical",
					"cluster-3-alerting",
					"cluster-4-not-connected",
					"cluster-5-last-seen",
				})
				Expect(page.HeaderName.Click()).Should(Succeed())
				AssertClusterOrder(page, []string{
					"cluster-5-last-seen",
					"cluster-4-not-connected",
					"cluster-3-alerting",
					"cluster-2-critical",
					"cluster-1-ready",
				})
			})
		})
	})

	Describe("Pagination", func() {
		BeforeEach(func() {
			for i := 1; i < 16; i++ {
				createCluster(db, "cluster"+strconv.Itoa(i), "", "Ready")
			}
		})

		Describe("How clicking the pagination controls should filter clusters", func() {
			It("Should have 10 clusters to begin with", func() {
				Eventually(page.ClustersList.All("tr.summary")).Should(HaveCount(10))
			})

			It("Should get the next 5 clusters when I click on the forward pagination control", func() {
				// wait for the next button to be on the page and click it
				Eventually(page.ClustersListPaginationNext).Should(BeFound())
				Expect(page.ClustersListPaginationNext.Click()).Should(Succeed())
				Eventually(page.ClustersList.All("tr.summary")).Should(HaveCount(5))
			})

			It("Should get the previous 10 clusters when I click on the previous pagination control", func() {
				// wait for the next button to be on the page and click it
				Eventually(page.ClustersListPaginationNext).Should(BeFound())
				Expect(page.ClustersListPaginationNext.Click()).Should(Succeed())
				// wait for the back button to be on the page and click it
				Eventually(page.ClustersListPaginationPrevious).Should(BeFound())
				Expect(page.ClustersListPaginationPrevious.Click()).Should(Succeed())
				Eventually(page.ClustersList.All("tr.summary")).Should(HaveCount(10))
			})

			It("Should go to the last page when I click on the last page control", func() {
				// wait for the last page button to be on the page and click it
				Eventually(page.ClustersListPaginationLast).Should(BeFound())
				Expect(page.ClustersListPaginationLast.Click()).Should(Succeed())
				Eventually(page.ClustersList.All("tr.summary")).Should(HaveCount(5))
			})

			It("Should go to the first page when I click on the first page control", func() {
				// wait for the last page button to be on the page and click it
				Eventually(page.ClustersListPaginationLast).Should(BeFound())
				Expect(page.ClustersListPaginationLast.Click()).Should(Succeed())
				// wait for the first page button to be on the page and click it
				Eventually(page.ClustersListPaginationFirst).Should(BeFound())
				Expect(page.ClustersListPaginationFirst.Click()).Should(Succeed())
				Eventually(page.ClustersList.All("tr.summary")).Should(HaveCount(10))
			})

			It("Should update the list of clusters on the page if the 20 clusters per page option is clicked", func() {
				Eventually(page.ClustersListPaginationPerPageDropdown).Should(BeFound())
				Expect(page.ClustersListPaginationPerPageDropdown.Click()).Should(Succeed())
				Expect(page.ClustersListPaginationPerPageDropdownSecond.Click()).Should(Succeed())
				Eventually(page.ClustersList.All("tr.summary")).Should(HaveCount(15))
			})
		})
	})

	Describe("Version(Nodes)", func() {
		BeforeEach(func() {
			// Similar control planes, different worker nodes
			createCluster(db, "cluster-1", "", "Last seen")
			createNodeInfo(db, "cluster-1", "cp-1", "v1.19.7", true)
			createNodeInfo(db, "cluster-1", "cp-2", "v1.19.7", true)
			createNodeInfo(db, "cluster-1", "worker-1", "v1.19.4", false)
			createNodeInfo(db, "cluster-1", "worker-2", "v1.19.4", false)

			// Different control planes and similar worker nodes
			createCluster(db, "cluster-2", "", "Last seen")
			createNodeInfo(db, "cluster-2", "cp-1", "v1.19.7", true)
			createNodeInfo(db, "cluster-2", "cp-2", "v1.19.4", true)
			createNodeInfo(db, "cluster-2", "worker-1", "v1.19.4", false)
			createNodeInfo(db, "cluster-2", "worker-2", "v1.19.4", false)

			// Similar control planes and worker nodes
			createCluster(db, "cluster-3", "", "Last seen")
			createNodeInfo(db, "cluster-3", "cp-1", "v1.19.7", true)
			createNodeInfo(db, "cluster-3", "worker-1", "v1.19.7", false)
			createNodeInfo(db, "cluster-3", "worker-2", "v1.19.7", false)

			// Similar worker nodes
			createCluster(db, "cluster-4", "", "Last seen")
			createNodeInfo(db, "cluster-4", "worker-1", "v1.19.7", false)
			createNodeInfo(db, "cluster-4", "worker-2", "v1.19.7", false)

			// Different worker nodes
			createCluster(db, "cluster-5", "", "Last seen")
			createNodeInfo(db, "cluster-5", "worker-1", "v1.19.7", false)
			createNodeInfo(db, "cluster-5", "worker-2", "v1.19.7", false)
			createNodeInfo(db, "cluster-5", "worker-3", "v1.19.4", false)
		})

		Describe("The column header", func() {
			It("should have Version ( Nodes ) text", func() {
				Eventually(page.HeaderNodeVersion).Should(HaveText("Version ( Nodes )"))
			})
		})

		Describe("The variations of versions", func() {
			It("should verify similar control planes and different worker nodes", func() {
				cluster := pages.FindClusterInList(page, "cluster-1")
				AssertRowCellContains(cluster.NodesVersions, "v1.19.7 ( 2CP )v1.19.4 ( 2 )")
			})

			It("should verify Different control planes and similar worker nodes", func() {
				cluster := pages.FindClusterInList(page, "cluster-2")
				AssertRowCellContains(cluster.NodesVersions, "v1.19.7 ( 1CP )v1.19.4 ( 1CP | 2 )")
			})

			It("should verify similar control planes and similar worker nodes", func() {
				cluster := pages.FindClusterInList(page, "cluster-3")
				AssertRowCellContains(cluster.NodesVersions, "v1.19.7 ( 1CP | 2 )")
			})

			It("should verify similar worker nodes", func() {
				cluster := pages.FindClusterInList(page, "cluster-4")
				AssertRowCellContains(cluster.NodesVersions, "v1.19.7 ( 2 )")
			})

			It("should verify different worker nodes", func() {
				cluster := pages.FindClusterInList(page, "cluster-5")
				AssertRowCellContains(cluster.NodesVersions, "v1.19.7 ( 2 )v1.19.4 ( 1 )")
			})
		})
	})

	Describe("View git repo", func() {

		BeforeEach(func() {
			// No flux instance installed
			createCluster(db, "no-flux-cluster", "", "Last seen")

			// One flux instance installed
			createCluster(db, "one-flux-cluster", "", "Last seen")
			createFluxInfo(db, "one-flux-cluster", "flux-1", "default", "git@github.com:weaveworks/fluxes-1.git", "master")

			// More than one flux instance installed
			createCluster(db, "two-flux-cluster", "", "Last seen")
			createFluxInfo(db, "two-flux-cluster", "flux-3", "wkp-flux", "git@github.com:weaveworks/fluxes-2.git", "main")
			createFluxInfo(db, "two-flux-cluster", "flux-4", "kube-system", "git@github.com:weaveworks/fluxes-3.git", "dev")
		})

		It("should show no button when no flux instance is installed", func() {
			cluster := pages.FindClusterInList(page, "no-flux-cluster")
			Eventually(cluster.GitRepoURL).Should(BeFound())
			Eventually(cluster.GitRepoURL, acceptancetest.ASSERTION_1SECOND_TIME_OUT).Should(HaveText("Repo not available"))
		})

		It("should show enabled button when one flux instance is installed", func() {
			cluster := pages.FindClusterInList(page, "one-flux-cluster")
			Eventually(cluster.GitRepoURL).Should(BeFound())
			Eventually(cluster.GitRepoURL, acceptancetest.ASSERTION_1SECOND_TIME_OUT).Should(BeEnabled())
			Eventually(cluster.GitRepoURL.Find("a"), acceptancetest.ASSERTION_1SECOND_TIME_OUT).Should(BeFound())
		})

		It("should show disabled button when more than one flux instance is installed", func() {
			cluster := pages.FindClusterInList(page, "two-flux-cluster")
			Eventually(cluster.GitRepoURL).Should(BeFound())
			Eventually(cluster.GitRepoURL, acceptancetest.ASSERTION_1SECOND_TIME_OUT).Should(HaveText("Repo not available"))
		})
	})

	Describe("The alerts widget!", func() {
		clusterName := "my-cluster"

		createRecentAlert := func(name, severity string, ago time.Duration) {
			createAlert(db, clusterName, name, severity, "", time.Second*30, time.Now().UTC().Add(ago*-1))
		}

		BeforeEach(func() {
			createCluster(db, clusterName, "", "Ready")
			createRecentAlert("alert1", "critical", time.Hour*2)
			createRecentAlert("alert2", "warning", time.Hour*1)
			createRecentAlert("alert3", "warning", time.Hour*3)
			Expect(intWebDriver.Navigate(uiURL + "/clusters/alerts")).To(Succeed())
		})

		It("should sort the alerts by starts at", func() {
			AssertAlertsOrder(page, []string{
				"alert2",
				"alert1",
				"alert3",
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
	defer func() {
		_ = srv.Shutdown(gcontext.Background())
	}()

	<-listenContext.Done()

	return listenError
}

func RunBroker(ctx gcontext.Context, dbURI string) error {
	srv, err := broker.NewServer(ctx, broker.ParamSet{
		Port:        brokerPort,
		DbURI:       dbURI,
		DbType:      "sqlite",
		PrivKeyFile: dbURI,
	})
	if err != nil {
		return err
	}
	return ListenAndServe(ctx, srv)
}

func RunCAPIServer(t *testing.T, ctx gcontext.Context, cl client.Client, discoveryClient discovery.DiscoveryInterface, db *gorm.DB) error {
	library := &templates.CRDLibrary{
		Log:       logr.Discard(),
		Client:    cl,
		Namespace: "default",
	}

	fakeAppsConfig := &wego_server.ApplicationsConfig{
		AppFactory: &apputilsfakes.FakeAppFactory{},
		KubeClient: cl,
	}

	return app.RunInProcessGateway(ctx, "0.0.0.0:"+capiServerPort, library, nil, cl, discoveryClient, db, "default", fakeAppsConfig, client.ObjectKey{Name: "entitlement", Namespace: "default"}, logr.Discard())
}

func RunUIServer(ctx gcontext.Context) {
	// is configured to proxy to
	// - 8000 for clusters-service
	// - 8090 for gitops-broker
	cmd := exec.CommandContext(ctx, "node", "server.js")
	cmd.Dir = getLocalPath("ui-cra")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(
		os.Environ(),
		[]string{
			"GITOPS_HOST=http://localhost:" + brokerPort,
			"CAPI_SERVER_HOST=http://localhost:" + capiServerPort,
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
	db, err := utils.OpenDebug(dbURI, false)
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
		filepath := acceptancetest.TakeScreenShot(acceptancetest.RandString(16)) //Save the screenshot of failure
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

	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		v1.AddToScheme,
		capiv1.AddToScheme,
		corev1.AddToScheme,
	}
	_ = schemeBuilder.AddToScheme(scheme)

	// Add entitlement secret
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "entitlement",
			Namespace: "default",
		},
		Type: "Opaque",
		Data: map[string][]byte{"entitlement": []byte(entitlement)},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(sec).
		Build()

	discoveryClient := discovery.NewDiscoveryClient(fakeclientset.NewSimpleClientset().Discovery().RESTClient())

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
		RunUIServer(ctx)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		_ = RunCAPIServer(t, ctx, cl, discoveryClient, db)
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
	SetDefaultEventuallyTimeout(acceptancetest.ASSERTION_5MINUTE_TIME_OUT)

	// Load up the acceptance suite suite
	mccpRunner := acceptancetest.DatabaseMCCPTestRunner{DB: db, Client: cl}

	acceptancetest.SetSeleniumServiceUrl(seleniumURL)
	acceptancetest.SetDefaultUIURL(uiURL)
	acceptancetest.DescribeSpecsMccpUi(mccpRunner)

	AfterSuite(func() {
		webDriver := acceptancetest.GetWebDriver()
		//Tear down the suite level setup
		if webDriver != nil {
			Expect(webDriver.Destroy()).To(Succeed())
		}

		if intWebDriver != nil {
			Expect(intWebDriver.Destroy()).To(Succeed())
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
