package acceptance

import (
	"fmt"
	"path"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func installViolatingDeployment(clusterName string, deploymentYaml string) {
	By(fmt.Sprintf("Install violating deployment to the %s cluster", clusterName), func() {
		Eventually(func(g Gomega) {
			_ = runCommandPassThrough("kubectl", "delete", "-f", deploymentYaml)
			g.Expect(runCommandPassThrough("kubectl", "apply", "-f", deploymentYaml)).ShouldNot(Succeed())
		}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(Succeed(), fmt.Sprintf("Test Postgres deployment should not be installed in the %s cluster", clusterName))
	})
}

func DescribeViolations(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Multi-Cluster Control Plane Violations", func() {

		Context("[UI] Violations can be seen in management cluster dashboard", func() {
			policiesYaml := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "policies.yaml")
			deploymentYaml := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "multi-container-manifest.yaml")

			policyName := "Containers Running With Privilege Escalation acceptance test"
			violationMsg := `Containers Running With Privilege Escalation acceptance test in deployment multi-container \(2 occurrences\)`
			voliationClusterName := "management"
			violationApplication := "default/multi-container"
			violationSeverity := "High"
			violationCategory := "weave.categories.pod-security"

			JustAfterEach(func() {
				_ = gitopsTestRunner.KubectlDelete([]string{}, policiesYaml)
				_ = gitopsTestRunner.KubectlDelete([]string{}, deploymentYaml)

			})

			It("Verify multiple occurrence violations can be monitored for violating resource", Label("integration", "violation"), func() {
				existingViolationCount := getViolationsCount()

				installTestPolicies("management", policiesYaml)
				installViolatingDeployment("management", deploymentYaml)

				pages.NavigateToPage(webDriver, "Violations")
				violationsPage := pages.GetViolationsPage(webDriver)

				By("And wait for violations to be visibe on the dashboard", func() {
					Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
					Eventually(violationsPage.ViolationHeader).Should(BeVisible())

					totalViolationCount := existingViolationCount + 2 // Container Running As Root + Containers Running With Privilege Escalation
					Eventually(func(g Gomega) string {
						g.Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
						time.Sleep(POLL_INTERVAL_1SECONDS)
						count, _ := violationsPage.ViolationCount.Text()
						return count

					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp(strconv.Itoa(totalViolationCount)), fmt.Sprintf("Dashboard failed to update with expected violations count: %d", totalViolationCount))

					Eventually(func(g Gomega) int {
						return violationsPage.CountViolations()
					}, ASSERTION_2MINUTE_TIME_OUT).Should(Equal(totalViolationCount), fmt.Sprintf("There should be %d policy enteries in policy table", totalViolationCount))

				})

				violationInfo := violationsPage.FindViolationInList(policyName)
				By(fmt.Sprintf("And verify '%s' violation Message", policyName), func() {
					Eventually(violationInfo.Message.Text).Should(MatchRegexp(violationMsg), fmt.Sprintf("Failed to list '%s' violation in vioilations table", violationMsg))
				})

				By(fmt.Sprintf("And verify '%s' violation Severity", policyName), func() {
					Eventually(violationInfo.Severity).Should(MatchText(violationSeverity), fmt.Sprintf("Failed to have expected vioilation Severity: %s", violationSeverity))
				})

				By(fmt.Sprintf("And verify '%s' violation cluster", policyName), func() {
					Eventually(violationInfo.Cluster).Should(MatchText(voliationClusterName), fmt.Sprintf("Failed to have expected violation cluster name: %s", voliationClusterName))
				})

				By(fmt.Sprintf("And verify '%s' violation application", policyName), func() {
					Eventually(violationInfo.Application).Should(MatchText(violationApplication), fmt.Sprintf("Failed to have expected violation Application: %s", violationApplication))
				})

				By(fmt.Sprintf("And navigate to '%s' Violation page", policyName), func() {
					Eventually(violationInfo.Message.Click).Should(Succeed(), fmt.Sprintf("Failed to navigate to %s violation detail page", violationMsg))
				})

				violationDetailPage := pages.GetViolationDetailPage(webDriver)
				By(fmt.Sprintf("And verify '%s' violation page", policyName), func() {
					Eventually(violationDetailPage.Header.Text).Should(MatchRegexp(policyName), "Failed to verify dashboard violation name ")
					Eventually(violationDetailPage.Title.Text).Should(MatchRegexp(policyName), "Failed to verify violation title on violation page")
					Eventually(violationDetailPage.Message.Text).Should(MatchRegexp(violationMsg), "Failed to verify violation Message on violation page")
					Eventually(violationDetailPage.ClusterName.Text).Should(MatchRegexp(voliationClusterName), "Failed to verify violation Cluster name on violation page")
					Eventually(violationDetailPage.Severity.Text).Should(MatchRegexp(violationSeverity), "Failed to verify violation Severity on violation page")
					Eventually(violationDetailPage.Category.Text).Should(MatchRegexp(violationCategory), "Failed to verify violation category on violation page")
					Eventually(violationDetailPage.Application.Text).Should(MatchRegexp(violationApplication), "Failed to verify violation application on violation page")
				})

				By(fmt.Sprintf("And verify '%s' violation Details", policyName), func() {
					occurenceCount := 2
					description := "Containers are running with PrivilegeEscalation configured."
					howToSolve := `spec:\s*containers:\s*securityContext:\s*allowPrivilegeEscalation: <value>`
					violatingEntity := `"name\\":\\"redis\\",\\"securityContext\\":{\\"allowPrivilegeEscalation\\":true}`

					Expect(violationDetailPage.OccurrencesCount.Text()).Should(MatchRegexp(strconv.Itoa(occurenceCount)), "Failed to verify violation occurrence count on violation page")
					Expect(violationDetailPage.Occurrences.Count()).Should(BeNumerically("==", occurenceCount), "Failed to verify number of violation occurrence enteries on violation page")
					for i := 0; i < occurenceCount; i++ {
						Expect(violationDetailPage.Occurrences.At(i).Text()).Should(MatchRegexp(fmt.Sprintf(`Container spec.template.spec.containers\[%d\] privilegeEscalation should be set to 'false'; detected 'true'`, i)), "Failed to verify number of violation occurrence enteries on violation page")
					}

					Expect(violationDetailPage.Description.Text()).Should(MatchRegexp(description), "Failed to verify violation Description on violation page")
					Expect(violationDetailPage.HowToSolve.Text()).Should(MatchRegexp(howToSolve), "Failed to verify violation 'How to solve' on violation page")
					Expect(violationDetailPage.ViolatingEntity.Text()).Should(MatchRegexp(violatingEntity), "Failed to verify 'Violating Entity' on violation page")
				})
			})
		})

		Context("[UI] Leaf cluster violations can be seen in management cluster", func() {
			var existingViolationCount int
			var mgmtClusterContext string
			var leafClusterContext string
			var leafClusterkubeconfig string
			var clusterBootstrapCopnfig string
			var gitopsCluster string
			patSecret := "violation-pat"
			bootstrapLabel := "bootstrap"
			leafClusterName := "wge-leaf-violation-kind"
			leafClusterNamespace := "default"

			policiesYaml := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "policies.yaml")
			deploymentYaml := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "postgres-manifest.yaml")
			policyName := "Container Image Pull Policy acceptance test"
			violationMsg := "Container Image Pull Policy acceptance test in deployment postgres"
			violationApplication := "default/postgres"
			violationSeverity := "Medium"
			violationCategory := "weave.categories.software-supply-chain"

			JustBeforeEach(func() {
				existingViolationCount = getViolationsCount()
				mgmtClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
				createCluster("kind", leafClusterName, "")
				leafClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
			})

			JustAfterEach(func() {
				useClusterContext(mgmtClusterContext)

				deleteSecret([]string{leafClusterkubeconfig, patSecret}, leafClusterNamespace)
				_ = gitopsTestRunner.KubectlDelete([]string{}, clusterBootstrapCopnfig)
				_ = gitopsTestRunner.KubectlDelete([]string{}, gitopsCluster)

				deleteClusters("kind", []string{leafClusterName}, "")
				_ = gitopsTestRunner.KubectlDelete([]string{}, policiesYaml)

			})

			It("Verify leaf cluster Violations can be monitored for violating resource via management cluster dashboard", Label("integration", "violation", "leaf-violation"), func() {
				leafClusterkubeconfig = createLeafClusterKubeconfig(leafClusterContext, leafClusterName, leafClusterNamespace)

				installPolicyAgent(leafClusterName)
				installTestPolicies(leafClusterName, policiesYaml)
				installViolatingDeployment(leafClusterName, deploymentYaml)

				useClusterContext(mgmtClusterContext)
				createPATSecret(leafClusterNamespace, patSecret)
				clusterBootstrapCopnfig = createClusterBootstrapConfig(leafClusterName, leafClusterNamespace, bootstrapLabel, patSecret)
				gitopsCluster = connectGitopsCuster(leafClusterName, leafClusterNamespace, bootstrapLabel, leafClusterkubeconfig)
				createLeafClusterSecret(leafClusterNamespace, leafClusterkubeconfig)

				By("Verify GitopsCluster status after creating kubeconfig secret", func() {
					pages.NavigateToPage(webDriver, "Clusters")
					clustersPage := pages.GetClustersPage(webDriver)
					pages.WaitForPageToLoad(webDriver)
					clusterInfo := clustersPage.FindClusterInList(leafClusterName)

					Eventually(clusterInfo.Status, ASSERTION_30SECONDS_TIME_OUT).Should(MatchText("Ready"))
				})

				By("And add kustomization bases for common resources for leaf cluster)", func() {
					addKustomizationBases(leafClusterName, leafClusterNamespace)
				})

				installTestPolicies("management", policiesYaml)
				installViolatingDeployment("management", deploymentYaml)

				pages.NavigateToPage(webDriver, "Violations")
				violationsPage := pages.GetViolationsPage(webDriver)

				By("And wait for violations to be visibe on the dashboard", func() {
					Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
					Eventually(violationsPage.ViolationHeader).Should(BeVisible())

					totalViolationCount := existingViolationCount + 1 + 1 // 1 management and 1 leaf violation
					Eventually(func(g Gomega) string {
						g.Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
						time.Sleep(POLL_INTERVAL_1SECONDS)
						count, _ := violationsPage.ViolationCount.Text()
						return count

					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp(strconv.Itoa(totalViolationCount)), fmt.Sprintf("Dashboard failed to update with expected violations count: %d", totalViolationCount))

					Eventually(func(g Gomega) int {
						return violationsPage.CountViolations()
					}, ASSERTION_2MINUTE_TIME_OUT).Should(Equal(totalViolationCount), fmt.Sprintf("There should be %d policy enteries in policy table", totalViolationCount))

				})

				By(fmt.Sprintf("And filter leaf cluster '%s' violations", leafClusterName), func() {
					filterID := "clusterName:" + leafClusterNamespace + `/` + leafClusterName
					searchPage := pages.GetSearchPage(webDriver)
					Eventually(searchPage.FilterBtn.Click).Should(Succeed(), "Failed to click filter buttton")
					searchPage.SelectFilter("cluster", filterID)

					Expect(searchPage.FilterBtn.Click()).Should(Succeed(), "Failed to click filter buttton")
				})

				violationInfo := violationsPage.FindViolationInList(policyName)
				By(fmt.Sprintf("And verify '%s' violation Message", policyName), func() {
					Eventually(violationInfo.Message.Text).Should(MatchRegexp(violationMsg), fmt.Sprintf("Failed to list '%s' violation in vioilations table", violationMsg))
				})

				By(fmt.Sprintf("And verify '%s' violation Severity", policyName), func() {
					Eventually(violationInfo.Severity).Should(MatchText(violationSeverity), fmt.Sprintf("Failed to have expected vioilation Severity: %s", violationSeverity))
				})

				By(fmt.Sprintf("And verify '%s' violation cluster", policyName), func() {
					Eventually(violationInfo.Cluster).Should(MatchText(leafClusterNamespace+`/`+leafClusterName), fmt.Sprintf("Failed to have expected violation cluster name: %s", leafClusterNamespace+`/`+leafClusterName))
				})

				By(fmt.Sprintf("And verify '%s' violation application", policyName), func() {
					Eventually(violationInfo.Application).Should(MatchText(violationApplication), fmt.Sprintf("Failed to have expected violation Application: %s", violationApplication))
				})

				By(fmt.Sprintf("And navigate to '%s' Violation page", policyName), func() {
					Eventually(violationInfo.Message.Click).Should(Succeed(), fmt.Sprintf("Failed to navigate to %s violation detail page", violationMsg))
				})

				violationDetailPage := pages.GetViolationDetailPage(webDriver)
				By(fmt.Sprintf("And verify '%s' violation page", policyName), func() {
					Eventually(violationDetailPage.Header.Text).Should(MatchRegexp(policyName), "Failed to verify dashboard violation name ")
					Eventually(violationDetailPage.Title.Text).Should(MatchRegexp(policyName), "Failed to verify violation title on violation page")
					Eventually(violationDetailPage.Message.Text).Should(MatchRegexp(violationMsg), "Failed to verify violation Message on violation page")
					Eventually(violationDetailPage.ClusterName.Text).Should(MatchRegexp(leafClusterNamespace+`/`+leafClusterName), "Failed to verify violation Cluster name on violation page")
					Eventually(violationDetailPage.Severity.Text).Should(MatchRegexp(violationSeverity), "Failed to verify violation Severity on violation page")
					Eventually(violationDetailPage.Category.Text).Should(MatchRegexp(violationCategory), "Failed to verify violation category on violation page")
					Eventually(violationDetailPage.Application.Text).Should(MatchRegexp(violationApplication), "Failed to verify violation application on violation page")
				})

				By(fmt.Sprintf("And verify '%s' violation Details", policyName), func() {
					description := "This Policy is to ensure you are setting a value for your imagePullPolicy."
					howToSolve := `spec:\s*containers:\s*- imagePullPolicy: <policy>`
					violatingEntity := `"name\\":\\"postgres\\",\\"namespace\\":\\"default\\"`
					Expect(violationDetailPage.Description.Text()).Should(MatchRegexp(description), "Failed to verify violation Description on violation page")
					Expect(violationDetailPage.HowToSolve.Text()).Should(MatchRegexp(howToSolve), "Failed to verify violation 'How to solve' on violation page")
					Expect(violationDetailPage.ViolatingEntity.Text()).Should(MatchRegexp(violatingEntity), "Failed to verify 'Violating Entity' on violation page")
				})
			})
		})
	})
}
