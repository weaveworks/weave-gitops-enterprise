package liveness

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/wks/common/database/utils"
)

func TestStartedHandler(t *testing.T) {
	handler := startedHandler(time.Now())
	req := httptest.NewRequest("GET", "http://example.com/started", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, resp.StatusCode, http.StatusOK)
	started, err := time.ParseDuration(string(body))
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), time.Now().Add(started), time.Millisecond)
}

func TestHealthzHandler(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	handler := healthzHandler()
	req := httptest.NewRequest("GET", "http://example.com/healthz", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, resp.StatusCode, http.StatusOK)
	assert.Equal(t, string(body), "ok")
}

func TestHealthzHandler_DatabaseNotReady(t *testing.T) {
	_, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)

	handler := healthzHandler()
	req := httptest.NewRequest("GET", "http://example.com/healthz", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, resp.StatusCode, http.StatusInternalServerError)
	assert.Equal(t, string(body), "database is not ready")
}
