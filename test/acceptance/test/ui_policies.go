package acceptance

import (
	"fmt"
	"path"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func installPolicyAgent(clusterName string) {
	ginkgo.By(fmt.Sprintf("And install cert-manager to %s cluster", clusterName), func() {
		stdOut, stdErr := runCommandAndReturnStringOutput("helm search repo profiles-catalog")
		if stdErr == "" && stdOut == "No results found" {
			err := runCommandPassThrough("helm", "repo", "add", "profiles-catalog", "https://raw.githubusercontent.com/weaveworks/weave-gitops-profile-examples/gh-pages")
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to add profiles repositoy")
		}

		err := runCommandPassThrough("helm", "upgrade", "--install", "cert-manager", "profiles-catalog/cert-manager", "--namespace", "cert-manager", "--create-namespace", "--version", "0.0.8", "--set", "installCRDs=true")
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to install cer-manager to leaf cluster: "+clusterName)
	})

	ginkgo.By(fmt.Sprintf("And install policy agent to %s cluster", clusterName), func() {
		err := runCommandPassThrough("helm", "upgrade", "--install", "weave-policy-agent", "profiles-catalog/weave-policy-agent", "--namespace", "policy-system", "--create-namespace", "--version", "0.5.x", "--set", "policy-agent.accountId=weaveworks", "--set", "policy-agent.clusterId="+clusterName)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to install policy agent to leaf cluster: "+clusterName)
	})
}

func installTestPolicies(clusterName string, policiesYaml string) {
	ginkgo.By(fmt.Sprintf("Add/Install test Policies to the %s cluster", clusterName), func() {
		err := runCommandPassThrough("kubectl", "apply", "-f", policiesYaml)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to install test policies to cluster:"+clusterName)
	})
}

func DescribePolicies(gitopsTestRunner GitopsTestRunner) {
	var _ = ginkgo.Describe("Multi-Cluster Control Plane Policies", func() {

		ginkgo.BeforeEach(func() {
			gomega.Expect(webDriver.Navigate(test_ui_url)).To(gomega.Succeed())

			if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
				loginUser()
			}
		})

		ginkgo.Context("[UI] Policies can be installed", func() {
			policiesYaml := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "policies.yaml")

			policyName := "Container Image Pull Policy acceptance test"
			policyID := "weave.policies.container-image-pull-policy-acceptance-test"
			policyClusterName := "management"
			policySeverity := "Medium"
			policyCategory := "weave.categories.software-supply-chain"
			policyTags := []string{"There is no tags for this policy"}
			policyTargetedKinds := []string{"Deployment", "Job", "ReplicationController", "ReplicaSet", "DaemonSet", "StatefulSet", "CronJob"}

			ginkgo.JustAfterEach(func() {
				_ = gitopsTestRunner.KubectlDelete([]string{}, policiesYaml)
			})

			ginkgo.It("Verify Policies can be installed  and dashboard is updated accordingly", ginkgo.Label("integration", "policy"), func() {
				existingPoliciesCount := getPoliciesCount()
				installTestPolicies("management", policiesYaml)

				pages.NavigateToPage(webDriver, "Policies")
				policiesPage := pages.GetPoliciesPage(webDriver)

				ginkgo.By("And wait for policies to be visibe on the dashboard", func() {
					gomega.Eventually(policiesPage.PolicyHeader).Should(matchers.BeVisible())

					totalPolicyCount := existingPoliciesCount + 4
					gomega.Eventually(func(g gomega.Gomega) int {
						gomega.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
						time.Sleep(POLL_INTERVAL_1SECONDS)
						return policiesPage.CountPolicies()
					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_3SECONDS).Should(gomega.Equal(totalPolicyCount), fmt.Sprintf("There should be %d policy enteries in policy table", totalPolicyCount))

				})

				policyInfo := policiesPage.FindPolicyInList(policyName)
				ginkgo.By(fmt.Sprintf("And verify '%s' policy Name", policyName), func() {
					gomega.Eventually(policyInfo.Name).Should(matchers.MatchText(policyName), fmt.Sprintf("Failed to list %s policy in  application table", policyName))
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' policy Category", policyName), func() {
					gomega.Eventually(policyInfo.Category).Should(matchers.MatchText(policyCategory), fmt.Sprintf("Failed to have expected %s policy Category: weave.categories.software-supply-chain", policyName))
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' policy Severity", policyName), func() {
					gomega.Eventually(policyInfo.Severity).Should(matchers.MatchText(policySeverity), fmt.Sprintf("Failed to have expected %s Policy Severity: Medium", policyName))
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' policy Cluster", policyName), func() {
					gomega.Eventually(policyInfo.Cluster).Should(matchers.MatchText(policyClusterName), fmt.Sprintf("Failed to have expected %[1]v policy Cluster: %[1]v", policyName))
				})

				ginkgo.By(fmt.Sprintf("And navigate to '%s' Policy page", policyName), func() {
					gomega.Eventually(policyInfo.Name.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s policy detail page", policyName))
				})

				policyDetailPage := pages.GetPolicyDetailPage(webDriver)
				ginkgo.By(fmt.Sprintf("And verify '%s' policy page", policyName), func() {
					gomega.Eventually(policyDetailPage.Header.Text).Should(gomega.MatchRegexp(policyName), "Failed to verify dashboard policy name ")
					gomega.Eventually(policyDetailPage.ID.Text).Should(gomega.MatchRegexp(policyID), "Failed to verify policy ID on policy page")
					gomega.Eventually(policyDetailPage.ClusterName.Text).Should(gomega.MatchRegexp(policyClusterName), "Failed to verify policy cluster on policy page")
					gomega.Eventually(policyDetailPage.Severity.Text).Should(gomega.MatchRegexp(policySeverity), "Failed to verify policy Severity on policy page")
					gomega.Eventually(policyDetailPage.Category.Text).Should(gomega.MatchRegexp(policyCategory), "Failed to verify policy category on policy page")

					gomega.Expect(policyDetailPage.GetTags()).Should(gomega.ConsistOf(policyTags), "Failed to verify policy Tags on policy page")
					gomega.Expect(policyDetailPage.GetTargetedK8sKind()).Should(gomega.ConsistOf(policyTargetedKinds), "Failed to verify policy Targeted K8s Kind on policy page")
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' policy Details", policyName), func() {
					description := "This Policy is to ensure you are setting a value for your imagePullPolicy."
					howToSolve := `spec:\s*containers:\s*- imagePullPolicy: <policy>`
					code := `result = {\s*15\s*"issue detected": true,\s*16\s*"msg": sprintf\("imagePolicyPolicy must be '%v'; found '%v'",\[policy, image_policy\]\),\s*17\s*"violating_key": sprintf\("spec.template.spec.containers\[%v\].imagePullPolicy", \[i]\),\s*18\s*"recommended_value": policy\s*19\s*}`

					gomega.Expect(policyDetailPage.Description.Text()).Should(gomega.MatchRegexp(description), "Failed to verify policy Description on policy page")
					gomega.Expect(policyDetailPage.HowToSolve.Text()).Should(gomega.MatchRegexp(howToSolve), "Failed to verify policy 'How to solve' on policy page")
					gomega.Expect(policyDetailPage.Code.Text()).Should(gomega.MatchRegexp(code), "Failed to verify 'Policy Code' on policy page")
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' policy parameters", policyName), func() {
					parameter := policyDetailPage.GetParameter("policy")
					gomega.Expect(parameter.Name.Text()).Should(gomega.MatchRegexp(`policy`), "Failed to verify parameter policy 'Name'")
					gomega.Expect(parameter.Type.Text()).Should(gomega.MatchRegexp(`string`), "Failed to verify parameter policy 'Type'")
					gomega.Expect(parameter.Value.Text()).Should(gomega.MatchRegexp(`Always`), "Failed to verify parameter policy 'Value'")
					gomega.Expect(parameter.Required.Text()).Should(gomega.MatchRegexp(`True`), "Failed to verify parameter policy 'Required'")

					parameter = policyDetailPage.GetParameter("exclude_namespace")
					namespaces := "test-systems"
					gomega.Expect(parameter.Name.Text()).Should(gomega.MatchRegexp(`exclude_namespace`), "Failed to verify parameter exclude_namespace 'Name'")
					gomega.Expect(parameter.Type.Text()).Should(gomega.MatchRegexp(`array`), "Failed to verify parameter exclude_namespace 'Type'")
					gomega.Expect(parameter.Value.Text()).Should(gomega.MatchRegexp(namespaces), "Failed to verify parameter exclude_namespace 'Value'")
					gomega.Expect(parameter.Required.Text()).Should(gomega.MatchRegexp(`True`), "Failed to verify parameter exclude_namespace 'Required'")

					parameter = policyDetailPage.GetParameter("exclude_label_key")
					gomega.Expect(parameter.Name.Text()).Should(gomega.MatchRegexp(`exclude_label_key`), "Failed to verify parameter exclude_label_key 'Name'")
					gomega.Expect(parameter.Type.Text()).Should(gomega.MatchRegexp(`string`), "Failed to verify parameter exclude_label_key 'Type'")
					gomega.Expect(parameter.Value.Text()).Should(gomega.MatchRegexp(`undefined`), "Failed to verify parameter exclude_label_key 'Value'")
					gomega.Expect(parameter.Required.Text()).Should(gomega.MatchRegexp(`False`), "Failed to verify parameter exclude_label_key 'Required'")

					parameter = policyDetailPage.GetParameter("exclude_label_value")
					gomega.Expect(parameter.Name.Text()).Should(gomega.MatchRegexp(`exclude_label_value`), "Failed to verify parameter exclude_label_value 'Name'")
					gomega.Expect(parameter.Type.Text()).Should(gomega.MatchRegexp(`string`), "Failed to verify parameter exclude_label_value 'Type'")
					gomega.Expect(parameter.Value.Text()).Should(gomega.MatchRegexp(`undefined`), "Failed to verify parameter exclude_label_value 'Value'")
					gomega.Expect(parameter.Required.Text()).Should(gomega.MatchRegexp(`False`), "Failed to verify parameter exclude_label_value 'Required'")
				})

				ginkgo.By("And again navigate to Polisies page via header link", func() {
					gomega.Expect(policiesPage.PolicyHeader.Click()).Should(gomega.Succeed(), "Failed to navigate to Policies pages via header link")
					pages.WaitForPageToLoad(webDriver)
				})
			})
		})

		ginkgo.Context("[UI] Policies can be installed on leaf cluster", func() {
			var mgmtClusterContext string
			var leafClusterContext string
			var leafClusterkubeconfig string
			var clusterBootstrapCopnfig string
			var gitopsCluster string
			patSecret := "policy-pat"
			bootstrapLabel := "bootstrap"
			leafClusterName := "wge-leaf-policy-kind"
			leafClusterNamespace := "default"

			policiesYaml := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "policies.yaml")
			policyName := "Container Running As Root acceptance test"
			policyID := "weave.policies.container-running-as-root-acceptance-test"
			policySeverity := "High"
			policyCategory := "weave.categories.pod-security"
			policyTags := []string{"pci-dss", "cis-benchmark", "mitre-attack", "nist800-190", "gdpr", "default"}
			policyTargetedKinds := []string{"Deployment", "Job", "ReplicationController", "ReplicaSet", "DaemonSet", "StatefulSet", "CronJob"}

			ginkgo.JustBeforeEach(func() {

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
				_ = gitopsTestRunner.KubectlDelete([]string{}, policiesYaml)

			})

			ginkgo.It("Verify Policies can be installed on leaf cluster and monitored via management cluster dashboard", ginkgo.Label("integration", "policy", "leaf-policy"), func() {
				existingPoliciesCount := getPoliciesCount()
				leafClusterkubeconfig = createLeafClusterKubeconfig(leafClusterContext, leafClusterName, leafClusterNamespace)

				installPolicyAgent(leafClusterName)
				installTestPolicies(leafClusterName, policiesYaml)

				useClusterContext(mgmtClusterContext)
				createPATSecret(leafClusterNamespace, patSecret)
				clusterBootstrapCopnfig = createClusterBootstrapConfig(leafClusterName, leafClusterNamespace, bootstrapLabel, patSecret)
				gitopsCluster = connectGitopsCuster(leafClusterName, leafClusterNamespace, bootstrapLabel, leafClusterkubeconfig)
				createLeafClusterSecret(leafClusterNamespace, leafClusterkubeconfig)

				waitForLeafClusterAvailability(leafClusterName, "Ready")
				addKustomizationBases("leaf", leafClusterName, leafClusterNamespace)

				installTestPolicies("management", policiesYaml)
				pages.NavigateToPage(webDriver, "Policies")
				policiesPage := pages.GetPoliciesPage(webDriver)

				ginkgo.By("And wait for policies to be visibe on the dashboard", func() {
					pages.NavigateToPage(webDriver, "Policies")
					gomega.Eventually(policiesPage.PolicyHeader).Should(matchers.BeVisible())

					totalPolicyCount := existingPoliciesCount + 8 // 4 management and 4 leaf policies
					gomega.Eventually(func(g gomega.Gomega) int {
						gomega.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
						time.Sleep(POLL_INTERVAL_1SECONDS)
						return policiesPage.CountPolicies()
					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_3SECONDS).Should(gomega.Equal(totalPolicyCount), fmt.Sprintf("There should be %d policy enteries in policy table", totalPolicyCount))

					// Wait for policy page to completely render policy information. Sometimes error appears momentarily due to RBAC reconciliation
					gomega.Eventually(func(g gomega.Gomega) bool {
						if !pages.ElementExist(policiesPage.AlertError) {
							return true
						}
						g.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
						return false
					}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.BeTrue(), "Policy page failed to render policies with complete policies information")
				})

				ginkgo.By(fmt.Sprintf("And filter leaf cluster '%s' policies", leafClusterName), func() {
					filterID := "clusterName: " + leafClusterNamespace + `/` + leafClusterName
					searchPage := pages.GetSearchPage(webDriver)
					searchPage.SelectFilter("cluster", filterID)
				})

				policyInfo := policiesPage.FindPolicyInList(policyName)
				ginkgo.By(fmt.Sprintf("And verify '%s' policy Name", policyName), func() {
					gomega.Eventually(policyInfo.Name).Should(matchers.MatchText(policyName), fmt.Sprintf("Failed to list %s policy in  application table", policyName))
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' policy Category", policyName), func() {
					gomega.Eventually(policyInfo.Category).Should(matchers.MatchText(policyCategory), fmt.Sprintf("Failed to have expected %s policy Category: weave.categories.pod-security", policyName))
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' policy Severity", policyName), func() {
					gomega.Eventually(policyInfo.Severity).Should(matchers.MatchText(policySeverity), fmt.Sprintf("Failed to have expected %s Policy Severity: Medium", policyName))
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' policy Cluster", policyName), func() {
					gomega.Eventually(policyInfo.Cluster).Should(matchers.MatchText(leafClusterName), fmt.Sprintf("Failed to have expected %[1]v policy Cluster: %[1]v", policyName))
				})

				ginkgo.By(fmt.Sprintf("And navigate to '%s' Policy page", policyName), func() {
					gomega.Eventually(policyInfo.Name.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s policy detail page", policyName))
				})

				policyDetailPage := pages.GetPolicyDetailPage(webDriver)
				ginkgo.By(fmt.Sprintf("And verify '%s' policy page", policyName), func() {
					gomega.Eventually(policyDetailPage.Header.Text).Should(gomega.MatchRegexp(policyName), "Failed to verify dashboard policy name ")
					gomega.Eventually(policyDetailPage.ID.Text).Should(gomega.MatchRegexp(policyID), "Failed to verify policy ID on policy page")
					gomega.Eventually(policyDetailPage.ClusterName.Text).Should(gomega.MatchRegexp(leafClusterName), "Failed to verify policy cluster on policy page")
					gomega.Eventually(policyDetailPage.Severity.Text).Should(gomega.MatchRegexp(policySeverity), "Failed to verify policy Severity on policy page")
					gomega.Eventually(policyDetailPage.Category.Text).Should(gomega.MatchRegexp(policyCategory), "Failed to verify policy category on policy page")

					gomega.Expect(policyDetailPage.GetTags()).Should(gomega.ConsistOf(policyTags), "Failed to verify policy Tags on policy page")
					gomega.Expect(policyDetailPage.GetTargetedK8sKind()).Should(gomega.ConsistOf(policyTargetedKinds), "Failed to verify policy Targeted K8s Kind on policy page")
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' policy Details", policyName), func() {
					description := "This Policy enforces that the securityContext.runAsNonRoot attribute is set to true."
					howToSolve := `spec:\s*securityContext:\s* runAsNonRoot: true`
					code := `result = {\s*20\s*"issue detected": true,\s*21\s*"msg": sprintf\("Container missing spec.template.spec.containers\[%v\].securityContext.runAsNonRoot while Pod.*\s*22\s*"violating_key": sprintf\("spec.template.spec.containers\[%v\].securityContext", \[i\]\)`

					gomega.Expect(policyDetailPage.Description.Text()).Should(gomega.MatchRegexp(description), "Failed to verify policy Description on policy page")
					gomega.Expect(policyDetailPage.HowToSolve.Text()).Should(gomega.MatchRegexp(howToSolve), "Failed to verify policy 'How to solve' on policy page")
					gomega.Expect(policyDetailPage.Code.Text()).Should(gomega.MatchRegexp(code), "Failed to verify 'Policy Code' on policy page")
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' policy parameters", policyName), func() {
					parameter := policyDetailPage.GetParameter("exclude_namespace")
					namespaces := "test-systems"
					gomega.Expect(parameter.Name.Text()).Should(gomega.MatchRegexp(`exclude_namespace`), "Failed to verify parameter exclude_namespace 'Name'")
					gomega.Expect(parameter.Type.Text()).Should(gomega.MatchRegexp(`array`), "Failed to verify parameter exclude_namespace 'Type'")
					gomega.Expect(parameter.Value.Text()).Should(gomega.MatchRegexp(namespaces), "Failed to verify parameter exclude_namespace 'Value'")
					gomega.Expect(parameter.Required.Text()).Should(gomega.MatchRegexp(`False`), "Failed to verify parameter exclude_namespace 'Required'")

					parameter = policyDetailPage.GetParameter("exclude_label_key")
					gomega.Expect(parameter.Name.Text()).Should(gomega.MatchRegexp(`exclude_label_key`), "Failed to verify parameter exclude_label_key 'Name'")
					gomega.Expect(parameter.Type.Text()).Should(gomega.MatchRegexp(`string`), "Failed to verify parameter exclude_label_key 'Type'")
					gomega.Expect(parameter.Value.Text()).Should(gomega.MatchRegexp(`undefined`), "Failed to verify parameter exclude_label_key 'Value'")
					gomega.Expect(parameter.Required.Text()).Should(gomega.MatchRegexp(`False`), "Failed to verify parameter exclude_label_key 'Required'")

					parameter = policyDetailPage.GetParameter("exclude_label_value")
					gomega.Expect(parameter.Name.Text()).Should(gomega.MatchRegexp(`exclude_label_value`), "Failed to verify parameter exclude_label_value 'Name'")
					gomega.Expect(parameter.Type.Text()).Should(gomega.MatchRegexp(`string`), "Failed to verify parameter exclude_label_value 'Type'")
					gomega.Expect(parameter.Value.Text()).Should(gomega.MatchRegexp(`undefined`), "Failed to verify parameter exclude_label_value 'Value'")
					gomega.Expect(parameter.Required.Text()).Should(gomega.MatchRegexp(`False`), "Failed to verify parameter exclude_label_value 'Required'")
				})

				ginkgo.By("And again navigate to Polisies page via header link", func() {
					gomega.Expect(policiesPage.PolicyHeader.Click()).Should(gomega.Succeed(), "Failed to navigate to Policies pages via header link")
					pages.WaitForPageToLoad(webDriver)
				})
			})
		})
	})
}
