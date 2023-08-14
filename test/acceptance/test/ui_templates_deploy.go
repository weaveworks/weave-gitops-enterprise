package acceptance

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func installIngressNginx(clusterName string, ingressNodePort int) {
	ginkgo.By(fmt.Sprintf("And install cert-manager to %s cluster for tls certificate creation", clusterName), func() {
		stdOut, _ := runCommandAndReturnStringOutput("helm search repo cert-manager")
		if !strings.Contains(stdOut, `cert-manager/cert-manager`) {
			err := runCommandPassThrough("helm", "repo", "add", "cert-manager", "https://charts.jetstack.io")
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to add cert-manage repositoy")
		}

		err := runCommandPassThrough("sh", "-c", `helm upgrade --install cert-manager cert-manager/cert-manager --namespace cert-manager --create-namespace --version 1.10.0 --wait --set installCRDs=true`)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to install cert-manager to leaf cluster '%s'", clusterName))
	})

	ginkgo.By(fmt.Sprintf("And install ingress-nginx to %s cluster for tls termination", clusterName), func() {
		stdOut, _ := runCommandAndReturnStringOutput("helm search repo ingress-nginx")
		if !strings.Contains(stdOut, `ingress-nginx/ingress-nginx`) {
			err := runCommandPassThrough("sh", "-c", "helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx")
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to add ingress-nginx repositoy")
		}

		err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx --namespace ingress-nginx --create-namespace --version 4.4.0 --wait --set controller.service.type=NodePort --set controller.service.nodePorts.https=%d, --set controller.extraArgs.v=4`, ingressNodePort))
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to install ingress-nginx to leaf cluster '%s'", clusterName))
	})
}

func verifySelfServicePodinfo(podinfo Application, webUrl, bgColour, message string) {
	ginkgo.By("And verify mvp self service podinfo", func() {
		windowName := "podinfo"
		currentWindow, err := webDriver.Session().GetWindow()
		gomega.Expect(err).To(gomega.BeNil(), "Failed to get current/active window")

		pages.OpenWindowInBg(webDriver, webUrl, windowName)
		gomega.Expect(webDriver.NextWindow()).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to switch to '%s' window", windowName))

		gomega.Eventually(func(g gomega.Gomega) {
			g.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
			g.Eventually(webDriver.Title, ASSERTION_5SECONDS_TIME_OUT).Should(gomega.MatchRegexp(strings.Join([]string{podinfo.TargetNamespace, podinfo.Name}, "-")))
		}, ASSERTION_1MINUTE_TIME_OUT).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to verify '%s' window title", windowName))

		gomega.Eventually(webDriver.Find(fmt.Sprintf(`div[style*="background-color: %s"]`, bgColour))).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify '%s' background colour", "podinfoApp"))
		gomega.Eventually(webDriver.Find(`div h1`).Text).Should(gomega.MatchRegexp(message), fmt.Sprintf("Failed to verify '%s' message", "podinfoApp"))
		time.Sleep(POLL_INTERVAL_1SECONDS)
		gomega.Eventually(webDriver.CloseWindow).Should(gomega.Succeed(), fmt.Sprintf("Failed to close '%s' window", windowName))
		gomega.Expect(webDriver.Session().SetWindow(currentWindow)).ShouldNot(gomega.HaveOccurred(), "Failed to switch to weave gitops enterprise dashboard")
	})
}

var _ = ginkgo.Describe("Multi-Cluster Control Plane GitOpsTemplates for deployments", ginkgo.Label("ui", "template", "deploy"), func() {
	var templateNamespaces []string

	ginkgo.BeforeEach(func() {
		gomega.Expect(webDriver.Navigate(testUiUrl)).To(gomega.Succeed())

		if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
			loginUser()
		}
	})

	ginkgo.AfterEach(func() {
		_ = runCommandPassThrough("kubectl", "delete", "CapiTemplate", "--all")
		_ = runCommandPassThrough("kubectl", "delete", "GitOpsTemplate", "--all")
		deleteNamespace(templateNamespaces)
	})

	ginkgo.Context("GitOps Template can create resources in the leaf cluster", ginkgo.Label("kind-leaf-cluster"), func() {
		var mgmtClusterContext string
		var leafClusterContext string
		var leafClusterkubeconfig string
		var clusterBootstrapCopnfig string
		var gitopsCluster string
		var existingAppCount int

		const leafIngressHost = "mvp.wge.com" // ingress host must be resolvable
		const leafIngressNodePort = 30081
		const leafIngressNamespace = "test-system"

		appNameSpace := leafIngressNamespace
		appTargetNamespace := leafIngressNamespace

		leafCluster := ClusterConfig{
			Type:      "other",
			Name:      "wge-leaf-mvp-kind",
			Namespace: "mvp-system",
		}

		bootstrapLabel := "bootstrap"
		patSecret := "leaf-pat"

		ginkgo.JustBeforeEach(func() {
			existingAppCount = getApplicationCount()
			mgmtClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
			// Create vanilla kind leaf cluster
			createCluster("kind", leafCluster.Name, "kind/extra-port-mapping-kind-config.yaml")
			leafClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
			// Creating namespace in leafcluster for ingress and helmrelease application
			createNamespace([]string{appNameSpace})
		})

		ginkgo.JustAfterEach(func() {
			useClusterContext(mgmtClusterContext)

			deleteSecret([]string{leafClusterkubeconfig, patSecret}, leafCluster.Namespace)
			_ = runCommandPassThrough("kubectl", "delete", "-f", clusterBootstrapCopnfig)
			_ = runCommandPassThrough("kubectl", "delete", "-f", gitopsCluster)

			deleteCluster("kind", leafCluster.Name, "")
			cleanGitRepository(path.Join("./clusters", leafCluster.Namespace))
			deleteNamespace([]string{leafCluster.Namespace})
		})

		ginkgo.It("Verify self service MVP deployment workflow for leaf cluster and management dashboard is updated accordingly", ginkgo.Label("smoke", "application"), func() {
			repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
			leafClusterkubeconfig = createLeafClusterKubeconfig(leafClusterContext, leafCluster.Name, leafCluster.Namespace)

			useClusterContext(mgmtClusterContext)
			createNamespace([]string{leafCluster.Namespace})

			createPATSecret(leafCluster.Namespace, patSecret)
			clusterBootstrapCopnfig = createClusterBootstrapConfig(leafCluster.Name, leafCluster.Namespace, bootstrapLabel, patSecret)
			gitopsCluster = connectGitopsCluster(leafCluster.Name, leafCluster.Namespace, bootstrapLabel, leafClusterkubeconfig)
			createLeafClusterSecret(leafCluster.Namespace, leafClusterkubeconfig)

			waitForLeafClusterAvailability(leafCluster.Name, "Ready")
			addKustomizationBases("leaf", leafCluster.Name, leafCluster.Namespace)

			ginkgo.By(fmt.Sprintf("And I verify %s GitopsCluster/leafCluster is bootstraped)", leafCluster.Name), func() {
				useClusterContext(leafClusterContext)
				verifyFluxControllers(GITOPS_DEFAULT_NAMESPACE)
				waitForGitRepoReady("flux-system", GITOPS_DEFAULT_NAMESPACE)
			})

			// Install ingress-nginx on leaf cluster for deployment UI
			installIngressNginx(leafCluster.Name, leafIngressNodePort)

			useClusterContext(mgmtClusterContext)
			pages.NavigateToPage(webDriver, "Applications")
			applicationsPage := pages.GetApplicationsPage(webDriver)

			ginkgo.By("And wait for existing applications to be visibe on the dashboard", func() {
				gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())
				existingAppCount += 2 // flux-system + clusters-bases-kustomization (leaf cluster)
			})

			// Add deployment to leafcluster using GitOpsTemplate
			templateName := "mvp-helmrelease-template"
			templateNamespace := "default"
			templateFiles := map[string]string{
				templateName: path.Join(testDataPath, "templates/application/mvp-helmrelease-template.yaml"),
			}

			installGitOpsTemplate(templateFiles)
			pages.NavigateToPage(webDriver, "Templates")
			waitForTemplatesToAppear(len(templateFiles))
			templatesPage := pages.GetTemplatesPage(webDriver)

			ginkgo.By("And I should choose a template", func() {
				templateRow := templatesPage.GetTemplateInformation(webDriver, templateName)
				gomega.Eventually(templateRow.CreateTemplate.Click, ASSERTION_2MINUTE_TIME_OUT).Should(gomega.Succeed())
			})

			createPage := pages.GetCreateClusterPage(webDriver)
			ginkgo.By("And wait for Create resource page to be fully rendered", func() {
				pages.WaitForPageToLoad(webDriver)
				gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
			})

			// Deployment application template parameters
			podinfo := Application{
				Name:            "podinfo",
				Namespace:       appNameSpace,
				Type:            "helm_release",
				Url:             "https://stefanprodan.github.io/podinfo",
				Chart:           "podinfo",
				Version:         "6.3.0",
				TargetNamespace: appTargetNamespace,
				Values:          fmt.Sprintf(`"{ui: {color: \"#7c4e34\", message: \"Hello World from %s cluster\"}}"`, leafCluster.Name),
			}
			// Ingress template parameters for deploymnet application
			appServiceName := strings.Join([]string{podinfo.TargetNamespace, podinfo.Name}, "-")
			appServicePort := "9898" // default podinfo service

			podinfoWebUrl := fmt.Sprintf(`https://%s:%d`, leafIngressHost, leafIngressNodePort)
			leafCluterPath := path.Join("clusters", leafCluster.Namespace, leafCluster.Name)
			templateRepoPath := path.Join(leafCluterPath, "mvp", podinfo.Namespace, podinfo.Name+".yaml")

			// Comment out this step if runnig the test locally. Local user is not a root user and failed to edit /etc/hosts file which fails the step.
			// Add host entry manually to the /etc/hosts file before running test locally.
			ginkgo.By("And set cluster hostname mapping in the /etc/hosts file for deploy service", func() {
				err := runCommandPassThrough(path.Join(getCheckoutRepoPath(), "test", "utils", "scripts", "hostname-to-ip.sh"), leafIngressHost)
				// Ignore error checking when running the test locally as the local test host user is not a root user, instead add the entry manually to /etc/hosts file
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to set deployment service hostname entry in /etc/hosts file")
			})

			// Not setting the helm chart values via parameters
			var parameters = []TemplateField{
				{
					Name:  "CLUSTER_PATH",
					Value: leafCluterPath,
				},
				{
					Name:  "RESOURCE_NAME",
					Value: podinfo.Name,
				},
				{
					Name:  "NAMESPACE",
					Value: podinfo.Namespace,
				},
				{
					Name:  "URL",
					Value: podinfo.Url,
				},
				{
					Name:  "CHART_NAME",
					Value: podinfo.Chart,
				},
				{
					Name:  "CHART_VERSION",
					Value: podinfo.Version,
				},
				{
					Name:  "TARGET_NAMESPACE",
					Value: podinfo.TargetNamespace,
				},
				{
					Name:  "HOST_NAME",
					Value: leafIngressHost,
				},
				{
					Name:  "SERVICE_NAME",
					Value: appServiceName,
				},
				{
					Name:  "SERVICE_PORT",
					Value: appServicePort,
				},
			}

			setParameterValues(createPage, parameters)

			preview := pages.GetPreview(webDriver)
			ginkgo.By("Then I should preview the PR", func() {
				gomega.Eventually(func(g gomega.Gomega) {
					g.Expect(createPage.PreviewPR.Click()).Should(gomega.Succeed())
					g.Expect(preview.Title.Text()).Should(gomega.MatchRegexp("PR Preview"))
					g.Expect(preview.Path.At(0)).Should(matchers.MatchText(templateRepoPath))
				}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Succeed(), "Failed to get PR preview")
			})

			ginkgo.By("Then verify all resources are labelled with template name and namespace", func() {
				// Verify resource definition preview
				gomega.Eventually(preview.GetPreviewTab("Resource Definition").Click).Should(gomega.Succeed(), "Failed to switch to 'RESOURCE DEFINITION' preview tab")
				previewText, _ := preview.Text.At(0).Text()

				re, _ := regexp.Compile(fmt.Sprintf("templates.weave.works/template-name: %s", templateName))
				matches := re.FindAllString(previewText, -1)
				gomega.Eventually(len(matches)).Should(gomega.Equal(3), "All resources should be labelled with template name")

				re, _ = regexp.Compile(fmt.Sprintf("templates.weave.works/template-namespace: %s", templateNamespace))
				matches = re.FindAllString(previewText, -1)
				gomega.Eventually(len(matches)).Should(gomega.Equal(3), "All resources should be labelled with template namespace")

				gomega.Eventually(preview.Close.Click).Should(gomega.Succeed(), "Failed to close the preview dialog")
			})

			// Pull request values
			pullRequest := PullRequest{
				Branch:  fmt.Sprintf("br-mvp-%s", strings.ToLower(randString(5))),
				Title:   "Self service MVP deployment",
				Message: "An exapmle self service MVP podinfo deployment",
			}
			_ = createGitopsPR(pullRequest)

			ginkgo.By("Then I should merge the pull request to start resource provisioning", func() {
				createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
				mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
			})

			ginkgo.By("Then force reconcile flux-system to immediately start resource provisioning", func() {
				useClusterContext(leafClusterContext)
				reconcile("reconcile", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				reconcile("reconcile", "", "kustomization", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				useClusterContext(mgmtClusterContext)
			})

			pages.NavigateToPage(webDriver, "Applications")
			ginkgo.By(fmt.Sprintf("And wait for %s application to be visibe on the dashboard", podinfo.Name), func() {
				gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())

				totalAppCount := existingAppCount + 1
				gomega.Eventually(applicationsPage.CountApplications, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))
			})

			verifyAppInformation(applicationsPage, podinfo, leafCluster, "Ready")

			// Verify podinfo application bakground colour and message
			applicationInfo := applicationsPage.FindApplicationInList(podinfo.Name)
			ginkgo.By(fmt.Sprintf("And navigate to %s application page", podinfo.Name), func() {
				gomega.Eventually(applicationInfo.Name.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s application detail page", podinfo.Name))
			})
			verifyAppPage(podinfo)

			bgColour := "rgb(52, 87, 124)"      // default podinfo background colour '#34577c'
			message := "greetings from podinfo" // default podinfo message
			verifySelfServicePodinfo(podinfo, podinfoWebUrl, bgColour, message)

			ginkgo.By(fmt.Sprintf("Then I should start editing the %s application", podinfo.Name), func() {
				appDetailPage := pages.GetApplicationsDetailPage(webDriver, podinfo.Type)
				gomega.Eventually(appDetailPage.Edit.Click, ASSERTION_30SECONDS_TIME_OUT).Should(gomega.Succeed(), "Failed to click 'EDIT' application button")
			})

			// Edit the podinfo application bakground colour and message
			createPage = pages.GetCreateClusterPage(webDriver)
			ginkgo.By("And wait for Edit resource page to be fully rendered", func() {
				pages.WaitForPageToLoad(webDriver)
				gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(podinfo.Name))
			})

			// Only editing setting helm chart values for the application
			parameters = []TemplateField{
				{
					Name:  "VALUES",
					Value: podinfo.Values,
				},
			}
			setParameterValues(createPage, parameters)

			_ = createGitopsPR(PullRequest{}) // accepts default pull request values
			ginkgo.By("Then I should merge the pull request to start resource provisioning", func() {
				createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
				mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
			})

			ginkgo.By("Then force reconcile flux-system to immediately start resource provisioning", func() {
				useClusterContext(leafClusterContext)
				reconcile("reconcile", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				reconcile("reconcile", "", "kustomization", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				useClusterContext(mgmtClusterContext)
			})

			// Verify edited podinfo application bakground colour and message
			bgColour = "rgb(124, 78, 52)"                                          // edited podinfo background colour '#7c4e34'
			message = fmt.Sprintf("Hello World from %s cluster", leafCluster.Name) // edited podinfo message
			verifySelfServicePodinfo(podinfo, podinfoWebUrl, bgColour, message)

			verifyDeleteApplication(applicationsPage, existingAppCount, podinfo.Name, templateRepoPath)
		})
	})
})
