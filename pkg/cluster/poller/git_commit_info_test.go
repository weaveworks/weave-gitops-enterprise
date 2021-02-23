package poller_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	httppayload "github.com/weaveworks/wks/common/http/payload"
	"github.com/weaveworks/wks/common/messaging/handlers"
	"github.com/weaveworks/wks/common/messaging/handlers/handlerstest"
	"github.com/weaveworks/wks/common/messaging/payload"
	clusterpoller "github.com/weaveworks/wks/pkg/cluster/poller"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGitCommitInfoPoller(t *testing.T) {
	testCases := []struct {
		name         string
		clusterState []runtime.Object
		httpResponse interface{}
		token        string
		expected     *payload.GitCommitInfo
	}{
		{
			name: "No service found",
			clusterState: []runtime.Object{
				&v1.Service{},
			},
			token:    "derp",
			expected: nil,
		},
		{
			name: "Service running",
			clusterState: []runtime.Object{
				&v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gitops-repo-broker",
						Namespace: "wkp-gitops-repo-broker",
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{
							{
								Name:     "http",
								Port:     8000,
								Protocol: "TCP",
							},
						},
						Type: v1.ServiceTypeClusterIP,
					},
				},
			},
			httpResponse: httppayload.BranchesView{
				Branches: []httppayload.BranchView{
					{
						Name: "master",
						Head: httppayload.CommitView{
							Author: httppayload.UserView{
								Name:  "foo",
								Email: "foo@users.noreply.github.com",
								When:  time.Now().UTC(),
							},
							Committer: httppayload.UserView{
								Name:  "GitHub",
								Email: "noreply@github.com",
								When:  time.Now().UTC(),
							},
							Hash:    "6c4b104bf502bb417aa2d73aac36fcd58f4e15df",
							Message: "Update gitops-params.yaml",
						},
					},
				},
			},
			token: "derp",
			expected: &payload.GitCommitInfo{
				Token: "derp",
				Commit: payload.CommitView{
					Sha: "6c4b104bf502bb417aa2d73aac36fcd58f4e15df",
					Author: payload.UserView{
						Name:  "foo",
						Email: "foo@users.noreply.github.com",
						Date:  time.Now().UTC(),
					},
					Committer: payload.UserView{
						Name:  "GitHub",
						Email: "noreply@github.com",
						Date:  time.Now().UTC(),
					},
					Message: "Update gitops-params.yaml",
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			clientset := fake.NewSimpleClientset(tt.clusterState...)
			client := handlerstest.NewFakeCloudEventsClient()
			sender := handlers.NewGitCommitInfoSender("test", client)
			httpClient := resty.New()
			httpmock.ActivateNonDefault(httpClient.GetClient())
			defer httpmock.DeactivateAndReset()
			if tt.httpResponse != nil {
				httpmock.RegisterResponder("GET", "http://gitops-repo-broker.wkp-gitops-repo-broker:8000/gitops/repo/branches", httpmock.NewJsonResponderOrPanic(200, tt.httpResponse))
			}
			poller := clusterpoller.NewGitCommitInfoPoller(tt.token, clientset, time.Second, httpClient, sender)

			go poller.Run(ctx.Done())
			time.Sleep(50 * time.Millisecond)
			cancel()

			if tt.expected == nil {
				client.AssertNoCloudEventsWereSent(t)
			} else {
				client.AssertGitCommitInfoWasSent(t, *tt.expected)
			}
		})
	}
}
