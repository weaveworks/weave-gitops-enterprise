package acceptance

import (
	"fmt"
	"path"
	"regexp"
	"strconv"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

var _ = ginkgo.Describe("Multi-Cluster Control Plane Cost Estimation", func() {
	DEPLOYMENT_APP := "my-mccp-cluster-service"

	ginkgo.BeforeEach(func() {
		gomega.Expect(webDriver.Navigate(testUiUrl)).To(gomega.Succeed())

		if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
			loginUser()
		}
	})

	ginkgo.AfterEach(func() {
		_ = runCommandPassThrough("kubectl", "delete", "CapiTemplate", "--all")
		_ = runCommandPassThrough("kubectl", "delete", "GitOpsTemplate", "--all")
	})

	ginkgo.Context("[UI] When Cost estimation feature is enabled", func() {

		ginkgo.JustBeforeEach(func() {
		})

		ginkgo.JustAfterEach(func() {

		})

		ginkgo.It("Verify capa EC2 cluster cost estimation", ginkgo.Label("integration", "cost"), func() {
			templateFiles := map[string]string{
				"capa-cluster-template": path.Join(testDataPath, "templates/cluster/aws/cluster-template-ec2.yaml"),
			}
			installGitOpsTemplate(templateFiles)

			pages.NavigateToPage(webDriver, "Templates")
			pages.WaitForPageToLoad(webDriver)

			templatesPage := pages.GetTemplatesPage(webDriver)
			ginkgo.By("And I should choose a template", func() {
				templateRow := templatesPage.GetTemplateInformation(webDriver, "capa-cluster-template")
				gomega.Expect(templateRow.CreateTemplate.Click()).To(gomega.Succeed())
			})

			createPage := pages.GetCreateClusterPage(webDriver)
			ginkgo.By("And wait for Create cluster page to be fully rendered", func() {
				pages.WaitForPageToLoad(webDriver)
				gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
			})

			// Parameter values
			leafCluster := ClusterConfig{
				Type:      "capi",
				Name:      "quick-capa-cluster",
				Namespace: "quick-capi",
			}
			k8Version := "v1.23.3"
			awsRegion := "us-east-1"
			controlPlaneMachineCount := "3"
			workerMachineCount := "3"
			expectedCost := 386 // for 3 control plane + 3 worker node in us-east-1

			parameters := []TemplateField{
				{
					Name:   "AWS_REGION",
					Value:  "",
					Option: awsRegion,
				},
				{
					Name:   "AWS_SSH_KEY_NAME",
					Value:  "",
					Option: "weave-gitops-pesto",
				},
				{
					Name:   "CLUSTER_NAME",
					Value:  leafCluster.Name,
					Option: "",
				},
				{
					Name:   "CONTROL_PLANE_MACHINE_COUNT",
					Value:  "",
					Option: controlPlaneMachineCount,
				},
				{
					Name:   "KUBERNETES_VERSION",
					Value:  "",
					Option: k8Version,
				},
				{
					Name:   "NAMESPACE",
					Value:  leafCluster.Namespace,
					Option: "",
				},
				{
					Name:   "WORKER_MACHINE_COUNT",
					Value:  "",
					Option: workerMachineCount,
				},
				{
					Name:   "COST_ESTIMATION_FILTERS",
					Value:  "tenancy=Dedicated&capacityStatus=UnusedCapacityReservation&operation=RunInstances",
					Option: "",
				},
			}

			setParameterValues(createPage, parameters)

			costEstimation := pages.GetCostEstimation(webDriver)
			ginkgo.By("And verify aws cluster cost estimation", func() {
				gomega.Eventually(costEstimation.Label).Should(matchers.BeVisible(), "Cost Estimation features is disabled")
				gomega.Eventually(costEstimation.Estimate.Click).Should(gomega.Succeed(), `Failed to click 'Get Estimation' button`)

				gomega.Eventually(costEstimation.Price).Should(matchers.MatchText(`\$\d+\.?\d*\sUSD`), `The price format text doesn't match`)
				re := regexp.MustCompile(`\d+`)
				priceTxt, _ := costEstimation.Price.Text()
				price, _ := strconv.ParseFloat(re.FindAllString(priceTxt, -1)[0], 32)
				gomega.Expect(price).Should(gomega.BeNumerically("~", expectedCost, 10), fmt.Sprintf("Cluster stimated cost should not exceeds the expected threshold boundaries: %d <=> %d", expectedCost-10, expectedCost+10))
			})

			// Now modify cluster for new estimated cost
			awsRegion = "ca-central-1"
			workerMachineCount = "5"
			expectedCost = 574 // for 3 control plane + 5 worker node in ca-central-1
			parameters = []TemplateField{
				{
					Name:   "AWS_REGION",
					Value:  "",
					Option: awsRegion,
				},
				{
					Name:   "WORKER_MACHINE_COUNT",
					Value:  "",
					Option: workerMachineCount,
				},
			}

			// Set new cluster parameters values
			setParameterValues(createPage, parameters)
			ginkgo.By("And verify aws cluster cost estimation after setting new parameter values", func() {
				gomega.Eventually(costEstimation.Estimate.Click).Should(gomega.Succeed(), `Failed to click 'Get Estimation' button`)
				gomega.Eventually(costEstimation.Price).Should(matchers.MatchText(`\$\d+\.?\d*\sUSD`), `The price format text doesn't match`)
				re := regexp.MustCompile(`\d+`)
				priceTxt, _ := costEstimation.Price.Text()
				price, _ := strconv.ParseFloat(re.FindAllString(priceTxt, -1)[0], 32)
				gomega.Expect(price).Should(gomega.BeNumerically("~", expectedCost, 10), fmt.Sprintf("Cluster estimated cost should not exceeds the expected threshold boundaries: %d <=> %d", expectedCost-10, expectedCost+10))
			})
		})

		ginkgo.It("Verify non-supported (eks) capa cluster cost estimation", ginkgo.Label("integration", "cost"), func() {
			templateFiles := map[string]string{
				"capa-cluster-template": path.Join(testDataPath, "templates/cluster/aws/cluster-template-eks.yaml"),
			}
			installGitOpsTemplate(templateFiles)

			pages.NavigateToPage(webDriver, "Templates")
			pages.WaitForPageToLoad(webDriver)

			templatesPage := pages.GetTemplatesPage(webDriver)
			ginkgo.By("And I should choose a template", func() {
				templateRow := templatesPage.GetTemplateInformation(webDriver, "capa-cluster-template-eks")
				gomega.Expect(templateRow.CreateTemplate.Click()).To(gomega.Succeed())
			})

			createPage := pages.GetCreateClusterPage(webDriver)
			ginkgo.By("And wait for Create cluster page to be fully rendered", func() {
				pages.WaitForPageToLoad(webDriver)
				gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
			})

			// Parameter values
			leafCluster := ClusterConfig{
				Type:      "capi",
				Name:      "quick-capa-cluster",
				Namespace: "quick-capi",
			}

			parameters := []TemplateField{
				{
					Name:   "CLUSTER_NAME",
					Value:  leafCluster.Name,
					Option: "",
				},
				{
					Name:   "NAMESPACE",
					Value:  leafCluster.Namespace,
					Option: "",
				},
			}

			setParameterValues(createPage, parameters)

			costEstimation := pages.GetCostEstimation(webDriver)
			ginkgo.By("And verify aws cluster cost estimation", func() {
				gomega.Eventually(costEstimation.Label).Should(matchers.BeVisible(), "Cost Estimation features is disabled")
				gomega.Eventually(costEstimation.Estimate.Click).Should(gomega.Succeed(), `Failed to click 'Get Estimation' button`)

				gomega.Eventually(costEstimation.Price).Should(matchers.MatchText(`\$0.00`), `Failed to verify expected estimated cluser cost`)
				gomega.Eventually(costEstimation.Message).Should(matchers.MatchText(`could not find infrastructure controlplane`), `The cluster cost estimation should fail due to infrastructure controlplane not found`)

			})
		})

		ginkgo.It("Verify capa machinepool cost estimation", ginkgo.Label("integration", "cost"), func() {
			templateFiles := map[string]string{
				"capa-cluster-template": path.Join(testDataPath, "templates/cluster/aws/cluster-template-machinepool.yaml"),
			}
			installGitOpsTemplate(templateFiles)

			pages.NavigateToPage(webDriver, "Templates")
			pages.WaitForPageToLoad(webDriver)

			templatesPage := pages.GetTemplatesPage(webDriver)
			ginkgo.By("And I should choose a template", func() {
				templateRow := templatesPage.GetTemplateInformation(webDriver, "capa-cluster-template-machinepool")
				gomega.Expect(templateRow.CreateTemplate.Click()).To(gomega.Succeed())
			})

			createPage := pages.GetCreateClusterPage(webDriver)
			ginkgo.By("And wait for Create cluster page to be fully rendered", func() {
				pages.WaitForPageToLoad(webDriver)
				gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
			})

			// Parameter values
			leafCluster := ClusterConfig{
				Type:      "capi",
				Name:      "quick-capa-cluster",
				Namespace: "quick-capi",
			}

			parameters := []TemplateField{
				{
					Name:   "CLUSTER_NAME",
					Value:  leafCluster.Name,
					Option: "",
				},
				{
					Name:   "NAMESPACE",
					Value:  leafCluster.Namespace,
					Option: "",
				},
				{
					Name:   "COST_ESTIMATION_FILTERS",
					Value:  "capacityStatus=UnusedCapacityReservation&operation=RunInstances",
					Option: "",
				},
			}

			setParameterValues(createPage, parameters)

			costEstimation := pages.GetCostEstimation(webDriver)
			ginkgo.By("And verify aws cluster cost estimation after setting new parameter values", func() {
				gomega.Eventually(costEstimation.Estimate.Click).Should(gomega.Succeed(), `Failed to click 'Get Estimation' button`)
				gomega.Eventually(costEstimation.Price).Should(matchers.MatchText(`\$\d+\.?\d*\sUSD`), `The price format text doesn't match`)
				re := regexp.MustCompile(`[-]?\d[\d]*\.?\d*`)
				priceTxt, _ := costEstimation.Price.Text()
				a := re.FindAllString(priceTxt, -1)
				fmt.Println(a)
				expectedCost := []int{144, 155}
				for i, p := range re.FindAllString(priceTxt, -1) {
					price, _ := strconv.ParseFloat(p, 32)
					gomega.Expect(price).Should(gomega.BeNumerically("~", expectedCost[i], 10), fmt.Sprintf("Cluster stimated cost should not exceeds the expected threshold boundaries: %d <=> %d", expectedCost[i]-10, expectedCost[i]+10))
				}
			})
		})
	})

	ginkgo.Context("[UI] When aws pricing secret is not available", func() {

		ginkgo.JustBeforeEach(func() {
			ginkgo.By("And create invalid aws-pricing secret", func() {
				// Delete existing valid secret
				deleteSecret([]string{"aws-pricing"}, GITOPS_DEFAULT_NAMESPACE)
				// Create new invalid secret
				err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl create secret generic aws-pricing --namespace=%s --from-literal="AWS_ACCESS_KEY_ID=YRTDAY7892NKETHCT7rHD" --from-literal="AWS_SECRET_ACCESS_KEY=78tkihRmK8UoXVT905WjFESzsl232fFENoUFz532"`,
					GITOPS_DEFAULT_NAMESPACE))
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to create secret for leaf cluster kubeconfig")
			})
			gomega.Expect(restartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE)).ShouldNot(gomega.HaveOccurred(), "Failed restart deployment successfully")
			loginUser()
		})

		ginkgo.JustAfterEach(func() {

			ginkgo.By("And create valid aws-pricing secret", func() {
				// Delete invalid secret
				deleteSecret([]string{"aws-pricing"}, GITOPS_DEFAULT_NAMESPACE)
				// Create valid secret
				err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl create secret generic aws-pricing --namespace=%s --from-literal="AWS_ACCESS_KEY_ID=%s" --from-literal="AWS_SECRET_ACCESS_KEY=%s"`,
					GITOPS_DEFAULT_NAMESPACE, GetEnv("AWS_ACCESS_KEY_ID", ""), GetEnv("AWS_SECRET_ACCESS_KEY", "")))
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to create secret for leaf cluster kubeconfig")
			})
			gomega.Expect(restartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE)).ShouldNot(gomega.HaveOccurred(), "Failed restart deployment successfully")
			loginUser()
		})

		ginkgo.It("Verify capa cost estimation with invalid pricing secrert", ginkgo.Label("integration", "cost"), func() {
			templateFiles := map[string]string{
				"capa-cluster-template": path.Join(testDataPath, "templates/cluster/aws/cluster-template-machinepool.yaml"),
			}
			installGitOpsTemplate(templateFiles)

			pages.NavigateToPage(webDriver, "Templates")
			pages.WaitForPageToLoad(webDriver)

			templatesPage := pages.GetTemplatesPage(webDriver)
			ginkgo.By("And I should choose a template", func() {
				templateRow := templatesPage.GetTemplateInformation(webDriver, "capa-cluster-template-machinepool")
				gomega.Expect(templateRow.CreateTemplate.Click()).To(gomega.Succeed())
			})

			createPage := pages.GetCreateClusterPage(webDriver)
			ginkgo.By("And wait for Create cluster page to be fully rendered", func() {
				pages.WaitForPageToLoad(webDriver)
				gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
			})

			// Parameter values
			leafCluster := ClusterConfig{
				Type:      "capi",
				Name:      "quick-capa-cluster",
				Namespace: "quick-capi",
			}

			parameters := []TemplateField{
				{
					Name:   "CLUSTER_NAME",
					Value:  leafCluster.Name,
					Option: "",
				},
			}

			setParameterValues(createPage, parameters)

			costEstimation := pages.GetCostEstimation(webDriver)
			ginkgo.By("And verify aws cluster cost estimation after setting new parameter values", func() {
				gomega.Eventually(costEstimation.Estimate.Click).Should(gomega.Succeed(), `Failed to click 'Get Estimation' button`)
				gomega.Eventually(costEstimation.Price).Should(matchers.MatchText(`\$0.00`), `Failed to verify expected estimated cluser cost`)

				gomega.Eventually(costEstimation.Message).Should(matchers.MatchText(`error getting prices for estimation`), `The cluster cost estimation should fail due to invalid secret`)
			})
		})
	})
})
