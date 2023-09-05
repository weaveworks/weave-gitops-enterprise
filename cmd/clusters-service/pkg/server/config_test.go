package server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/testing/protocmp"

	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func TestGetConfig(t *testing.T) {
	testCases := []struct {
		name                 string
		repoUrl              string
		gitHostTypesEnv      string
		expectedGitHostTypes map[string]string
	}{
		{
			name:            "value set",
			repoUrl:         "https://github.com/user/blog",
			gitHostTypesEnv: "example.com=github",
			expectedGitHostTypes: map[string]string{
				"example.com":       "github",
				"dev.azure.com":     "azure-devops",
				"github.com":        "github",
				"gitlab.com":        "gitlab",
				"ssh.dev.azure.com": "azure-devops",
			},
		},
		{
			name:    " not set",
			repoUrl: "",
			expectedGitHostTypes: map[string]string{
				"dev.azure.com":     "azure-devops",
				"github.com":        "github",
				"gitlab.com":        "gitlab",
				"ssh.dev.azure.com": "azure-devops",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Fake user setting things
			viper.SetDefault("capi-templates-repository-url", tt.repoUrl)
			viper.SetDefault("git-host-types", tt.gitHostTypesEnv)

			s := createServer(t, serverOptions{})

			res, _ := s.GetConfig(context.Background(), &capiv1_protos.GetConfigRequest{})

			if diff := cmp.Diff(tt.repoUrl, res.RepositoryUrl, protocmp.Transform()); diff != "" {
				t.Fatalf("repository URL didn't match expected:\n%s", diff)
			}

			if diff := cmp.Diff(tt.expectedGitHostTypes, res.GitHostTypes, protocmp.Transform()); diff != "" {
				t.Fatalf("githosttypes didn't match expected:\n%s", diff)
			}
		})
	}
}
