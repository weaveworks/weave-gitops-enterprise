package profiling

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultPprofHandler(t *testing.T) {

	t.Run("should create pprof handler", func(t *testing.T) {
		_, h := NewDefaultPprofHandler()

		ts := httptest.NewServer(h)
		defer ts.Close()

		resp, err := http.Get(ts.URL)
		require.NoError(t, err)
		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		pprofPage := string(b)

		assert.Contains(t, pprofPage, "<title>/debug/pprof/</title>")
	})
}
