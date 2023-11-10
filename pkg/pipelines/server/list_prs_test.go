package server_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops-enterprise/internal/grpctesting"
	"github.com/weaveworks/weave-gitops-enterprise/internal/pipetesting"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/git"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetListPullRequest(t *testing.T) {
	ctx := context.Background()
	kclient := fake.NewClientBuilder().WithScheme(grpctesting.BuildScheme()).Build()

	pipelineNamespace := pipetesting.NewNamespace(ctx, t, kclient)
	targetNamespace := pipetesting.NewNamespace(ctx, t, kclient)
	secret := createSecret(ctx, t, kclient, "github-token", pipelineNamespace.Name, map[string][]byte{
		"token": []byte("github-token"),
	})
	hr := createHelmRelease(ctx, t, kclient, "app-1", targetNamespace.Name)

	envName := "env-1"
	p := newPipeline("pipe-2", pipelineNamespace.Name, targetNamespace.Name, envName, hr)
	p.Spec.Promotion = &ctrl.Promotion{
		Strategy: ctrl.Strategy{
			PullRequest: &ctrl.PullRequestPromotion{
				SecretRef: meta.LocalObjectReference{
					Name: secret.Name,
				},
			},
		},
	}
	require.NoError(t, kclient.Create(ctx, p))

	prs := []*git.PullRequest{gitfakes.NewPullRequest(0, "testing", pipelineNamespace.Name+"/pipe-2/env-1", "http://example.com/foo/bar/pulls/1", false, "main")}
	fakeGitProvider := gitfakes.NewFakeGitProvider("", nil, nil, nil, prs)

	factory := grpctesting.MakeClustersManager(kclient, nil, "management", fmt.Sprintf("%s/cluster-1", pipelineNamespace.Name))
	serverClient := pipetesting.SetupServer(t, factory, kclient, "management", "", fakeGitProvider)

	res, err := serverClient.ListPullRequests(context.Background(), &pb.ListPullRequestsRequest{
		Name:      p.Name,
		Namespace: pipelineNamespace.Name,
	})
	require.NoError(t, err)

	assert.Equal(t, map[string]string{
		"env-1": "http://example.com/foo/bar/pulls/1",
	}, res.PullRequests)
}

func createSecret(ctx context.Context, t *testing.T, k client.Client, name string, ns string, data map[string][]byte) *corev1.Secret {
	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}

	require.NoError(t, k.Create(ctx, s))

	return s
}
