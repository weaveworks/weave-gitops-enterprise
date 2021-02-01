package alertmanager

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const exampleAlertJson = `
{
	"version": "4",
	"groupKey": "{}:{alertname=\"TestAlert\"}",
	"status": "firing",
	"receiver": "team",
	"groupLabels": {},
	"commonLabels": {},
	"commonAnnotations": {},
	"externalURL": "http://127.0.0.1:9093",
	"alerts": [
		{
		"status": "firing",
		"labels": {},
		"annotations": {},
		"startsAt": "2019-01-03T22:20:08.822+09:00",
		"endsAt": "0001-01-01T00:00:00Z",
		"generatorURL": "http://127.0.0.1:9090/graph"
		}
	]
}
`

func TestWebhookHandler(t *testing.T) {
	req, err := http.NewRequest("", "", bytes.NewBuffer([]byte(exampleAlertJson)))
	require.Nil(t, err)

	response, ce := executeRequest(req)
	assert.Equal(t, "http://127.0.0.1:9093", ce.Source())
	assert.Equal(t, http.StatusOK, response.Code)
}

func executeRequest(req *http.Request) (*httptest.ResponseRecorder, event.Event) {
	rec := httptest.NewRecorder()
	var alertEvent event.Event
	handler := http.HandlerFunc(NewWebhookHandler(func(ce event.Event) {
		alertEvent = ce
	}))
	handler.ServeHTTP(rec, req)
	return rec, alertEvent
}
