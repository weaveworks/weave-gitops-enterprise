package acceptance

import (
	"fmt"
	"path"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func DescribeViolations(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Multi-Cluster Control Plane Violations", func() {

		Context("[UI] Violations can be seen in management cluster dashboard", func() {
			policiesYaml := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "policies.yaml")
			deploymentYaml := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "postgres-manifest.yaml")

			policyName := "Container Image Pull Policy acceptance-test"
			policyID := "weave.policies.container-image-pull-policy-acceptance-test"
			violationMsg := "Container Image Pull Policy acceptance-test in deployment postgres"
			voliationClusterName := "management"
			violationApplication := "default/postgres"
			violationSeverity := "Medium"
			violationCategory := "weave.categories.software-supply-chain"

			JustAfterEach(func() {
				_ = gitopsTestRunner.KubectlDelete([]string{}, policiesYaml)
			})

			It("Verify Violations can be monitored for violating resource", Label("integration", "violation"), func() {
				existingViolationCount := getViolationsCount()

				installTestPolicies("management", policiesYaml)

				pages.NavigateToPage(webDriver, "Violations")
				violationsPage := pages.GetViolationsPage(webDriver)

				By("Install violating Postgres deployment to the management cluster", func() {
					Expect(waitForResource("policy", policyID, "default", "", ASSERTION_1MINUTE_TIME_OUT))
					Expect(gitopsTestRunner.KubectlApply([]string{}, deploymentYaml)).ShouldNot(Succeed(), "Failed to install test Postgres deployment to management cluster")
				})

				By("And wait for violations to be visibe on the dashboard", func() {
					Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
					Eventually(violationsPage.ViolationHeader).Should(BeVisible())

					totalViolationCount := existingViolationCount + 1
					Eventually(violationsPage.ViolationCount, ASSERTION_2MINUTE_TIME_OUT).Should(MatchText(strconv.Itoa(totalViolationCount)), fmt.Sprintf("Dashboard failed to update with expected violations count: %d", totalViolationCount))
					Eventually(func(g Gomega) int {
						return violationsPage.CountViolations()
					}, ASSERTION_2MINUTE_TIME_OUT).Should(Equal(totalViolationCount), fmt.Sprintf("There should be %d policy enteries in policy table", totalViolationCount))
				})

				violationInfo := violationsPage.FindViolationInList(policyName)
				By(fmt.Sprintf("And verify '%s' violation Message", policyName), func() {
					Eventually(violationInfo.Message.Text).Should(MatchRegexp(violationMsg), fmt.Sprintf("Failed to list %s violation in vioilations table", violationMsg))
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
