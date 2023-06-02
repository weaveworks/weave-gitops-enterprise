package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	expReqs = []handlerConfig{
		{Path: "/test/1", Method: "GET", Code: 200, ReturnData: "hello world!", SleepDuration: 45 * time.Millisecond, NumberRequests: 10},
	}

	expMetrics = []string{
		`# HELP http_request_duration_seconds The latency of the HTTP requests.`,
		`# TYPE http_request_duration_seconds histogram`,
		`http_request_duration_seconds_bucket{code="200",handler="/test/1",method="GET",service="",le="+Inf"} 10`,
		`http_request_duration_seconds_count{code="200",handler="/test/1",method="GET",service=""} 10`,

		`# HELP http_requests_inflight The number of inflight requests being handled at the same time.`,
		`# TYPE http_requests_inflight gauge`,
		`http_requests_inflight{handler="/test/1",service=""} 0`,

		`# HELP http_response_size_bytes The size of the HTTP responses.`,
		`# TYPE http_response_size_bytes histogram`,
		`http_response_size_bytes_bucket{code="200",handler="/test/1",method="GET",service="",le="+Inf"} 10`,
		`http_response_size_bytes_sum{code="200",handler="/test/1",method="GET",service=""} 120`,
		`http_response_size_bytes_count{code="200",handler="/test/1",method="GET",service=""} 10`,
	}
)

type testServer struct{ server *httptest.Server }

func (t testServer) Close()      { t.server.Close() }
func (t testServer) URL() string { return t.server.URL }

// handlerConfig is the configuration the servers will need to set up to properly
// execute the tests.
type handlerConfig struct {
	Path           string
	Code           int
	Method         string
	ReturnData     string
	SleepDuration  time.Duration
	NumberRequests int
}

func TestNewMetricsServer(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	t.Run("should create new metrics server", func(t *testing.T) {
		NewPrometheusServer(Options{
			ServerAddress: "localhost:8080",
		}, prometheus.Gatherers{
			prometheus.DefaultGatherer,
		})

		// Get metrics.
		r, err := http.NewRequest(http.MethodGet, "http://localhost:8080/metrics", nil)
		require.NoError(err)
		resp, err := http.DefaultClient.Do(r)
		require.NoError(err)

		// Check.
		b, err := io.ReadAll(resp.Body)
		require.NoError(err)
		metrics := string(b)

		assert.Contains(metrics, "# HELP go_gc_duration_seconds")
	})
}

func TestWithMetrics(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	t.Run("should record metrics for http server", func(t *testing.T) {
		// create metrics server
		NewPrometheusServer(Options{
			ServerAddress: "localhost:8080",
		}, prometheus.Gatherers{
			prometheus.DefaultGatherer,
		})
		// create http server with metrics recorder
		next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(http.StatusOK)
			_, _ = rw.Write([]byte("hello world!"))
		})
		handler := WithHttpMetrics(next)
		s := testServer{server: httptest.NewServer(handler)}
		t.Cleanup(func() { s.Close() })

		// generate some http traffic
		for _, config := range expReqs {
			for i := 0; i < config.NumberRequests; i++ {
				r, err := http.NewRequest(config.Method, s.URL()+config.Path, nil)
				require.NoError(err)
				resp, err := http.DefaultClient.Do(r)
				require.NoError(err)

				assert.Equal(config.Code, resp.StatusCode)
				b, err := io.ReadAll(resp.Body)
				require.NoError(err)
				assert.Equal(config.ReturnData, string(b))
			}
		}

		// assert metrics has been created included the previous generated http requests
		r, err := http.NewRequest(http.MethodGet, "http://localhost:8080/metrics", nil)
		require.NoError(err)
		resp, err := http.DefaultClient.Do(r)
		require.NoError(err)

		b, err := io.ReadAll(resp.Body)
		require.NoError(err)
		metrics := string(b)

		for _, expMetric := range expMetrics {
			assert.Contains(metrics, expMetric)
		}
	})
}
