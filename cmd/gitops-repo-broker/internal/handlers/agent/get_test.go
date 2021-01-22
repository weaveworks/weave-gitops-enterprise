package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	stream := "derp"

	req, err := http.NewRequest("GET", "/gitops/api/agent.yaml?token=derp", nil)
	require.Nil(t, err)

	response := executeRequest(req)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Contains(t, response.Body.String(), stream)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(Get)
	handler.ServeHTTP(rec, req)

	return rec
}
