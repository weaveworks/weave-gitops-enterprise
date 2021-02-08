package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/wks/common/database/models"
	"github.com/weaveworks/wks/common/database/utils"
	"gorm.io/gorm"
)

const dbTestName = "test.db"

func TestGetCluster_NilDB(t *testing.T) {
	response := executeGet(t, nil, json.MarshalIndent, "")
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "{\"message\":\"The database has not been initialised.\"}\n", response.Body.String())
}

func TestGetCluster_NoTables(t *testing.T) {
	db, err := utils.Open("")
	defer os.Remove(dbTestName)
	assert.NoError(t, err)

	response := executeGet(t, db, json.MarshalIndent, "")
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "{\"message\":\"The database has not been initialised.\"}\n", response.Body.String())
}

func TestGetCluster_JSONError(t *testing.T) {
	db, err := utils.Open("")
	defer os.Remove(dbTestName)
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	jsonError := func(v interface{}, prefix, indent string) ([]byte, error) {
		return nil, errors.New("oops")
	}

	response := executeGet(t, db, jsonError, "")
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "{\"message\":\"oops\"}\n", response.Body.String())
}

func TestGetCluster(t *testing.T) {
	db, err := utils.Open(dbTestName)
	defer os.Remove(dbTestName)
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	// No data
	response := executeGet(t, db, json.MarshalIndent, "")
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "[]", response.Body.String())

	// Pop a cluster in
	myCluster := models.Cluster{Name: "My Cluster"}
	db.Create(&myCluster)
	response = executeGet(t, db, json.MarshalIndent, "")
	assert.Equal(t, http.StatusOK, response.Code)
	var clusters []models.Cluster
	err = json.Unmarshal(response.Body.Bytes(), &clusters)
	assert.NoError(t, err)
	assert.Equal(t, "My Cluster", clusters[0].Name)
}

func executeGet(t *testing.T, db *gorm.DB, fn MarshalIndent, url string) *httptest.ResponseRecorder {
	req, err := http.NewRequest("GET", url, nil)
	require.Nil(t, err)
	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(NewGetClusters(db, fn))
	handler.ServeHTTP(rec, req)
	return rec
}
