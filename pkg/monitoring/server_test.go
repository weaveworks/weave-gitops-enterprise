package monitoring

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	require := require.New(t)

	t.Run("cannot create management server without valid options", func(t *testing.T) {
		_, err := NewServer(Options{}, nil)
		require.Error(err)

		_, err = NewServer(Options{Enabled: false}, nil)
		require.Error(err)

		_, err = NewServer(Options{Enabled: true, ServerAddress: ""}, nil)
		require.Error(err)

		_, err = NewServer(Options{Enabled: true, ServerAddress: "localhost:9090"}, nil)
		require.Error(err)

	})

	t.Run("can create server valid options", func(t *testing.T) {
		handlers := map[string]http.Handler{"/test": stringHandler("this is my test")}
		_, err := NewServer(Options{
			Enabled:       true,
			ServerAddress: "localhost:8080",
		}, handlers)
		require.NoError(err)

		r, err := http.NewRequest(http.MethodGet, "http://localhost:8080/test", nil)
		require.NoError(err)
		resp, err := http.DefaultClient.Do(r)
		require.NoError(err)

		// Check.
		b, err := io.ReadAll(resp.Body)
		require.NoError(err)
		msg := string(b)

		require.Contains(msg, "this is my test")
	})

}

type stringHandler string

func (s stringHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(s))
}
