package server

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func TestGetConfig(t *testing.T) {
	testCases := []struct {
		name  string
		value string
	}{
		{
			name:  "value set",
			value: "https://github.com/user/blog",
		},
		{
			name:  "value not set",
			value: "",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("CAPI_TEMPLATES_REPOSITORY_URL", tt.value)
			defer os.Unsetenv("CAPI_TEMPLATES_REPOSITORY_URL")

			s := createServer(t, nil, "", "", nil, nil, "", nil)

			res, _ := s.GetConfig(context.Background(), &capiv1_protos.GetConfigRequest{})

			if diff := cmp.Diff(tt.value, res.RepositoryURL, protocmp.Transform()); diff != "" {
				t.Fatalf("repository URL didn't match expected:\n%s", diff)
			}
		})
	}
}
