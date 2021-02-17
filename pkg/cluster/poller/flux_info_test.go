package poller_test

import (
	"context"
	"testing"
	"time"

	"github.com/weaveworks/wks/common/messaging/handlers"
	"github.com/weaveworks/wks/common/messaging/handlers/handlerstest"
	"github.com/weaveworks/wks/common/messaging/payload"
	clusterpoller "github.com/weaveworks/wks/pkg/cluster/poller"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestFluxInfoPoller(t *testing.T) {
	testCases := []struct {
		name         string
		clusterState []runtime.Object // state of cluster before executing tests
		expected     *payload.FluxInfo
	}{
		{
			name: "No flux deployments",
			clusterState: []runtime.Object{
				&v1.Namespace{},
			},
			expected: nil,
		},
		{
			name: "One namespace and one flux deployment with the expected info",
			clusterState: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "wkp-flux",
						UID:  "f72c7ce4-afd1-4840-bd50-fb4fabc99859",
					},
				},
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "flux",
						Namespace: "wkp-flux",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"name": "flux"},
						},
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Name:  "flux",
										Image: "docker.io/weaveworks/wkp-jk-init:v2.0.3-RC.1-2-gd677dc0a",
										Args: []string{
											"--memcached-service=",
											"--ssh-keygen-dir=/var/fluxd/keygen",
											"--sync-garbage-collection=true",
											"--git-poll-interval=10s",
											"--sync-interval=10s",
											"--manifest-generation=true",
											"--listen-metrics=:3031",
											"--git-url=git@github.com:dinosk/fluxes-1.git",
											"--git-branch=master",
											"--registry-exclude-image=*"},
									},
								},
							},
						},
					},
				},
			},
			expected: &payload.FluxInfo{
				Token: "derp",
				Deployments: []payload.FluxDeploymentInfo{
					{
						Name:      "flux",
						Namespace: "wkp-flux",
						Args: []string{
							"--memcached-service=",
							"--ssh-keygen-dir=/var/fluxd/keygen",
							"--sync-garbage-collection=true",
							"--git-poll-interval=10s",
							"--sync-interval=10s",
							"--manifest-generation=true",
							"--listen-metrics=:3031",
							"--git-url=git@github.com:dinosk/fluxes-1.git",
							"--git-branch=master",
							"--registry-exclude-image=*"},
						Image: "docker.io/weaveworks/wkp-jk-init:v2.0.3-RC.1-2-gd677dc0a",
					},
				},
			},
		},
		{
			name: "One namespace and one flux deployment with missing info",
			clusterState: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "wkp-flux",
						UID:  "f72c7ce4-afd1-4840-bd50-fb4fabc99859",
					},
				},
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "flux",
						Namespace: "wkp-flux",
					},
					Spec: appsv1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Name:  "flux",
										Image: "docker.io/weaveworks/wkp-jk-init:v2.0.3-RC.1-2-gd677dc0a",
									},
								},
							},
						},
					},
				},
			},
			expected: nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			clientset := fake.NewSimpleClientset(tt.clusterState...)
			client := handlerstest.NewFakeCloudEventsClient()
			sender := handlers.NewFluxInfoSender("test", client)
			poller := clusterpoller.NewFluxInfoPoller("derp", clientset, time.Second, sender)

			go poller.Run(ctx.Done())
			time.Sleep(50 * time.Millisecond)
			cancel()

			if tt.expected == nil {
				client.AssertNoCloudEventsWereSent(t)
			} else {
				client.AssertFluxInfoWasSent(t, *tt.expected)
			}
		})
	}
}
