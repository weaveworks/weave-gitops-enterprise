package server_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/weaveworks/pipeline-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/internal/grpctesting"
	"github.com/weaveworks/weave-gitops-enterprise/internal/pipetesting"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestApprovePipeline(t *testing.T) {
	ctx := context.Background()

	kclient := fake.NewClientBuilder().WithScheme(grpctesting.BuildScheme()).Build()

	pipelineNamespace := pipetesting.NewNamespace(ctx, t, kclient)
	targetNamespace := pipetesting.NewNamespace(ctx, t, kclient)

	factory := grpctesting.MakeClustersManager(kclient, "management", fmt.Sprintf("%s/cluster-1", pipelineNamespace.Name))

	// Setup pipeline controller server
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "https://github.com/my-project/pulls/1")
		w.WriteHeader(http.StatusCreated)
	}))
	defer s.Close()

	serverClient := pipetesting.SetupServer(t, factory, kclient, "management", s.URL, nil)

	hr := createHelmRelease(ctx, t, kclient, "app-1", targetNamespace.Name)

	envName := "env-1"

	// create a pipeline that's ready to be approved
	p := newPipeline("pipe-1", pipelineNamespace.Name, targetNamespace.Name, envName, hr)
	p.Status.Environments = map[string]*v1alpha1.EnvironmentStatus{
		envName: {
			WaitingApproval: v1alpha1.WaitingApproval{
				Revision: "1.2.3",
			},
		},
	}
	require.NoError(t, kclient.Create(ctx, p))

	resp, err := serverClient.ApprovePromotion(context.Background(), &pb.ApprovePromotionRequest{
		Name:      p.Name,
		Namespace: pipelineNamespace.Name,
		Env:       envName,
		Revision:  "1.2.1",
	})
	require.NoError(t, err)

	require.Equal(t, "https://github.com/my-project/pulls/1", resp.PullRequestUrl)
}
