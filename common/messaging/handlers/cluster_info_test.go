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

func TestClusterInfoSender(t *testing.T) {
	testCases := []struct {
		name string
		info payload.ClusterInfo
		// error to be returned by the fake client
		clientErr error
		// error returned from Send method
		err      error
		expected *payload.ClusterInfo
	}{
		{
			name: "ClusterInfo gets sent successfully",
			info: payload.ClusterInfo{
				Cluster: payload.Cluster{
					ID:   "foo",
					Type: "existinginfra",
					Nodes: []payload.Node{
						{
							MachineID:      "111",
							IsControlPlane: true,
							KubeletVersion: "1.20.1",
						},
						{
							MachineID:      "222",
							IsControlPlane: false,
							KubeletVersion: "1.20.1",
						},
					},
				},
			},
			clientErr: nil,
			err:       nil,
			expected: &payload.ClusterInfo{
				Cluster: payload.Cluster{
					ID:   "foo",
					Type: "existinginfra",
					Nodes: []payload.Node{
						{
							MachineID:      "111",
							IsControlPlane: true,
							KubeletVersion: "1.20.1",
						},
						{
							MachineID:      "222",
							IsControlPlane: false,
							KubeletVersion: "1.20.1",
						},
					},
				},
			},
		},
		{
			name: "ClusterInfo does not get sent successfully",
			info: payload.ClusterInfo{
				Cluster: payload.Cluster{
					ID: "bar",
				},
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
			sender := handlers.NewClusterInfoSender("test", client)
			err := sender.Send(context.TODO(), tc.info)
			assert.Equal(t, err, tc.err)
			if tc.expected == nil {
				client.AssertNoCloudEventsWereSent(t)
			} else {
				client.AssertClusterInfoWasSent(t, *tc.expected)
			}
		})
	}
}
