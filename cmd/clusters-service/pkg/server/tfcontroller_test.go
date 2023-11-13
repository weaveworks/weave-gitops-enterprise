package server

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/testing/protocmp"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git/gitfakes"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func TestCreateTerraformPullRequest(t *testing.T) {
	viper.SetDefault("terraform-repository-path", "clusters/my-cluster/cluster-templates")
	testCases := []struct {
		name           string
		clusterState   []runtime.Object
		provider       git.Provider
		pruneEnvVar    string
		req            *capiv1_protos.CreateTfControllerPullRequestRequest
		expected       string
		committedFiles []*capiv1_protos.CommitFile
		err            error
		dbRows         int
	}{
		{
			name:   "validation errors",
			req:    &capiv1_protos.CreateTfControllerPullRequestRequest{},
			err:    errors.New("2 errors occurred:\ntemplate name must be specified\nparameter values must be specified"),
			dbRows: 0,
		},
		{
			name: "name validation errors",
			clusterState: []runtime.Object{
				makeClusterTemplates(t),
			},
			req: &capiv1_protos.CreateTfControllerPullRequestRequest{
				Name: "cluster-template-1",
				ParameterValues: map[string]string{
					"RESOURCE_NAME": "foo bar bad name",
					"NAMESPACE":     "default",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a cluster through a CAPI template",
				CommitMessage: "Add cluster manifest",
				Namespace:     "default",
			},
			err:    errors.New(`validation error rendering template cluster-template-1, invalid value for metadata.name: "foo bar bad name", a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`),
			dbRows: 0,
		},
		{
			name: "pull request failed",
			clusterState: []runtime.Object{
				makeClusterTemplates(t),
			},
			provider: gitfakes.NewFakeGitProvider("", nil, errors.New("oops"), nil, nil),
			req: &capiv1_protos.CreateTfControllerPullRequestRequest{
				Name: "cluster-template-1",
				ParameterValues: map[string]string{
					"RESOURCE_NAME": "foo",
					"NAMESPACE":     "default",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a resource through Terraform template",
				CommitMessage: "Add terraform template",
				Namespace:     "default",
			},
			dbRows: 0,
			err:    errors.New(`rpc error: code = Unauthenticated desc = failed to access repo https://github.com/org/repo.git: oops`),
		},
		{
			name: "create pull request",
			clusterState: []runtime.Object{
				makeClusterTemplates(t),
			},
			provider: gitfakes.NewFakeGitProvider("https://github.com/org/repo/pull/1", nil, nil, nil, nil),
			req: &capiv1_protos.CreateTfControllerPullRequestRequest{
				Name: "cluster-template-1",
				ParameterValues: map[string]string{
					"RESOURCE_NAME": "foo",
					"NAMESPACE":     "default",
				},
				RepositoryUrl: "https://github.com/org/repo.git",
				HeadBranch:    "feature-01",
				BaseBranch:    "main",
				Title:         "New Cluster",
				Description:   "Creates a resource through Terraform template",
				CommitMessage: "Add terraform template",
				Namespace:     "default",
			},
			dbRows:   1,
			expected: "https://github.com/org/repo/pull/1",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			viper.SetDefault("runtime-namespace", "default")
			// setup
			ts := httptest.NewServer(makeServeMux(t))
			hr := makeTestHelmRepository(ts.URL, func(hr *sourcev1.HelmRepository) {
				hr.Name = "weaveworks-charts"
				hr.Namespace = "default"
			})
			tt.clusterState = append(tt.clusterState, hr)
			s := createServer(t, serverOptions{
				clusterState: tt.clusterState,
				namespace:    "default",
				provider:     tt.provider,
			})

			// request
			createPullRequestResponse, err := s.CreateTfControllerPullRequest(context.Background(), tt.req)

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
				fakeGitProvider := (tt.provider).(*gitfakes.FakeGitProvider)
				if diff := cmp.Diff(prepCommitedFiles(t, ts.URL, tt.committedFiles), fakeGitProvider.GetCommittedFiles()); len(tt.committedFiles) > 0 && diff != "" {
					t.Fatalf("committed files do not match expected committed files:\n%s", diff)
				}
			}
		})
	}
}
