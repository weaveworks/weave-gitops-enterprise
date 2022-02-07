package server

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fluxcd/go-git-providers/gitprovider"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/charts"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/models"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
)

func TestCreatePullRequest(t *testing.T) {
	testCases := []struct {
		name           string
		clusterState   []runtime.Object
		provider       git.Provider
		pruneEnvVar    string
		req            *capiv1_protos.CreatePullRequestRequest
		expected       string
		committedFiles []CommittedFile
		err            error
		dbRows         int
	}{
		{
			name:   "validation errors",
			req:    &capiv1_protos.CreatePullRequestRequest{},
			err:    errors.New("2 errors occurred:\ntemplate name must be specified\nparameter values must be specified"),
			dbRows: 0,
		},
		{
			name: "name validation errors",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			req: &capiv1_protos.CreatePullRequestRequest{
				TemplateName: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "foo bar bad name",
					"NAMESPACE":    "default",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
			},
			err:    errors.New(`validation error rendering template cluster-template-1, invalid value for metadata.name: "foo bar bad name", a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`),
			dbRows: 0,
		},
		{
			name: "pull request failed",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			provider: NewFakeGitProvider("", nil, errors.New("oops")),
			req: &capiv1_protos.CreatePullRequestRequest{
				TemplateName: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "foo",
					"NAMESPACE":    "default",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
			},
			dbRows: 0,
			err:    errors.New(`rpc error: code = Unauthenticated desc = failed to access repo https://github.com/org/repo.git: oops`),
		},
		{
			name: "create pull request",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				TemplateName: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "foo",
					"NAMESPACE":    "default",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
			},
			dbRows:   1,
			expected: "https://github.com/org/repo/pull/1",
		},
		{
			name: "default profile values",
			clusterState: []runtime.Object{
				makeTemplateConfigMap("template1", makeTemplate(t)),
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil),
			req: &capiv1_protos.CreatePullRequestRequest{
				TemplateName: "cluster-template-1",
				ParameterValues: map[string]string{
					"CLUSTER_NAME": "dev",
					"NAMESPACE":    "default",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
				Values: []*capiv1_protos.ProfileValues{
					{
						Name:    "demo-profile",
						Version: "0.0.1",
						Values:  base64.StdEncoding.EncodeToString([]byte(``)),
					},
				},
			},
			dbRows: 1,
			committedFiles: []CommittedFile{
				{
					Path: ".weave-gitops/apps/capi/dev.yaml",
					Content: `apiVersion: fooversion
kind: fookind
metadata:
  annotations:
    capi.weave.works/display-name: ClusterName
    kustomize.toolkit.fluxcd.io/prune: disabled
  name: dev
`,
				},
				{
					Path: ".weave-gitops/clusters/dev/system/demo-profile.yaml",
					Content: `apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: weaveworks-charts
  namespace: default
spec:
  interval: 10m0s
  url: http://127.0.0.1:%s/charts
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: dev-demo-profile
  namespace: wego-system
spec:
  chart:
    spec:
      chart: demo-profile
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta1
        kind: HelmRepository
        name: weaveworks-charts
        namespace: default
      version: 0.0.1
  interval: 1m0s
  values:
    favoriteDrink: coffee
status: {}
`,
				},
			},
			expected: "https://github.com/org/repo/pull/1",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("RUNTIME_NAMESPACE", "default") // needs to match the helm repo namespace
			defer os.Unsetenv("RUNTIME_NAMESPACE")
			// setup
			ts := httptest.NewServer(makeServeMux(t))
			hr := makeTestHelmRepository(ts.URL, func(hr *sourcev1beta1.HelmRepository) {
				hr.Name = "weaveworks-charts"
				hr.Namespace = "default"
			})
			tt.clusterState = append(tt.clusterState, hr)
			db := createDatabase(t)
			s := createServer(t, tt.clusterState, "capi-templates", "default", tt.provider, db, "", hr)

			// request
			createPullRequestResponse, err := s.CreatePullRequest(context.Background(), tt.req)

			// Check the response looks good
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to create a pull request:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("got the wrong error:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, createPullRequestResponse.WebUrl, protocmp.Transform()); diff != "" {
					t.Fatalf("pull request url didn't match expected:\n%s", diff)
				}
				fakeGitProvider := (tt.provider).(*FakeGitProvider)
				if diff := cmp.Diff(tt.committedFiles, fakeGitProvider.GetCommittedFiles()); len(tt.committedFiles) > 0 && diff != "" {
					if !strings.Contains(diff, "url") {
						t.Fatalf("committed files do not match expected committed files:\n%s", diff)
					}
				}
			}

			// Check the db looks good
			var clusters []models.Cluster
			tx := db.Find(&clusters)
			if tx.Error != nil {
				t.Fatalf("error querying db:\n%v", tx.Error)
			}
			if diff := cmp.Diff(len(clusters), tt.dbRows); diff != "" {
				t.Fatalf("Rows mismatch:\n%s\nwas: %d", diff, len(clusters))
			}
		})
	}
}

func TestGetKubeconfig(t *testing.T) {
	testCases := []struct {
		name                    string
		clusterState            []runtime.Object
		clusterObjectsNamespace string // Namespace that cluster objects are created in
		req                     *capiv1_protos.GetKubeconfigRequest
		ctx                     context.Context
		expected                []byte
		err                     error
	}{
		{
			name: "get kubeconfig as JSON",
			clusterState: []runtime.Object{
				makeSecret("dev-kubeconfig", "default", "value", "foo"),
			},
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.MD{}),
			expected: []byte(fmt.Sprintf(`{"kubeconfig":"%s"}`, base64.StdEncoding.EncodeToString([]byte("foo")))),
		},
		{
			name: "get kubeconfig as binary",
			clusterState: []runtime.Object{
				makeSecret("dev-kubeconfig", "default", "value", "foo"),
			},
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.Pairs("accept", "application/octet-stream")),
			expected: []byte("foo"),
		},
		{
			name:                    "secret not found",
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			err: errors.New("unable to get secret \"dev-kubeconfig\" for Kubeconfig: secrets \"dev-kubeconfig\" not found"),
		},
		{
			name: "secret found but is missing key",
			clusterState: []runtime.Object{
				makeSecret("dev-kubeconfig", "default", "val", "foo"),
			},
			clusterObjectsNamespace: "default",
			req: &capiv1_protos.GetKubeconfigRequest{
				ClusterName: "dev",
			},
			err: errors.New("secret \"default/dev-kubeconfig\" was found but is missing key \"value\""),
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("CAPI_CLUSTERS_NAMESPACE", tt.clusterObjectsNamespace)
			defer os.Unsetenv("CAPI_CLUSTERS_NAMESPACE")

			db := createDatabase(t)
			gp := NewFakeGitProvider("", nil, nil)
			s := createServer(t, tt.clusterState, "capi-templates", "default", gp, db, tt.clusterObjectsNamespace, nil)

			res, err := s.GetKubeconfig(tt.ctx, tt.req)

			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to get the kubeconfig:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("got the wrong error:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, res.Data, protocmp.Transform()); diff != "" {
					t.Fatalf("kubeconfig didn't match expected:\n%s", diff)
				}
			}
		})
	}
}

func TestDeleteClustersPullRequest(t *testing.T) {
	testCases := []struct {
		name     string
		dbState  []interface{}
		provider git.Provider
		req      *capiv1_protos.DeleteClustersPullRequestRequest
		expected string
		err      error
	}{
		{
			name: "validation errors",
			req:  &capiv1_protos.DeleteClustersPullRequestRequest{},
			err:  errors.New("at least one cluster name must be specified"),
		},
		{
			name:     "cluster does not exist",
			dbState:  []interface{}{},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil),
			req: &capiv1_protos.DeleteClustersPullRequestRequest{
				ClusterNames:  []string{"foo"},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-02",
				BaseBranch:    "feature-01",
				Title:         "Delete Cluster",
				Description:   "Deletes a cluster",
				CommitMessage: "Remove cluster manifest",
			},
			err: gorm.ErrRecordNotFound,
		},
		{
			name: "create delete pull request",
			dbState: []interface{}{
				&models.Cluster{Name: "foo", Token: "foo-token"},
				&models.Cluster{Name: "bar", Token: "bar-token"},
			},
			provider: NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil),
			req: &capiv1_protos.DeleteClustersPullRequestRequest{
				ClusterNames:  []string{"foo", "bar"},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-02",
				BaseBranch:    "feature-01",
				Title:         "Delete Cluster",
				Description:   "Deletes a cluster",
				CommitMessage: "Remove cluster manifest",
			},
			expected: "https://github.com/org/repo/pull/1",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			db := createDatabase(t)
			s := createServer(t, []runtime.Object{}, "capi-templates", "default", tt.provider, db, "", nil)
			for _, o := range tt.dbState {
				db.Create(o)
			}

			// delete request
			deletePullRequestResponse, err := s.DeleteClustersPullRequest(context.Background(), tt.req)

			// Check the response looks good
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to create a pull request:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("got the wrong error:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, deletePullRequestResponse.WebUrl, protocmp.Transform()); diff != "" {
					t.Fatalf("pull request url didn't match expected:\n%s", diff)
				}

				var clusters []models.Cluster
				db.Preload(clause.Associations).Find(&clusters)
				for _, cluster := range clusters {
					if len(cluster.PullRequests) != 1 {
						t.Fatalf("got the wrong number of pull requests:%d", len(cluster.PullRequests))
					}
					if cluster.PullRequests[0].Type != "delete" {
						t.Fatalf("got the wrong type of pull request:%s", cluster.PullRequests[0].Type)
					}
				}
			}
		})
	}
}

func createDatabase(t *testing.T) *gorm.DB {
	db, err := utils.OpenDebug("", os.Getenv("DEBUG_SERVER_DB") == "true")
	if err != nil {
		t.Fatal(err)
	}
	err = utils.MigrateTables(db)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func makeNamespace(n string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: n,
		},
	}
}

func makeSecret(n string, ns string, s ...string) *corev1.Secret {
	data := make(map[string][]byte)
	for i := 0; i < len(s); i += 2 {
		data[s[i]] = []byte(s[i+1])
	}

	nsObj := makeNamespace(ns)

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: nsObj.GetName(),
		},
		Data: data,
	}
}

func NewFakeGitProvider(url string, repo *git.GitRepo, err error) git.Provider {
	return &FakeGitProvider{
		url:  url,
		repo: repo,
		err:  err,
	}
}

type FakeGitProvider struct {
	url            string
	repo           *git.GitRepo
	err            error
	committedFiles []gitprovider.CommitFile
}

func (p *FakeGitProvider) WriteFilesToBranchAndCreatePullRequest(ctx context.Context, req git.WriteFilesToBranchAndCreatePullRequestRequest) (*git.WriteFilesToBranchAndCreatePullRequestResponse, error) {
	if p.err != nil {
		return nil, p.err
	}
	p.committedFiles = append(p.committedFiles, req.Files...)
	return &git.WriteFilesToBranchAndCreatePullRequestResponse{WebURL: p.url}, nil
}

func (p *FakeGitProvider) CloneRepoToTempDir(req git.CloneRepoToTempDirRequest) (*git.CloneRepoToTempDirResponse, error) {
	if p.err != nil {
		return nil, p.err
	}
	return &git.CloneRepoToTempDirResponse{Repo: p.repo}, nil
}

func (p *FakeGitProvider) GetRepository(ctx context.Context, gp git.GitProvider, url string) (gitprovider.OrgRepository, error) {
	if p.err != nil {
		return nil, p.err
	}
	return nil, nil
}

func (p *FakeGitProvider) GetCommittedFiles() []CommittedFile {
	var committedFiles []CommittedFile
	for _, f := range p.committedFiles {
		committedFiles = append(committedFiles, CommittedFile{
			Path:    *f.Path,
			Content: *f.Content,
		})
	}
	return committedFiles
}

type CommittedFile struct {
	Path    string
	Content string
}

func makeServeMux(t *testing.T, opts ...func(*repo.IndexFile)) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/charts/index.yaml", func(w http.ResponseWriter, req *http.Request) {
		b, err := yaml.Marshal(makeTestChartIndex(opts...))
		if err != nil {
			t.Fatal(err)
		}

		_, err = w.Write(b)
		if err != nil {
			t.Fatal(err)
		}
	})
	mux.Handle("/", http.FileServer(http.Dir("../charts/testdata")))
	return mux
}

func makeTestChartIndex(opts ...func(*repo.IndexFile)) *repo.IndexFile {
	ri := &repo.IndexFile{
		APIVersion: "v1",
		Entries: map[string]repo.ChartVersions{
			"demo-profile": {
				{
					Metadata: &chart.Metadata{
						Annotations: map[string]string{
							charts.ProfileAnnotation: "demo-profile",
						},
						Description: "Simple demo profile",
						Home:        "https://example.com/testing",
						Name:        "demo-profile",
						Sources: []string{
							"https://example.com/testing",
						},
						Version: "0.0.1",
					},
					Created: time.Now(),
					Digest:  "aaff4545f79d8b2913a10cb400ebb6fa9c77fe813287afbacf1a0b897cdffffff",
					URLs: []string{
						"/charts/demo-profile-0.1.0.tgz",
					},
				},
			},
		},
	}
	for _, o := range opts {
		o(ri)
	}
	return ri
}
