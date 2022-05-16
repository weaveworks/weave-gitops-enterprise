package agent

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/models"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
	"gorm.io/gorm"
)

func TestGet(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	db.Create(&models.Cluster{Name: "My Cluster", Token: "derp"})

	encodedToken := base64.StdEncoding.EncodeToString([]byte("derp"))
	response := executeGet(t, db, "foo-nat-url", "foo-am-url", "/gitops/api/agent.yaml?token=derp")
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Contains(t, response.Body.String(), encodedToken)
	assert.Contains(t, response.Body.String(), "foo-nat-url")
	assert.Contains(t, response.Body.String(), "foo-am-url")

	amURL := "http://example.com:9090"
	response = executeGet(t, db, "foo-nat-url", "foo-am-url", fmt.Sprintf("/gitops/api/agent.yaml?token=derp&alertmanagerURL=%v", amURL))
	assert.Contains(t, response.Body.String(), amURL)

	escapedURL := url.QueryEscape(amURL)
	response = executeGet(t, db, "foo-nat-url", "foo-am-url", fmt.Sprintf("/gitops/api/agent.yaml?token=derp&alertmanagerURL=%v", escapedURL))
	assert.Contains(t, response.Body.String(), amURL)
}

func TestGet_ClusterNotFound(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	db.Create(&models.Cluster{Name: "My Cluster", Token: "notderp"})

	response := executeGet(t, db, "foo-nat-url", "foo-am-url", "/gitops/api/agent.yaml?token=derp")
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func executeGet(t *testing.T, db *gorm.DB, natsURL, alertmanagerURL, url string) *httptest.ResponseRecorder {
	req, err := http.NewRequest("GET", url, nil)
	require.Nil(t, err)
	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(NewGetHandler(db, natsURL, alertmanagerURL, ""))
	handler.ServeHTTP(rec, req)
	return rec
}

func TestRenderTemplate(t *testing.T) {
	out, err := renderTemplate("foo", "bar-image", "nats-url-ewq", "am-url-ewq")
	encodedToken := base64.StdEncoding.EncodeToString([]byte("foo"))
	require.NoError(t, err)
	assert.Contains(t, out, fmt.Sprintf("token: %s", encodedToken))
	assert.Contains(t, out, "image: weaveworks/wkp-agent:bar-image")
	assert.Contains(t, out, "--nats-url=nats-url-ewq")
	assert.Contains(t, out, "--alertmanager-url=am-url-ewq")
}
