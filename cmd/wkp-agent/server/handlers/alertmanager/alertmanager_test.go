package alertmanager

import (
	"bytes"
	"fmt"
	"github.com/prometheus/alertmanager/api/v2/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const alertResponse = `
[
  {
    "annotations": {
      "message": "This is an alert meant to ensure that the entire alerting pipeline is functional.\nThis alert is always firing, therefore it should always be firing in Alertmanager\nand always fire against a receiver. There are integrations with various notification\nmechanisms that send a notification when this alert is not firing. For example the\n\"DeadMansSnitch\" integration in PagerDuty.\n"
    },
    "endsAt": "2021-02-01T14:08:22.565Z",
    "fingerprint": "68fbcb30c99fc94d",
    "receivers": [
      {
        "name": "null"
      }
    ],
    "startsAt": "2021-01-27T08:55:22.565Z",
    "status": {
      "inhibitedBy": [],
      "silencedBy": [],
      "state": "active"
    },
    "updatedAt": "2021-02-01T14:04:22.572Z",
    "generatorURL": "/prometheus/graph?g0.expr=vector%281%29&g0.tab=1",
    "labels": {
      "alertname": "Watchdog",
      "prometheus": "wkp-prometheus/prometheus-operator-kube-p-prometheus",
      "severity": "none"
    }
  }
]
`

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

func TestGetAlertsAsEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Contains(t, req.URL.String(), "/api/v2/alerts")
		// Send response to be tested
		rw.Header().Set("Content-Type", "application/json")
		rw.Write([]byte(alertResponse))
	}))
	// Close the server when test finishes
	defer server.Close()

	ev, err := GetAlertsAsEvent(fmt.Sprintf("%v/api/v2", server.URL))
	assert.NoError(t, err)
	alerts := models.GettableAlerts{}
	err = ev.DataAs(&alerts)
	assert.NoError(t, err)
	assert.Len(t, alerts, 1)
	assert.Equal(t, alerts[0].Labels["alertname"], "Watchdog")
}

func TestToCloudEvent(t *testing.T) {
	fp := "wiggly"
	alerts := models.GettableAlerts{&models.GettableAlert{Fingerprint: &fp}}
	ev, err := ToCloudEvent("ewq", alerts)
	assert.NoError(t, err)
	assert.Equal(t, "ewq", ev.Source())

	newAlerts := models.GettableAlerts{}
	err = ev.DataAs(&newAlerts)
	assert.NoError(t, err)
	assert.Len(t, newAlerts, 1)
	assert.Equal(t, fp, *newAlerts[0].Fingerprint)
}
