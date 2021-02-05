package healthcheck

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var started = time.Now()

func TestStarted(t *testing.T) {
	req, err := http.NewRequest("GET", "/gitops/started", nil)
	require.Nil(t, err)

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(Started(started))
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHealthz(t *testing.T) {
	req, err := http.NewRequest("GET", "/gitops/healthz", nil)
	require.Nil(t, err)

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(Healthz(started))
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRedirect(t *testing.T) {
	req, err := http.NewRequest("GET", "/gitops/redirect", nil)
	require.Nil(t, err)

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(Redirect)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusFound, rec.Code)
}
