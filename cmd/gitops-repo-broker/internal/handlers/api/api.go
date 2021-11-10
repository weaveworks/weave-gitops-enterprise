package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops-repo-broker/internal/common"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops-repo-broker/internal/handlers/api/database"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops-repo-broker/internal/handlers/api/views"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/models"
	"gorm.io/gorm"
)

var (
	ErrNilDB          = errors.New("The database has not been initialised.")
	ErrInvalidPayload = errors.New("Invalid payload")
	ErrRowUnpack      = errors.New("Error constructing response")
)

// Signature for json.MarshalIndent accepted as a method parameter for unit tests
type MarshalIndent func(v interface{}, prefix, indent string) ([]byte, error)

// Signature for json.Unmarshal accepted as a method parameter for unit test
type Unmarshal func(data []byte, v interface{}) error

// Signature for api.Generate accepted as a method parameter for unit test
type GenerateToken func() (string, error)

// Handlers

func FindCluster(db *gorm.DB, marshalIndentFn MarshalIndent) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		cluster, err := getClusterFromRequest(r, db)
		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}
		if cluster == nil {
			common.WriteError(w, fmt.Errorf("cluster not found"), http.StatusNotFound)
			return
		}

		respondWithJSON(w, http.StatusOK, cluster, marshalIndentFn)
	}
}

func ListClusters(db *gorm.DB, marshalIndentFn MarshalIndent) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		// Read sort-by-column from the url string if provided
		sortColumn := ""
		sortByParam, ok := r.URL.Query()["sortBy"]
		if ok {
			p := sortByParam[0]
			log.Debugf("sorting by column: %s\n", p)
			if p == "Name" || p == "ClusterStatus" || p == "Token" || p == "IngressURL" {
				sortColumn = sortByParam[0]
			}
		}

		// Read sort order from the url string if provided
		sortOrder := ""
		sortOrderParam, ok := r.URL.Query()["order"]
		if ok {
			log.Debugf("sorting by order: %s\n", sortOrderParam)
			if sortOrderParam[0] == "ASC" || sortOrderParam[0] == "DESC" {
				sortOrder = sortOrderParam[0]
			}
		}

		var pagination database.Pagination
		if perPage := r.URL.Query().Get("per_page"); perPage != "" {
			if size, err := strconv.Atoi(perPage); err == nil {
				pagination.Size = size
			}
		}

		if page := r.URL.Query().Get("page"); page != "" {
			if pageNo, err := strconv.Atoi(page); err == nil {
				pagination.Page = pageNo
			}
		}

		response, err := database.GetClusters(db,
			database.GetClustersRequest{
				SortColumn: sortColumn,
				SortOrder:  sortOrder,
				Pagination: pagination,
			})
		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}

		res := views.ClustersResponse{
			Clusters: response.Clusters,
		}
		data, err := marshalIndentFn(res, "", " ")
		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Range", fmt.Sprintf("clusters */%d", response.Total))
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(data))
	}
}

func RegisterCluster(db *gorm.DB, validate *validator.Validate, unmarshalFn Unmarshal, marshalFn MarshalIndent, generateTokenFn GenerateToken) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			common.WriteError(w, err, http.StatusBadRequest)
			return
		}

		crr := &views.ClusterRegistrationRequest{}
		if err := unmarshalFn(reqBody, crr); err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}

		err = validate.Struct(crr)
		if err != nil {
			log.Errorf("Failed to validate payload: %v", err)
			validationErrors := err.(validator.ValidationErrors)
			combinedError := fmt.Errorf("%v, %v", ErrInvalidPayload, validationErrors)
			common.WriteError(w, combinedError, http.StatusBadRequest)
			return
		}

		t, err := generateTokenFn()
		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}

		c := &models.Cluster{
			Name:       crr.Name,
			IngressURL: crr.IngressURL,
			Token:      t,
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(c).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}

		res := views.ClusterRegistrationResponse{
			ID:         c.ID,
			Name:       c.Name,
			IngressURL: c.IngressURL,
			Token:      c.Token,
		}
		output, err := marshalFn(res, "", " ")
		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", output)
	}
}

func UpdateCluster(db *gorm.DB, validate *validator.Validate, unmarshalFn Unmarshal, marshalFn MarshalIndent) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			common.WriteError(w, err, http.StatusBadRequest)
			return
		}

		updated := &views.ClusterUpdateRequest{}
		if err := unmarshalFn(reqBody, updated); err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}

		err = validate.Struct(updated)
		if err != nil {
			log.Errorf("Failed to validate payload: %v", err)
			validationErrors := err.(validator.ValidationErrors)
			combinedError := fmt.Errorf("%v, %v", ErrInvalidPayload, validationErrors)
			common.WriteError(w, combinedError, http.StatusBadRequest)
			return
		}

		id, err := getClusterIDFromRequest(r)
		if err != nil {
			log.Errorf("Failed to get id param from path: %v", err)
			combinedError := fmt.Errorf("%v, %v", ErrInvalidPayload, "Error parsing 'id' param from path")
			common.WriteError(w, combinedError, http.StatusBadRequest)
			return
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			cluster := models.Cluster{}
			if err := tx.First(&cluster, id).Error; err != nil {
				return err
			}

			cluster.Name = updated.Name
			cluster.IngressURL = updated.IngressURL
			if err := tx.Save(cluster).Error; err != nil {
				return err
			}

			clusterView, err := database.GetCluster(tx, uint(id))
			if err != nil {
				return err
			}
			respondWithJSON(w, http.StatusOK, clusterView, marshalFn)
			return nil
		})

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				common.WriteError(w, fmt.Errorf("cluster not found"), http.StatusNotFound)
				return
			}
			common.WriteError(w, err, http.StatusInternalServerError)
		}
	}
}

func ListAlerts(db *gorm.DB, marshalIndentFn MarshalIndent) func(w http.ResponseWriter, r *http.Request) {
	alertsQuery := `
		SELECT
			a.id,
			a.annotations,
			a.ends_at,
			a.fingerprint,
			a.inhibited_by,
			a.silenced_by,
			a.severity,
			a.state,
			a.starts_at,
			a.updated_at,
			a.labels,
			c.id as ClusterID,
			c.name as ClusterName,
			c.ingress_url as ClusterIngressURL
		FROM
			alerts a,
			clusters c
		WHERE
			a.cluster_token = c.token and
			a.severity != 'none' and a.severity is not null
		ORDER BY
			a.starts_at DESC
		`

	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}
		var rows []views.AlertsClusterRow
		if db.Dialector.Name() == "postgres" {
			DB, err := db.DB()
			if err != nil {
				return
			}

			log.Debugf("raw query: %s\n", alertsQuery)
			result, err := DB.Query(alertsQuery)
			if err != nil {
				return
			}
			for result.Next() {
				rows = append(rows, database.AlertListRowScan(result))
			}
		} else {
			err := db.Transaction(func(tx *gorm.DB) error {
				if err := tx.Raw(alertsQuery).Scan(&rows).Error; err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				log.Errorf("Failed to query for alerts: %v", err)
				common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
				return
			}
		}

		res, err := toAlertResponse(rows)
		if err != nil {
			log.Errorf("Failed to generate response for alerts: %v", err)
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		respondWithJSON(w, http.StatusOK, res, marshalIndentFn)
	}
}

func UnregisterCluster(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		id, err := getClusterIDFromRequest(r)
		if err != nil {
			log.Errorf("Parameter 'id' is not a uint: %v", err)
			common.WriteError(w, err, http.StatusBadRequest)
			return
		}

		cluster, err := database.GetCluster(db, uint(id))
		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			log.Errorf("Failed to load cluster(%d): %v", id, err)
			return
		}

		if cluster == nil {
			w.WriteHeader(http.StatusNotFound)
			log.Errorf("Cluster(%d) was not found", id)
			return
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			dependentObjectsToDelete := []interface{}{
				&models.Event{},
				&models.ClusterInfo{},
				&models.Alert{},
				&models.FluxInfo{},
				&models.GitCommit{},
				&models.Workspace{},
			}

			for _, o := range dependentObjectsToDelete {
				if err := tx.Where("cluster_token = ?", cluster.Token).Delete(o).Error; err != nil {
					return fmt.Errorf("failed to delete %T records when unregistering Cluster %q: %w", o, cluster.Token, err)
				}
			}

			if err := tx.Delete(&models.Cluster{}, id).Error; err != nil {
				return fmt.Errorf("failed to delete Cluster(%d) record when unregistering Cluster %q: %w", id, cluster.Token, err)
			}

			return nil
		})
		if err != nil {
			log.Errorf("Failed to unregister Cluster(%d): %v", id, err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// helpers
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}, marshalIndentFn MarshalIndent) {
	response, err := marshalIndentFn(payload, "", " ")
	if err != nil {
		common.WriteError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(response)
	if err != nil {
		common.WriteError(w, err, http.StatusInternalServerError)
	}
}

func getClusterFromRequest(r *http.Request, db *gorm.DB) (*views.ClusterView, error) {
	id, err := getClusterIDFromRequest(r)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse id: %v", err)
	}

	return database.GetCluster(db, uint(id))
}

func getClusterIDFromRequest(r *http.Request) (uint64, error) {
	vars := mux.Vars(r)
	idString := vars["id"]
	return strconv.ParseUint(idString, 10, 64)
}

func toAlertResponse(rows []views.AlertsClusterRow) (*views.AlertsResponse, error) {

	res := views.AlertsResponse{
		Alerts: []views.AlertView{},
	}

	for _, r := range rows {
		var labels map[string]interface{}
		var annotations map[string]interface{}
		if err := json.Unmarshal(r.Labels, &labels); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(r.Annotations, &annotations); err != nil {
			return nil, err
		}
		res.Alerts = append(res.Alerts, views.AlertView{
			ID:          r.ID,
			Fingerprint: r.Fingerprint,
			State:       r.State,
			Severity:    r.Severity,
			InhibitedBy: r.InhibitedBy,
			SilencedBy:  r.SilencedBy,
			Annotations: annotations,
			Labels:      labels,
			StartsAt:    r.StartsAt,
			UpdatedAt:   r.UpdatedAt,
			EndsAt:      r.EndsAt,
			Cluster: views.ClusterView{
				ID:         r.ClusterID,
				Name:       r.ClusterName,
				IngressURL: r.ClusterIngressURL,
			},
		})
	}

	return &res, nil
}
