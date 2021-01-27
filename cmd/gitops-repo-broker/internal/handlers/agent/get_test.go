package agent

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	encodedToken := base64.StdEncoding.EncodeToString([]byte("derp"))
	response := executeGet(t, "foo-nat-url", "/gitops/api/agent.yaml?token=derp")
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Contains(t, response.Body.String(), encodedToken)
	assert.Contains(t, response.Body.String(), "foo-nat-url")
}

func executeGet(t *testing.T, natsURL, url string) *httptest.ResponseRecorder {
	req, err := http.NewRequest("GET", url, nil)
	require.Nil(t, err)
	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(NewGetHandler(natsURL))
	handler.ServeHTTP(rec, req)
	return rec
}

func TestRenderTemplate(t *testing.T) {
	out, err := renderTemplate("foo", "bar-image", "nats-url-ewq")
	encodedToken := base64.StdEncoding.EncodeToString([]byte("foo"))
	require.NoError(t, err)
	assert.Contains(t, out, fmt.Sprintf("token: %s", encodedToken))
	assert.Contains(t, out, "image: weaveworks/wkp-agent:bar-image")
	assert.Contains(t, out, "--nats-url=nats-url-ewq")
}
