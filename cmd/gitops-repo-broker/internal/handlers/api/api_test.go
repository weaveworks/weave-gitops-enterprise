package api

import (
	"encoding/json"
	"github.com/weaveworks/wks/common/database/models"
	"github.com/weaveworks/wks/common/database/utils"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const dbTestName = "test.db"

func TestGetCluster(t *testing.T) {
	db, err := utils.Open(dbTestName)
	defer os.Remove(dbTestName)
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	// No data
	response := executeGet(t, dbTestName, "")
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "[]", response.Body.String())

	// Pop a cluster in
	myCluster := models.Cluster{Name: "My Cluster"}
	db.Create(&myCluster)
	response = executeGet(t, dbTestName, "")
	assert.Equal(t, http.StatusOK, response.Code)
	var clusters []models.Cluster
	err = json.Unmarshal(response.Body.Bytes(), &clusters)
	assert.NoError(t, err)
	assert.Equal(t, "My Cluster", clusters[0].Name)
}

func executeGet(t *testing.T, dbURI, url string) *httptest.ResponseRecorder {
	req, err := http.NewRequest("GET", url, nil)
	require.Nil(t, err)
	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(NewGetClusters(dbURI))
	handler.ServeHTTP(rec, req)
	return rec
}
