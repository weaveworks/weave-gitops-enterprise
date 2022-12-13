package acceptance

import (
	"context"
	"fmt"
	"path"
	"regexp"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Gitops miscellaneous CLI tests", ginkgo.Label("cli"), func() {

	var stdOut string
	var stdErr string

	ginkgo.BeforeEach(func() {

	})

	ginkgo.AfterEach(func() {

	})

	ginkgo.Context("[CLI] When no clusters are available in the management cluster", ginkgo.Label("cli", "cluster"), func() {

		ginkgo.It("Verify gitops lists no clusters", func() {
			ginkgo.By("And gitops state is reset", func() {
				resetControllers("enterprise")
				verifyEnterpriseControllers("my-mccp", "", GITOPS_DEFAULT_NAMESPACE)
				checkClusterService(wgeEndpointUrl)
			})

			stdOut, _ = runGitopsCommand(`get cluster`)

			ginkgo.By("Then gitops lists no clusters", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`management\s+Ready`))
			})
		})
	})

	ginkgo.Context("[CLI] When profiles are available in the management cluster", ginkgo.Label("profile"), func() {

		ginkgo.It("Verify gitops can list profiles from default profile repository", func() {
			ginkgo.By("And wait for cluster-service to cache profiles", func() {
				gomega.Expect(waitForGitopsResources(context.Background(), Request{Path: `charts/list?repository.name=weaveworks-charts&repository.namespace=flux-system&repository.cluster.name=management`}, POLL_INTERVAL_5SECONDS, ASSERTION_15MINUTE_TIME_OUT)).To(gomega.Succeed(), "Failed to get a successful response from /v1/charts")
			})

			stdOut, _ = runGitopsCommand(`get profiles`)

			ginkgo.By("Then gitops lists profiles with default values", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`cert-manager\s+[,.\d\w\s]+0.0.8,0.0.7[,.\d\w- ]+layer-0`))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`weave-policy-agent\s+[,.\d\w\s]+0.4.0[,.\d\w ]+layer-1`))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`metallb\s+[,.\d\w\s]+0.0.2,0.0.1[,.\d\w ]+layer-0`))
			})
		})

		ginkgo.It("Verify gitops can list profiles from any profile repository", func() {
			createNamespace([]string{"test-profiles"})
			defer deleteNamespace([]string{"test-profiles"})

			addSource("helm", "profiles-catalog", "test-profiles", "https://raw.githubusercontent.com/weaveworks/profiles-catalog/gh-pages", "", "")
			defer deleteSource("helm", "profiles-catalog", "test-profiles", "")

			ginkgo.By("And wait for cluster-service to cache profiles", func() {
				gomega.Expect(waitForGitopsResources(context.Background(), Request{Path: `charts/list?repository.name=profiles-catalog&repository.namespace=test-profiles&repository.cluster.name=management`}, POLL_INTERVAL_5SECONDS, ASSERTION_15MINUTE_TIME_OUT)).To(gomega.Succeed(), "Failed to get a successful response from /v1/charts")
			})

			stdOut, _ = runGitopsCommand(`get profiles --cluster-name management --repo-name profiles-catalog --repo-namespace test-profiles`)

			ginkgo.By("Then gitops lists profiles without defaults", func() {
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`dex\s+[,.\d\w\s]+0.0.11,0.0.10-0`))
				gomega.Eventually(stdOut).Should(gomega.MatchRegexp(`secrets-store-config\s+[,.\d\w\s]+0.0.1[,.\d\w- ]+layer-4`))
			})
		})
	})

	ginkgo.Context("[CLI] When entitlement is available in the cluster", ginkgo.Label("entitlement"), func() {
		var resourceName string
		DEPLOYMENT_APP := "my-mccp-cluster-service"

		checkEntitlement := func(typeEntitelment string, beFound bool) {
			checkOutput := func() bool {
				cmd := fmt.Sprintf(`get %s`, resourceName)
				stdOut, stdErr = runGitopsCommand(cmd, ASSERTION_1MINUTE_TIME_OUT)

				msg := stdErr + " " + stdOut

				if typeEntitelment == "expired" {
					re := regexp.MustCompile(`Your entitlement for Weave GitOps Enterprise has expired`)
					return re.MatchString(msg)
				}
				re := regexp.MustCompile(`No entitlement was found for Weave GitOps Enterprise`)
				return re.MatchString(msg)

			}

			matcher := gomega.BeFalse
			if beFound {
				matcher = gomega.BeTrue
			}

			resourceName = "templates"
			logger.Infof("Running 'gitops get %s --endpoint %s'", resourceName, wgeEndpointUrl)
			gomega.Eventually(checkOutput, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(matcher())
			resourceName = "credentials"
			logger.Infof("Running 'gitops get %s --endpoint %s'", resourceName, wgeEndpointUrl)
			gomega.Eventually(checkOutput, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(matcher())
			resourceName = "clusters"
			logger.Infof("Running 'gitops get %s --endpoint %s'", resourceName, wgeEndpointUrl)
			gomega.Eventually(checkOutput, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(matcher())
		}

		ginkgo.JustBeforeEach(func() {

		})

		ginkgo.JustAfterEach(func() {
			ginkgo.By("When I apply the valid entitlement", func() {
				gomega.Expect(runCommandPassThrough("kubectl", "apply", "-f", path.Join(testDataPath, "entitlement/entitlement-secret.yaml")), "Failed to create/configure entitlement")
			})

			ginkgo.By("Then I restart the cluster service pod for valid entitlemnt to take effect", func() {
				gomega.Expect(restartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
			})

			ginkgo.By("And the Cluster service is healthy", func() {
				checkClusterService(wgeEndpointUrl)
			})

			ginkgo.By("And I should not see the error or warning message for valid entitlement", func() {
				checkEntitlement("expired", false)
				checkEntitlement("invalid", false)
			})

			deleteIPCredentials("AWS")
			// Login to the dashbord because the logout automatically when the cluster service restarts for entitlement checking
			loginUser()
		})

		ginkgo.It("Verify cluster service acknowledges the entitlement presences", func() {

			ginkgo.By("When I delete the entitlement", func() {
				gomega.Expect(runCommandPassThrough("kubectl", "delete", "-f", path.Join(testDataPath, "entitlement/entitlement-secret.yaml")), "Failed to delete entitlement secret")
			})

			ginkgo.By("Then I restart the cluster service pod for missing entitlemnt to take effect", func() {
				gomega.Expect(restartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE)).ShouldNot(gomega.HaveOccurred(), "Failed restart deployment successfully")
			})

			ginkgo.By("And I should see the error message for missing entitlement", func() {
				checkEntitlement("missing", true)
			})

			ginkgo.By("When I apply the expired entitlement", func() {
				gomega.Expect(runCommandPassThrough("kubectl", "apply", "-f", path.Join(testDataPath, "entitlement/entitlement-secret-expired.yaml")), "Failed to create/configure entitlement")
			})

			ginkgo.By("Then I restart the cluster service pod for expired entitlemnt to take effect", func() {
				gomega.Expect(restartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
			})

			ginkgo.By("And I should see the warning message for expired entitlement", func() {
				checkEntitlement("expired", true)
			})

			ginkgo.By("When I apply the invalid entitlement", func() {
				gomega.Expect(runCommandPassThrough("kubectl", "apply", "-f", path.Join(testDataPath, "entitlement/entitlement-secret-invalid.yaml")), "Failed to create/configure entitlement")
			})

			ginkgo.By("Then I restart the cluster service pod for invalid entitlemnt to take effect", func() {
				gomega.Expect(restartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
			})

			ginkgo.By("And I should see the error message for invalid entitlement", func() {
				checkEntitlement("invalid", true)
			})
		})
	})

})
