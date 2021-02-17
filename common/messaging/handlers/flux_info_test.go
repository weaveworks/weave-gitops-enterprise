package handlers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/wks/common/messaging/handlers"
	"github.com/weaveworks/wks/common/messaging/handlers/handlerstest"
	"github.com/weaveworks/wks/common/messaging/payload"
)

func TestFluxInfoSender(t *testing.T) {
	testCases := []struct {
		name string
		info payload.FluxInfo
		// error to be returned by the fake client
		clientErr error
		// error returned from Send method
		err      error
		expected *payload.FluxInfo
	}{
		{
			name: "FluxInfo gets sent successfully",
			info: payload.FluxInfo{
				Token: "derp",
				Deployments: []payload.FluxDeploymentInfo{
					{
						Name: "flux",
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
			clientErr: nil,
			err:       nil,
			expected: &payload.FluxInfo{
				Token: "derp",
				Deployments: []payload.FluxDeploymentInfo{
					{
						Name: "flux",
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
			name: "ClusterInfo does not get sent successfully",
			info: payload.FluxInfo{
				Token: "derp",
			},
			clientErr: errors.New("oops"),
			err:       errors.New("oops"),
			expected:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := handlerstest.NewFakeCloudEventsClient()
			client.SetupErrorForSend(tc.clientErr)
			sender := handlers.NewFluxInfoSender("test", client)
			err := sender.Send(context.TODO(), tc.info)
			assert.Equal(t, err, tc.err)
			if tc.expected == nil {
				client.AssertNoCloudEventsWereSent(t)
			} else {
				client.AssertFluxInfoWasSent(t, *tc.expected)
			}
		})
	}
}
