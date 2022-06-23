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

func DescribePolicies(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Multi-Cluster Control Plane Policies", func() {

		Context("[UI] Policies can be installed", func() {
			policiesYaml := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "policies.yaml")

			policyName := "Container Image Pull Policy acceptance-test"
			policyID := "weave.policies.container-image-pull-policy-acceptance-test"
			policyClusterName := "management"
			policySeverity := "Medium"
			policyCategory := "weave.categories.software-supply-chain"
			policyTags := []string{"There is no tags for this policy"}
			policyTargetedKinds := []string{"Deployment", "Job", "ReplicationController", "ReplicaSet", "DaemonSet", "StatefulSet", "CronJob"}

			JustAfterEach(func() {
				_ = gitopsTestRunner.KubectlDelete([]string{}, policiesYaml)
			})

			It("Verify Policies can be installed  and dashboard is updated accordingly", Label("integration", "policy"), func() {
				existingPoliciesCount := 1

				pages.NavigateToPage(webDriver, "Policies")
				policiesPage := pages.GetPoliciesPage(webDriver)
				By("And wait for Applications page to be fully rendered", func() {
					pages.WaitForPageToLoad(webDriver)
					existingPoliciesCount = policiesPage.CountPolicies()
				})

				By("Add/Install test Policies to the management cluster", func() {
					Expect(gitopsTestRunner.KubectlApply([]string{}, policiesYaml), "Failed to install test policies to management cluster")

				})

				By("And wait for policies to be visibe on the dashboard", func() {
					Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
					Eventually(policiesPage.PolicyHeader).Should(BeVisible())

					totalPolicyCount := existingPoliciesCount + 3
					Eventually(policiesPage.PolicyCount, ASSERTION_2MINUTE_TIME_OUT).Should(MatchText(strconv.Itoa(totalPolicyCount)), fmt.Sprintf("Dashboard failed to update with expected policies count: %d", totalPolicyCount))
					Eventually(func(g Gomega) int {
						return policiesPage.CountPolicies()
					}, ASSERTION_2MINUTE_TIME_OUT).Should(Equal(totalPolicyCount), fmt.Sprintf("There should be %d policy enteries in policy table", totalPolicyCount))
				})

				policyInfo := policiesPage.FindPolicyInList(policyName)
				By(fmt.Sprintf("And verify '%s' policy Name", policyName), func() {
					Eventually(policyInfo.Name).Should(MatchText(policyName), fmt.Sprintf("Failed to list %s policy in  application table", policyName))
				})

				By(fmt.Sprintf("And verify '%s' policy Category", policyName), func() {
					Eventually(policyInfo.Category).Should(MatchText(policyCategory), fmt.Sprintf("Failed to have expected %s policy Category: weave.categories.software-supply-chain", policyName))
				})

				By(fmt.Sprintf("And verify '%s' policy Severity", policyName), func() {
					Eventually(policyInfo.Severity).Should(MatchText(policySeverity), fmt.Sprintf("Failed to have expected %s Policy Severity: Medium", policyName))
				})

				By(fmt.Sprintf("And verify '%s' policy Cluster", policyName), func() {
					Eventually(policyInfo.Cluster).Should(MatchText(policyClusterName), fmt.Sprintf("Failed to have expected %[1]v policy Cluster: %[1]v", policyName))
				})

				By(fmt.Sprintf("And navigate to '%s' Policy page", policyName), func() {
					Eventually(policyInfo.Name.Click).Should(Succeed(), fmt.Sprintf("Failed to navigate to %s policy detail page", policyName))
				})

				policyDetailPage := pages.GetPolicyDetailPage(webDriver)
				By(fmt.Sprintf("And verify '%s' policy page", policyName), func() {
					Eventually(policyDetailPage.Header.Text).Should(MatchRegexp(policyName), "Failed to verify dashboard policy name ")
					Eventually(policyDetailPage.Title.Text).Should(MatchRegexp(policyName), "Failed to verify policy title on policy page")
					Eventually(policyDetailPage.ID.Text).Should(MatchRegexp(policyID), "Failed to verify policy ID on policy page")
					Eventually(policyDetailPage.ClusterName.Text).Should(MatchRegexp(policyClusterName), "Failed to verify policy cluster on policy page")
					Eventually(policyDetailPage.Severity.Text).Should(MatchRegexp(policySeverity), "Failed to verify policy Severity on policy page")
					Eventually(policyDetailPage.Category.Text).Should(MatchRegexp(policyCategory), "Failed to verify policy category on policy page")

					Expect(policyDetailPage.GetTags()).Should(ConsistOf(policyTags), "Failed to verify policy Tags on policy page")
					Expect(policyDetailPage.GetTargetedK8sKind()).Should(ConsistOf(policyTargetedKinds), "Failed to verify policy Targeted K8s Kind on policy page")
				})

				By(fmt.Sprintf("And verify '%s' policy Details", policyName), func() {
					description := "This Policy is to ensure you are setting a value for your imagePullPolicy."
					howToSolve := `spec:\s*containers:\s*- imagePullPolicy: <policy>`
					code := `result = {\s*15\s*"issue detected": true,\s*16\s*"msg": sprintf\("imagePolicyPolicy must be '%v'; found '%v'",\[policy, image_policy\]\),\s*17\s*"violating_key": sprintf\("spec.template.spec.containers\[%v\].imagePullPolicy", \[i]\),\s*18\s*"recommended_value": policy\s*19\s*}`

					Expect(policyDetailPage.Description.Text()).Should(MatchRegexp(description), "Failed to verify policy Description on policy page")
					Expect(policyDetailPage.HowToSolve.Text()).Should(MatchRegexp(howToSolve), "Failed to verify policy 'How to solve' on policy page")
					Expect(policyDetailPage.Code.Text()).Should(MatchRegexp(code), "Failed to verify 'Policy Code' on policy page")
				})

				By(fmt.Sprintf("And verify '%s' policy parameters", policyName), func() {
					parameter := policyDetailPage.GetParameter("policy")
					Expect(parameter.Name.Text()).Should(MatchRegexp(`policy`), "Failed to verify parameter policy 'Name'")
					Expect(parameter.Type.Text()).Should(MatchRegexp(`string`), "Failed to verify parameter policy 'Type'")
					Expect(parameter.Value.Text()).Should(MatchRegexp(`Always`), "Failed to verify parameter policy 'Value'")
					Expect(parameter.Required.Text()).Should(MatchRegexp(`True`), "Failed to verify parameter policy 'Required'")

					parameter = policyDetailPage.GetParameter("exclude_namespace")
					namespaces := "test-systems"
					Expect(parameter.Name.Text()).Should(MatchRegexp(`exclude_namespace`), "Failed to verify parameter exclude_namespace 'Name'")
					Expect(parameter.Type.Text()).Should(MatchRegexp(`array`), "Failed to verify parameter exclude_namespace 'Type'")
					Expect(parameter.Value.Text()).Should(MatchRegexp(namespaces), "Failed to verify parameter exclude_namespace 'Value'")
					Expect(parameter.Required.Text()).Should(MatchRegexp(`True`), "Failed to verify parameter exclude_namespace 'Required'")

					parameter = policyDetailPage.GetParameter("exclude_label_key")
					Expect(parameter.Name.Text()).Should(MatchRegexp(`exclude_label_key`), "Failed to verify parameter exclude_label_key 'Name'")
					Expect(parameter.Type.Text()).Should(MatchRegexp(`string`), "Failed to verify parameter exclude_label_key 'Type'")
					Expect(parameter.Value.Text()).Should(MatchRegexp(`undefined`), "Failed to verify parameter exclude_label_key 'Value'")
					Expect(parameter.Required.Text()).Should(MatchRegexp(`False`), "Failed to verify parameter exclude_label_key 'Required'")

					parameter = policyDetailPage.GetParameter("exclude_label_value")
					Expect(parameter.Name.Text()).Should(MatchRegexp(`exclude_label_value`), "Failed to verify parameter exclude_label_value 'Name'")
					Expect(parameter.Type.Text()).Should(MatchRegexp(`string`), "Failed to verify parameter exclude_label_value 'Type'")
					Expect(parameter.Value.Text()).Should(MatchRegexp(`undefined`), "Failed to verify parameter exclude_label_value 'Value'")
					Expect(parameter.Required.Text()).Should(MatchRegexp(`False`), "Failed to verify parameter exclude_label_value 'Required'")
				})

				By("And again navigate to Polisies page via header link", func() {
					Expect(policiesPage.PolicyHeader.Click()).Should(Succeed(), "Failed to navigate to Policies pages via header link")
					pages.WaitForPageToLoad(webDriver)
				})
			})
		})
	})
}
