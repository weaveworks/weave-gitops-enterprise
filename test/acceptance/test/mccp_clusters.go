package acceptance

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
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
		KubeconfigPath: os.Getenv("MCCP_GCE_LEAF_KUBECONFIG"),
	},
	"eks": {
		Status:          "Critical alerts",
		IsWKP:           false,
		AlertManagerURL: "http://acmeprom-kube-prometheus-s-alertmanager.default:9093/api/v2",
		KubeconfigPath:  os.Getenv("MCCP_EKS_LEAF_KUBECONFIG"),
	},
	"kind-wkp": {
		Status:         "Ready",
		IsWKP:          true,
		KubeconfigPath: os.Getenv("MCCP_KIND_WKP_LEAF_KUBECONFIG"),
	},
}

func ClusterStatusFromList(clustersPage *pages.ClustersPage, clusterName string) *agouti.Selection {
	return pages.FindClusterInList(clustersPage, clusterName).Status
}

func createClusterEntry(webDriver *agouti.Page, clusterName string) (*pages.ClustersPage, *pages.ClusterConnectionPage) {

	//To check if page is loaded in its entirety
	clustersPage := pages.GetClustersPage(webDriver)
	clusterConnectionPage := pages.GetClusterConnectionPage(webDriver)

	var count string
	var expectedCount string

	By("And wait for the page to be fully loaded", func() {
		Eventually(clustersPage.SupportEmailLink).Should(BeVisible())
		Eventually(clustersPage.ClusterCount).Should(MatchText(`[0-9]+`))
		time.Sleep(UI_POLL_INTERVAL)
		count, _ = clustersPage.ClusterCount.Text()
		tmpCount, _ := strconv.Atoi(count)
		expectedCount = strconv.Itoa(tmpCount + 1)
	})

	By("When I click Connect a cluster button", func() {
		Eventually(clustersPage.ConnectClusterButton).Should(HaveText(expectedConnectClusterLabel))
		Expect(clustersPage.ConnectClusterButton.Click()).To(Succeed())
	})

	By("And I see the connection dialog", func() {
		Eventually(clusterConnectionPage.ClusterConnectionPopup).Should(BeFound())
	})

	By("And I enter the cluster name and ingress url", func() {
		clusterConnectionPage.ClusterName.SendKeys(clusterName)
		clusterConnectionPage.ClusterIngressURL.SendKeys("https://google.com")
	})

	By("And I click Save & next button", func() {
		Expect(clusterConnectionPage.ClusterSaveAndNext.Click()).To(Succeed())
	})

	By("And I see cluster is added to the list", func() {
		Eventually(clustersPage.ClusterCount, ASSERTION_1MINUTE_TIME_OUT).Should(HaveText(expectedCount))
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

func connectACluster(webDriver *agouti.Page, mccpTestRunner MCCPTestRunner, leaf LeafSpec) (*pages.ClustersPage, string, string) {

	tokenURLRegex := `https?:\/\/[-a-zA-Z0-9@:%._\+~#=]+\/gitops\/api\/agent\.yaml\?token=[0-9a-zA-Z]+`
	var tokenURL []string

	clusterName := RandString(32)
	fmt.Printf("Generated a new cluster name! %s\n", clusterName)
	clustersPage, clusterConnectionPage := createClusterEntry(webDriver, clusterName)
	commandEnv := getCommandEnv(leaf)

	By("And next page shows me kubectl command to apply on cluster to connect", func() {
		// Refresh the wkp-agent state
		err := mccpTestRunner.KubectlDeleteAllAgents(commandEnv)
		if err != nil {
			fmt.Println("Failed to delete the wkp-agent")
		}

		Eventually(clusterConnectionPage.ConnectionInstructions).Should(MatchText(`kubectl apply -f "` + tokenURLRegex + `"`))
		command, err := clusterConnectionPage.ConnectionInstructions.Text()
		if err == nil {
			log.Printf("Command :%s", command)
		}

		var rgx = regexp.MustCompile(`kubectl apply -f "(` + tokenURLRegex + `)"`)
		tokenURL = rgx.FindStringSubmatch(command)

		fmt.Printf("Connecting up %s with token %s\n", clusterName, tokenURL[1])
		fmt.Printf("Leaf is WKP cluster? %v\n", leaf.IsWKP)
		fmt.Printf("Leaf has alertmanager url? %v\n", leaf.AlertManagerURL)
		manifestURL := tokenURL[1]
		if leaf.AlertManagerURL != "" {
			manifestURL = fmt.Sprintf("%s&alertmanagerURL=%s", manifestURL, leaf.AlertManagerURL)
		}

		err = mccpTestRunner.KubectlApply(commandEnv, manifestURL)
		if err != nil {
			fmt.Printf(`Failed to install the wkp-agent by applying given command: %s`, command)
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

	// If leaf is not a WKP cluster skip the git activity check
	if leaf.IsWKP {
		By("And it should be showing the git activity", func() {
			Eventually(pages.FindClusterInList(clustersPage, clusterName).GitActivity.Find("svg"), ASSERTION_2MINUTE_TIME_OUT).
				Should(BeFound())
			Eventually(pages.FindClusterInList(clustersPage, clusterName).GitActivity.Find("circle")).
				Should(BeFound())
		})
	}

	return clustersPage, clusterName, tokenURL[1]
}

func Clear(field *agouti.Selection) {
	_ = field.SendKeys(strings.Repeat("\uE003", 100))
}

func ClearIngressURL(webDriver *agouti.Page, clusterName string) {
	clustersPage := pages.GetClustersPage(webDriver)
	Expect(pages.FindClusterInList(clustersPage, clusterName).EditCluster.Click()).To(Succeed())
	clusterConnectionPage := pages.GetClusterConnectionPage(webDriver)
	Eventually(clusterConnectionPage.ClusterConnectionPopup).Should(BeFound())

	Eventually(clusterConnectionPage.ClusterIngressURL).Should(BeFound())
	Clear(clusterConnectionPage.ClusterIngressURL)

	Eventually(clusterConnectionPage.ClusterSaveAndNext).Should(BeFound())
	Expect(clusterConnectionPage.ClusterSaveAndNext.Click()).To(Succeed())

	Eventually(pages.GetClusterConnectionPage(webDriver).ButtonClose).Should(BeFound())
	Expect(pages.GetClusterConnectionPage(webDriver).ButtonClose.Click()).To(Succeed())
}

func DescribeMCCPClusters(mccpTestRunner MCCPTestRunner) {

	var _ = Describe("Multi-Cluster Control Plane Clusters", func() {

		BeforeEach(func() {

			By("Given Kubernetes cluster is setup", func() {
				//TODO - Verify that cluster is up and running using kubectl
			})
			initializeWebdriver()
		})

		AfterEach(func() {
			TakeNextScreenshot()
			//Tear down
			//Expect(webDriver.Destroy()).To(Succeed())
		})

		It("Verify MCCP page structure first time with no cluster configured", func() {

			By("And MCCP state is reset", func() {
				mccpTestRunner.ResetDatabase()
				mccpTestRunner.VerifyMCCPPodsRunning()
				mccpTestRunner.VerifyMCCPPodsRunning()
			})

			clustersPage := pages.GetClustersPage(webDriver)
			By("Then I should see the correct clusters count next to the clusters header", func() {
				Eventually(clustersPage.ClusterCount).Should(MatchText(`[0-9]+`))
			})

			By("And should have 'Connect a cluster' button", func() {
				Eventually(clustersPage.ConnectClusterButton).Should(HaveText(expectedConnectClusterLabel))
			})

			By("And should have clusters list table", func() {
				Eventually(clustersPage.ClustersListSection).Should(BeFound())
			})

			By("And should have No clusters configured text", func() {
				Eventually(clustersPage.NoClusterConfigured).Should(HaveText("No clusters configured"))
			})

			By("And should have support email", func() {
				Expect(clustersPage.SupportEmailLink.Attribute("href")).To(HaveSuffix("mailto:support@weave.works"))
			})

			By("And should have No alerts firing message", func() {
				Expect(webDriver.Navigate(GetWkpUrl() + "/clusters/alerts")).To(Succeed())
				Eventually(clustersPage.NoFiringAlertMessage).Should(HaveText("No alerts firing"))
			})

		})

		It("Verify connect a cluster dialog ui layout", func() {
			clustersPage := pages.GetClustersPage(webDriver)

			By("When I click Connect a cluster button", func() {
				Eventually(clustersPage.ConnectClusterButton).Should(HaveText(expectedConnectClusterLabel))
				Expect(clustersPage.ConnectClusterButton.Click()).To(Succeed())
			})

			clusterConnectionPage := pages.GetClusterConnectionPage(webDriver)

			By("Then I should see the connection dialog", func() {
				Eventually(clusterConnectionPage.ClusterConnectionPopup).Should(BeFound())
			})

			By("And should have input name", func() {
				Eventually(clusterConnectionPage.ClusterName).Should(BeFound())
			})

			By("And should have input ingress URL", func() {
				Eventually(clusterConnectionPage.ClusterIngressURL).Should(BeFound())
			})

			By("And should have Save & next button", func() {
				Eventually(clusterConnectionPage.ClusterSaveAndNext).Should(HaveText("Save & next"))
			})
		})

		It("Verify connect a cluster input field validation", func() {
			clusterNameMax := RandString(300)
			fmt.Printf("Generated a new cluster name! %s\n", clusterNameMax)
			clustersPage, clusterConnectionPage := createClusterEntry(webDriver, clusterNameMax)

			By("And the cluster connection popup is closed", func() {
				Expect(clusterConnectionPage.ButtonClose.Click()).To(Succeed())
				Eventually(clusterConnectionPage.ClusterConnectionPopup).ShouldNot(BeFound())
			})

			By("And I should see the cluster name being shortened to the first 256 characters added", func() {
				cluster := pages.FindClusterInList(clustersPage, clusterNameMax[0:256])
				Eventually(cluster.Name, ASSERTION_1MINUTE_TIME_OUT).
					Should(HaveText(clusterNameMax[0:256]))
			})

			By("When I click Connect a cluster button", func() {
				Eventually(clustersPage.ConnectClusterButton).Should(HaveText(expectedConnectClusterLabel))
				Expect(clustersPage.ConnectClusterButton.Click()).To(Succeed())
			})

			By("Then I should see the connection dialog", func() {
				Eventually(clusterConnectionPage.ClusterConnectionPopup).Should(BeFound())
			})

			By("And I enter an empty cluster name", func() {
				Clear(clusterConnectionPage.ClusterName)
			})

			By("And I see Save & next button disabled", func() {
				Eventually(clusterConnectionPage.ButtonClusterSaveAndNext).ShouldNot(BeEnabled())
			})

			By("And I enter an all-spaces cluster name", func() {
				clusterConnectionPage.ClusterName.SendKeys("     ")
			})

			By("And I see Save & next button disabled", func() {
				Eventually(clusterConnectionPage.ButtonClusterSaveAndNext).ShouldNot(BeEnabled())
			})
		})

		It("Verify that clusters table have correct column headers ", func() {

			clustersPage := pages.GetClustersPage(webDriver)

			By("Then I should see clusters table with Name column", func() {
				Eventually(clustersPage.HeaderName).Should(HaveText("Name"))
			})

			By("And with Status column", func() {
				Eventually(clustersPage.HeaderStatus).Should(HaveText("Status"))
			})

			By("And with Latest git activity column", func() {
				Eventually(clustersPage.HeaderGitActivity).Should(HaveText("Latest git activity"))
			})

			By("And with Version (Nodes) column", func() {
				Eventually(clustersPage.HeaderNodeVersion).Should(HaveText("Version ( Nodes )"))
			})
		})

		It("Verify Not connected cluster status", func() {

			clusterName := RandString(32)
			fmt.Printf("Generated a new cluster name! %s\n", clusterName)
			clustersPage, clusterConnectionPage := createClusterEntry(webDriver, clusterName)

			By("And I close the connection dialog", func() {

				Eventually(clusterConnectionPage.ButtonClose).Should(BeFound())
				Expect(clusterConnectionPage.ButtonClose.Click()).To(Succeed())
				Eventually(clusterConnectionPage.ClusterConnectionPopup).ShouldNot(BeFound())
			})

			By("And I should see the cluster appears in the clusters list along with ready status", func() {
				Eventually(ClusterStatusFromList(clustersPage, clusterName), ASSERTION_1MINUTE_TIME_OUT).
					Should(HaveText("Not connected"))
			})

		})

		It("Verify last seen status", func() {

			clustersPage, clusterName, tokenURL := connectACluster(webDriver, mccpTestRunner, leaves["self"])

			By("And I disconnect the cluster", func() {
				mccpTestRunner.KubectlDelete([]string{}, tokenURL)
			})

			By("Then I should see the cluster status is changed to Last seen", func() {
				mccpTestRunner.TimeTravelToLastSeen()
				Eventually(ClusterStatusFromList(clustersPage, clusterName), ASSERTION_5MINUTE_TIME_OUT).
					Should(MatchText(`Last seen(\r\n|\r|\n)\d minutes ago`))
			})
		})

		It("Verify user can connect a cluster", func() {
			connectACluster(webDriver, mccpTestRunner, leaves["self"])
		})

		It("Verify alerts widget with firing alerts", func() {
			if getEnv("ACCEPTANCE_TESTS_DATABASE_TYPE", "") == "postgres" {
				Skip("This test case runs only with sqlite")
			}
			mccpTestRunner.ResetDatabase()
			mccpTestRunner.VerifyMCCPPodsRunning()
			mccpTestRunner.VerifyMCCPPodsRunning()
			clustersPage := pages.GetClustersPage(webDriver)
			Expect(webDriver.Navigate(GetWkpUrl() + "/clusters/alerts")).To(Succeed())
			Eventually(clustersPage.NoFiringAlertMessage).Should(BeFound())

			Expect(webDriver.Navigate(GetWkpUrl())).To(Succeed())
			Eventually(clustersPage.NoClusterConfigured).Should(HaveText("No clusters configured"))

			clustersPage, clusterName, _ := connectACluster(webDriver, mccpTestRunner, leaves["self"])

			alerts := [3]string{"AlertOne", "AlertTwo", "AlertThree"}
			messages := [3]string{"Critical Alert One", "Critical Alert Two", "Critical Alert Three"}
			severity := [3]string{"critical", "warning", "critical"}

			By("when I navigate to the alerts page..", func() {
				Expect(webDriver.Navigate(GetWkpUrl() + "/clusters/alerts")).To(Succeed())
			})

			By("And when a critical alert fires", func() {
				for i := 0; i < len(alerts); i++ {
					mccpTestRunner.FireAlert(alerts[i], severity[i], messages[i], time.Second*15)
				}

				Eventually(clustersPage.FiringAlerts, ASSERTION_1MINUTE_TIME_OUT).Should(BeFound())
			})

			By("Then alerts appear in the firing alerts widget with message, cluster name and timestamp", func() {

				for i := 0; i < len(alerts); i++ {
					alert := pages.FindAlertInFiringAlertsWidget(clustersPage, alerts[i])
					Expect(alert).ShouldNot(BeNil())
					Eventually(alert.Severity).Should(HaveText(severity[i]))
					Eventually(alert.ClusterName).Should(HaveText(clusterName))
					Eventually(alert.TimeStamp).Should(MatchText(`(a|\d) (minutes?|few seconds) ago`))
				}
			})

			By("And once alerts are resolved they disappear from the firing alerts widget", func() {
				Eventually(clustersPage.NoFiringAlertMessage, ASSERTION_1MINUTE_TIME_OUT).Should(BeFound())
			})
		})

		It("Verify that cluster status is changed to Alerting and then to Critical alerts ", func() {

			clustersPage, clusterName, _ := connectACluster(webDriver, mccpTestRunner, leaves["self"])

			By("And system raises a warning alert", func() {
				mccpTestRunner.FireAlert("ExampleAlert", "warning", "oh no", time.Second*30)
			})

			By("Then I should see the cluster status is changed to Alerting", func() {
				Eventually(ClusterStatusFromList(clustersPage, clusterName), ASSERTION_1MINUTE_TIME_OUT).
					Should(HaveText("Alerting"))
			})

			By("And when warning alert is resolved after 15s", func() {
				mccpTestRunner.TimeTravelToAlertsResolved()
				Eventually(ClusterStatusFromList(clustersPage, clusterName), ASSERTION_1MINUTE_TIME_OUT).
					Should(HaveText("Ready"))
			})

			By("And system raises a critical alert", func() {
				mccpTestRunner.FireAlert("ExampleAlert", "critical", "oh no", time.Second*30)
			})

			By("Then I should see the cluster status changes to Critical alerts", func() {
				Eventually(ClusterStatusFromList(clustersPage, clusterName), ASSERTION_1MINUTE_TIME_OUT).
					Should(HaveText("Critical alerts"))
			})

			By("And when alert is resolved then I should see the cluster status changes back to ready", func() {
				mccpTestRunner.TimeTravelToAlertsResolved()
				Eventually(ClusterStatusFromList(clustersPage, clusterName), ASSERTION_1MINUTE_TIME_OUT).
					Should(HaveText("Ready"))
			})
		})

		It("Verify that the ingress URL can be added and removed", func() {
			clusterName := RandString(32)
			fmt.Printf("Generated a new cluster name! %s\n", clusterName)
			clustersPage, clusterConnectionPage := createClusterEntry(webDriver, clusterName)

			By("And the cluster connection popup is closed", func() {
				Expect(clusterConnectionPage.ButtonClose.Click()).To(Succeed())
				Eventually(clusterConnectionPage.ClusterConnectionPopup).ShouldNot(BeFound())
			})

			By("Then the name of the cluster should be a link", func() {
				link := pages.FindClusterInList(clustersPage, clusterName).Name.Find("a")
				Expect(link).To(BeFound())
				url, _ := link.Attribute("href")
				Expect(url).To(Equal("https://google.com/"))
			})

			By("And when the ingress URL is cleared", func() {
				ClearIngressURL(webDriver, clusterName)
			})

			By("Then the name of the cluster should not be a link", func() {
				Eventually(pages.FindClusterInList(clustersPage, clusterName).Name.Find("a")).ShouldNot(BeFound())
			})
		})

		It("Verify clicking on alert name in alerts widget will take to the cluster page", func() {
			if getEnv("ACCEPTANCE_TESTS_DATABASE_TYPE", "") == "postgres" {
				Skip("This test case runs only with sqlite")
			}
			mccpTestRunner.ResetDatabase()
			clustersPage := pages.GetClustersPage(webDriver)

			Expect(webDriver.Navigate(GetWkpUrl() + "/clusters/alerts")).To(Succeed())
			Eventually(clustersPage.NoFiringAlertMessage).Should(BeFound())

			Expect(webDriver.Navigate(GetWkpUrl())).To(Succeed())
			Eventually(clustersPage.NoClusterConfigured).Should(HaveText("No clusters configured"))
			clustersPage, clusterName, _ := connectACluster(webDriver, mccpTestRunner, leaves["self"])

			Expect(webDriver.Navigate(GetWkpUrl() + "/clusters/alerts")).To(Succeed())

			alert := "MyAlert"
			message := "My Critical Alert"

			By("And when an alert fires", func() {

				mccpTestRunner.FireAlert(alert, "critical", message, time.Second*15)
				Eventually(clustersPage.FiringAlerts, ASSERTION_1MINUTE_TIME_OUT).Should(BeFound())

			})

			By("Then alerts appear in the firing alerts widget with hyper link cluster name ", func() {

				alert := pages.FindAlertInFiringAlertsWidget(clustersPage, alert)
				Eventually(alert.ClusterName).Should(HaveText(clusterName))
				Expect(webDriver).To(HaveWindowCount(1))
				Expect(alert.ClusterName.Click()).To(Succeed())
				Expect(webDriver).To(HaveWindowCount(2))
				Eventually(clustersPage.NoFiringAlertMessage, ASSERTION_1MINUTE_TIME_OUT).Should(BeFound())
			})
		})

		It("Verify disconnect cluster", func() {
			clusterName := RandString(32)
			fmt.Printf("Generated a new cluster name! %s\n", clusterName)
			clustersPage, clusterConnectionPage := createClusterEntry(webDriver, clusterName)
			confirmDisconnectClusterDialog := pages.GetConfirmDisconnectClusterDialog(webDriver)

			By("And the cluster connection popup is closed", func() {
				Expect(clusterConnectionPage.ButtonClose.Click()).To(Succeed())
				Eventually(clusterConnectionPage.ClusterConnectionPopup).ShouldNot(BeFound())
			})

			By("And when I click edit cluster I should see disconnect cluster tab", func() {
				Expect(pages.FindClusterInList(clustersPage, clusterName).EditCluster.Click()).To(Succeed())

				Eventually(clusterConnectionPage.ClusterConnectionPopup).Should(BeFound())
				Eventually(clusterConnectionPage.DisconnectTab).Should(BeFound())
			})

			By("And I open Disconnect tab and click Remove cluster from the MCCP button", func() {
				Expect(clusterConnectionPage.DisconnectTab.Click()).Should(Succeed())

				Eventually(clusterConnectionPage.ButtonRemoveCluster).Should(BeFound())
				Expect(clusterConnectionPage.ButtonRemoveCluster.Click()).To(Succeed())
			})

			By("Then I should see an alert popup with Remove button", func() {
				Eventually(confirmDisconnectClusterDialog.AlertPopup, ASSERTION_1MINUTE_TIME_OUT).Should(BeFound())
				Eventually(confirmDisconnectClusterDialog.ButtonRemove).Should(BeFound())
			})

			By("And I click remove button and the alert is closed", func() {
				Expect(confirmDisconnectClusterDialog.ButtonRemove.Click()).To(Succeed())
				Eventually(confirmDisconnectClusterDialog.AlertPopup).ShouldNot(BeFound())
			})

			By("Then I should see the cluster removed from the table", func() {
				Eventually(pages.FindClusterInList(clustersPage, clusterName).Name, ASSERTION_1MINUTE_TIME_OUT).
					ShouldNot(BeFound())
			})
		})

		It("Verify team workspaces variations", func() {
			if getEnv("CONNECT_KIND_WKP_LEAF_TEST", "") == "" {
				Skip("set CONNECT_KIND_WKP_LEAF_TEST env var to run this test")
			}

			clusterName := RandString(32)
			fmt.Printf("Generated a new cluster name! %s\n", clusterName)
			leaf := leaves["kind-wkp"]
			clustersPage, clusterName, _ := connectACluster(webDriver, mccpTestRunner, leaf)
			cluster := pages.FindClusterInList(clustersPage, clusterName)
			commandEnv := getCommandEnv(leaf)

			By("And I add a new workspace to the cluster", func() {
				// In acceptance test this has to be the host cluster.
				mccpTestRunner.AddWorkspace(commandEnv, clusterName)
			})

			By("Then I found the new workspace added to the Team Workspaces column", func() {
				Eventually(cluster.TeamWorkspaces, ASSERTION_1MINUTE_TIME_OUT).Should(MatchText("mccp-devs-workspace"))
			})

			By("And the workspaces should be a link", func() {
				link := cluster.TeamWorkspaces.Find("a")
				Expect(link).To(BeFound())
				url, _ := link.Attribute("href")
				Expect(url).To(Equal("https://google.com/workspaces"))
			})

			By("And when the ingress URL is cleared", func() {
				ClearIngressURL(webDriver, clusterName)
			})

			By("Then the team workspaces should not be a link", func() {
				Eventually(cluster.TeamWorkspaces.Find("a")).ShouldNot(BeFound())
			})
		})

		XIt("Verify user can connect a GCE cluster", func() {
			if getEnv("CONNECT_GCE_LEAF_TEST", "") == "" {
				Skip("set CONNECT_GCE_LEAF_TEST env var to run this test")
			}
			connectACluster(webDriver, mccpTestRunner, leaves["gce"])
		})

		It("Verify user can connect an EKS cluster", func() {
			if getEnv("CONNECT_EKS_LEAF_TEST", "") == "" {
				Skip("set CONNECT_EKS_LEAF_TEST env var to run this test")
			}
			connectACluster(webDriver, mccpTestRunner, leaves["eks"])
		})

		It("Verify user can connect a kind cluster with cluster components installed", func() {
			if getEnv("CONNECT_KIND_WKP_LEAF_TEST", "") == "" {
				Skip("set CONNECT_KIND_WKP_LEAF_TEST env var to run this test")
			}
			connectACluster(webDriver, mccpTestRunner, leaves["kind-wkp"])
		})
	})
}
