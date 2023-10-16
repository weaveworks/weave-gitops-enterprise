package profiling

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPprofServer(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	t.Run("should create new pprof server", func(t *testing.T) {

		NewPprofServer(Options{
			ServerAddress: "localhost:8080",
		})

		// Get metrics.
		r, err := http.NewRequest(http.MethodGet, "http://localhost:8080/debug/pprof", nil)
		require.NoError(err)
		resp, err := http.DefaultClient.Do(r)
		require.NoError(err)

		// Check.
		b, err := io.ReadAll(resp.Body)
		require.NoError(err)
		metrics := string(b)

		assert.Contains(metrics, "Types of profiles available")
	})
}
