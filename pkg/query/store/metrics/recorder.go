package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	dataStoreSubsystem = "datastore"

	StoreRolesAction            = "StoreRoles"
	StoreRoleBindingsAction     = "StoreRoleBindings"
	StoreObjectsAction          = "StoreObjects"
	DeleteObjectsAction         = "DeleteObjects"
	DeleteAllObjectsAction      = "DeleteAllObjects"
	DeleteRolesAction           = "DeleteRoles"
	DeleteAllRolesAction        = "DeleteAllRoles"
	DeleteRoleBindingsAction    = "DeleteRoleBindings"
	DeleteAllRoleBindingsAction = "DeleteAllRoleBindings"
	GetObjectsAction            = "GetObjects"
	GetObjectByIdAction         = "GetObjectByID"
	GetRolesAction              = "GetRoles"
	GetRoleBindingsAction       = "GetRoleBindings"
	GetAccessRulesAction        = "GetAccessRules"

	FailedLabel  = "error"
	SuccessLabel = "success"
)

// TODO review visibility
var DatastoreLatencyHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Subsystem: dataStoreSubsystem,
	Name:      "latency_seconds",
	Help:      "datastore latency",
	Buckets:   prometheus.LinearBuckets(0.01, 0.01, 10),
}, []string{"action", "status"})

// TODO review visibility
var DatastoreInflightRequests = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Subsystem: dataStoreSubsystem,
	Name:      "inflight_requests",
	Help:      "number of datastore in-flight requests.",
}, []string{"action"})

func init() {
	prometheus.MustRegister(DatastoreLatencyHistogram)
	prometheus.MustRegister(DatastoreInflightRequests)
}

func DataStoreSetLatency(action string, status string, duration time.Duration) {
	DatastoreLatencyHistogram.WithLabelValues(action, status).Observe(duration.Seconds())
}

func DataStoreInflightRequests(action string, number float64) {
	DatastoreInflightRequests.WithLabelValues(action).Add(number)
}
