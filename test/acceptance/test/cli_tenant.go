package acceptance

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"text/template"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func verifyTenantYaml(contents string, tenantName string, namespaces []string, sa bool, saName string, saRoles map[string][]string, groups []string) {
	ginkgo.By("Then I should verify generated tenant yaml file", func() {
		for _, namespace := range namespaces {
			// Namespace
			gomega.Eventually(contents).Should(gomega.MatchRegexp(fmt.Sprintf(`kind: Namespace[\s\w\d./:-]*toolkit.fluxcd.io\/tenant: %s[\s\w\d./:-]*name: %s\s`, tenantName, namespace)))
			// developmentRBAC
			if sa {
				gomega.Eventually(contents).Should(gomega.MatchRegexp(fmt.Sprintf(`kind: ServiceAccount[\s\w\d./:-]*toolkit.fluxcd.io\/tenant: %s[\s\w\d./:-]*name: %s\s[\s\w\d./:-]*namespace: %s\s`, tenantName, saName, namespace)))
			}
			for k, roles := range saRoles {
				for _, r := range roles {
					gomega.Eventually(contents).Should(gomega.MatchRegexp(fmt.Sprintf(`kind: RoleBinding[\s\w\d./:-]*toolkit.fluxcd.io\/tenant: %s[\s\w\d./:-]*namespace: %s\s[\s\w\d./:-]*kind: %s[\s\w\d./:-]*name: %s\s`, tenantName, namespace, k, r)))
				}
			}
			// teamsRBAC
			gomega.Eventually(contents).Should(gomega.MatchRegexp(fmt.Sprintf(`kind: Role[\s\w\d./:-]*toolkit.fluxcd.io\/tenant: %[1]v[\s\w\d./:-]*name: %[1]v-team\s[\s\w\d./:-]*namespace: %[2]v\s[\s\w\d./:-]*apiGroups`, tenantName, namespace)))
			for _, group := range groups {
				gomega.Eventually(contents).Should(gomega.MatchRegexp(fmt.Sprintf(`kind: RoleBinding[\s\w\d./:-]*toolkit.fluxcd.io\/tenant: %[2]v[\s\w\d./:-]*namespace: %[1]v\s[\s\w\d./:-]*kind: Role[\s\w\d./:-]*name: %[2]v-team\s[\s\w\d./:-]*kind: Group[\s\w\d./:-]*name: %[3]v\s`, namespace, tenantName, group)))
			}
			// policy
			gomega.Eventually(contents).Should(gomega.MatchRegexp(fmt.Sprintf(`kind: Policy[\s\w\d./:-]*name: weave.policies.tenancy.%s-allowed-application-deploy\s`, tenantName)))
			gomega.Eventually(contents).Should(gomega.MatchRegexp(fmt.Sprintf(`kind: Policy[\s\w\d./:-|(),"%%{}!-]*labels:[\s\w\d./:-]*toolkit.fluxcd.io/tenant: %s\s`, tenantName)))
			gomega.Eventually(contents).Should(gomega.MatchRegexp(fmt.Sprintf(`kind: Policy[\s\w\d./:-|(),"%%{}!-]*parameters:[\s\w\d./:-]*name: namespaces[\s\w\d./:-]*value:[\s\w\d./:-]*%s\s`, namespace)))
			gomega.Eventually(contents).Should(gomega.MatchRegexp(fmt.Sprintf(`kind: Policy[\s\w\d./:-|(),"%%{}!-]*parameters:[\s\w\d./:-]*name: service_account_name[\s\w\d./:-]*value: %s\s`, saName)))
			gomega.Eventually(contents).Should(gomega.MatchRegexp(fmt.Sprintf(`kind: Policy[\s\w\d./:-|(),"%%{}!-]*targets:[\s\w\d./:-]*HelmRelease[\s\w\d./:-]*Kustomization[\s\w\d./:-\[\]]*namespaces:[\s\w\d./:-]*%s\s`, namespace)))
		}
	})
}

func verifyTenatResources(tenantName string, namespaces []string, sa bool) {
	ginkgo.By("Then I should verify tenant resources", func() {
		for _, namespace := range namespaces {
			// Namespace
			_, stdErr := runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl get namespace -l toolkit.fluxcd.io/tenant=%s -n %s`, tenantName, namespace))
			gomega.Expect(stdErr).Should(gomega.BeEmpty(), "Tenant namespace is not created")
			// Service Account
			if sa {
				_, stdErr = runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl get serviceaccount  -l toolkit.fluxcd.io/tenant=%s -n %s`, tenantName, namespace))
				gomega.Expect(stdErr).Should(gomega.BeEmpty(), "Tenant serviceaccount is not created")
			}
			// Role
			_, stdErr = runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl get role -l toolkit.fluxcd.io/tenant=%s -n %s`, tenantName, namespace))
			gomega.Expect(stdErr).Should(gomega.BeEmpty(), "Tenant Role '%s' is not created")
			// Rolbinding
			_, stdErr = runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl get rolebinding -l toolkit.fluxcd.io/tenant=%s -n %s`, tenantName, namespace))
			gomega.Expect(stdErr).Should(gomega.BeEmpty(), "Tenant RolBinding '%s' is not created")
			// Policy
			_, stdErr = runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl get policy -l toolkit.fluxcd.io/tenant=%s -n %s`, tenantName, namespace))
			gomega.Expect(stdErr).Should(gomega.BeEmpty(), "Tenant Policy '%s' is not created")
		}
	})
}

func DescribeCliTenant(gitopsTestRunner GitopsTestRunner) {
	var _ = ginkgo.Describe("Gitops CLI Tenants Tests", ginkgo.Ordered, func() {
		var tenantYaml string
		var stdOut string
		var stdErr string

		ginkgo.BeforeEach(ginkgo.OncePerOrdered, func() {
			// Delete the oidc user default roles/rolebindings because the same user is used as a tenant
			_ = runCommandPassThrough("kubectl", "delete", "-f", path.Join(getCheckoutRepoPath(), "test", "utils", "data", "user-role-bindings.yaml"))
		})

		ginkgo.AfterEach(ginkgo.OncePerOrdered, func() {
			// Create the oidc user default roles/rolebindings after tenant tests completed
			_ = runCommandPassThrough("kubectl", "apply", "-f", path.Join(getCheckoutRepoPath(), "test", "utils", "data", "user-role-bindings.yaml"))
		})

		ginkgo.Context("[CLI] When input tenant definition yaml is available", ginkgo.Ordered, func() {

			ginkgo.JustBeforeEach(func() {
				tenantYaml = path.Join("/tmp", "generated-tenant.yaml")
			})

			ginkgo.JustAfterEach(func() {
				deleteTenants([]string{tenantYaml})
			})

			ginkgo.It("Verify a single tenant resources can be exported", ginkgo.Label("tenant"), func() {
				tenatDefination := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "tenancy", "single-tenant.yaml")

				// verify tenants resources are exported to terminal
				stdOut, stdErr = runGitopsCommand(fmt.Sprintf(`create tenants --from-file %s --export`, tenatDefination))
				gomega.Expect(stdErr).Should(gomega.BeEmpty(), "gitops create tenant command failed with an error")

				verifyTenantYaml(string(stdOut), "test-team", []string{"test-system"}, true, "test-team", map[string][]string{"ClusterRole": {"cluster-admin"}}, []string{"weaveworks:QA"})

				// verify tenants resources are exported to output file
				_, stdErr = runGitopsCommand(fmt.Sprintf(`create tenants --from-file %s --export > %s`, tenatDefination, tenantYaml))
				gomega.Expect(stdErr).Should(gomega.BeEmpty(), "gitops create tenant command failed with an error")

				contents, err := ioutil.ReadFile(tenantYaml)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Faile to read generated tentant yaml file")

				verifyTenantYaml(string(contents), "test-team", []string{"test-system"}, true, "test-team", map[string][]string{"ClusterRole": {"cluster-admin"}}, []string{"weaveworks:QA"})
			})

			ginkgo.It("Verify global service account reconciliation by tenant", ginkgo.Label("tenant"), func() {
				tenatDefination := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "tenancy", "single-tenant-sa.yaml")

				_, stdErr = runGitopsCommand(fmt.Sprintf(`create tenants --from-file %s --export > %s`, tenatDefination, tenantYaml))
				gomega.Expect(stdErr).Should(gomega.BeEmpty(), "gitops create tenant command failed with an error")

				contents, err := ioutil.ReadFile(tenantYaml)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Faile to read generated tentant yaml file")

				verifyTenantYaml(string(contents), "dev-team", []string{"dev-system"}, false, "reconcilerServiceAccount", map[string][]string{"ClusterRole": {"cluster-admin"}}, []string{"weaveworks:Pesto", "developers"})
			})

			ginkgo.It("Verify global service account reconciliation by tenant with deployment RBAC", ginkgo.Label("tenant"), func() {
				tenatDefination := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "tenancy", "single-tenant-deployment-sa.yaml")

				_, stdErr = runGitopsCommand(fmt.Sprintf(`create tenants --from-file %s --export > %s`, tenatDefination, tenantYaml))
				gomega.Expect(stdErr).Should(gomega.BeEmpty(), "gitops create tenant command failed with an error")

				contents, err := ioutil.ReadFile(tenantYaml)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Faile to read generated tentant yaml file")

				verifyTenantYaml(string(contents), "test-team", []string{"test-system"}, false, "reconcilerServiceAccount", map[string][]string{"Role": {"foo-role"}}, []string{"wge-test"})
			})

			ginkgo.It("Verify creating tenant resource using kubeconfig ", ginkgo.Label("tenant"), func() {
				_ = gitopsTestRunner.KubectlDelete([]string{}, tenantYaml)
				tenatDefination := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "tenancy", "multiple-tenant.yaml")

				// Export tenants resources to output file (required to delete tenant resources after test completion)
				_, stdErr = runGitopsCommand(fmt.Sprintf(`create tenants --from-file %s --export > %s`, tenatDefination, tenantYaml))
				gomega.Expect(stdErr).Should(gomega.BeEmpty(), "gitops create tenant command failed with an error")

				// Create tenant resource using default kubeconfig
				_, stdErr = runGitopsCommand(fmt.Sprintf(`create tenants --from-file %s`, tenatDefination))
				gomega.Expect(stdErr).Should(gomega.BeEmpty(), "gitops create tenant command failed with an error")

				verifyTenatResources("test-team", []string{"test-system", "test-kustomization"}, true)
				verifyTenatResources("dev-team", []string{"dev-system"}, true)
			})

			ginkgo.It("Verify tenant can only install the application from allowed repositories", ginkgo.Label("tenant"), func() {
				tenantYaml = createTenant(path.Join(getCheckoutRepoPath(), "test", "utils", "data", "tenancy", "multiple-tenant.yaml"))

				// Adding not allowed git repository source
				namespace := "dev-system"
				sourceURL := "https://github.com/stefanprodan/podinfo"
				_, stdErr := runCommandAndReturnStringOutput(fmt.Sprintf("flux create source git my-podinfo --url=%s --branch=master --interval=30s --namespace %s", sourceURL, namespace))
				gomega.Expect(stdErr).Should(gomega.MatchRegexp(`admission webhook "admission.agent.weaveworks" denied the request`), fmt.Sprintf("The restricted git repository source '%s' should not be allowd", sourceURL))

				// Adding not allowed helm repository source
				sourceURL = "https://raw.githubusercontent.com/weaveworks/profiles-catalog/gh-pages"
				_, stdErr = runCommandAndReturnStringOutput(fmt.Sprintf("flux create source helm profiles-catalog --url=%s --interval=30s --namespace %s", sourceURL, namespace))
				gomega.Expect(stdErr).Should(gomega.MatchRegexp(`admission webhook "admission.agent.weaveworks" denied the request`), fmt.Sprintf("The restricted git repository source '%s' should not be allowd", sourceURL))
			})

			ginkgo.It("Verify tenant can only add allowed GitopsCluster to multi-cluster setup", ginkgo.Label("tenant"), func() {
				leafCluster := ClusterConfig{
					Type:      "other",
					Name:      "wge-leaf-tenant-kind",
					Namespace: "dev-system",
				}

				patSecret := "application-pat"
				bootstrapLabel := "bootstrap"
				leafClusterkubeconfig := "wge-leaf-tenant-kind-kubeconfig"

				tenantYaml = createTenant(path.Join(getCheckoutRepoPath(), "test", "utils", "data", "tenancy", "multiple-tenant.yaml"))
				createPATSecret(leafCluster.Namespace, patSecret)
				defer deleteSecret([]string{patSecret}, leafCluster.Namespace)
				clusterBootstrapCopnfig := createClusterBootstrapConfig(leafCluster.Name, leafCluster.Namespace, bootstrapLabel, patSecret)
				defer func() {
					_ = gitopsTestRunner.KubectlDelete([]string{}, clusterBootstrapCopnfig)
				}()

				ginkgo.By(fmt.Sprintf("Add GitopsCluster resource for %s cluster to management cluster", leafCluster.Name), func() {
					contents, err := ioutil.ReadFile(path.Join(getCheckoutRepoPath(), "test/utils/data/gitops-cluster.yaml"))
					gomega.Expect(err).To(gomega.BeNil(), "Failed to read GitopsCluster template yaml")

					t := template.Must(template.New("gitops-cluster").Parse(string(contents)))

					// Prepare  data to insert into the template.
					type TemplateInput struct {
						ClusterName      string
						NameSpace        string
						Bootstrap        string
						KubeconfigSecret string
					}
					input := TemplateInput{leafCluster.Name, leafCluster.Namespace, bootstrapLabel, leafClusterkubeconfig}

					gitopsCluster := path.Join("/tmp", leafCluster.Name+"-gitops-cluster.yaml")

					f, err := os.Create(gitopsCluster)
					gomega.Expect(err).To(gomega.BeNil(), "Failed to create GitopsCluster manifest yaml")

					err = t.Execute(f, input)
					f.Close()
					gomega.Expect(err).To(gomega.BeNil(), "Failed to generate GitopsCluster manifest yaml")

					_, stdErr := runCommandAndReturnStringOutput("kubectl apply -f " + gitopsCluster)
					gomega.Expect(stdErr).Should(gomega.MatchRegexp(fmt.Sprintf(`cluster secretRef %s is not allowed for namespace dev-system`, leafClusterkubeconfig)), fmt.Sprintf("Failed to create GitopsCluster resource for  cluster: %s", leafCluster.Name))
				})
			})
		})
	})
}
