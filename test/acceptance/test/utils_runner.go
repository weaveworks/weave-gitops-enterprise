package acceptance

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"text/template"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/capi"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/types"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Interface that can be implemented either with:
// - "Real" commands like "exec(kubectl...)"
// - "Mock" commands like db.Create(cluster_info...)
type GitopsTestRunner interface {
	ResetControllers(controllers string)
	VerifyWegoPodsRunning()
	FireAlert(name, severity, message string, fireFor time.Duration) error
	KubectlApply(env []string, manifest string) error
	KubectlDelete(env []string, manifest string) error
	KubectlDeleteAllAgents(env []string) error
	TimeTravelToLastSeen() error
	TimeTravelToAlertsResolved() error
	AddWorkspace(env []string, clusterName string) error
	CreateApplyCapitemplates(templateCount int, templateFile string) []string
	DeleteApplyCapiTemplates(templateFiles []string)
	CreateIPCredentials(infrastructureProvider string)
	DeleteIPCredentials(infrastructureProvider string)
	CheckClusterService(capiEndpointURL string)
	RestartDeploymentPods(env []string, appName string, namespace string) error
}

// "DB" backend that creates/delete rows

type DatabaseGitopsTestRunner struct {
	DB     *gorm.DB
	Client goclient.Client
}

func (b DatabaseGitopsTestRunner) TimeTravelToLastSeen() error {
	oneMinuteAgo := time.Now().UTC().Add(time.Minute * -2)
	b.DB.Exec("update cluster_info set updated_at = ?", oneMinuteAgo)
	return nil
}

func (b DatabaseGitopsTestRunner) TimeTravelToAlertsResolved() error {
	b.DB.Where("1 = 1").Delete(&models.Alert{})
	return nil
}

func (b DatabaseGitopsTestRunner) ResetControllers(controllers string) {
	b.DB.Where("1 = 1").Delete(&models.Cluster{})
}

func (b DatabaseGitopsTestRunner) VerifyWegoPodsRunning() {

}

func (b DatabaseGitopsTestRunner) KubectlApply(env []string, manifest string) error {
	u, err := url.Parse(manifest)
	if err != nil {
		return err
	}
	token := u.Query()["token"][0]

	b.DB.Create(&models.ClusterInfo{
		UID:          types.UID(RandString(10)),
		ClusterToken: token,
		UpdatedAt:    time.Now().UTC(),
	})
	b.DB.Create(&models.GitCommit{
		ClusterToken: token,
		Sha:          "abcdef123456",
		AuthorName:   "Alice",
		AuthorEmail:  "alice@acme.org",
		AuthorDate:   time.Now().UTC().Add(time.Hour * -1),
		Message:      "Fixed it",
	})
	b.DB.Create(&models.FluxInfo{
		ClusterToken: token,
		Name:         "flux",
		Namespace:    "wkp-flux",
		RepoURL:      "git@github.com:wkp/my-cluster",
		RepoBranch:   "main",
	})
	return nil
}

func (b DatabaseGitopsTestRunner) KubectlDelete(env []string, tokenURL string) error {
	//
	// No more cluster_infos will be created anyway..
	// FIXME: maybe we add a polling loop that keeps creating cluster_info while its connected
	//
	return nil
}

func (b DatabaseGitopsTestRunner) KubectlDeleteAllAgents(env []string) error {
	// No more cluster_infos will be created anyway..
	return nil
}

func (b DatabaseGitopsTestRunner) FireAlert(name, severity, message string, fireFor time.Duration) error {
	var firstCluster models.Cluster
	b.DB.Last(&firstCluster)

	//
	// FIXME: we shouldn't need this. The UI should stop showing the alerts after 30s anyway
	// But its not filtering on endsAt right now.
	//
	go func() {
		time.Sleep(fireFor)
		b.DB.Where("1 = 1").Delete(&models.Alert{})
	}()

	labels := fmt.Sprintf(`{ "alertname": "%s", "severity": "%s" }`, name, severity)
	annotations := fmt.Sprintf(`{ "message": "%s" }`, message)
	b.DB.Create(&models.Alert{
		ClusterToken: firstCluster.Token,
		UpdatedAt:    time.Now().UTC(),
		Labels:       datatypes.JSON(labels),
		Annotations:  datatypes.JSON(annotations),
		Severity:     severity,
		StartsAt:     time.Now().UTC().Add(fireFor * -1),
		EndsAt:       time.Now().UTC().Add(fireFor),
	})

	return nil
}

func (b DatabaseGitopsTestRunner) AddWorkspace(env []string, clusterName string) error {
	var firstCluster models.Cluster
	b.DB.Where("Name = ?", clusterName).First(&firstCluster)

	b.DB.Create(&models.Workspace{
		ClusterToken: firstCluster.Token,
		Name:         "mccp-devs-workspace",
		Namespace:    "wkp-workspace",
	})

	return nil
}

func (b DatabaseGitopsTestRunner) CreateApplyCapitemplates(templateCount int, templateFile string) []string {
	templateFiles, err := generateTestCapiTemplates(templateCount, templateFile)
	Expect(err).To(BeNil(), "Failed to generate CAPITemplate template test files")
	By("Apply/Install CAPITemplate templates", func() {
		for _, fileName := range templateFiles {
			template, err := capi.ParseFile(fileName)
			Expect(err).To(BeNil(), "Failed to parse CAPITemplate template files")
			err = b.Client.Create(context.Background(), template)
			Expect(err).To(BeNil(), "Failed to create CAPITemplate template files")
		}
	})

	return templateFiles
}

func (b DatabaseGitopsTestRunner) DeleteApplyCapiTemplates(templateFiles []string) {
	By("Delete CAPITemplate templates", func() {
		for _, fileName := range templateFiles {
			template, err := capi.ParseFile(fileName)
			Expect(err).To(BeNil(), "Failed to parse CAPITemplate template files")
			err = b.Client.Delete(context.Background(), template)
			Expect(err).To(BeNil(), "Failed to delete CAPITemplate template files")
		}
	})
}

func (b DatabaseGitopsTestRunner) CheckClusterService(capiEndpointURL string) {

}

func (b DatabaseGitopsTestRunner) RestartDeploymentPods(env []string, appName string, namespace string) error {
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
	scriptPath := path.Join(getCheckoutRepoPath(), "test", "utils", "scripts", "wego-enterprise.sh")
	_ = runCommandPassThrough(scriptPath, "reset_controllers", controllers)
}

func (b RealGitopsTestRunner) VerifyWegoPodsRunning() {
	verifyEnterpriseControllers("my-mccp", "", GITOPS_DEFAULT_NAMESPACE)
}

func (b RealGitopsTestRunner) KubectlApply(env []string, manifest string) error {
	return runCommandPassThroughWithEnv(env, "kubectl", "apply", "-f", manifest)
}

func (b RealGitopsTestRunner) KubectlDelete(env []string, manifest string) error {
	return runCommandPassThroughWithEnv(env, "kubectl", "delete", "-f", manifest)
}

func (b RealGitopsTestRunner) KubectlDeleteAllAgents(env []string) error {
	return runCommandPassThroughWithEnv(env, "kubectl", "delete", "-n", "wkp-agent", "deploy", "wkp-agent")
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
	client := &http.Client{}
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

func (b RealGitopsTestRunner) AddWorkspace(env []string, clusterName string) error {
	return runCommandPassThroughWithEnv(env, "kubectl", "apply", "-f", "../../utils/data/mccp-workspace.yaml")
}

// This function will crete the test capiTemplate files and do the kubectl apply for capiserver availability
func (b RealGitopsTestRunner) CreateApplyCapitemplates(templateCount int, templateFile string) []string {
	templateFiles, err := generateTestCapiTemplates(templateCount, templateFile)
	Expect(err).To(BeNil(), "Failed to generate CAPITemplate template test files")

	By("Apply/Install CAPITemplate templates", func() {
		for _, fileName := range templateFiles {
			err = runCommandPassThrough("kubectl", "apply", "-f", fileName)
			Expect(err).To(BeNil(), "Failed to apply/install CAPITemplate template files")
		}
	})

	return templateFiles
}

// This function deletes the test capiTemplate files and do the kubectl delete to clean the cluster
func (b RealGitopsTestRunner) DeleteApplyCapiTemplates(templateFiles []string) {
	By("Delete CAPITemplate templates", func() {

		for _, fileName := range templateFiles {
			err := b.KubectlDelete([]string{}, fileName)
			Expect(err).To(BeNil(), "Failed to delete CAPITemplate template")
		}
	})

	err := deleteFile(templateFiles)
	Expect(err).To(BeNil(), "Failed to delete CAPITemplate template test files")
}

func (b RealGitopsTestRunner) CheckClusterService(capiEndpointURL string) {
	output := func() string {
		stdOut, stdErr := runCommandAndReturnStringOutput(fmt.Sprintf(`curl -s -o /dev/null -v -w %%{http_code} %s/v1/templates`, capiEndpointURL), ASSERTION_30SECONDS_TIME_OUT)
		if stdErr != "" {
			logger.Warnf("Clusters service not ready yet: %q", stdErr)
		}
		return stdOut
	}
	Eventually(output, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp("200"), "Cluster Service is not healthy")
}

func (b RealGitopsTestRunner) RestartDeploymentPods(env []string, appName string, namespace string) error {
	// Restart the deployment pods
	err := runCommandPassThroughWithEnv(env, "kubectl", "rollout", "restart", "deployment", appName, "-n", namespace)
	if err == nil {
		// Wait for all the deployments replicas to rolled out successfully
		err = runCommandPassThroughWithEnv(env, "kubectl", "rollout", "status", "deployment", appName, "-n", namespace)
	}
	return err
}

func (b RealGitopsTestRunner) CreateIPCredentials(infrastructureProvider string) {
	testDataPath := path.Join(getCheckoutRepoPath(), "test", "utils", "data")
	if infrastructureProvider == "AWS" {
		By("Install AWSClusterStaticIdentity CRD", func() {
			_, _ = runCommandAndReturnStringOutput(fmt.Sprintf("kubectl apply -f %s/infrastructure.cluster.x-k8s.io_awsclusterstaticidentities.yaml", testDataPath))
			_, _ = runCommandAndReturnStringOutput("kubectl wait --for=condition=established --timeout=90s crd/awsclusterstaticidentities.infrastructure.cluster.x-k8s.io")
		})

		By("Install AWSClusterRoleIdentity CRD", func() {
			_, _ = runCommandAndReturnStringOutput(fmt.Sprintf("kubectl apply -f %s/infrastructure.cluster.x-k8s.io_awsclusterroleidentities.yaml", testDataPath))
			_, _ = runCommandAndReturnStringOutput("kubectl wait --for=condition=established --timeout=90s crd/awsclusterroleidentities.infrastructure.cluster.x-k8s.io")
		})

		By("Create AWS Secret, AWSClusterStaticIdentity and AWSClusterRoleIdentity)", func() {
			_, _ = runCommandAndReturnStringOutput(fmt.Sprintf("kubectl apply -f %s/aws_cluster_credentials.yaml", testDataPath))
		})

	} else if infrastructureProvider == "AZURE" {
		By("Install AzureClusterIdentity CRD", func() {
			_, _ = runCommandAndReturnStringOutput(fmt.Sprintf("kubectl apply -f %s/infrastructure.cluster.x-k8s.io_azureclusteridentities.yaml", testDataPath))
			_, _ = runCommandAndReturnStringOutput("kubectl wait --for=condition=established --timeout=90s crd/azureclusteridentities.infrastructure.cluster.x-k8s.io")
		})

		By("Create Azure Secret and AzureClusterIdentity)", func() {
			_, _ = runCommandAndReturnStringOutput(fmt.Sprintf("kubectl apply -f %s/azure_cluster_credentials.yaml", testDataPath))
		})
	}

}

func (b RealGitopsTestRunner) DeleteIPCredentials(infrastructureProvider string) {
	testDataPath := path.Join(getCheckoutRepoPath(), "test", "utils", "data")
	if infrastructureProvider == "AWS" {
		By("Delete AWS identities and CRD", func() {
			_ = b.KubectlDelete([]string{}, fmt.Sprintf("%s/aws_cluster_credentials.yaml", testDataPath))
			_ = b.KubectlDelete([]string{}, fmt.Sprintf("%s/infrastructure.cluster.x-k8s.io_awsclusterroleidentities.yaml", testDataPath))
			_ = b.KubectlDelete([]string{}, fmt.Sprintf("%s/infrastructure.cluster.x-k8s.io_awsclusterstaticidentities.yaml", testDataPath))
		})

	} else if infrastructureProvider == "AZURE" {
		By("Delete Azure identities and CRD", func() {
			_ = b.KubectlDelete([]string{}, fmt.Sprintf("%s/azure_cluster_credentials.yaml", testDataPath))
			_ = b.KubectlDelete([]string{}, fmt.Sprintf("%s/infrastructure.cluster.x-k8s.io_azureclusteridentities.yaml", testDataPath))
		})
	}
}

// This function generates multiple capitemplate files from a single capitemplate to be used as test data
func generateTestCapiTemplates(templateCount int, templateFile string) (templateFiles []string, err error) {
	// Read input capitemplate
	contents, err := ioutil.ReadFile(path.Join(getCheckoutRepoPath(), "test", "utils", "data", templateFile))

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

		fileName := fmt.Sprintf("%s%d", templateFile, i)

		f, err := os.Create(path.Join("/tmp", fileName))
		if err != nil {
			return templateFiles, err
		}
		templateFiles = append(templateFiles, f.Name())

		err = t.Execute(f, input)
		if err != nil {
			logger.Infoln("Executing template:", err)
		}

		f.Close()
	}

	return templateFiles, nil
}
