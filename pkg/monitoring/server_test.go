package monitoring

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/monitoring/metrics"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/monitoring/profiling"
)

func TestNewServer(t *testing.T) {
	t.Run("cannot create monitoring server without server address", func(t *testing.T) {
		_, err := NewServer(Options{Enabled: true, ServerAddress: ""})
		require.Error(t, err)
	})

	t.Run("can create server valid options", func(t *testing.T) {
		s, err := NewServer(Options{
			Enabled:       true,
			ServerAddress: "localhost:8080",
			MetricsOptions: metrics.Options{
				Enabled: true,
			},
			ProfilingOptions: profiling.Options{
				Enabled: true,
			},
		})
		require.NoError(t, err)
		defer func(s *http.Server, ctx context.Context) {
			err := s.Shutdown(ctx)
			require.NoError(t, err)
		}(s, context.Background())

		mockServer := httptest.NewServer(s.Handler)
		defer mockServer.Close()

		r, err := http.Get(mockServer.URL + "/metrics") // Adjust the URL path as needed
		require.NoError(t, err)
		require.Equal(t, r.StatusCode, http.StatusOK)

		r, err = http.Get(mockServer.URL + "/debug/pprof") // Adjust the URL path as needed
		require.NoError(t, err)
		require.Equal(t, r.StatusCode, http.StatusOK)
	})

}
