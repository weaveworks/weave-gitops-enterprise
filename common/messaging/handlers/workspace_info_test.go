package handlers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/handlers"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/handlers/handlerstest"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/payload"
)

func TestWorkspaceInfoSender(t *testing.T) {
	testCases := []struct {
		name string
		info payload.WorkspaceInfo
		// error to be returned by the fake client
		clientErr error
		// error returned from Send method
		err      error
		expected *payload.WorkspaceInfo
	}{
		{
			name: "WorkspaceInfo gets sent successfully",
			info: payload.WorkspaceInfo{
				Token: "derp",
				Workspaces: []payload.Workspace{
					{
						Name:      "foo-ws",
						Namespace: "foo-ns",
					},
					{
						Name:      "bar-ws",
						Namespace: "bar-ns",
					},
				},
			},
			expected: &payload.WorkspaceInfo{
				Token: "derp",
				Workspaces: []payload.Workspace{
					{
						Name:      "foo-ws",
						Namespace: "foo-ns",
					},
					{
						Name:      "bar-ws",
						Namespace: "bar-ns",
					},
				},
			},
		},
		{
			name: "WorkspaceInfo does not get sent successfully",
			info: payload.WorkspaceInfo{
				Token: "derp",
				Workspaces: []payload.Workspace{
					{
						Name:      "foo-ws",
						Namespace: "foo-ns",
					},
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
			sender := handlers.NewWorkspaceInfoSender("test", client)
			err := sender.Send(context.TODO(), tc.info)
			assert.Equal(t, err, tc.err)
			if tc.expected == nil {
				client.AssertNoCloudEventsWereSent(t)
			} else {
				client.AssertWorkspaceInfoWasSent(t, *tc.expected)
			}
		})
	}
}
