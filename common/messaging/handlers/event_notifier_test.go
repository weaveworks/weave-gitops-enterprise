package handlers_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/handlers"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/handlers/handlerstest"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/payload"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEventNotifier(t *testing.T) {
	testCases := []struct {
		name string
		obj  interface{}
		// error to be returned by the fake client
		clientErr error
		// error returned from Notify method
		err      error
		expected *payload.KubernetesEvent
	}{
		{
			name:      "Object that is not a Kubernetes Event gets ignored",
			obj:       &v1.Pod{},
			clientErr: nil,
			err:       nil,
			expected:  nil,
		},
		{
			name: "Object that is a Kubernetes Event gets sent successfully",
			obj: &v1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
			clientErr: nil,
			err:       nil,
			expected: &payload.KubernetesEvent{
				Token: "derp",
				Event: v1.Event{
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
				},
			},
		},
		{
			name: "Object that is a Kubernetes Event does not get sent successfully",
			obj: &v1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
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
			notifier := handlers.NewEventNotifier("derp", "test", client)
			err := notifier.Notify("add", tc.obj)
			assert.Equal(t, err, tc.err)
			if tc.expected == nil {
				client.AssertNoCloudEventsWereSent(t)
			} else {
				client.AssertEventWasSent(t, *tc.expected)
			}
		})
	}
}
