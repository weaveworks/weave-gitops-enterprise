package payload

import (
	models "github.com/prometheus/alertmanager/api/v2/models"
)

type PrometheusAlerts struct {
	Token  string                `json:"token"`
	Alerts models.GettableAlerts `json:"alerts"`
}
