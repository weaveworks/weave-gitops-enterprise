package alertmanager

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/alertmanager/api/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/payload"
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

	ev, err := GetAlertsAsEvent("derp", fmt.Sprintf("%v/api/v2", server.URL))
	assert.NoError(t, err)
	pa := payload.PrometheusAlerts{}
	err = ev.DataAs(&pa)
	assert.NoError(t, err)
	assert.Len(t, pa.Alerts, 1)
	assert.Equal(t, pa.Alerts[0].Labels["alertname"], "Watchdog")
}

func TestToCloudEvent(t *testing.T) {
	fp := "wiggly"
	alerts := models.GettableAlerts{&models.GettableAlert{Fingerprint: &fp}}
	pa := &payload.PrometheusAlerts{
		Token:  "derp",
		Alerts: alerts,
	}
	ev, err := ToCloudEvent("ewq", pa)
	assert.NoError(t, err)
	assert.Equal(t, "ewq", ev.Source())

	npa := payload.PrometheusAlerts{}
	err = ev.DataAs(&npa)
	assert.NoError(t, err)
	assert.Len(t, npa.Alerts, 1)
	assert.Equal(t, fp, *npa.Alerts[0].Fingerprint)
}
