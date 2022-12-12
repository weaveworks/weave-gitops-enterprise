package acceptance

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"text/template"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/capi/v1alpha1"
)

// Interface that can be implemented either with:
// - "Real" commands like "exec(kubectl...)"
// - "Mock" commands like db.Create(cluster_info...)
type GitopsTestRunner interface {
	ResetControllers(controllers string)
	VerifyWegoPodsRunning()
	FireAlert(name, severity, message string, fireFor time.Duration) error
	KubectlApply(env []string, manifest string) error
	KubectlApplyInsecure(env []string, manifest string) error
	KubectlDelete(env []string, manifest string) error
	KubectlDeleteInsecure(env []string, manifest string) error
	TimeTravelToLastSeen() error
	TimeTravelToAlertsResolved() error
	CreateApplyCapitemplates(templateCount int, templateFile string) []string
	DeleteApplyCapiTemplates(templateFiles []string)
	CreateIPCredentials(infrastructureProvider string)
	DeleteIPCredentials(infrastructureProvider string)
	RestartDeploymentPods(appName string, namespace string) error
}

type DatabaseGitopsTestRunner struct {
	Client goclient.Client
}

func (b DatabaseGitopsTestRunner) TimeTravelToLastSeen() error {
	return nil
}

func (b DatabaseGitopsTestRunner) TimeTravelToAlertsResolved() error {
	return nil
}

func (b DatabaseGitopsTestRunner) ResetControllers(controllers string) {

}

func (b DatabaseGitopsTestRunner) VerifyWegoPodsRunning() {

}

func (b DatabaseGitopsTestRunner) KubectlApply(env []string, manifest string) error {
	return nil
}

func (b DatabaseGitopsTestRunner) KubectlApplyInsecure(env []string, manifest string) error {
	return b.KubectlApply(env, manifest)
}

func (b DatabaseGitopsTestRunner) KubectlDelete(env []string, tokenURL string) error {
	//
	// No more cluster_infos will be created anyway..
	// FIXME: maybe we add a polling loop that keeps creating cluster_info while its connected
	//
	return nil
}

func (b DatabaseGitopsTestRunner) KubectlDeleteInsecure(env []string, tokenURL string) error {
	return b.KubectlDelete(env, tokenURL)
}

func (b DatabaseGitopsTestRunner) FireAlert(name, severity, message string, fireFor time.Duration) error {
	return nil
}

func (b DatabaseGitopsTestRunner) CreateApplyCapitemplates(templateCount int, templateFile string) []string {
	templateFiles, err := generateTestCapiTemplates(templateCount, templateFile)
	gomega.Expect(err).To(gomega.BeNil(), "Failed to generate CAPITemplate template test files by database test runner")
	ginkgo.By("Apply/Install CAPITemplate templates", func() {
		for _, fileName := range templateFiles {
			capiTemplate, err := parseCAPITemplateFromFile(fileName)
			gomega.Expect(err).To(gomega.BeNil(), "Failed to parse CAPITemplate template files")
			err = b.Client.Create(context.Background(), capiTemplate)
			gomega.Expect(err).To(gomega.BeNil(), "Failed to create CAPITemplate template files")
		}
	})

	return templateFiles
}

func (b DatabaseGitopsTestRunner) DeleteApplyCapiTemplates(templateFiles []string) {
	ginkgo.By("Delete CAPITemplate templates", func() {
		for _, fileName := range templateFiles {
			capiTemplate, err := parseCAPITemplateFromFile(fileName)
			gomega.Expect(err).To(gomega.BeNil(), "Failed to parse CAPITemplate template files")
			err = b.Client.Delete(context.Background(), capiTemplate)
			gomega.Expect(err).To(gomega.BeNil(), "Failed to delete CAPITemplate template files")
		}
	})
}

func (b DatabaseGitopsTestRunner) RestartDeploymentPods(appName string, namespace string) error {
	return nil
}

func (b DatabaseGitopsTestRunner) CreateIPCredentials(infrastructureProvider string) {

}

func (b DatabaseGitopsTestRunner) DeleteIPCredentials(infrastructureProvider string) {

}

// "Real" backend that call kubectl and posts to alertmanagement

type RealGitopsTestRunner struct{}

func (b RealGitopsTestRunner) TimeTravelToLastSeen() error {
	return nil
}

func (b RealGitopsTestRunner) TimeTravelToAlertsResolved() error {
	return nil
}

func (b RealGitopsTestRunner) ResetControllers(controllers string) {
	scriptPath := path.Join(testScriptsPath, "wego-enterprise.sh")
	_ = runCommandPassThrough(scriptPath, "reset_controllers", controllers)
}

func (b RealGitopsTestRunner) VerifyWegoPodsRunning() {
	verifyEnterpriseControllers("my-mccp", "", GITOPS_DEFAULT_NAMESPACE)
	CheckClusterService(wge_endpoint_url)
}

func (b RealGitopsTestRunner) KubectlApply(env []string, url string) error {
	return runCommandPassThroughWithEnv(env, "kubectl", "apply", "-f", url)
}

func (b RealGitopsTestRunner) KubectlApplyInsecure(env []string, url string) error {
	err := runCommandPassThrough("curl", "--insecure", "-o", "/tmp/manifest.yaml", url)
	if err != nil {
		return fmt.Errorf("failed to curl manifest: %w", err)
	}
	return b.KubectlApply(env, "/tmp/manifest.yaml")
}

func (b RealGitopsTestRunner) KubectlDelete(env []string, url string) error {
	return runCommandPassThroughWithEnv(env, "kubectl", "delete", "-f", url)
}

func (b RealGitopsTestRunner) KubectlDeleteInsecure(env []string, url string) error {
	err := runCommandPassThrough("curl", "--insecure", "-o", "/tmp/manifest.yaml", url)
	if err != nil {
		return fmt.Errorf("failed to curl manifest: %w", err)
	}
	return b.KubectlDelete(env, "/tmp/manifest.yaml")
}

func (b RealGitopsTestRunner) FireAlert(name, severity, message string, fireFor time.Duration) error {
	const alertTemplate = `
    [
      {
        "labels": {
          "alertname": "{{ .Name }}",
          "severity": "{{ .Severity }}"
        },
        "annotations": {
          "message": "{{ .Message }}"
        },
        "startsAt": "{{ .StartsAt }}",
        "endsAt": "{{ .EndsAt }}"
      }
    ]
    `

	t, err := template.New("alert").Parse(alertTemplate)
	if err != nil {
		return err
	}
	var populated bytes.Buffer
	err = t.Execute(&populated, struct {
		Name     string
		Severity string
		Message  string
		StartsAt string
		EndsAt   string
	}{
		name,
		severity,
		message,
		time.Now().UTC().Add(fireFor * -1).Format(time.RFC3339),
		time.Now().UTC().Add(fireFor).Format(time.RFC3339),
	})

	if err != nil {
		return err
	}

	fmt.Print(populated.String())
	req, err := http.NewRequest("POST", test_ui_url+"/alertmanager/api/v2/alerts", &populated)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("alertmanager didn't like the alert: %v", resp.StatusCode)
	}

	return nil
}

// This function will crete the test capiTemplate files and do the kubectl apply for capiserver availability
func (b RealGitopsTestRunner) CreateApplyCapitemplates(templateCount int, templateFile string) []string {
	templateFiles, err := generateTestCapiTemplates(templateCount, templateFile)
	gomega.Expect(err).To(gomega.BeNil(), "Failed to generate CAPITemplate template test files by real test runner")

	ginkgo.By("Apply/Install CAPITemplate templates", func() {
		for _, fileName := range templateFiles {
			err = runCommandPassThrough("kubectl", "apply", "-f", fileName)
			gomega.Expect(err).To(gomega.BeNil(), "Failed to apply/install CAPITemplate template files")
		}
	})

	return templateFiles
}

// This function deletes the test capiTemplate files and do the kubectl delete to clean the cluster
func (b RealGitopsTestRunner) DeleteApplyCapiTemplates(templateFiles []string) {
	ginkgo.By("Delete CAPITemplate templates", func() {

		for _, fileName := range templateFiles {
			err := b.KubectlDelete([]string{}, fileName)
			gomega.Expect(err).To(gomega.BeNil(), "Failed to delete CAPITemplate template")
		}
	})

	err := deleteFile(templateFiles)
	gomega.Expect(err).To(gomega.BeNil(), "Failed to delete CAPITemplate template test files")
}

func (b RealGitopsTestRunner) RestartDeploymentPods(appName string, namespace string) error {
	// Restart the deployment pods
	var err error
	for i := 1; i < 5; i++ {
		time.Sleep(POLL_INTERVAL_1SECONDS)
		err = runCommandPassThrough("kubectl", "rollout", "restart", "deployment", appName, "-n", namespace)
		if err == nil {
			// Wait for all the deployments replicas to rolled out successfully
			err = runCommandPassThrough("kubectl", "rollout", "status", "deployment", appName, "-n", namespace)
			break
		}
	}

	return err
}

func (b RealGitopsTestRunner) CreateIPCredentials(infrastructureProvider string) {
	if infrastructureProvider == "AWS" {
		// CAPA installs the AWS identity crds
		if capi_provider != "capa" {
			ginkgo.By("Install AWSClusterStaticIdentity CRD", func() {
				_, _ = runCommandAndReturnStringOutput(fmt.Sprintf("kubectl apply -f %s/capi-multi-tenancy/infrastructure.cluster.x-k8s.io_awsclusterstaticidentities.yaml", testDataPath))
				_, _ = runCommandAndReturnStringOutput("kubectl wait --for=condition=established --timeout=90s crd/awsclusterstaticidentities.infrastructure.cluster.x-k8s.io", ASSERTION_2MINUTE_TIME_OUT)
			})

			ginkgo.By("Install AWSClusterRoleIdentity CRD", func() {
				_, _ = runCommandAndReturnStringOutput(fmt.Sprintf("kubectl apply -f %s/capi-multi-tenancy/infrastructure.cluster.x-k8s.io_awsclusterroleidentities.yaml", testDataPath))
				_, _ = runCommandAndReturnStringOutput("kubectl wait --for=condition=established --timeout=90s crd/awsclusterroleidentities.infrastructure.cluster.x-k8s.io", ASSERTION_2MINUTE_TIME_OUT)
			})
		}

		ginkgo.By("Create AWS Secret, AWSClusterStaticIdentity and AWSClusterRoleIdentity)", func() {
			_, _ = runCommandAndReturnStringOutput("kubectl create namespace capa-system")
			_, _ = runCommandAndReturnStringOutput(fmt.Sprintf("kubectl apply -f %s/capi-multi-tenancy/aws_cluster_credentials.yaml", testDataPath), ASSERTION_30SECONDS_TIME_OUT)
		})

	} else if infrastructureProvider == "AZURE" {
		ginkgo.By("Install AzureClusterIdentity CRD", func() {
			_, _ = runCommandAndReturnStringOutput(fmt.Sprintf("kubectl apply -f %s/capi-multi-tenancy/infrastructure.cluster.x-k8s.io_azureclusteridentities.yaml", testDataPath))
			_, _ = runCommandAndReturnStringOutput("kubectl wait --for=condition=established --timeout=90s crd/azureclusteridentities.infrastructure.cluster.x-k8s.io", ASSERTION_2MINUTE_TIME_OUT)
		})

		ginkgo.By("Create Azure Secret and AzureClusterIdentity)", func() {
			_, _ = runCommandAndReturnStringOutput(fmt.Sprintf("kubectl apply -f %s/capi-multi-tenancy/azure_cluster_credentials.yaml", testDataPath), ASSERTION_30SECONDS_TIME_OUT)
		})
	}

}

func (b RealGitopsTestRunner) DeleteIPCredentials(infrastructureProvider string) {
	if infrastructureProvider == "AWS" {
		ginkgo.By("Delete AWS identities and CRD", func() {
			// Identity crds are installed as part of CAPA installation
			_ = b.KubectlDelete([]string{}, fmt.Sprintf("%s/capi-multi-tenancy/aws_cluster_credentials.yaml", testDataPath))
			if capi_provider != "capa" {
				_ = b.KubectlDelete([]string{}, fmt.Sprintf("%s/capi-multi-tenancy/infrastructure.cluster.x-k8s.io_awsclusterroleidentities.yaml", testDataPath))
				_ = b.KubectlDelete([]string{}, fmt.Sprintf("%s/capi-multi-tenancy/infrastructure.cluster.x-k8s.io_awsclusterstaticidentities.yaml", testDataPath))
				_, _ = runCommandAndReturnStringOutput("kubectl delete namespace capa-system")
			}
		})

	} else if infrastructureProvider == "AZURE" {
		ginkgo.By("Delete Azure identities and CRD", func() {
			_ = b.KubectlDelete([]string{}, fmt.Sprintf("%s/capi-multi-tenancy/azure_cluster_credentials.yaml", testDataPath))
			_ = b.KubectlDelete([]string{}, fmt.Sprintf("%s/capi-multi-tenancy/infrastructure.cluster.x-k8s.io_azureclusteridentities.yaml", testDataPath))
		})
	}
}

// This function generates multiple capitemplate files from a single capitemplate to be used as test data
func generateTestCapiTemplates(templateCount int, templateFile string) (templateFiles []string, err error) {
	// Read input capitemplate
	contents, err := ioutil.ReadFile(path.Join(testDataPath, templateFile))

	if err != nil {
		return templateFiles, err
	}

	// Prepare  data to insert into the template.
	type TemplateInput struct {
		Count int
	}

	// Create a new template and parse the letter into it.
	t := template.Must(template.New("capi-template").Parse(string(contents)))

	// Execute the template for each count.
	for i := 0; i < templateCount; i++ {
		input := TemplateInput{i}

		fileName := fmt.Sprintf("%s%d", filepath.Base(templateFile), i)

		f, err := os.Create(path.Join("/tmp", fileName))
		if err != nil {
			return templateFiles, err
		}
		templateFiles = append(templateFiles, f.Name())

		if err = t.Execute(f, input); err != nil {
			logger.Infoln("Executing template:", err)
		}

		f.Close()
	}

	return templateFiles, nil
}

func parseCAPITemplateFromBytes(b []byte) (*capiv1.CAPITemplate, error) {
	var c capiv1.CAPITemplate
	err := yaml.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil

}

func parseCAPITemplateFromFile(filename string) (*capiv1.CAPITemplate, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parseCAPITemplateFromBytes(b)
}
