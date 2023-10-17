package monitoring

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/monitoring/metrics"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/monitoring/profiling"
)

func TestNewServer(t *testing.T) {
	require := require.New(t)

	t.Run("cannot create management server without valid options", func(t *testing.T) {
		_, err := NewServer(Options{})
		require.Error(err)

		_, err = NewServer(Options{Enabled: false})
		require.Error(err)

		_, err = NewServer(Options{Enabled: true, ServerAddress: ""})
		require.Error(err)
	})

	t.Run("can create server valid options", func(t *testing.T) {
		_, err := NewServer(Options{
			Enabled:       true,
			ServerAddress: "localhost:8080",
			MetricsOptions: metrics.Options{
				Enabled: true,
			},
			ProfilingOptions: profiling.Options{
				Enabled: true,
			},
		})
		require.NoError(err)

		r, err := http.NewRequest(http.MethodGet, "http://localhost:8080/metrics", nil)
		require.NoError(err)
		_, err = http.DefaultClient.Do(r)
		require.NoError(err)

		r, err = http.NewRequest(http.MethodGet, "http://localhost:8080/debug/pprof", nil)
		require.NoError(err)
		_, err = http.DefaultClient.Do(r)
		require.NoError(err)
	})

}
