package poller

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/wks/common/messaging/handlers"
	"github.com/weaveworks/wks/common/messaging/handlers/handlerstest"
	"github.com/weaveworks/wks/common/messaging/payload"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestParseLogs(t *testing.T) {
	logs := `
ts=2021-03-02T12:59:05.544496465Z caller=loop.go:127 component=sync-loop event=refreshed url=ssh://git@github.com/foot/wk-simon.git branch=master HEAD=abc
ts=2021-03-02T13:59:05.544496465Z caller=loop.go:127 component=sync-loop event=refreshed url=ssh://git@github.com/foot/wk-simon.git branch=master HEAD=abc
ts=2021-03-02T14:59:05.544496465Z caller=loop.go:127 component=sync-loop event=refreshed url=ssh://git@github.com/foot/wk-simon.git branch=master HEAD=def
`

	details, err := constructLogDetails(strings.NewReader(logs))
	assert.NoError(t, err)
	assert.Len(t, details, 2)
	assert.Equal(t, details[0].Branch, "master")
	assert.Equal(t, details[0].Head, "abc")
	assert.Equal(t, details[1].Head, "def")
}

var namespaceAndDeployment = []runtime.Object{
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
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "flux",
					},
				},
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
}

func TestFluxInfoPoller(t *testing.T) {
	testCases := []struct {
		name              string
		clusterState      []runtime.Object // state of cluster before executing tests
		expected          *payload.FluxInfo
		expectedErrorLogs []string
	}{
		{
			name: "No flux deployments",
			clusterState: []runtime.Object{
				&v1.Namespace{},
			},
			expected: nil,
			expectedErrorLogs: []string{
				"Unable to create fluxInfo object: No flux deployments detected",
			},
		},
		{
			name:         "One namespace and one flux deployment with the expected info",
			clusterState: namespaceAndDeployment,
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
			expectedErrorLogs: []string{
				"Failed to query Flux logs: No pods found for deployment",
			},
		},
		{
			name: "One namespace and one flux deployment with the expected info + a pod",
			clusterState: append(namespaceAndDeployment,
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "flux-5968779dc7-wrg4x",
						Namespace: "wkp-flux",
						Labels: map[string]string{
							"name": "flux",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "flux",
								Image: "flux-image",
							},
						},
					},
				},
			),
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
						// Found some logs but none watched what we were after.
						Syncs: []payload.FluxLogInfo{},
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
			expectedErrorLogs: []string{
				"Unable to create fluxInfo object: No flux deployments detected",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			logsHook := logrustest.NewGlobal()
			defer logsHook.Reset()

			ctx, cancel := context.WithCancel(context.Background())
			clientset := fake.NewSimpleClientset(tt.clusterState...)
			client := handlerstest.NewFakeCloudEventsClient()
			sender := handlers.NewFluxInfoSender("test", client)
			poller := NewFluxInfoPoller("derp", clientset, time.Second, sender)

			go poller.Run(ctx.Done())
			time.Sleep(50 * time.Millisecond)
			cancel()

			if tt.expected == nil {
				client.AssertNoCloudEventsWereSent(t)
			} else {
				client.AssertFluxInfoWasSent(t, *tt.expected)
			}

			assertErrorsLogged(t, logsHook, tt.expectedErrorLogs)
		})
	}
}

func assertErrorsLogged(t *testing.T, logsHook *logrustest.Hook, expectedErrorLogs []string) {
	actualErrorLogs := []string{}
	for _, en := range logsHook.Entries {
		if en.Level == logrus.ErrorLevel {
			actualErrorLogs = append(actualErrorLogs, en.Message)
		}
	}
	assert.Subset(t, expectedErrorLogs, actualErrorLogs)
}
