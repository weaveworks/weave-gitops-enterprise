package metrics

import (
	"reflect"
	"testing"
)

func TestParseMetrics(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput []Metric
		expectedError  bool
	}{
		{
			name: "counter metric",
			input: `
# HELP rest_client_requests_total Number of HTTP requests, partitioned by status code, method, and host.
# TYPE rest_client_requests_total counter
rest_client_requests_total{code="200",host="10.96.0.1:443",method="GET"} 649
`,
			expectedOutput: []Metric{
				{
					Name: "rest_client_requests_total",
					Help: "Number of HTTP requests, partitioned by status code, method, and host.",
					Type: "COUNTER",
					Labels: map[string]string{
						"code":   "200",
						"host":   "10.96.0.1:443",
						"method": "GET",
					},
					Value: 649.0,
				},
			},
			expectedError: false,
		},
		{
			name: "gauge metric",
			input: `
# HELP controller_runtime_active_workers Number of currently used workers per controller
# TYPE controller_runtime_active_workers gauge
controller_runtime_active_workers{controller="fluxshardset"} 0
`,
			expectedOutput: []Metric{
				{
					Name:   "controller_runtime_active_workers",
					Help:   "Number of currently used workers per controller",
					Type:   "GAUGE",
					Labels: map[string]string{"controller": "fluxshardset"},
					Value:  0.0,
				},
			},
			expectedError: false,
		},
		{
			name: "histogram metric",
			input: `
# HELP workqueue_work_duration_seconds How long in seconds processing an item from workqueue takes.
# TYPE workqueue_work_duration_seconds histogram
workqueue_work_duration_seconds_bucket{name="fluxshardset",le="1e-08"} 0
workqueue_work_duration_seconds_bucket{name="fluxshardset",le="1e-07"} 0
workqueue_work_duration_seconds_bucket{name="fluxshardset",le="1e-06"} 0
workqueue_work_duration_seconds_bucket{name="fluxshardset",le="9.999999999999999e-06"} 0
workqueue_work_duration_seconds_bucket{name="fluxshardset",le="9.999999999999999e-05"} 0
workqueue_work_duration_seconds_bucket{name="fluxshardset",le="0.001"} 0
workqueue_work_duration_seconds_bucket{name="fluxshardset",le="0.01"} 0
workqueue_work_duration_seconds_bucket{name="fluxshardset",le="0.1"} 2
workqueue_work_duration_seconds_bucket{name="fluxshardset",le="1"} 2
workqueue_work_duration_seconds_bucket{name="fluxshardset",le="10"} 2
workqueue_work_duration_seconds_bucket{name="fluxshardset",le="+Inf"} 2
workqueue_work_duration_seconds_sum{name="fluxshardset"} 0.043736973
workqueue_work_duration_seconds_count{name="fluxshardset"} 2
`,
			expectedOutput: []Metric{
				{
					Name:   "workqueue_work_duration_seconds",
					Help:   "How long in seconds processing an item from workqueue takes.",
					Type:   "HISTOGRAM",
					Labels: map[string]string{"name": "fluxshardset"},
					Value:  0.043736973,
					Count:  2,
				},
			},
			expectedError: false,
		},
		{
			name: "summary metric",
			input: `
# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 4.8959e-05
go_gc_duration_seconds{quantile="0.25"} 7.39e-05
go_gc_duration_seconds{quantile="0.5"} 0.000214533
go_gc_duration_seconds{quantile="0.75"} 0.000365024
go_gc_duration_seconds{quantile="1"} 0.000474971
go_gc_duration_seconds_sum 0.003078904
go_gc_duration_seconds_count 14
			`,
			expectedOutput: []Metric{
				{
					Name:   "go_gc_duration_seconds",
					Help:   "A summary of the pause duration of garbage collection cycles.",
					Type:   "SUMMARY",
					Labels: map[string]string{},
					Value:  0.003078904,
					Count:  14,
				},
			},
		},
	}

	for _, test := range tests {
		output, err := parseMetrics([]byte(test.input))

		if (err != nil) != test.expectedError {
			t.Errorf("Unexpected error status for input:\n%s\nError: %v", test.input, err)
			continue
		}

		if !reflect.DeepEqual(output, test.expectedOutput) {
			t.Errorf("Mismatch for input:\n%s\nExpected: %v\nGot: %v", test.input, test.expectedOutput, output)
		}
	}
}
