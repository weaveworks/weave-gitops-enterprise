package api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/wks/cmd/gitops-repo-broker/internal/handlers/api"
	"github.com/weaveworks/wks/common/database/models"
	"github.com/weaveworks/wks/common/database/utils"
	"gorm.io/gorm"
)

const dbTestName = "test.db"

func TestListClusters_NilDB(t *testing.T) {
	response := executeGet(t, nil, json.MarshalIndent, "")
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "{\"message\":\"The database has not been initialised.\"}\n", response.Body.String())
}

func TestListClusters_NoTables(t *testing.T) {
	db, err := utils.Open("")
	defer os.Remove(dbTestName)
	assert.NoError(t, err)

	response := executeGet(t, db, json.MarshalIndent, "")
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "{\"message\":\"The database has not been initialised.\"}\n", response.Body.String())
}

func TestListClusters_JSONError(t *testing.T) {
	db, err := utils.Open("")
	defer os.Remove(dbTestName)
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	jsonError := func(v interface{}, prefix, indent string) ([]byte, error) {
		return nil, errors.New("oops")
	}

	// No data
	response := executeGet(t, db, jsonError, "")
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "{\"message\":\"oops\"}\n", response.Body.String())

	// With data
	db.Create(&models.Cluster{Name: "MyCluster"})
	response = executeGet(t, db, jsonError, "")
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "{\"message\":\"oops\"}\n", response.Body.String())
}

func TestListClusters(t *testing.T) {
	db, err := utils.Open(dbTestName)
	defer os.Remove(dbTestName)
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	// No data
	response := executeGet(t, db, json.MarshalIndent, "")
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "{\n \"clusters\": []\n}", response.Body.String())

	// Register a cluster
	db.Create(&models.Cluster{Name: "My Cluster", Token: "derp"})
	response = executeGet(t, db, json.MarshalIndent, "")
	assert.Equal(t, http.StatusOK, response.Code)
	var res api.ClustersResponse
	err = json.Unmarshal(response.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.Equal(t, api.ClustersResponse{
		Clusters: []api.Cluster{
			{
				Name: "My Cluster",
				Type: "",
			},
		},
	}, res)

	// Agent sends cluster info
	db.Create(&models.ClusterInfo{
		UID:   "123",
		Token: "derp",
		Type:  "existingInfra",
	})
	db.Create(&models.NodeInfo{
		UID:            "456",
		ClusterInfoUID: "123",
		Token:          "derp",
		Name:           "wks-1",
		IsControlPlane: true,
		KubeletVersion: "v1.19.7",
	})
	db.Create(&models.NodeInfo{
		UID:            "789",
		ClusterInfoUID: "123",
		Token:          "derp",
		Name:           "wks-2",
		IsControlPlane: false,
		KubeletVersion: "v1.19.7",
	})
	response = executeGet(t, db, json.MarshalIndent, "")
	assert.Equal(t, http.StatusOK, response.Code)
	err = json.Unmarshal(response.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.Equal(t, api.ClustersResponse{
		Clusters: []api.Cluster{
			{
				Name: "My Cluster",
				Type: "existingInfra",
				Nodes: []api.Node{
					{
						Name:           "wks-1",
						IsControlPlane: true,
						KubeletVersion: "v1.19.7",
					},
					{
						Name:           "wks-2",
						IsControlPlane: false,
						KubeletVersion: "v1.19.7",
					},
				},
			},
		},
	}, res)
}

func executeGet(t *testing.T, db *gorm.DB, fn api.MarshalIndent, url string) *httptest.ResponseRecorder {
	req, err := http.NewRequest("GET", url, nil)
	require.Nil(t, err)
	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(api.ListClusters(db, fn))
	handler.ServeHTTP(rec, req)
	return rec
}

func TestRegisterCluster_NilDB(t *testing.T) {
	response := executePost(t, nil, nil, nil, nil, nil)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "{\"message\":\"The database has not been initialised.\"}\n", response.Body.String())
}

func TestRegisterCluster_IOError(t *testing.T) {
	db, err := utils.Open("")
	defer os.Remove(dbTestName)
	assert.NoError(t, err)

	response := executePost(t, FakeErrorReader{}, db, json.Unmarshal, json.MarshalIndent, nil)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "{\"message\":\"oops\"}\n", response.Body.String())
}

func TestRegisterCluster_JSONError(t *testing.T) {
	db, err := utils.Open("")
	defer os.Remove(dbTestName)
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	unmarshalFnError := func(data []byte, v interface{}) error {
		return errors.New("unmarshal error")
	}

	marshalIndentFnError := func(v interface{}, prefix, indent string) ([]byte, error) {
		return nil, errors.New("marshal error")
	}

	// Request body
	data, _ := json.MarshalIndent(api.ClusterRegistrationRequest{
		Name: "derp",
	}, "", " ")

	// Unmarshal error
	response := executePost(t, bytes.NewReader(data), db, unmarshalFnError, nil, nil)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "{\"message\":\"unmarshal error\"}\n", response.Body.String())

	// MarshalIndent error
	response = executePost(t, bytes.NewReader(data), db, json.Unmarshal, marshalIndentFnError, api.Generate)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "{\"message\":\"marshal error\"}\n", response.Body.String())
}

func TestRegisterCluster_TokenGenerationError(t *testing.T) {
	db, err := utils.Open("")
	defer os.Remove(dbTestName)
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	// Request body
	data, _ := json.MarshalIndent(api.ClusterRegistrationRequest{
		Name: "derp",
	}, "", " ")
	response := executePost(t, bytes.NewReader(data), db, json.Unmarshal, json.MarshalIndent, NewFakeTokenGenerator("", errors.New("error generating token")).Generate)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "{\"message\":\"error generating token\"}\n", response.Body.String())
}

func TestRegisterCluster_ValidateRequestBody(t *testing.T) {
	db, err := utils.Open("")
	defer os.Remove(dbTestName)
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	// Request body
	data, _ := json.MarshalIndent(api.ClusterRegistrationRequest{
		Name:       "derp",
		IngressURL: "not a url",
	}, "", " ")
	response := executePost(t, bytes.NewReader(data), db, json.Unmarshal, json.MarshalIndent, NewFakeTokenGenerator("fake token", nil).Generate)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "{\"message\":\"Invalid payload\"}\n", response.Body.String())
}

func TestRegisterCluster(t *testing.T) {
	db, err := utils.Open("")
	defer os.Remove(dbTestName)
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	// Request body
	data, _ := json.MarshalIndent(api.ClusterRegistrationRequest{
		Name:       "derp",
		IngressURL: "http://localhost:8000/ui",
	}, "", " ")
	response := executePost(t, bytes.NewReader(data), db, json.Unmarshal, json.MarshalIndent, NewFakeTokenGenerator("fake token", nil).Generate)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "{\n \"name\": \"derp\",\n \"ingressUrl\": \"http://localhost:8000/ui\",\n \"token\": \"fake token\"\n}", response.Body.String())
}

func executePost(t *testing.T, r io.Reader, db *gorm.DB, unmarshalFn api.Unmarshal, marshalFn api.MarshalIndent, generateTokenFn api.GenerateToken) *httptest.ResponseRecorder {
	req, err := http.NewRequest("POST", "", r)
	require.Nil(t, err)
	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(api.RegisterCluster(db, validator.New(), unmarshalFn, marshalFn, generateTokenFn))
	handler.ServeHTTP(rec, req)
	return rec
}

type FakeErrorReader struct {
}

func (r FakeErrorReader) Read(b []byte) (n int, err error) {
	return 0, errors.New("oops")
}

type FakeTokenGenerator struct {
	token string
	err   error
}

func NewFakeTokenGenerator(token string, err error) FakeTokenGenerator {
	return FakeTokenGenerator{
		token: token,
		err:   err,
	}
}

func (f FakeTokenGenerator) Generate() (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.token, nil
}
