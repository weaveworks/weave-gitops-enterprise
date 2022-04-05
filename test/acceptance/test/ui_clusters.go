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

func deleteClusterEntry(webDriver *agouti.Page, clusterNames []string) {
	for _, clusterName := range clusterNames {
		clustersPage := pages.GetClustersPage(webDriver)
		clusterConnectionPage := pages.GetClusterConnectionPage(webDriver)
		confirmDisconnectClusterDialog := pages.GetConfirmDisconnectClusterDialog(webDriver)

		logger.Tracef("Deleting cluster entry: %s", clusterName)
		Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())

		By("And wait for the page to be fully loaded", func() {
			Eventually(clustersPage.SupportEmailLink).Should(BeVisible())
			Eventually(clustersPage.ClusterCount).Should(MatchText(`[0-9]+`))
			Eventually(clustersPage.ClustersListSection).Should(BeFound())
			pages.ScrollWindow(webDriver, WINDOW_SIZE_X, WINDOW_SIZE_Y)
		})

		By("And when I click edit cluster I should see disconnect cluster tab", func() {

			if len(clusterName) > 256 {
				clusterName = clusterName[0:256]
			}
			clustersPage.WaitForClusterToAppear(webDriver, clusterName)
			Expect(pages.FindClusterInList(clustersPage, clusterName).EditCluster.Click()).To(Succeed())

			Eventually(clusterConnectionPage.ClusterConnectionPopup).Should(BeFound())
			Eventually(clusterConnectionPage.DisconnectTab).Should(BeFound())
		})

		By("And I open Disconnect tab and click Remove cluster from the wego button", func() {
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
	}
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

func DescribeClusters(gitopsTestRunner GitopsTestRunner) {

	var _ = Describe("Multi-Cluster Control Plane Clusters", func() {

		BeforeEach(func() {
			Expect(webDriver.Navigate(test_ui_url)).To(Succeed())
		})

		It("@integration Verify Weave Gitops Enterprise version", func() {
			By("And I verify the version", func() {
				clustersPage := pages.GetClustersPage(webDriver)
				Eventually(clustersPage.Version).Should(BeFound())
				Expect(clustersPage.Version.Text()).Should(MatchRegexp(enterpriseChartVersion()), "Expected Weave Gitops enterprise version is not found")
			})
		})

		It("Verify page structure first time with no cluster configured", func() {
			if GetEnv("ACCEPTANCE_TESTS_DATABASE_TYPE", "") == "postgres" {
				Skip("This test case runs only with sqlite")
			}

			By("And wego enterprise state is reset", func() {
				gitopsTestRunner.ResetControllers("enterprise")
				gitopsTestRunner.VerifyWegoPodsRunning()
				Eventually(webDriver.Refresh).ShouldNot(HaveOccurred())
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
				Expect(webDriver.Navigate(test_ui_url + "/clusters/alerts")).To(Succeed())
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

			By("And should have SAVE & NEXT button", func() {
				Eventually(clusterConnectionPage.ClusterSaveAndNext).Should(HaveText("SAVE & NEXT"))
			})
		})

		It("Verify connect a cluster input field validation", func() {
			clusterNameMax := RandString(300)
			logger.Tracef("Generated a new cluster name! %s", clusterNameMax)
			clustersPage, clusterConnectionPage := createClusterEntry(webDriver, clusterNameMax)

			By("And the cluster connection popup is closed", func() {
				Expect(clusterConnectionPage.ButtonClose.Click()).To(Succeed())
				Eventually(clusterConnectionPage.ClusterConnectionPopup).ShouldNot(BeFound())
			})

			By("And I should see the cluster name being shortened to the first 256 characters added", func() {
				cluster := pages.FindClusterInList(clustersPage, clusterNameMax[0:256])
				Eventually(cluster.Name, ASSERTION_1MINUTE_TIME_OUT).Should(HaveText(clusterNameMax[0:256]))
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

			By("And I see SAVE & NEXT button disabled", func() {
				Eventually(clusterConnectionPage.ButtonClusterSaveAndNext).ShouldNot(BeEnabled())
			})

			By("And I enter an all-spaces cluster name", func() {
				_ = clusterConnectionPage.ClusterName.SendKeys("     ")
			})

			By("And I see SAVE & NEXT button disabled", func() {
				Eventually(clusterConnectionPage.ButtonClusterSaveAndNext).ShouldNot(BeEnabled())
			})

			deleteClusterEntry(webDriver, []string{clusterNameMax})
		})

		It("Verify that clusters table have correct column headers ", func() {

			clustersPage := pages.GetClustersPage(webDriver)

			By("Then I should see clusters table with Name column", func() {
				Eventually(clustersPage.HeaderName).Should(HaveText("Name"))
			})

			By("And with Status column", func() {
				Eventually(clustersPage.HeaderStatus).Should(HaveText("Status"))
			})
		})

		It("Verify Not connected cluster status", func() {

			clusterName := RandString(32)
			logger.Tracef("Generated a new cluster name! %s", clusterName)
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

			clustersPage, clusterName, tokenURL := connectACluster(webDriver, gitopsTestRunner, leaves["self"])

			By("And I disconnect the cluster", func() {
				_ = gitopsTestRunner.KubectlDeleteInsecure([]string{}, tokenURL)
			})

			By("Then I should see the cluster status is changed to Last seen", func() {
				_ = gitopsTestRunner.TimeTravelToLastSeen()
				Eventually(ClusterStatusFromList(clustersPage, clusterName), ASSERTION_5MINUTE_TIME_OUT).
					Should(MatchText(`Last seen(\r\n|\r|\n)\d minutes ago`))
			})
		})

		It("Verify user can connect a cluster", func() {
			connectACluster(webDriver, gitopsTestRunner, leaves["self"])
		})

		It("Verify alerts widget with firing alerts", func() {
			if GetEnv("ACCEPTANCE_TESTS_DATABASE_TYPE", "") == "postgres" {
				Skip("This test case runs only with sqlite")
			}

			Skip("Alertmanager is not accessible via ingress anymore!")

			gitopsTestRunner.ResetControllers("enterprise")
			gitopsTestRunner.VerifyWegoPodsRunning()
			Eventually(webDriver.Refresh).ShouldNot(HaveOccurred())

			clustersPage := pages.GetClustersPage(webDriver)
			pages.NavigateToPage(webDriver, "Alerts")
			Eventually(clustersPage.NoFiringAlertMessage).Should(BeFound())

			Expect(webDriver.Navigate(test_ui_url)).To(Succeed())
			Eventually(clustersPage.NoClusterConfigured).Should(HaveText("No clusters configured"))

			clustersPage, clusterName, _ := connectACluster(webDriver, gitopsTestRunner, leaves["self"])

			alerts := [3]string{"AlertOne", "AlertTwo", "AlertThree"}
			messages := [3]string{"Critical Alert One", "Critical Alert Two", "Critical Alert Three"}
			severity := [3]string{"critical", "warning", "critical"}

			By("when I navigate to the alerts page..", func() {
				pages.NavigateToPage(webDriver, "Alerts")
			})

			By("And when a critical alert fires", func() {
				for i := 0; i < len(alerts); i++ {
					Expect(gitopsTestRunner.FireAlert(alerts[i], severity[i], messages[i], time.Second*15)).To(Succeed())
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

			Skip("Alertmanager is not accessible via ingress anymore!")

			clustersPage, clusterName, _ := connectACluster(webDriver, gitopsTestRunner, leaves["self"])

			By("And system raises a warning alert", func() {
				Expect(gitopsTestRunner.FireAlert("ExampleAlert", "warning", "oh no", time.Second*30)).To(Succeed())
			})

			By("Then I should see the cluster status is changed to Alerting", func() {
				Eventually(ClusterStatusFromList(clustersPage, clusterName), ASSERTION_1MINUTE_TIME_OUT).
					Should(HaveText("Alerting"))
			})

			By("And when warning alert is resolved after 15s", func() {
				_ = gitopsTestRunner.TimeTravelToAlertsResolved()
				Eventually(ClusterStatusFromList(clustersPage, clusterName), ASSERTION_1MINUTE_TIME_OUT).
					Should(HaveText("Ready"))
			})

			By("And system raises a critical alert", func() {
				Expect(gitopsTestRunner.FireAlert("ExampleAlert", "critical", "oh no", time.Second*30)).To(Succeed())
			})

			By("Then I should see the cluster status changes to Critical alerts", func() {
				Eventually(ClusterStatusFromList(clustersPage, clusterName), ASSERTION_1MINUTE_TIME_OUT).
					Should(HaveText("Critical alerts"))
			})

			By("And when alert is resolved then I should see the cluster status changes back to ready", func() {
				_ = gitopsTestRunner.TimeTravelToAlertsResolved()
				Eventually(ClusterStatusFromList(clustersPage, clusterName), ASSERTION_1MINUTE_TIME_OUT).
					Should(HaveText("Ready"))
			})
		})

		It("Verify that the ingress URL can be added and removed", func() {
			clusterName := RandString(32)
			logger.Tracef("Generated a new cluster name! %s", clusterName)
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
			if GetEnv("ACCEPTANCE_TESTS_DATABASE_TYPE", "") == "postgres" {
				Skip("This test case runs only with sqlite")
			}

			Skip("Alertmanager is not accessible via ingress anymore!")

			gitopsTestRunner.ResetControllers("enterprise")
			gitopsTestRunner.VerifyWegoPodsRunning()
			Eventually(webDriver.Refresh).ShouldNot(HaveOccurred())

			clustersPage := pages.GetClustersPage(webDriver)

			pages.NavigateToPage(webDriver, "Alerts")
			Eventually(clustersPage.NoFiringAlertMessage).Should(BeFound())

			Expect(webDriver.Navigate(test_ui_url)).To(Succeed())
			Eventually(clustersPage.NoClusterConfigured).Should(HaveText("No clusters configured"))
			clustersPage, clusterName, _ := connectACluster(webDriver, gitopsTestRunner, leaves["self"])

			pages.NavigateToPage(webDriver, "Alerts")

			alert := "MyAlert"
			message := "My Critical Alert"

			By("And when an alert fires", func() {
				Expect(gitopsTestRunner.FireAlert(alert, "critical", message, time.Second*15)).To(Succeed())
				Eventually(clustersPage.FiringAlerts, ASSERTION_1MINUTE_TIME_OUT).Should(BeFound())
			})

			By("Then alerts appear in the firing alerts widget with hyper link cluster name ", func() {
				alert := pages.FindAlertInFiringAlertsWidget(clustersPage, alert)
				Eventually(alert.ClusterName).Should(HaveText(clusterName))

				winCount, _ := webDriver.WindowCount()
				Expect(alert.ClusterName.Click()).To(Succeed())
				Expect(webDriver).To(HaveWindowCount(winCount + 1))
				Expect(webDriver.NextWindow()).ShouldNot(HaveOccurred(), "Failed to switch to cluster page window")
				Expect(webDriver.CloseWindow()).ShouldNot(HaveOccurred())
				Expect(webDriver.SwitchToWindow(WGE_WINDOW_NAME)).ShouldNot(HaveOccurred(), "Failed to switch to weave gitops enterprise window")

				Eventually(clustersPage.NoFiringAlertMessage, ASSERTION_1MINUTE_TIME_OUT).Should(BeFound())
			})
		})

		It("Verify disconnect cluster", func() {
			clusterName := RandString(32)
			logger.Tracef("Generated a new cluster name! %s", clusterName)
			_, clusterConnectionPage := createClusterEntry(webDriver, clusterName)

			By("And the cluster connection popup is closed", func() {
				Expect(clusterConnectionPage.ButtonClose.Click()).To(Succeed())
				Eventually(clusterConnectionPage.ClusterConnectionPopup).ShouldNot(BeFound())
			})

			deleteClusterEntry(webDriver, []string{clusterName})
		})

		It("@gce Verify user can connect a GCE cluster", func() {
			connectACluster(webDriver, gitopsTestRunner, leaves["gce"])
		})

		It("@eks Verify user can connect an EKS cluster", func() {
			connectACluster(webDriver, gitopsTestRunner, leaves["eks"])
		})
	})
}
