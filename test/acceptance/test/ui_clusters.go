package acceptance

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

var expectedConnectClusterLabel = "CONNECT A CLUSTER"

type LeafSpec struct {
	Status          string
	IsWKP           bool
	AlertManagerURL string
	KubeconfigPath  string
}

var leaves = map[string]LeafSpec{
	"self": {
		Status:          "Ready",
		IsWKP:           false,
		AlertManagerURL: "http://my-prom-kube-prometheus-st-alertmanager.prom:9093/api/v2",
		KubeconfigPath:  "",
	},
	"gce": {
		Status:         "Critical alerts",
		IsWKP:          true,
		KubeconfigPath: os.Getenv("GCE_LEAF_KUBECONFIG"),
	},
	"eks": {
		Status:          "Critical alerts",
		IsWKP:           false,
		AlertManagerURL: "http://acmeprom-kube-prometheus-s-alertmanager.default:9093/api/v2",
		KubeconfigPath:  os.Getenv("EKS_LEAF_KUBECONFIG"),
	},
}

func ClusterStatusFromList(clustersPage *pages.ClustersPage, clusterName string) *agouti.Selection {
	return pages.FindClusterInList(clustersPage, clusterName).Status
}

func createClusterEntry(webDriver *agouti.Page, clusterName string) (*pages.ClustersPage, *pages.ClusterConnectionPage) {
	pages.NavigateToPage(webDriver, "Cluster")
	clustersPage := pages.GetClustersPage(webDriver)
	clusterConnectionPage := pages.GetClusterConnectionPage(webDriver)

	var count string
	var expectedCount string

	By("And wait for the page to be fully loaded", func() {
		Eventually(clustersPage.SupportEmailLink).Should(BeVisible())
		Eventually(clustersPage.ClusterCount).Should(MatchText(`[0-9]+`))
		Eventually(clustersPage.ClustersListSection).Should(BeFound())
	})

	By("When I click Connect a cluster button", func() {
		Eventually(clustersPage.ConnectClusterButton).Should(HaveText(expectedConnectClusterLabel))
		Expect(clustersPage.ConnectClusterButton.Click()).To(Succeed())
	})

	By("And I see the connection dialog", func() {
		Eventually(clusterConnectionPage.ClusterConnectionPopup).Should(BeFound())
	})

	By("And I enter the cluster name and ingress url", func() {
		_ = clusterConnectionPage.ClusterName.SendKeys(clusterName)
		_ = clusterConnectionPage.ClusterIngressURL.SendKeys("https://google.com")
	})

	By("And cluster count before adding new cluster entery", func() {
		time.Sleep(POLL_INTERVAL_1SECONDS) // Sometimes UI took bit longer to update the cluster count
		count, _ = clustersPage.ClusterCount.Text()
		tmpCount, _ := strconv.Atoi(count)
		expectedCount = strconv.Itoa(tmpCount + 1)
	})

	By("And I click SAVE & NEXT button", func() {
		Expect(clusterConnectionPage.ClusterSaveAndNext.Click()).To(Succeed())
	})

	By("And I see cluster is added to the list", func() {
		Eventually(clustersPage.ClusterCount, ASSERTION_2MINUTE_TIME_OUT).Should(HaveText(expectedCount))
	})

	return clustersPage, clusterConnectionPage
}

func getCommandEnv(leaf LeafSpec) []string {
	commandEnv := []string{}
	if leaf.KubeconfigPath != "" {
		commandEnv = os.Environ()
		kubeconfigIndex := -1
		for i, ev := range commandEnv {
			if strings.HasPrefix(ev, "KUBECONFIG=") {
				kubeconfigIndex = i
			}
		}
		if kubeconfigIndex > -1 {
			// Remove the element at index i from a.
			commandEnv[kubeconfigIndex] = commandEnv[len(commandEnv)-1] // Copy last element to index i.
			commandEnv[len(commandEnv)-1] = ""                          // Erase last element (write zero value).
			commandEnv = commandEnv[:len(commandEnv)-1]                 // Truncate slice.
		}
		commandEnv = append(commandEnv, fmt.Sprintf("KUBECONFIG=%s", leaf.KubeconfigPath))
	}
	return commandEnv
}

func connectACluster(webDriver *agouti.Page, gitopsTestRunner GitopsTestRunner, leaf LeafSpec) (*pages.ClustersPage, string, string) {
	By("when I navigate to the cluster page..", func() {
		pages.NavigateToPage(webDriver, "Cluster")
	})

	tokenURLRegex := `https?:\/\/[-a-zA-Z0-9@:%._\+~#=]+\/gitops\/api\/agent\.yaml\?token=[0-9a-zA-Z]+`
	var tokenURL []string

	clusterName := RandString(32)
	logger.Tracef("Generated a new cluster name! %s", clusterName)
	clustersPage, clusterConnectionPage := createClusterEntry(webDriver, clusterName)
	commandEnv := getCommandEnv(leaf)

	By("And next page shows me kubectl command to apply on cluster to connect", func() {
		// Refresh the wkp-agent state
		err := gitopsTestRunner.KubectlDeleteAllAgents(commandEnv)
		if err != nil {
			logger.Tracef("Failed to delete the wkp-agent")
		}

		Eventually(clusterConnectionPage.ConnectionInstructions).Should(MatchText(`kubectl apply -f "` + tokenURLRegex + `"`))
		command, err := clusterConnectionPage.ConnectionInstructions.Text()
		if err == nil {
			logger.Tracef("Command :%s", command)
		}

		var rgx = regexp.MustCompile(`kubectl apply -f "(` + tokenURLRegex + `)"`)
		tokenURL = rgx.FindStringSubmatch(command)

		logger.Tracef("Connecting up %s with token %s", clusterName, tokenURL[1])
		logger.Tracef("Leaf is WKP cluster? %v", leaf.IsWKP)
		logger.Tracef("Leaf has alertmanager url? %v", leaf.AlertManagerURL)
		manifestURL := tokenURL[1]
		if leaf.AlertManagerURL != "" {
			manifestURL = fmt.Sprintf("%s&alertmanagerURL=%s", manifestURL, leaf.AlertManagerURL)
		}

		err = gitopsTestRunner.KubectlApplyInsecure(commandEnv, manifestURL)
		if err != nil {
			logger.Errorf(`Failed to install the wkp-agent by (tls insecurely) applying given url: %s, %s`, manifestURL, err)
		}
	})

	By("Then I should see the cluster status changes to Connected", func() {
		Eventually(clusterConnectionPage.ConnectionStatus, ASSERTION_6MINUTE_TIME_OUT).Should(MatchText(`Connected`))
		Expect(clusterConnectionPage.ButtonClose.Click()).To(Succeed())
		Eventually(clusterConnectionPage.ClusterConnectionPopup).ShouldNot(BeFound())
	})

	By("And I should see the cluster appears in the clusters list with the expected status", func() {
		Eventually(ClusterStatusFromList(clustersPage, clusterName), ASSERTION_1MINUTE_TIME_OUT).
			Should(HaveText(leaf.Status))
	})

	return clustersPage, clusterName, tokenURL[1]
}
