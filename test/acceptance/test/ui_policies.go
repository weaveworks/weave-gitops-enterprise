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

func installPolicyAgent(clusterName string) {
	By(fmt.Sprintf("And install cert-manager to %s cluster", clusterName), func() {
		stdOut, stdErr := runCommandAndReturnStringOutput("helm search repo charts-profile")
		if stdErr == "" && stdOut == "No results found" {
			err := runCommandPassThrough("helm", "repo", "add", "charts-profile", "https://s3.us-east-1.amazonaws.com/weaveworks-wkp/charts-profile/")
			Expect(err).ShouldNot(HaveOccurred(), "Failed to add profiles repositoy")
		}

		err := runCommandPassThrough("helm", "upgrade", "--install", "cert-manager", "charts-profile/cert-manager", "--namespace", "cert-manager", "--create-namespace", "--version", "0.0.7", "--set", "installCRDs=true")
		Expect(err).ShouldNot(HaveOccurred(), "Failed to install cer-manager to leaf cluster: "+clusterName)
	})

	By(fmt.Sprintf("And install policy agent to %s cluster", clusterName), func() {
		err := runCommandPassThrough("helm", "upgrade", "--install", "weave-policy-agent", "charts-profile/weave-policy-agent", "--namespace", "policy-system", "--create-namespace", "--version", "0.3.x", "--set", "accountId=weaveworks", "--set", "clusterId="+clusterName)
		Expect(err).ShouldNot(HaveOccurred(), "Failed to install policy agent to leaf cluster: "+clusterName)
	})
}

func installTestPolicies(clusterName string, policiesYaml string) {
	By(fmt.Sprintf("Add/Install test Policies to the %s cluster", clusterName), func() {
		err := runCommandPassThrough("kubectl", "apply", "-f", policiesYaml)
		Expect(err).ShouldNot(HaveOccurred(), "Failed to install test policies to cluster:"+clusterName)
	})
}

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
				existingPoliciesCount := getPoliciesCount()
				installTestPolicies("management", policiesYaml)

				pages.NavigateToPage(webDriver, "Policies")
				policiesPage := pages.GetPoliciesPage(webDriver)

				By("And wait for policies to be visibe on the dashboard", func() {
					Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
					Eventually(policiesPage.PolicyHeader).Should(BeVisible())

					totalPolicyCount := existingPoliciesCount + 3
					Eventually(func(g Gomega) string {
						g.Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
						time.Sleep(POLL_INTERVAL_1SECONDS)
						count, _ := policiesPage.PolicyCount.Text()
						return count

					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp(strconv.Itoa(totalPolicyCount)), fmt.Sprintf("Dashboard failed to update with expected policies count: %d", totalPolicyCount))

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

		Context("[UI] Policies can be installed on leaf cluster", func() {
			var mgmtClusterContext string
			var leafClusterContext string
			var leafClusterkubeconfig string
			leafClusterName := "wge-leaf-policy-kind"
			leafClusterNamespace := "default"

			policiesYaml := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "policies.yaml")
			policyName := "Container Running As Root acceptance test"
			policyID := "weave.policies.container-running-as-root-acceptance-test"
			policySeverity := "High"
			policyCategory := "weave.categories.pod-security"
			policyTags := []string{"pci-dss", "cis-benchmark", "mitre-attack", "nist800-190", "gdpr", "default"}
			policyTargetedKinds := []string{"Deployment", "Job", "ReplicationController", "ReplicaSet", "DaemonSet", "StatefulSet", "CronJob"}

			JustBeforeEach(func() {

				mgmtClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
				createCluster("kind", leafClusterName, "")
				leafClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
			})

			JustAfterEach(func() {
				useClusterContext(mgmtClusterContext)

				deleteKubeconfigSecret([]string{leafClusterkubeconfig}, leafClusterNamespace)
				deleteGitopsCluster([]string{leafClusterName}, leafClusterNamespace)

				deleteClusters("kind", []string{leafClusterName})

				_ = gitopsTestRunner.KubectlDelete([]string{}, policiesYaml)

			})

			It("Verify Policies can be installed on leaf cluster and monitored via management cluster dashboard", Label("integration", "policy", "leaf-policy"), func() {
				existingPoliciesCount := getPoliciesCount()
				leafClusterkubeconfig = createLeafClusterKubeconfig(leafClusterContext, leafClusterName, leafClusterNamespace)

				installPolicyAgent(leafClusterName)
				installTestPolicies(leafClusterName, policiesYaml)

				useClusterContext(mgmtClusterContext)
				connectGitopsCuster(leafClusterName, leafClusterNamespace, leafClusterkubeconfig)
				createLeafClusterSecret(leafClusterNamespace, leafClusterkubeconfig)

				By("Verify GitopsCluster status after creating kubeconfig secret", func() {
					pages.NavigateToPage(webDriver, "Clusters")
					clustersPage := pages.GetClustersPage(webDriver)
					pages.WaitForPageToLoad(webDriver)
					clusterInfo := clustersPage.FindClusterInList(leafClusterName)

					Eventually(clusterInfo.Status, ASSERTION_30SECONDS_TIME_OUT).Should(MatchText("Ready"))
				})

				By("And add kustomization bases for common resources for leaf cluster)", func() {
					addKustomizationBases(leafClusterName)
				})

				installTestPolicies("management", policiesYaml)
				pages.NavigateToPage(webDriver, "Policies")
				policiesPage := pages.GetPoliciesPage(webDriver)

				By("And wait for policies to be visibe on the dashboard", func() {
					pages.NavigateToPage(webDriver, "Policies")
					Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
					Eventually(policiesPage.PolicyHeader).Should(BeVisible())

					totalPolicyCount := existingPoliciesCount + 6 // 3 management and 3 leaf policies

					Eventually(func(g Gomega) string {
						g.Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
						time.Sleep(POLL_INTERVAL_1SECONDS)
						count, _ := policiesPage.PolicyCount.Text()
						return count

					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp(strconv.Itoa(totalPolicyCount)), fmt.Sprintf("Dashboard failed to update with expected policies count: %d", totalPolicyCount))

					Eventually(func(g Gomega) int {
						return policiesPage.CountPolicies()
					}, ASSERTION_2MINUTE_TIME_OUT).Should(Equal(totalPolicyCount), fmt.Sprintf("There should be %d policy enteries in policy table", totalPolicyCount))
				})

				By(fmt.Sprintf("And filter leaf cluster '%s' policies", leafClusterName), func() {
					filterID := "clusterName:" + leafClusterNamespace + `/` + leafClusterName
					searchPage := pages.GetSearchPage(webDriver)
					Eventually(searchPage.FilterBtn.Click).Should(Succeed(), "Failed to click filter buttton")
					searchPage.SelectFilter("cluster", filterID)

					Expect(searchPage.FilterBtn.Click()).Should(Succeed(), "Failed to click filter buttton")
				})

				policyInfo := policiesPage.FindPolicyInList(policyName)
				By(fmt.Sprintf("And verify '%s' policy Name", policyName), func() {
					Eventually(policyInfo.Name).Should(MatchText(policyName), fmt.Sprintf("Failed to list %s policy in  application table", policyName))
				})

				By(fmt.Sprintf("And verify '%s' policy Category", policyName), func() {
					Eventually(policyInfo.Category).Should(MatchText(policyCategory), fmt.Sprintf("Failed to have expected %s policy Category: weave.categories.pod-security", policyName))
				})

				By(fmt.Sprintf("And verify '%s' policy Severity", policyName), func() {
					Eventually(policyInfo.Severity).Should(MatchText(policySeverity), fmt.Sprintf("Failed to have expected %s Policy Severity: Medium", policyName))
				})

				By(fmt.Sprintf("And verify '%s' policy Cluster", policyName), func() {
					Eventually(policyInfo.Cluster).Should(MatchText(leafClusterName), fmt.Sprintf("Failed to have expected %[1]v policy Cluster: %[1]v", policyName))
				})

				By(fmt.Sprintf("And navigate to '%s' Policy page", policyName), func() {
					Eventually(policyInfo.Name.Click).Should(Succeed(), fmt.Sprintf("Failed to navigate to %s policy detail page", policyName))
				})

				policyDetailPage := pages.GetPolicyDetailPage(webDriver)
				By(fmt.Sprintf("And verify '%s' policy page", policyName), func() {
					Eventually(policyDetailPage.Header.Text).Should(MatchRegexp(policyName), "Failed to verify dashboard policy name ")
					Eventually(policyDetailPage.Title.Text).Should(MatchRegexp(policyName), "Failed to verify policy title on policy page")
					Eventually(policyDetailPage.ID.Text).Should(MatchRegexp(policyID), "Failed to verify policy ID on policy page")
					Eventually(policyDetailPage.ClusterName.Text).Should(MatchRegexp(leafClusterName), "Failed to verify policy cluster on policy page")
					Eventually(policyDetailPage.Severity.Text).Should(MatchRegexp(policySeverity), "Failed to verify policy Severity on policy page")
					Eventually(policyDetailPage.Category.Text).Should(MatchRegexp(policyCategory), "Failed to verify policy category on policy page")

					Expect(policyDetailPage.GetTags()).Should(ConsistOf(policyTags), "Failed to verify policy Tags on policy page")
					Expect(policyDetailPage.GetTargetedK8sKind()).Should(ConsistOf(policyTargetedKinds), "Failed to verify policy Targeted K8s Kind on policy page")
				})

				By(fmt.Sprintf("And verify '%s' policy Details", policyName), func() {
					description := "This Policy enforces that the securityContext.runAsNonRoot attribute is set to true."
					howToSolve := `spec:\s*securityContext:\s* runAsNonRoot: true`
					code := `result = {\s*20\s*"issue detected": true,\s*21\s*"msg": sprintf\("Container missing spec.template.spec.containers\[%v\].securityContext.runAsNonRoot while Pod.*\s*22\s*"violating_key": sprintf\("spec.template.spec.containers\[%v\].securityContext", \[i\]\)`

					Expect(policyDetailPage.Description.Text()).Should(MatchRegexp(description), "Failed to verify policy Description on policy page")
					Expect(policyDetailPage.HowToSolve.Text()).Should(MatchRegexp(howToSolve), "Failed to verify policy 'How to solve' on policy page")
					Expect(policyDetailPage.Code.Text()).Should(MatchRegexp(code), "Failed to verify 'Policy Code' on policy page")
				})

				By(fmt.Sprintf("And verify '%s' policy parameters", policyName), func() {
					parameter := policyDetailPage.GetParameter("exclude_namespace")
					namespaces := "test-systems"
					Expect(parameter.Name.Text()).Should(MatchRegexp(`exclude_namespace`), "Failed to verify parameter exclude_namespace 'Name'")
					Expect(parameter.Type.Text()).Should(MatchRegexp(`array`), "Failed to verify parameter exclude_namespace 'Type'")
					Expect(parameter.Value.Text()).Should(MatchRegexp(namespaces), "Failed to verify parameter exclude_namespace 'Value'")
					Expect(parameter.Required.Text()).Should(MatchRegexp(`False`), "Failed to verify parameter exclude_namespace 'Required'")

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
