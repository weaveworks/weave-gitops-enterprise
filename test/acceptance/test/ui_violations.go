package acceptance

import (
	"fmt"
	"path"
	"strconv"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func installViolatingDeployment(clusterName string, deploymentYaml string) {
	ginkgo.By(fmt.Sprintf("Install violating deployment to the %s cluster", clusterName), func() {
		gomega.Eventually(func(g gomega.Gomega) {
			_ = runCommandPassThrough("kubectl", "delete", "-f", deploymentYaml)
			g.Expect(runCommandPassThrough("kubectl", "apply", "-f", deploymentYaml)).ShouldNot(gomega.Succeed())
		}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Succeed(), fmt.Sprintf("Test Postgres deployment should not be installed in the %s cluster", clusterName))
	})
}

func DescribeViolations(gitopsTestRunner GitopsTestRunner) {
	var _ = ginkgo.Describe("Multi-Cluster Control Plane Violations", func() {

		ginkgo.BeforeEach(func() {
			gomega.Expect(webDriver.Navigate(test_ui_url)).To(gomega.Succeed())

			if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
				loginUser()
			}
		})

		ginkgo.Context("[UI] Violations can be seen in management cluster dashboard", func() {
			policiesYaml := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "policies.yaml")
			// Just specify policy config yaml path
			policyConfigYaml := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "policy-config.yaml")
			deploymentYaml := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "multi-container-manifest.yaml")

			policyName := "Containers Running With Privilege Escalation acceptance test"
			violationMsg := `Containers Running With Privilege Escalation acceptance test in deployment multi-container \(2 occurrences\)`
			voliationClusterName := "management"
			violationApplication := "default/multi-container"
			violationSeverity := "High"
			violationCategory := "weave.categories.pod-security"
			configPolicy := "Containers Minimum Replica Count acceptance test"
			policyConfigViolationMsg := `Containers Minimum Replica Count acceptance test in deployment podinfo (1 occurrences)`

			ginkgo.JustAfterEach(func() {
				// Delete the Policy config
				_ = gitopsTestRunner.KubectlDelete([]string{}, policyConfigYaml)

				_ = gitopsTestRunner.KubectlDelete([]string{}, policiesYaml)

				_ = gitopsTestRunner.KubectlDelete([]string{}, deploymentYaml)

			})

			ginkgo.It("Verify multiple occurrence violations can be monitored for violating resource", ginkgo.Label("integration", "violation"), func() {
				existingViolationCount := getViolationsCount()

				installTestPolicies("management", policiesYaml)
				// Add/Install Policy config to management cluster
				installPolicyConfig("management", policyConfigYaml)
				installViolatingDeployment("management", deploymentYaml)

				pages.NavigateToPage(webDriver, "Violations")
				violationsPage := pages.GetViolationsPage(webDriver)

				ginkgo.By("And wait for violations to be visibe on the dashboard", func() {
					gomega.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
					gomega.Eventually(violationsPage.ViolationHeader).Should(matchers.BeVisible())

					totalViolationCount := existingViolationCount + 3 // Container Running As Root + Containers Running With Privilege Escalation + Containers Minimum Replica Count
					gomega.Eventually(func(g gomega.Gomega) int {
						gomega.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
						time.Sleep(POLL_INTERVAL_1SECONDS)
						return violationsPage.CountViolations()
					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_3SECONDS).Should(gomega.Equal(totalViolationCount), fmt.Sprintf("There should be %d policy enteries in policy table, but found %d", totalViolationCount, existingViolationCount))

				})

				violationInfo := violationsPage.FindViolationInList(policyName)

				ginkgo.By(fmt.Sprintf("And verify '%s' violation Message", policyName), func() {
					gomega.Eventually(violationInfo.Message.Text).Should(gomega.MatchRegexp(violationMsg), fmt.Sprintf("Failed to list '%s' violation in vioilations table", violationMsg))
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' violation cluster", policyName), func() {
					gomega.Eventually(violationInfo.Cluster).Should(matchers.MatchText(voliationClusterName), fmt.Sprintf("Failed to have expected violation cluster name: %s", voliationClusterName))
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' violation application", policyName), func() {
					gomega.Eventually(violationInfo.Application).Should(matchers.MatchText(violationApplication), fmt.Sprintf("Failed to have expected violation Application: %s", violationApplication))
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' violation Severity", policyName), func() {
					gomega.Eventually(violationInfo.Severity).Should(matchers.MatchText(violationSeverity), fmt.Sprintf("Failed to have expected vioilation Severity: %s", violationSeverity))
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' violation Validated Policy", policyName), func() {
					gomega.Eventually(violationInfo.ValidatedPolicy).Should(matchers.MatchText(policyName), fmt.Sprintf("Failed to have expected vioilation Valodate Policy: %s", policyName))
				})

				ginkgo.By(fmt.Sprintf("And navigate to '%s' Violation page", policyName), func() {
					gomega.Eventually(violationInfo.Message.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s violation detail page", violationMsg))
				})

				violationDetailPage := pages.GetViolationDetailPage(webDriver)
				ginkgo.By(fmt.Sprintf("And verify '%s' violation page", policyName), func() {
					gomega.Eventually(violationDetailPage.Header.Text).Should(gomega.MatchRegexp(policyName), "Failed to verify dashboard violation name ")
					gomega.Eventually(violationDetailPage.ClusterName.Text).Should(gomega.MatchRegexp(voliationClusterName), "Failed to verify violation Cluster name on violation page")
					gomega.Eventually(violationDetailPage.Severity.Text).Should(gomega.MatchRegexp(violationSeverity), "Failed to verify violation Severity on violation page")
					gomega.Eventually(violationDetailPage.Category.Text).Should(gomega.MatchRegexp(violationCategory), "Failed to verify violation category on violation page")
					gomega.Eventually(violationDetailPage.Application.Text).Should(gomega.MatchRegexp(violationApplication), "Failed to verify violation application on violation page")
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' violation Details", policyName), func() {
					occurenceCount := 2
					description := "Containers are running with PrivilegeEscalation configured."
					howToSolve := `spec:\s*containers:\s*securityContext:\s*allowPrivilegeEscalation: <value>`
					violatingEntity := `"name\\":\\"redis\\",\\"securityContext\\":{\\"allowPrivilegeEscalation\\":true}`

					gomega.Expect(violationDetailPage.OccurrencesCount.Text()).Should(gomega.MatchRegexp(strconv.Itoa(occurenceCount)), "Failed to verify violation occurrence count on violation page")
					gomega.Expect(violationDetailPage.Occurrences.Count()).Should(gomega.BeNumerically("==", occurenceCount), "Failed to verify number of violation occurrence enteries on violation page")
					for i := 0; i < occurenceCount; i++ {
						gomega.Expect(violationDetailPage.Occurrences.At(i).Text()).Should(gomega.MatchRegexp(fmt.Sprintf(`Container spec.template.spec.containers\[%d\] privilegeEscalation should be set to 'false'; detected 'true'`, i)), "Failed to verify number of violation occurrence enteries on violation page")
					}

					gomega.Expect(violationDetailPage.Description.Text()).Should(gomega.MatchRegexp(description), "Failed to verify violation Description on violation page")
					gomega.Expect(violationDetailPage.HowToSolve.Text()).Should(gomega.MatchRegexp(howToSolve), "Failed to verify violation 'How to solve' on violation page")
					gomega.Expect(violationDetailPage.ViolatingEntity.Text()).Should(gomega.MatchRegexp(violatingEntity), "Failed to verify 'Violating Entity' on violation page")
				})

				verifyPolicyConfigInAppViolationsDetails(configPolicy, policyConfigViolationMsg)
			})
		})

		ginkgo.Context("[UI] Leaf cluster violations can be seen in management cluster", func() {
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
			// Just specify policy config yaml path
			policyConfigYaml := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "policy-config.yaml")
			deploymentYaml := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "postgres-manifest.yaml")
			policyName := "Container Image Pull Policy acceptance test"
			violationMsg := "Container Image Pull Policy acceptance test in deployment postgres"
			violationApplication := "default/postgres"
			violationSeverity := "Medium"
			violationCategory := "weave.categories.software-supply-chain"
			configPolicy := "Containers Minimum Replica Count acceptance test"
			policyConfigViolationMsg := `Containers Minimum Replica Count acceptance test in deployment podinfo (1 occurrences)`

			ginkgo.JustBeforeEach(func() {
				existingViolationCount = getViolationsCount()
				mgmtClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
				createCluster("kind", leafClusterName, "")
				leafClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
			})

			ginkgo.JustAfterEach(func() {
				useClusterContext(mgmtClusterContext)

				deleteSecret([]string{leafClusterkubeconfig, patSecret}, leafClusterNamespace)
				_ = gitopsTestRunner.KubectlDelete([]string{}, clusterBootstrapCopnfig)
				_ = gitopsTestRunner.KubectlDelete([]string{}, gitopsCluster)

				deleteCluster("kind", leafClusterName, "")
				// Delete the Policy config
				_ = gitopsTestRunner.KubectlDelete([]string{}, policyConfigYaml)
				_ = gitopsTestRunner.KubectlDelete([]string{}, policiesYaml)

			})

			ginkgo.FIt("Verify leaf cluster Violations can be monitored for violating resource via management cluster dashboard", ginkgo.Label("integration", "violation", "leaf-violation"), func() {
				leafClusterkubeconfig = createLeafClusterKubeconfig(leafClusterContext, leafClusterName, leafClusterNamespace)

				installPolicyAgent(leafClusterName)
				installTestPolicies(leafClusterName, policiesYaml)
				// Add/Install Policy config to leaf cluster
				installPolicyConfig(leafClusterName, policyConfigYaml)
				installViolatingDeployment(leafClusterName, deploymentYaml)

				useClusterContext(mgmtClusterContext)
				createPATSecret(leafClusterNamespace, patSecret)
				clusterBootstrapCopnfig = createClusterBootstrapConfig(leafClusterName, leafClusterNamespace, bootstrapLabel, patSecret)
				gitopsCluster = connectGitopsCluster(leafClusterName, leafClusterNamespace, bootstrapLabel, leafClusterkubeconfig)
				createLeafClusterSecret(leafClusterNamespace, leafClusterkubeconfig)

				waitForLeafClusterAvailability(leafClusterName, "Ready")
				addKustomizationBases("leaf", leafClusterName, leafClusterNamespace)

				installTestPolicies("management", policiesYaml)
				// Add/Install Policy config to management cluster
				installPolicyConfig(leafClusterName, policyConfigYaml)
				installViolatingDeployment("management", deploymentYaml)

				pages.NavigateToPage(webDriver, "Violations")
				violationsPage := pages.GetViolationsPage(webDriver)

				ginkgo.By("And wait for violations to be visibe on the dashboard", func() {
					gomega.Eventually(violationsPage.ViolationHeader).Should(matchers.BeVisible())

					totalViolationCount := existingViolationCount + 2 + 2 // 2 management and 2 leaf violation
					gomega.Eventually(func(g gomega.Gomega) int {
						gomega.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
						time.Sleep(POLL_INTERVAL_1SECONDS)
						return violationsPage.CountViolations()
					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_3SECONDS).Should(gomega.Equal(totalViolationCount), fmt.Sprintf("There should be %d policy enteries in policy table , but found %d", totalViolationCount, existingViolationCount))

				})

				ginkgo.By(fmt.Sprintf("And add filter leaf cluster '%s' violations", leafClusterName), func() {
					filterID := "clusterName: " + leafClusterNamespace + `/` + leafClusterName
					searchPage := pages.GetSearchPage(webDriver)
					searchPage.SelectFilter("cluster", filterID)
				})

				violationInfo := violationsPage.FindViolationInList(policyName)
				ginkgo.By(fmt.Sprintf("And verify '%s' violation Message", policyName), func() {
					gomega.Eventually(violationInfo.Message.Text).Should(gomega.MatchRegexp(violationMsg), fmt.Sprintf("Failed to list '%s' violation in vioilations table", violationMsg))
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' violation cluster", policyName), func() {
					gomega.Eventually(violationInfo.Cluster).Should(matchers.MatchText(leafClusterNamespace+`/`+leafClusterName), fmt.Sprintf("Failed to have expected violation cluster name: %s", leafClusterNamespace+`/`+leafClusterName))
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' violation application", policyName), func() {
					gomega.Eventually(violationInfo.Application).Should(matchers.MatchText(violationApplication), fmt.Sprintf("Failed to have expected violation Application: %s", violationApplication))
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' violation Severity", policyName), func() {
					gomega.Eventually(violationInfo.Severity).Should(matchers.MatchText(violationSeverity), fmt.Sprintf("Failed to have expected vioilation Severity: %s", violationSeverity))
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' violation Validated Policy", policyName), func() {
					gomega.Eventually(violationInfo.ValidatedPolicy).Should(matchers.MatchText(policyName), fmt.Sprintf("Failed to have expected vioilation Valodate Policy: %s", policyName))
				})

				ginkgo.By(fmt.Sprintf("And navigate to '%s' Violation page", policyName), func() {
					gomega.Eventually(violationInfo.Message.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s violation detail page", violationMsg))
				})

				violationDetailPage := pages.GetViolationDetailPage(webDriver)
				ginkgo.By(fmt.Sprintf("And verify '%s' violation page", policyName), func() {
					gomega.Eventually(violationDetailPage.Header.Text).Should(gomega.MatchRegexp(policyName), "Failed to verify dashboard violation name ")
					gomega.Eventually(violationDetailPage.ClusterName.Text).Should(gomega.MatchRegexp(leafClusterNamespace+`/`+leafClusterName), "Failed to verify violation Cluster name on violation page")
					gomega.Eventually(violationDetailPage.Severity.Text).Should(gomega.MatchRegexp(violationSeverity), "Failed to verify violation Severity on violation page")
					gomega.Eventually(violationDetailPage.Category.Text).Should(gomega.MatchRegexp(violationCategory), "Failed to verify violation category on violation page")
					gomega.Eventually(violationDetailPage.Application.Text).Should(gomega.MatchRegexp(violationApplication), "Failed to verify violation application on violation page")
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' violation Details", policyName), func() {
					description := "This Policy is to ensure you are setting a value for your imagePullPolicy."
					howToSolve := `spec:\s*containers:\s*- imagePullPolicy: <policy>`
					violatingEntity := `"name\\":\\"postgres\\",\\"namespace\\":\\"default\\"`
					gomega.Expect(violationDetailPage.Description.Text()).Should(gomega.MatchRegexp(description), "Failed to verify violation Description on violation page")
					gomega.Expect(violationDetailPage.HowToSolve.Text()).Should(gomega.MatchRegexp(howToSolve), "Failed to verify violation 'How to solve' on violation page")
					gomega.Expect(violationDetailPage.ViolatingEntity.Text()).Should(gomega.MatchRegexp(violatingEntity), "Failed to verify 'Violating Entity' on violation page")
				})
				verifyPolicyConfigInAppViolationsDetails(configPolicy, policyConfigViolationMsg)
			})
		})
	})
}
