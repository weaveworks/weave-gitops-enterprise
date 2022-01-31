package api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"gorm.io/datatypes"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/gorilla/mux"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/handlers/api"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/handlers/api/views"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/models"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
	"gorm.io/gorm"
)

func errorBody(message string) interface{} {
	return map[string]interface{}{"message": message}
}

var noInit = errorBody("The database has not been initialised.")
var noSuchTable = errorBody("no such table: clusters")
var now = time.Now()

func assertEqualCmp(t *testing.T, want, got interface{}) {
	//
	// Use go-cmp to diff things, works for time.Time in an unmarshaled json struct
	// https://github.com/stretchr/testify/issues/502
	//
	// Using cmp.Diff like this gives quite nice output
	diff := cmp.Diff(want, got)
	assert.True(t, diff == "", diff)
}

func doRequest(t *testing.T, handler http.HandlerFunc, method, path, url, data string) (*httptest.ResponseRecorder, interface{}) {
	body := bytes.NewReader([]byte(data))
	req, err := http.NewRequest(method, url, body)
	require.Nil(t, err)
	rec := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc(path, handler)
	router.ServeHTTP(rec, req)
	if rec.Header().Get("Content-Type") == "application/json" {
		var res interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &res)
		require.NoError(t, err)
		return rec, res
	}
	return rec, rec.Body.String()
}

func TestNilDb(t *testing.T) {
	nilDbHandlers := []http.HandlerFunc{
		api.FindCluster(nil, nil),
		api.ListAlerts(nil, nil),
		api.ListClusters(nil, nil),
		api.RegisterCluster(nil, nil, nil, nil, nil),
		api.UpdateCluster(nil, nil, nil, nil),
		api.UnregisterCluster(nil),
	}
	for i, fn := range nilDbHandlers {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			response, body := doRequest(t, fn, "GET", "/", "/", "")
			assert.Equal(t, http.StatusInternalServerError, response.Code)
			assert.Equal(t, noInit, body)
		})
	}
}

func TestNoTables(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	assert.NoError(t, err)

	noTablesHandlers := []struct {
		handler http.HandlerFunc
		message interface{}
	}{
		{api.FindCluster(db, nil), noInit},
		{api.ListAlerts(db, nil), noInit},
		{api.ListClusters(db, nil), noInit},
		{
			api.RegisterCluster(db, validator.New(), json.Unmarshal, nil, NewFakeTokenGenerator("derp", nil).Generate),
			noSuchTable,
		},
		{api.UpdateCluster(db, validator.New(), json.Unmarshal, nil), noSuchTable},
	}

	for i, tt := range noTablesHandlers {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			response, body := doRequest(t, tt.handler, "GET", "/{id:[0-9]+}", "/1", `{ "id": 1, "name": "ewq" }`)
			assert.Equal(t, http.StatusInternalServerError, response.Code)
			assert.Equal(t, tt.message, body)
		})
	}
}

func TestJSONMarshalErrors(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)
	myCluster := models.Cluster{Name: "MyCluster"}
	result := db.Create(&myCluster)
	assert.NoError(t, result.Error)

	marshalError := func(v interface{}, prefix, indent string) ([]byte, error) {
		return nil, errors.New("oops")
	}

	unmarshallErrors := []struct {
		handler http.HandlerFunc
		data    string
	}{
		{api.FindCluster(db, marshalError), ""},
		{api.ListAlerts(db, marshalError), ""},
		{api.ListClusters(db, marshalError), ""},
		{api.RegisterCluster(db, validator.New(), json.Unmarshal, marshalError, NewFakeTokenGenerator("derp", nil).Generate), `{ "name": "ewq" }`},
		{api.UpdateCluster(db, validator.New(), json.Unmarshal, marshalError), `{ "id": 1, "name": "ewq2" }`},
	}

	for i, tt := range unmarshallErrors {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			response, body := doRequest(
				t,
				tt.handler,
				"GET",
				"/{id:[0-9]+}",
				fmt.Sprintf("/%d", myCluster.ID),
				tt.data,
			)
			assert.Equal(t, http.StatusInternalServerError, response.Code)
			assert.Equal(t, errorBody("oops"), body)
		})
	}
}

func TestGetCluster(t *testing.T) {
	requestTests := []struct {
		description  string
		path         string
		responseCode int
		response     interface{}
		clusters     []models.Cluster
	}{
		{"404 if no :id", "/", 404, "404 page not found\n", nil},
		{"404 if no cluster in db", "/1", 404, errorBody("cluster not found"), nil},
		{
			"200 if cluster is in db",
			"/1",
			200,
			map[string]interface{}{
				"id":         float64(1),
				"name":       "ewq",
				"token":      "",
				"type":       "",
				"ingressUrl": "",
				"status":     "notConnected",
				"updatedAt":  "0001-01-01T00:00:00Z",
			},
			[]models.Cluster{{Name: "ewq"}},
		},
		{
			"Get the correct cluster",
			"/2",
			200,
			map[string]interface{}{
				"id":         float64(2),
				"name":       "dsa",
				"token":      "dsa",
				"type":       "",
				"ingressUrl": "",
				"status":     "notConnected",
				"updatedAt":  "0001-01-01T00:00:00Z",
			},
			[]models.Cluster{{Name: "ewq", Token: "ewq"}, {Name: "dsa", Token: "dsa"}},
		},
	}

	for _, rt := range requestTests {
		t.Run(rt.description, func(t *testing.T) {
			db, err := utils.Open("", "sqlite", "", "", "")
			assert.NoError(t, err)
			err = utils.MigrateTables(db)
			assert.NoError(t, err)
			if rt.clusters != nil {
				result := db.Create(&rt.clusters)
				assert.NoError(t, result.Error)
			}
			response, body := doRequest(
				t,
				api.FindCluster(db, json.MarshalIndent),
				"GET",
				"/{id:[0-9]+}",
				rt.path,
				"",
			)
			assert.Equal(t, rt.responseCode, response.Code)
			assert.Equal(t, rt.response, body)
		})
	}
}

func TestUpdateCluster_ValidateRequestBody(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	// Request body
	data, _ := json.MarshalIndent(views.ClusterUpdateRequest{
		Name:       "derp",
		IngressURL: "not a url",
	}, "", " ")
	response := executeUpdate(t, bytes.NewReader(data), db, json.Unmarshal, json.MarshalIndent, NewFakeTokenGenerator("fake token", nil).Generate)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "{\"message\":\"Invalid payload, Error parsing 'url' in field 'Ingress URL'\"}\n", response.Body.String())

	data, _ = json.MarshalIndent(views.ClusterUpdateRequest{
		Name:       "derp",
		IngressURL: "",
	}, "", " ")
	response = executeUpdate(t, bytes.NewReader(data), db, json.Unmarshal, json.MarshalIndent, NewFakeTokenGenerator("fake token", nil).Generate)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "{\"message\":\"Invalid payload, Error parsing 'id' param from path\"}\n", response.Body.String())

}

func TestUpdateCluster(t *testing.T) {
	requestTests := []struct {
		description  string
		path         string
		data         interface{}
		responseCode int
		response     interface{}
		clusters     []models.Cluster
		getResponse  interface{}
	}{
		{"404 if no :id", "/", nil, 404, "404 page not found\n", nil, nil},
		{
			"404 if no cluster in db",
			"/1",
			map[string]interface{}{
				"name": "ewq1",
			},
			404,
			errorBody("cluster not found"),
			nil,
			nil,
		},
		{
			"200 if cluster is in db",
			"/1",
			map[string]interface{}{
				"name": "ewq1",
			},
			200,
			map[string]interface{}{
				"id":         float64(1),
				"name":       "ewq1",
				"token":      "ewq",
				"type":       "",
				"ingressUrl": "",
				"status":     "notConnected",
				"updatedAt":  "0001-01-01T00:00:00Z",
			},
			[]models.Cluster{{Name: "ewq", Token: "ewq"}},
			map[string]interface{}{
				"id":         float64(1),
				"name":       "ewq1",
				"token":      "ewq",
				"type":       "",
				"ingressUrl": "",
				"status":     "notConnected",
				"updatedAt":  "0001-01-01T00:00:00Z",
			},
		},
		{
			"Update the correct cluster",
			"/2",
			map[string]interface{}{
				"name": "dsa2",
			},
			200,
			map[string]interface{}{
				"id":         float64(2),
				"name":       "dsa2",
				"token":      "dsa",
				"type":       "",
				"ingressUrl": "",
				"status":     "notConnected",
				"updatedAt":  "0001-01-01T00:00:00Z",
			},
			[]models.Cluster{{Name: "ewq", Token: "ewq"}, {Name: "dsa", Token: "dsa"}},
			map[string]interface{}{
				"id":         float64(2),
				"name":       "dsa2",
				"token":      "dsa",
				"type":       "",
				"ingressUrl": "",
				"status":     "notConnected",
				"updatedAt":  "0001-01-01T00:00:00Z",
			},
		},
		{
			"Can't update token",
			"/2",
			map[string]interface{}{
				"name":  "dsa",
				"token": "newtoken",
			},
			200,
			map[string]interface{}{
				"id":         float64(2),
				"name":       "dsa",
				"token":      "dsa",
				"type":       "",
				"ingressUrl": "",
				"status":     "notConnected",
				"updatedAt":  "0001-01-01T00:00:00Z",
			},
			[]models.Cluster{{Name: "ewq", Token: "ewq"}, {Name: "dsa", Token: "dsa"}},
			map[string]interface{}{
				"id":         float64(2),
				"name":       "dsa",
				"token":      "dsa",
				"type":       "",
				"ingressUrl": "",
				"status":     "notConnected",
				"updatedAt":  "0001-01-01T00:00:00Z",
			},
		},
	}
	for _, rt := range requestTests {
		t.Run(rt.description, func(t *testing.T) {
			db, err := utils.Open("", "sqlite", "", "", "")
			assert.NoError(t, err)
			err = utils.MigrateTables(db)
			assert.NoError(t, err)
			if rt.clusters != nil {
				result := db.Create(&rt.clusters)
				assert.NoError(t, result.Error)
			}
			dataStr := ""
			if rt.data != nil {
				dataBytes, err := json.Marshal(rt.data)
				assert.NoError(t, err)
				dataStr = string(dataBytes)
			}
			response, body := doRequest(
				t,
				api.UpdateCluster(db, validator.New(), json.Unmarshal, json.MarshalIndent),
				"PUT",
				"/{id:[0-9]+}",
				rt.path,
				dataStr,
			)
			assert.Equal(t, rt.responseCode, response.Code)
			assert.Equal(t, rt.response, body)

			if rt.getResponse != nil {
				response, body := doRequest(
					t,
					api.FindCluster(db, json.MarshalIndent),
					"GET",
					"/{id:[0-9]+}",
					rt.path,
					"",
				)
				assert.Equal(t, 200, response.Code)
				assert.Equal(t, rt.getResponse, body)
			}
		})
	}
}

func TestListClusters(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
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
	var res views.ClustersResponse
	err = json.Unmarshal(response.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.Equal(t, views.ClustersResponse{
		Clusters: []views.ClusterView{
			{
				ID:       1,
				Name:     "My Cluster",
				Token:    "derp",
				Status:   "notConnected",
				Type:     "",
				FluxInfo: nil,
			},
		},
	}, res)

	// Agent sends cluster info
	db.Create(&models.ClusterInfo{
		UID:          "123",
		ClusterToken: "derp",
		UpdatedAt:    now,
		Type:         "existingInfra",
	})
	db.Create(&models.NodeInfo{
		UID:            "456",
		ClusterInfoUID: "123",
		ClusterToken:   "derp",
		Name:           "wks-1",
		IsControlPlane: true,
		KubeletVersion: "v1.19.7",
	})
	db.Create(&models.NodeInfo{
		UID:            "789",
		ClusterInfoUID: "123",
		ClusterToken:   "derp",
		Name:           "wks-2",
		IsControlPlane: false,
		KubeletVersion: "v1.19.7",
	})
	db.Create(&models.FluxInfo{
		ClusterToken: "derp",
		Name:         "flux",
		Namespace:    "wkp-flux",
		Args:         "--memcached-service=,--ssh-keygen-dir=/var/fluxd/keygen,--sync-garbage-collection=true,--git-poll-interval=10s,--sync-interval=10s,--manifest-generation=true,--listen-metrics=:3031,--git-url=git@github.com:weaveworks/fluxes-1.git,--git-branch=master,--registry-exclude-image=*",
		Image:        "docker.io/weaveworks/wkp-jk-init:v2.0.3-RC.1-2-gd677dc0a",
		RepoURL:      "git@github.com:weaveworks/fluxes-1.git",
		RepoBranch:   "master",
	})
	db.Create(&models.Workspace{
		ClusterToken: "derp",
		Name:         "foo",
		Namespace:    "wkp-workspaces",
	})

	response = executeGet(t, db, json.MarshalIndent, "")
	assert.Equal(t, http.StatusOK, response.Code)
	err = json.Unmarshal(response.Body.Bytes(), &res)
	assert.NoError(t, err)
	assertEqualCmp(t, views.ClustersResponse{
		Clusters: []views.ClusterView{
			{
				ID:        1,
				Name:      "My Cluster",
				Token:     "derp",
				Type:      "existingInfra",
				Status:    "ready",
				UpdatedAt: now,
				Nodes: []views.NodeView{
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
				FluxInfo: []views.FluxInfoView{
					{
						Name:       "flux",
						Namespace:  "wkp-flux",
						RepoURL:    "git@github.com:weaveworks/fluxes-1.git",
						RepoBranch: "master",
						LogInfo:    datatypes.JSON{'n', 'u', 'l', 'l'},
					},
				},
				Workspaces: []views.WorkspaceView{
					{
						Name:      "foo",
						Namespace: "wkp-workspaces",
					},
				},
			},
		},
	}, res)
}

func TestListClustersSorting(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	// No data
	response := executeGet(t, db, json.MarshalIndent, "")
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "{\n \"clusters\": []\n}", response.Body.String())

	defaultOrder := []views.ClusterView{
		{
			ID:       1,
			Name:     "My Cluster",
			Token:    "derp",
			Status:   "notConnected",
			Type:     "",
			FluxInfo: nil,
		},
		{
			ID:       4,
			Name:     "ec2-cluster",
			Token:    "nerp",
			Status:   "notConnected",
			Type:     "",
			FluxInfo: nil,
		},
		{
			ID:       2,
			Name:     "gcp-cluster",
			Token:    "berp",
			Status:   "notConnected",
			Type:     "",
			FluxInfo: nil,
		},
		{
			ID:       3,
			Name:     "packet-cluster",
			Token:    "merp",
			Status:   "notConnected",
			Type:     "",
			FluxInfo: nil,
		},
	}

	tests := []struct {
		clusters       []models.Cluster
		sortQuery      string
		expectedResult []views.ClusterView
		reversed       bool
	}{
		{
			clusters: []models.Cluster{
				{
					Name:  "My Cluster",
					Token: "derp",
				},
				{
					Name:  "gcp-cluster",
					Token: "berp",
				},
				{
					Name:  "packet-cluster",
					Token: "merp",
				},
				{
					Name:  "ec2-cluster",
					Token: "nerp",
				},
			},
			sortQuery:      "",
			expectedResult: defaultOrder,
			reversed:       false,
		},
		{
			clusters:       []models.Cluster{},
			sortQuery:      "/clusters?sortBy=Name&order=ASC",
			expectedResult: defaultOrder,
			reversed:       false,
		},
		{
			clusters:       []models.Cluster{},
			sortQuery:      "/clusters?sortBy=Name&order=DESC",
			expectedResult: defaultOrder,
			reversed:       true,
		},
		{
			clusters:  []models.Cluster{},
			sortQuery: "/clusters?sortBy=Token&order=ASC",
			expectedResult: []views.ClusterView{
				{
					ID:       2,
					Name:     "gcp-cluster",
					Token:    "berp",
					Status:   "notConnected",
					Type:     "",
					FluxInfo: nil,
				},
				{
					ID:       1,
					Name:     "My Cluster",
					Token:    "derp",
					Status:   "notConnected",
					Type:     "",
					FluxInfo: nil,
				},
				{
					ID:       3,
					Name:     "packet-cluster",
					Token:    "merp",
					Status:   "notConnected",
					Type:     "",
					FluxInfo: nil,
				},
				{
					ID:       4,
					Name:     "ec2-cluster",
					Token:    "nerp",
					Status:   "notConnected",
					Type:     "",
					FluxInfo: nil,
				},
			},
		},
		{
			clusters: []models.Cluster{
				{
					Name:       "Thy Cluster",
					Token:      "perp",
					IngressURL: "http://googel",
				},
				{
					Name:       "Bye Cluster",
					Token:      "lerp",
					IngressURL: "http://googel",
				},
			},
			sortQuery: "/clusters?sortBy=IngressURL&order=ASC",
			expectedResult: []views.ClusterView{
				{
					ID:       1,
					Name:     "My Cluster",
					Token:    "derp",
					Status:   "notConnected",
					Type:     "",
					FluxInfo: nil,
				},
				{
					ID:       4,
					Name:     "ec2-cluster",
					Token:    "nerp",
					Status:   "notConnected",
					Type:     "",
					FluxInfo: nil,
				},
				{
					ID:       2,
					Name:     "gcp-cluster",
					Token:    "berp",
					Status:   "notConnected",
					Type:     "",
					FluxInfo: nil,
				},
				{
					ID:       3,
					Name:     "packet-cluster",
					Token:    "merp",
					Status:   "notConnected",
					Type:     "",
					FluxInfo: nil,
				},
				{
					ID:         6,
					Name:       "Bye Cluster",
					Token:      "lerp",
					Status:     "notConnected",
					Type:       "",
					FluxInfo:   nil,
					IngressURL: "http://googel",
				},
				{
					ID:         5,
					Name:       "Thy Cluster",
					Token:      "perp",
					Status:     "notConnected",
					Type:       "",
					FluxInfo:   nil,
					IngressURL: "http://googel",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.sortQuery, func(t *testing.T) {
			if len(test.clusters) > 0 {
				db.CreateInBatches(test.clusters, 100)
			}

			// Test query without any parameters
			response = executeGet(t, db, json.MarshalIndent, test.sortQuery)
			assert.Equal(t, http.StatusOK, response.Code)
			var res views.ClustersResponse
			err = json.Unmarshal(response.Body.Bytes(), &res)
			assert.NoError(t, err)

			// If the expected result is reused in a following test case, it'll now be reversed
			if test.reversed {
				reverseAny(test.expectedResult)
			}

			assert.Equal(t, views.ClustersResponse{
				Clusters: test.expectedResult,
			}, res)
		})
	}
}

func TestListCluster_MultipleFluxInfo(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
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
	var res views.ClustersResponse
	err = json.Unmarshal(response.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.Equal(t, views.ClustersResponse{
		Clusters: []views.ClusterView{
			{
				ID:       1,
				Name:     "My Cluster",
				Token:    "derp",
				Status:   "notConnected",
				Type:     "",
				FluxInfo: nil,
			},
		},
	}, res)

	db.Create(&models.FluxInfo{
		ClusterToken: "derp",
		Name:         "flux",
		Namespace:    "wkp-flux",
		Args:         "--memcached-service=,--ssh-keygen-dir=/var/fluxd/keygen,--sync-garbage-collection=true,--git-poll-interval=10s,--sync-interval=10s,--manifest-generation=true,--listen-metrics=:3031,--git-url=git@github.com:weaveworks/fluxes-1.git,--git-branch=master,--registry-exclude-image=*",
		Image:        "docker.io/weaveworks/wkp-jk-init:v2.0.3-RC.1-2-gd677dc0a",
		RepoURL:      "git@github.com:weaveworks/fluxes-1.git",
		RepoBranch:   "master",
		Syncs:        datatypes.JSON{'n', 'u', 'l', 'l'},
	})
	db.Create(&models.FluxInfo{
		ClusterToken: "derp",
		Name:         "flux-namespaced",
		Namespace:    "default",
		Args:         "--memcached-service=,--ssh-keygen-dir=/var/fluxd/keygen,--sync-garbage-collection=true,--git-poll-interval=10s,--sync-interval=10s,--manifest-generation=true,--listen-metrics=:3031,--git-url=git@github.com:weaveworks/fluxes-1.git,--git-branch=master,--registry-exclude-image=*",
		Image:        "docker.io/fluxcd/flux:v0.8.1",
		RepoURL:      "git@github.com:weaveworks/fluxes-2.git",
		RepoBranch:   "dev",
		Syncs:        datatypes.JSON{'n', 'u', 'l', 'l'},
	})
	db.Create(&models.FluxInfo{
		ClusterToken: "derp",
		Name:         "flux-system",
		Namespace:    "kube-system",
		Args:         "--memcached-service=,--ssh-keygen-dir=/var/fluxd/keygen,--sync-garbage-collection=true,--git-poll-interval=10s,--sync-interval=10s,--manifest-generation=true,--listen-metrics=:3031,--git-url=git@github.com:weaveworks/fluxes-1.git,--git-branch=master,--registry-exclude-image=*",
		Image:        "docker.io/fluxcd/flux:v0.8.1",
		RepoURL:      "git@github.com:weaveworks/fluxes-3.git",
		RepoBranch:   "main",
		Syncs:        datatypes.JSON{'n', 'u', 'l', 'l'},
	})

	response = executeGet(t, db, json.MarshalIndent, "")
	assert.Equal(t, http.StatusOK, response.Code)
	err = json.Unmarshal(response.Body.Bytes(), &res)
	assert.NoError(t, err)
	assertEqualCmp(t, views.ClustersResponse{
		Clusters: []views.ClusterView{
			{
				ID:     1,
				Name:   "My Cluster",
				Token:  "derp",
				Status: "notConnected",
				FluxInfo: []views.FluxInfoView{
					{
						Name:       "flux",
						Namespace:  "wkp-flux",
						RepoURL:    "git@github.com:weaveworks/fluxes-1.git",
						RepoBranch: "master",
						LogInfo:    datatypes.JSON{'n', 'u', 'l', 'l'},
					},
					{
						Name:       "flux-namespaced",
						Namespace:  "default",
						RepoURL:    "git@github.com:weaveworks/fluxes-2.git",
						RepoBranch: "dev",
						LogInfo:    datatypes.JSON{'n', 'u', 'l', 'l'},
					},
					{
						Name:       "flux-system",
						Namespace:  "kube-system",
						RepoURL:    "git@github.com:weaveworks/fluxes-3.git",
						RepoBranch: "main",
						LogInfo:    datatypes.JSON{'n', 'u', 'l', 'l'},
					},
				},
			},
		},
	}, res)
}

func TestListCluster_MultipleWorkspaces(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
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
	var res views.ClustersResponse
	err = json.Unmarshal(response.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.Equal(t, views.ClustersResponse{
		Clusters: []views.ClusterView{
			{
				ID:       1,
				Name:     "My Cluster",
				Token:    "derp",
				Status:   "notConnected",
				Type:     "",
				FluxInfo: nil,
			},
		},
	}, res)

	db.Create(&models.Workspace{
		ClusterToken: "derp",
		Name:         "ws-1",
		Namespace:    "wkp-workspaces",
	})
	db.Create(&models.Workspace{
		ClusterToken: "derp",
		Name:         "ws-2",
		Namespace:    "wkp-workspaces",
	})
	db.Create(&models.Workspace{
		ClusterToken: "derp",
		Name:         "ws-3",
		Namespace:    "wkp-workspaces",
	})

	response = executeGet(t, db, json.MarshalIndent, "")
	assert.Equal(t, http.StatusOK, response.Code)
	err = json.Unmarshal(response.Body.Bytes(), &res)
	assert.NoError(t, err)
	assertEqualCmp(t, views.ClustersResponse{
		Clusters: []views.ClusterView{
			{
				ID:     1,
				Name:   "My Cluster",
				Token:  "derp",
				Status: "notConnected",
				Workspaces: []views.WorkspaceView{
					{
						Name:      "ws-1",
						Namespace: "wkp-workspaces",
					},
					{
						Name:      "ws-2",
						Namespace: "wkp-workspaces",
					},
					{
						Name:      "ws-3",
						Namespace: "wkp-workspaces",
					},
				},
			},
		},
	}, res)
}

func TestListClusters_Pagination(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	assert.NoError(t, err)

	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	var clusters []*models.Cluster
	for i := 1; i <= 25; i++ {
		token := fmt.Sprintf("Token-%d", i)
		c := &models.Cluster{Name: fmt.Sprintf("Cluster %3d", i), Token: token}
		db.Create(c)
		ci := &models.ClusterInfo{UID: uuid.NewUUID(), ClusterToken: token, Type: "aws"}
		db.Create(ci)
		for k := 1; k <= 3; k++ {
			ni := &models.NodeInfo{UID: uuid.NewUUID(), ClusterToken: token, IsControlPlane: true, ClusterInfo: *ci}
			db.Create(ni)
		}
		for k := 1; k <= 5; k++ {
			ni := &models.NodeInfo{UID: uuid.NewUUID(), ClusterToken: token, IsControlPlane: false, ClusterInfo: *ci}
			db.Create(ni)
		}
		clusters = append(clusters, c)
	}

	// Default pagination size is 10
	response := executeGet(t, db, json.MarshalIndent, "/clusters")
	assert.Equal(t, http.StatusOK, response.Code)
	var res views.ClustersResponse
	err = json.Unmarshal(response.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(res.Clusters))
	assert.Equal(t, getIDsFromModels(clusters[0:10]), getIDsFromViews(res.Clusters))
	assert.Equal(t, response.Result().Header["Content-Range"][0], "clusters */25")

	response = executeGet(t, db, json.MarshalIndent, "/clusters?per_page=5&page=1")
	assert.Equal(t, http.StatusOK, response.Code)
	_ = json.Unmarshal(response.Body.Bytes(), &res)
	assert.Equal(t, 10, len(res.Clusters))
	assert.Equal(t, getIDsFromModels(clusters[0:10]), getIDsFromViews(res.Clusters))
	assert.Equal(t, response.Result().Header["Content-Range"][0], "clusters */25")

	response = executeGet(t, db, json.MarshalIndent, "/clusters?per_page=15&page=1")
	assert.Equal(t, http.StatusOK, response.Code)
	_ = json.Unmarshal(response.Body.Bytes(), &res)
	assert.Equal(t, 15, len(res.Clusters))
	assert.Equal(t, getIDsFromModels(clusters[0:15]), getIDsFromViews(res.Clusters))
	assert.Equal(t, response.Result().Header["Content-Range"][0], "clusters */25")

	response = executeGet(t, db, json.MarshalIndent, "/clusters?per_page=15&page=2")
	assert.Equal(t, http.StatusOK, response.Code)
	_ = json.Unmarshal(response.Body.Bytes(), &res)
	assert.Equal(t, 10, len(res.Clusters))
	assert.Equal(t, getIDsFromModels(clusters[15:25]), getIDsFromViews(res.Clusters))
	assert.Equal(t, response.Result().Header["Content-Range"][0], "clusters */25")

	response = executeGet(t, db, json.MarshalIndent, "/clusters?per_page=15&page=3")
	assert.Equal(t, http.StatusOK, response.Code)
	_ = json.Unmarshal(response.Body.Bytes(), &res)
	assert.Equal(t, 0, len(res.Clusters))
	assert.Equal(t, response.Result().Header["Content-Range"][0], "clusters */25")

	response = executeGet(t, db, json.MarshalIndent, "/clusters?page=0")
	assert.Equal(t, http.StatusOK, response.Code)
	_ = json.Unmarshal(response.Body.Bytes(), &res)
	assert.Equal(t, 10, len(res.Clusters))
	assert.Equal(t, getIDsFromModels(clusters[0:10]), getIDsFromViews(res.Clusters))
	assert.Equal(t, response.Result().Header["Content-Range"][0], "clusters */25")
}

func getIDsFromModels(clusters []*models.Cluster) []int {
	var ids []int
	for _, c := range clusters {
		ids = append(ids, int(c.ID))
	}
	return ids
}

func getIDsFromViews(views []views.ClusterView) []int {
	var ids []int

	for _, v := range views {
		ids = append(ids, int(v.ID))
	}

	return ids
}

func executeGet(t *testing.T, db *gorm.DB, fn api.MarshalIndent, url string) *httptest.ResponseRecorder {
	req, err := http.NewRequest("GET", url, nil)
	require.Nil(t, err)
	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(api.ListClusters(db, fn))
	handler.ServeHTTP(rec, req)
	return rec
}

func TestRegisterCluster_IOError(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	assert.NoError(t, err)

	response := executePost(t, FakeErrorReader{}, db, json.Unmarshal, json.MarshalIndent, nil)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "{\"message\":\"oops\"}\n", response.Body.String())
}

func TestRegisterCluster_JSONError(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
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
	data, _ := json.MarshalIndent(views.ClusterRegistrationRequest{
		Name: "derp",
	}, "", " ")

	// Unmarshal error
	response := executePost(t, bytes.NewReader(data), db, unmarshalFnError, nil, nil)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "{\"message\":\"unmarshal error\"}\n", response.Body.String())

	// MarshalIndent error
	response = executePost(t, bytes.NewReader(data), db, json.Unmarshal, marshalIndentFnError, utils.Generate)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "{\"message\":\"marshal error\"}\n", response.Body.String())
}

func TestRegisterCluster_TokenGenerationError(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	// Request body
	data, _ := json.MarshalIndent(views.ClusterRegistrationRequest{
		Name: "derp",
	}, "", " ")
	response := executePost(t, bytes.NewReader(data), db, json.Unmarshal, json.MarshalIndent, NewFakeTokenGenerator("", errors.New("error generating token")).Generate)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "{\"message\":\"error generating token\"}\n", response.Body.String())
}

func TestRegisterCluster_ValidateRequestBody(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	// Request body
	data, _ := json.MarshalIndent(views.ClusterRegistrationRequest{
		Name:       "derp",
		IngressURL: "not a url",
	}, "", " ")
	response := executePost(t, bytes.NewReader(data), db, json.Unmarshal, json.MarshalIndent, NewFakeTokenGenerator("fake token", nil).Generate)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "{\"message\":\"Invalid payload, Error parsing 'url' in field 'Ingress URL'\"}\n", response.Body.String())
}

func TestRegisterCluster(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	// Request body
	data, _ := json.MarshalIndent(views.ClusterRegistrationRequest{
		Name:       "derp",
		IngressURL: "http://localhost:8000/ui",
	}, "", " ")
	response := executePost(t, bytes.NewReader(data), db, json.Unmarshal, json.MarshalIndent, NewFakeTokenGenerator("fake token", nil).Generate)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "{\n \"id\": 1,\n \"name\": \"derp\",\n \"ingressUrl\": \"http://localhost:8000/ui\",\n \"token\": \"fake token\"\n}", response.Body.String())
}

func executePost(t *testing.T, r io.Reader, db *gorm.DB, unmarshalFn api.Unmarshal, marshalFn api.MarshalIndent, generateTokenFn api.GenerateToken) *httptest.ResponseRecorder {
	req, err := http.NewRequest("POST", "", r)
	require.Nil(t, err)
	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(api.RegisterCluster(db, validator.New(), unmarshalFn, marshalFn, generateTokenFn))
	handler.ServeHTTP(rec, req)
	return rec
}

func executeUpdate(t *testing.T, r io.Reader, db *gorm.DB, unmarshalFn api.Unmarshal, marshalFn api.MarshalIndent, generateTokenFn api.GenerateToken) *httptest.ResponseRecorder {
	req, err := http.NewRequest("POST", "", r)
	require.Nil(t, err)
	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(api.UpdateCluster(db, validator.New(), unmarshalFn, marshalFn))
	handler.ServeHTTP(rec, req)
	return rec
}

func TestListAlerts(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	annotations := map[string]interface{}{"anno": "foo"}
	labels := map[string]interface{}{"labels": "bar"}
	annotationJSON, err := toJSON(annotations)
	assert.NoError(t, err)
	labelsJSON, err := toJSON(labels)
	assert.NoError(t, err)

	c := models.Cluster{Name: "My Cluster", Token: "derp"}
	db.Create(&c)
	a := models.Alert{
		ClusterToken: c.Token,
		Fingerprint:  "123",
		State:        "active",
		Severity:     "foo",
		InhibitedBy:  "bar",
		SilencedBy:   "baz",
		Annotations:  datatypes.JSON(annotationJSON),
		Labels:       datatypes.JSON(labelsJSON),
		StartsAt:     time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		EndsAt:       time.Now().UTC(),
	}
	db.Create(&a)
	response, _ := doRequest(t, api.ListAlerts(db, json.MarshalIndent), "GET", "/", "/", "")
	assert.Equal(t, http.StatusOK, response.Code)

	var payload views.AlertsResponse
	err = json.Unmarshal(response.Body.Bytes(), &payload)
	assert.NoError(t, err)
	assert.Len(t, payload.Alerts, 1)
	alert := payload.Alerts[0]
	assert.Equal(t, a.ID, alert.ID)
	assert.Equal(t, a.Fingerprint, alert.Fingerprint)
	assert.Equal(t, a.State, alert.State)
	assert.Equal(t, a.Severity, alert.Severity)
	assert.Equal(t, a.InhibitedBy, alert.InhibitedBy)
	assert.Equal(t, a.SilencedBy, alert.SilencedBy)
	assert.Equal(t, annotations, alert.Annotations)
	assert.Equal(t, labels, alert.Labels)
	assert.Equal(t, a.StartsAt, alert.StartsAt)
	assert.Equal(t, a.UpdatedAt, alert.UpdatedAt)
	assert.Equal(t, a.EndsAt, alert.EndsAt)
	assert.Equal(t, c.Name, alert.Cluster.Name)
}

func TestListClusters_StatusCritical(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	// Register a cluster
	db.Create(&models.Cluster{Name: "My Cluster", Token: "derp"})
	rightNow := time.Now()
	// Agent sends cluster info
	db.Create(&models.ClusterInfo{
		ClusterToken: "derp",
		UpdatedAt:    rightNow,
	})

	// Add a critical alert
	myCriticalAlert := models.Alert{
		ID:           135,
		ClusterToken: "derp",
		Severity:     "critical",
	}
	db.Create(&myCriticalAlert)

	// Add a non-critical alert
	myNonCriticalAlert := models.Alert{
		ID:           246,
		ClusterToken: "derp",
		Severity:     "info",
	}
	db.Create(&myNonCriticalAlert)

	response := executeGet(t, db, json.MarshalIndent, "")
	assert.Equal(t, http.StatusOK, response.Code)
	var res views.ClustersResponse
	err = json.Unmarshal(response.Body.Bytes(), &res)
	assert.NoError(t, err)
	assertEqualCmp(t, views.ClustersResponse{
		Clusters: []views.ClusterView{
			{
				ID:        1,
				Token:     "derp",
				Name:      "My Cluster",
				Status:    "critical",
				UpdatedAt: rightNow,
				FluxInfo:  nil,
			},
		},
	}, res)
}

func TestListClusters_StatusAlerting(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	// Register a cluster
	db.Create(&models.Cluster{Name: "My Cluster", Token: "derp"})

	rightNow := time.Now()
	// Agent sends cluster info
	db.Create(&models.ClusterInfo{
		ClusterToken: "derp",
		UpdatedAt:    rightNow,
	})

	// Add a non-critical alert
	myNonCriticalAlert := models.Alert{
		ID:           246,
		ClusterToken: "derp",
		Severity:     "info",
	}

	db.Create(&myNonCriticalAlert)

	response := executeGet(t, db, json.MarshalIndent, "")
	assert.Equal(t, http.StatusOK, response.Code)
	var res views.ClustersResponse
	err = json.Unmarshal(response.Body.Bytes(), &res)
	assert.NoError(t, err)
	assertEqualCmp(t, views.ClustersResponse{
		Clusters: []views.ClusterView{
			{
				ID:        1,
				Token:     "derp",
				Name:      "My Cluster",
				Status:    "alerting",
				UpdatedAt: rightNow,
				FluxInfo:  nil,
			},
		},
	}, res)
}

func TestListClusters_StatusLastSeen(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	// Register a cluster
	db.Create(&models.Cluster{Name: "My Cluster", Token: "derp"})
	// Add last updated 5 minutes ago
	rightNow := time.Now()
	count := 5
	then := rightNow.Add(time.Duration(-count) * time.Minute)

	// Agent sends cluster info
	db.Create(&models.ClusterInfo{
		ClusterToken: "derp",
		UpdatedAt:    then,
	})

	response := executeGet(t, db, json.MarshalIndent, "")
	assert.Equal(t, http.StatusOK, response.Code)
	var res views.ClustersResponse
	err = json.Unmarshal(response.Body.Bytes(), &res)
	assert.NoError(t, err)
	assertEqualCmp(t, views.ClustersResponse{
		Clusters: []views.ClusterView{
			{
				ID:        1,
				Token:     "derp",
				Name:      "My Cluster",
				Status:    "lastSeen",
				UpdatedAt: then,
				FluxInfo:  nil,
			},
		},
	}, res)
}

func TestListClusters_PullRequests(t *testing.T) {
	db := createDatabase(t)

	// Register a cluster
	c := models.Cluster{Name: "My Cluster", Token: "derp", CAPIName: "fooname", CAPINamespace: "default"}
	db.Create(&c)
	pr := models.PullRequest{URL: "boop", Type: "create"}
	db.Create(&pr)
	c.PullRequests = append(c.PullRequests, &pr)
	db.Save(&c)

	c2 := models.Cluster{Name: "My Cluster2", Token: "derp2", CAPIName: "fooname2", CAPINamespace: "default"}
	db.Create(&c2)
	pr2 := models.PullRequest{URL: "boop2", Type: "delete"}
	db.Create(&pr2)
	c2.PullRequests = append(c2.PullRequests, &pr2)
	db.Save(&c2)

	c3 := models.Cluster{Name: "My Cluster3", Token: "derp3", CAPIName: "fooname3", CAPINamespace: "default"}
	capiCluster3 := models.CAPICluster{Name: "fooname3", ClusterToken: "derp3", Namespace: "default"}
	db.Create(&c3)
	pr3a := models.PullRequest{URL: "boop3", Type: "create"}
	pr3b := models.PullRequest{URL: "boop4", Type: "delete"}
	db.Create(&pr3a)
	db.Create(&pr3b)
	c3.PullRequests = append(c3.PullRequests, &pr3a)
	c3.PullRequests = append(c3.PullRequests, &pr3b)
	db.Save(&c3)
	db.Create(&capiCluster3)

	// Different namespace and different status
	c4 := models.Cluster{Name: "My Cluster4", Token: "derp4", CAPIName: "fooname4", CAPINamespace: "mccp"}
	db.Create(&c4)
	pr4 := models.PullRequest{URL: "boop5", Type: "create"}
	db.Create(&pr4)
	c4.PullRequests = append(c4.PullRequests, &pr4)
	db.Save(&c4)
	now := time.Now()
	fmt.Println("now:", now)
	then := now.AddDate(0, 0, 20)
	ci := models.ClusterInfo{ClusterToken: "derp4", UpdatedAt: then}
	db.Create(&ci)

	response := executeGet(t, db, json.MarshalIndent, "")
	assert.Equal(t, http.StatusOK, response.Code)

	var res views.ClustersResponse
	err := json.Unmarshal(response.Body.Bytes(), &res)
	assert.NoError(t, err)

	assertEqualCmp(t, views.ClustersResponse{
		Clusters: []views.ClusterView{
			{
				ID:          1,
				Token:       "derp",
				Name:        "My Cluster",
				Status:      "pullRequestCreated",
				FluxInfo:    nil,
				PullRequest: &views.PullRequestView{URL: "boop", Type: "create"},
			},
			{
				ID:          2,
				Token:       "derp2",
				Name:        "My Cluster2",
				Status:      "pullRequestCreated",
				FluxInfo:    nil,
				PullRequest: &views.PullRequestView{URL: "boop2", Type: "delete"},
			},
			{
				ID:            3,
				Token:         "derp3",
				Name:          "My Cluster3",
				CAPIName:      "fooname3",
				CAPINamespace: "default",
				Status:        "pullRequestCreated",
				FluxInfo:      nil,
				PullRequest:   &views.PullRequestView{URL: "boop4", Type: "delete"},
			},
			{
				ID:          4,
				Token:       "derp4",
				Name:        "My Cluster4",
				Status:      "ready",
				UpdatedAt:   then,
				FluxInfo:    nil,
				PullRequest: &views.PullRequestView{URL: "boop5", Type: "create"},
			},
		},
	}, res)
}

func TestListClusters_ClusterFound(t *testing.T) {
	db := createDatabase(t)

	// Register a cluster
	c := models.Cluster{Name: "My Cluster", Token: "derp", CAPIName: "fooname", CAPINamespace: "default"}
	db.Create(&c)
	pr := models.PullRequest{}
	db.Create(&pr)
	c.PullRequests = append(c.PullRequests, &pr)
	db.Save(&c)
	db.Create(&models.CAPICluster{Name: c.CAPIName, Namespace: c.CAPINamespace, Object: datatypes.JSON(`"derp"`)})

	response := executeGet(t, db, json.MarshalIndent, "")
	assert.Equal(t, http.StatusOK, response.Code)

	var res views.ClustersResponse
	err := json.Unmarshal(response.Body.Bytes(), &res)
	assert.NoError(t, err)

	assertEqualCmp(t, views.ClustersResponse{
		Clusters: []views.ClusterView{
			{
				ID:            1,
				Token:         "derp",
				Name:          "My Cluster",
				Status:        "clusterFound",
				FluxInfo:      nil,
				CAPICluster:   datatypes.JSON(`"derp"`),
				CAPIName:      "fooname",
				CAPINamespace: "default",
				PullRequest:   &views.PullRequestView{URL: ""},
			},
		},
	}, res)
}

func TestListClusters_StatusNotConnected(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)

	// Register a cluster
	db.Create(&models.Cluster{Name: "My Cluster", Token: "derp"})

	// Add last updated 40 minutes ago
	rightNow := time.Now()
	count := 40
	then := rightNow.Add(time.Duration(-count) * time.Minute)

	// Agent sends cluster info
	db.Create(&models.ClusterInfo{
		ClusterToken: "derp",
		UpdatedAt:    then,
	})

	response := executeGet(t, db, json.MarshalIndent, "")
	assert.Equal(t, http.StatusOK, response.Code)
	var res views.ClustersResponse
	err = json.Unmarshal(response.Body.Bytes(), &res)
	assert.NoError(t, err)
	assertEqualCmp(t, views.ClustersResponse{
		Clusters: []views.ClusterView{
			{
				ID:        1,
				Token:     "derp",
				Name:      "My Cluster",
				Type:      "",
				Status:    "notConnected",
				UpdatedAt: then,
				FluxInfo:  nil,
			},
		},
	}, res)
}

func TestUnregisterCluster(t *testing.T) {
	testCases := []struct {
		name                 string
		path                 string
		clustersBefore       []models.Cluster // clusters in db before DELETE
		dependentStateBefore []interface{}    // dependent state in db before DELETE
		responseCode         int
		clustersAfter        []models.Cluster // clusters in db after DELETE
		dependentStateAfter  []interface{}    // dependent state in db after DELETE
	}{
		{
			name: "Unregister an existing cluster",
			path: "/1",
			clustersBefore: []models.Cluster{
				{Name: "foo", Token: "t1", PullRequests: []*models.PullRequest{{Model: models.Model{ID: 1}}}},
				{Name: "bar", Token: "t2", PullRequests: []*models.PullRequest{{Model: models.Model{ID: 2}}}},
			},
			dependentStateBefore: []interface{}{
				&models.Event{UID: "foo", ClusterToken: "t1"}, &models.Event{UID: "bar", ClusterToken: "t2"},
				&models.ClusterInfo{UID: "foo", ClusterToken: "t1"}, &models.ClusterInfo{UID: "bar", ClusterToken: "t2"},
				&models.NodeInfo{UID: "foo", ClusterToken: "t1", ClusterInfoUID: "foo"}, &models.NodeInfo{UID: "bar", ClusterToken: "t2", ClusterInfoUID: "bar"},
				&models.Alert{ID: 1, ClusterToken: "t1"}, &models.Alert{ID: 2, ClusterToken: "t2"},
				&models.FluxInfo{Name: "foo", ClusterToken: "t1"}, &models.FluxInfo{Name: "bar", ClusterToken: "t2"},
				&models.GitCommit{Sha: "foo", ClusterToken: "t1"}, &models.GitCommit{Sha: "bar", ClusterToken: "t2"},
				&models.Workspace{Name: "foo", ClusterToken: "t1"}, &models.Workspace{Name: "bar", ClusterToken: "t2"},
			},
			responseCode: http.StatusNoContent,
			clustersAfter: []models.Cluster{
				{Name: "bar", Token: "t2", PullRequests: []*models.PullRequest{{Model: models.Model{ID: 2}}}},
			},
			dependentStateAfter: []interface{}{
				models.Event{UID: "bar", ClusterToken: "t2"},
				models.ClusterInfo{UID: "bar", ClusterToken: "t2"},
				models.NodeInfo{UID: "bar", ClusterToken: "t2"},
				models.Alert{ID: 2, ClusterToken: "t2"},
				models.FluxInfo{Name: "bar", ClusterToken: "t2"},
				models.GitCommit{Sha: "bar", ClusterToken: "t2"},
				models.Workspace{Name: "bar", ClusterToken: "t2"},
				models.PullRequest{Model: models.Model{ID: 1}}, models.PullRequest{Model: models.Model{ID: 2}},
				ClusterPullRequests{ClusterID: 2, PullRequestID: 2},
			},
		},
		{
			name:           "Unregister a non-existing cluster",
			path:           "/2",
			clustersBefore: []models.Cluster{{Name: "foo"}},
			responseCode:   http.StatusNotFound,
		},
		{
			name:         "Id param not a unint number",
			path:         "/foo",
			responseCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := utils.OpenDebug("", false)
			assert.NoError(t, err)
			err = utils.MigrateTables(db)
			assert.NoError(t, err)

			if tc.clustersBefore != nil {
				// Setup state before DELETE
				tx := db.Create(tc.clustersBefore)
				require.NoError(t, tx.Error)
			}
			for _, o := range tc.dependentStateBefore {
				tx := db.Create(o)
				require.NoError(t, tx.Error)
			}

			// Unregister cluster
			response, _ := doRequest(t, api.UnregisterCluster(db), "DELETE", "/{id}", tc.path, "")
			assert.Equal(t, tc.responseCode, response.Code)

			if tc.responseCode >= 200 && tc.responseCode <= 299 {
				// Expect 404 when getting cluster, if previous request was successful
				response, _ := doRequest(t, api.FindCluster(db, json.MarshalIndent), "GET", "/{id}", tc.path, "")
				assert.Equal(t, http.StatusNotFound, response.Code)

				var expectedEvents []models.Event
				var expectedClusterInfo []models.ClusterInfo
				var expectedNodeInfo []models.NodeInfo
				var expectedAlerts []models.Alert
				var expectedFluxInfo []models.FluxInfo
				var expectedGitCommits []models.GitCommit
				var expectedWorkspaces []models.Workspace
				var expectedPullRequests []models.PullRequest
				var expectedClusterPullRequests []ClusterPullRequests
				for _, i := range tc.dependentStateAfter {
					switch v := i.(type) {
					case models.Event:
						expectedEvents = append(expectedEvents, i.(models.Event))
					case models.ClusterInfo:
						expectedClusterInfo = append(expectedClusterInfo, i.(models.ClusterInfo))
					case models.NodeInfo:
						expectedNodeInfo = append(expectedNodeInfo, i.(models.NodeInfo))
					case models.Alert:
						expectedAlerts = append(expectedAlerts, i.(models.Alert))
					case models.FluxInfo:
						expectedFluxInfo = append(expectedFluxInfo, i.(models.FluxInfo))
					case models.GitCommit:
						expectedGitCommits = append(expectedGitCommits, i.(models.GitCommit))
					case models.Workspace:
						expectedWorkspaces = append(expectedWorkspaces, i.(models.Workspace))
					case models.PullRequest:
						expectedPullRequests = append(expectedPullRequests, i.(models.PullRequest))
					case ClusterPullRequests:
						expectedClusterPullRequests = append(expectedClusterPullRequests, i.(ClusterPullRequests))
					default:
						fmt.Printf("Unknown type %T!\n", v)
					}
				}

				var actualEvents []models.Event
				result := db.Find(&actualEvents)
				assert.NoError(t, result.Error)
				assert.Len(t, actualEvents, len(expectedEvents))
				assert.Subset(t, actualEvents, expectedEvents)

				var actualClusterInfo []models.ClusterInfo
				result = db.Find(&actualClusterInfo)
				assert.NoError(t, result.Error)
				assert.Len(t, actualClusterInfo, len(expectedClusterInfo))
				assert.Equal(t, actualClusterInfo[0].UID, expectedClusterInfo[0].UID)

				var actualNodeInfo []models.NodeInfo
				result = db.Find(&actualNodeInfo)
				assert.NoError(t, result.Error)
				assert.Len(t, actualNodeInfo, len(expectedNodeInfo))
				assert.Equal(t, actualNodeInfo[0].Name, expectedNodeInfo[0].Name)

				var actualAlerts []models.Alert
				result = db.Find(&actualAlerts)
				assert.NoError(t, result.Error)
				assert.Len(t, actualAlerts, len(expectedAlerts))
				assert.Equal(t, actualAlerts[0].ID, expectedAlerts[0].ID)

				var actualFluxInfo []models.FluxInfo
				result = db.Find(&actualFluxInfo)
				assert.NoError(t, result.Error)
				assert.Len(t, actualFluxInfo, len(expectedFluxInfo))
				assert.Subset(t, actualFluxInfo, expectedFluxInfo)

				var actualGitCommits []models.GitCommit
				result = db.Find(&actualGitCommits)
				assert.NoError(t, result.Error)
				assert.Len(t, actualGitCommits, len(expectedGitCommits))
				assert.Subset(t, actualGitCommits, expectedGitCommits)

				var actualWorkspaces []models.Workspace
				result = db.Find(&actualWorkspaces)
				assert.NoError(t, result.Error)
				assert.Len(t, actualWorkspaces, len(expectedWorkspaces))
				assert.Subset(t, actualWorkspaces, expectedWorkspaces)

				var actualPullRequests []models.PullRequest
				result = db.Find(&actualPullRequests)
				assert.NoError(t, result.Error)
				assert.Len(t, actualPullRequests, len(expectedPullRequests))
				assert.Equal(t, actualPullRequests[0].ID, expectedPullRequests[0].ID)

				var actualClusterPullRequests []ClusterPullRequests
				result = db.Find(&actualClusterPullRequests)
				assert.NoError(t, result.Error)
				assert.Len(t, actualClusterPullRequests, len(expectedClusterPullRequests))
				assert.Equal(t, actualClusterPullRequests[0].ClusterID, expectedClusterPullRequests[0].ClusterID)
				assert.Equal(t, actualClusterPullRequests[0].PullRequestID, expectedClusterPullRequests[0].PullRequestID)
			}
		})
	}
}

func createDatabase(t *testing.T) *gorm.DB {
	db, err := utils.OpenDebug("", true)
	assert.NoError(t, err)
	err = utils.MigrateTables(db)
	assert.NoError(t, err)
	return db
}

func toJSON(obj interface{}) ([]byte, error) {
	output := bytes.NewBufferString("")
	encoder := json.NewEncoder(output)
	if err := encoder.Encode(obj); err != nil {
		return []byte{}, err
	}
	return output.Bytes(), nil
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

func reverseAny(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

// So we can query the table usually managed by gorm
type ClusterPullRequests struct {
	ClusterID     int `gorm:"primaryKey"`
	PullRequestID int `gorm:"primaryKey"`
}
