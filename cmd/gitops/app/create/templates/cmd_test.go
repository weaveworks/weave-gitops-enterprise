package templates

import (
	"encoding/base64"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	templatesv1 "github.com/weaveworks/templates-controller/apis/core"
	gapiv1 "github.com/weaveworks/templates-controller/apis/gitops/v1alpha2"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"google.golang.org/protobuf/testing/protocmp"
	"helm.sh/helm/v3/pkg/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var defaultParams map[string]string = map[string]string{
	"CLUSTER_NAME":       "test-cluster",
	"RESOURCE_NAME":      "test-resource",
	"NAMESPACE":          "test-namespace",
	"GIT_REPO_NAMESPACE": "test-git-repo-namespace",
	"GIT_REPO_NAME":      "test-git-repo-name",
	"PATH":               "clusters/out.yaml",
}

var testSettings *cli.EnvSettings = &cli.EnvSettings{
	RepositoryConfig: "testdata/repositories.yaml",
	RepositoryCache:  "testdata/repository",
}

func Test_parseTemplate(t *testing.T) {
	type args struct {
		templateFile string
	}

	tests := []struct {
		name     string
		args     args
		expected *gapiv1.GitOpsTemplate
		err      error
	}{
		{
			name: "valid template",
			args: args{
				templateFile: "testdata/template.yaml",
			},
			expected: &gapiv1.GitOpsTemplate{
				TypeMeta:   metav1.TypeMeta{Kind: "GitOpsTemplate", APIVersion: "templates.weave.works/v1alpha2"},
				ObjectMeta: metav1.ObjectMeta{Name: "test-template", Namespace: "default"},
				Spec: templatesv1.TemplateSpec{
					Description: "This is a sample WGE template to test parsing functionality.",
					Params: []templatesv1.TemplateParam{
						{Name: "CLUSTER_NAME", Description: "Name of the cluster.", Required: false},
						{Name: "RESOURCE_NAME", Description: "Name of the template.", Required: false},
						{Name: "NAMESPACE", Description: "Namespace to create the resource in.", Required: false},
						{Name: "GIT_REPO_NAMESPACE", Description: "Namespace of the configuring git repository object.", Required: false},
						{Name: "GIT_REPO_NAME", Description: "Name of the configuring git repository.", Required: false},
						{Name: "PATH", Description: "Path to the generated resource.", Required: false},
					},
					ResourceTemplates: []templatesv1.ResourceTemplate{
						{
							Path: "${PATH}",
							Content: []templatesv1.ResourceTemplateContent{
								{
									RawExtension: runtime.RawExtension{
										Raw: []byte(`{"apiVersion":"kustomize.toolkit.fluxcd.io/v1beta2","kind":"Kustomization","metadata":{"name":"${RESOURCE_NAME}","namespace":"${NAMESPACE}"},"spec":{"interval":"1h","path":"${TEMPLATE_PATH}","sourceRef":{"kind":"GitRepository","name":"${GIT_REPO_NAME}","namespace":"${GIT_REPO_NAMESPACE}"}}}`),
									},
								},
							},
						},
					},
				},
			},
			err: nil,
		},
		{
			name: "invalid template",
			args: args{
				templateFile: "testdata/invalid-template.yaml",
			},
			expected: nil,
			err:      errors.New("failed to read template file testdata/invalid-template.yaml: open testdata/invalid-template.yaml: no such file or directory"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTemplate(tt.args.templateFile)
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to parse template:\n%v", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("returned error didn't match expected:\n%v", diff)
				}
			}
			if diff := cmp.Diff(tt.expected, result, protocmp.Transform()); diff != "" {
				t.Fatalf("result didn't match expected:\n%s", diff)
			}
		})
	}
}

func Test_CreateCommand(t *testing.T) {
	cmd := CreateCommand
	cmd.SetArgs([]string{})
	cmd.SetOut(io.Discard)
	err := cmd.Execute()
	assert.ErrorContains(t, err, "must specify template file")
}

func Test_initializeConfig(t *testing.T) {
	cmd := &cobra.Command{}
	configPath = "testdata/config.yaml"
	config = Config{}

	err := initializeConfig(cmd)
	if err != nil {
		t.Errorf("Error initializing config: %v", err)
	}

	assert.Equal(t, config.TemplateFile, "template.yaml")

	expectedParams := []string{
		"CLUSTER_NAME=test-cluster",
		"RESOURCE_NAME=test-resource",
		"NAMESPACE=test-namespace",
		"GIT_REPO_NAMESPACE=test-git-repo-namespace",
		"GIT_REPO_NAME=test-git-repo-name",
		"PATH=clusters/out.yaml",
	}

	if diff := cmp.Diff(expectedParams, config.ParameterValues); diff != "" {
		t.Fatalf("result didn't match expected:\n%s", diff)
	}
}

func TestGenerateFilesLocally(t *testing.T) {
	tmpl, err := parseTemplate("testdata/template.yaml")
	assert.NoError(t, err)

	// don't have to specify any helm settings if no profiles are around
	files, err := generateFilesLocally(tmpl, defaultParams, "test-repo", nil, nil, logr.Discard())
	assert.NoError(t, err)

	expectedFiles := []string{
		"clusters/out.yaml",
	}

	actualFilenames := []string{}
	for _, file := range files {
		actualFilenames = append(actualFilenames, file.Path)
	}

	if diff := cmp.Diff(expectedFiles, actualFilenames); diff != "" {
		t.Fatalf("result didn't match expected:\n%s", diff)
	}
}

func TestGenerateFilesLocallyWithCharts(t *testing.T) {
	tmpl, err := parseTemplate("testdata/template-with-charts.yaml")
	assert.NoError(t, err)

	profiles := []*capiv1_proto.ProfileValues{
		{
			Name:    "test-profile",
			Values:  base64.StdEncoding.EncodeToString([]byte("foo: bar")),
			Version: "0.0.7",
			HelmRepository: &capiv1_proto.HelmRepositoryRef{
				Name:      "test-repo",
				Namespace: "default",
			},
		},
	}

	files, err := generateFilesLocally(tmpl, defaultParams, "test-repo", profiles, testSettings, logr.Discard())
	require.NoError(t, err)

	expectedFiles := []string{
		"clusters/out.yaml",
		"test-namespace/test-resource/profiles.yaml",
	}

	actualFilenames := []string{}
	for _, file := range files {
		actualFilenames = append(actualFilenames, file.Path)
	}

	assert.Contains(t, *files[1].Content, "version: 0.0.8")
	assert.Contains(t, *files[1].Content, "test-repo")
	assert.Contains(t, *files[1].Content, "foo: bar")

	if diff := cmp.Diff(expectedFiles, actualFilenames); diff != "" {
		t.Fatalf("result didn't match expected:\n%s", diff)
	}
}

func TestRunWithProfiles(t *testing.T) {
	// make a temp dir to store the output
	tmpDir := t.TempDir()

	// Set and restore some env
	keyCache := "HELM_REPOSITORY_CACHE"
	keyConfig := "HELM_REPOSITORY_CONFIG"
	// Set up the test paths for helm repo and cache
	t.Setenv(keyCache, testSettings.RepositoryCache)
	t.Setenv(keyConfig, testSettings.RepositoryConfig)

	cmd := CreateCommand
	cmd.SetArgs([]string{
		"--config", "testdata/config-with-values.yaml",
		"--output-dir", tmpDir,
	})
	cmd.SetOut(io.Discard)
	err := cmd.Execute()
	assert.NoError(t, err)

	expectedFiles := []string{
		"clusters/out.yaml",
		"test-namespace/test-resource/profiles.yaml",
	}

	actualFilenames := []string{}
	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			actualFilenames = append(actualFilenames, strings.TrimPrefix(path, tmpDir+"/"))
		}
		return nil
	})
	assert.NoError(t, err)

	if diff := cmp.Diff(expectedFiles, actualFilenames); diff != "" {
		t.Fatalf("result didn't match expected:\n%s", diff)
	}

	// check that the profiles.yaml file contains the expected values
	profilesFile, err := os.ReadFile(filepath.Join(tmpDir, "test-namespace/test-resource/profiles.yaml"))
	assert.NoError(t, err)
	assert.Contains(t, string(profilesFile), "version: '>0.1'")
	assert.Contains(t, string(profilesFile), "cert-manager")
	assert.Contains(t, string(profilesFile), "foo: bar")
}
