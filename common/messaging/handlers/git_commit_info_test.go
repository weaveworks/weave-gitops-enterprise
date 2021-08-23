package handlers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/handlers"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/handlers/handlerstest"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/payload"
)

func TestGitCommitInfoSender(t *testing.T) {
	testCases := []struct {
		name string
		info payload.GitCommitInfo
		// error to be returned by the fake client
		clientErr error
		// error returned from Send method
		err      error
		expected *payload.GitCommitInfo
	}{
		{
			name: "GitCommitInfo gets sent successfully",
			info: payload.GitCommitInfo{
				Token: "derp",
				Commit: payload.CommitView{
					Sha:     "123",
					Message: "Fixing prod",
					Author: payload.UserView{
						Name:  "foo",
						Email: "foo@weave.works",
						Date:  time.Now(),
					},
					Committer: payload.UserView{
						Name:  "bar",
						Email: "bar@weave.works",
						Date:  time.Now(),
					},
				},
			},
			clientErr: nil,
			err:       nil,
			expected: &payload.GitCommitInfo{
				Token: "derp",
				Commit: payload.CommitView{
					Sha:     "123",
					Message: "Fixing prod",
					Author: payload.UserView{
						Name:  "foo",
						Email: "foo@weave.works",
						Date:  time.Now(),
					},
					Committer: payload.UserView{
						Name:  "bar",
						Email: "bar@weave.works",
						Date:  time.Now(),
					},
				},
			},
		},
		{
			name: "GitCommitInfo does not get sent successfully",
			info: payload.GitCommitInfo{
				Token: "derp",
				Commit: payload.CommitView{
					Sha: "123",
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
			sender := handlers.NewGitCommitInfoSender("test", client)
			err := sender.Send(context.TODO(), tc.info)
			assert.Equal(t, err, tc.err)
			if tc.expected == nil {
				client.AssertNoCloudEventsWereSent(t)
			} else {
				client.AssertGitCommitInfoWasSent(t, *tc.expected)
			}
		})
	}
}
