package api

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/cmd/gitops-repo-broker/internal/common"
	"github.com/weaveworks/wks/common/database/models"
	"gorm.io/gorm"
)

var (
	ErrNilDB          = errors.New("The database has not been initialised.")
	ErrInvalidPayload = errors.New("Invalid payload")
)

// Signature for json.MarshalIndent accepted as a method parameter for unit tests
type MarshalIndent func(v interface{}, prefix, indent string) ([]byte, error)

// Signature for json.Unmarshal accepted as a method parameter for unit test
type Unmarshal func(data []byte, v interface{}) error

// Signature for api.Generate accepted as a method parameter for unit test
type GenerateToken func() (string, error)

// DB helpers (FIXME: maybe move somewhere)

func getClusters(db *gorm.DB, extraQuery string, extraValues ...interface{}) ([]ClusterView, error) {
	type ClusterNodeRow struct {
		ID             uint
		Name           string
		Token          string
		Type           string
		NodeName       string
		IsControlPlane bool
		KubeletVersion string
	}

	var rows []ClusterNodeRow
	if err := db.Raw(`
			SELECT
				c.id AS ID,
				c.name AS Name, 
				c.token AS Token, 
				ci.type AS Type, 
				ni.name AS NodeName, 
				ni.is_control_plane AS IsControlPlane, 
				ni.kubelet_version AS KubeletVersion
			FROM 
				clusters c 
				LEFT JOIN cluster_info ci ON c.token = ci.token 
				LEFT JOIN node_info ni ON c.token = ni.token
			WHERE c.deleted_at IS NULL
		`+extraQuery, extraValues).Scan(&rows).Error; err != nil {
		return nil, ErrNilDB
	}

	clusters := map[string]*ClusterView{}
	for _, r := range rows {
		if cl, ok := clusters[r.Name]; !ok {
			// Add new cluster with node to map
			c := ClusterView{
				ID:    r.ID,
				Name:  r.Name,
				Token: r.Token,
				Type:  r.Type,
			}
			// Do not add nodes if they don't exist yet
			if r.NodeName != "" {
				c.Nodes = append(c.Nodes, NodeView{
					Name:           r.NodeName,
					IsControlPlane: r.IsControlPlane,
					KubeletVersion: r.KubeletVersion,
				})
			}
			clusters[r.Name] = &c
		} else {
			// Update existing cluster in map with node
			n := NodeView{
				Name:           r.NodeName,
				IsControlPlane: r.IsControlPlane,
				KubeletVersion: r.KubeletVersion,
			}
			cl.Nodes = append(cl.Nodes, n)
		}
	}

	clusterList := []ClusterView{}
	for _, c := range clusters {
		clusterList = append(clusterList, *c)
	}

	return clusterList, nil
}

func getCluster(db *gorm.DB, id uint) (*ClusterView, error) {
	clusters, err := getClusters(db, " AND c.id = ?", id)
	if err != nil {
		return nil, err
	}
	if len(clusters) == 0 {
		return nil, nil
	}
	return &clusters[0], nil
}

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

		clusters, err := getClusters(db, "")
		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}

		res := ClustersResponse{
			Clusters: clusters,
		}
		data, err := marshalIndentFn(res, "", " ")
		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}
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

		crr := &ClusterRegistrationRequest{}
		if err := unmarshalFn(reqBody, crr); err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}

		err = validate.Struct(crr)
		if err != nil {
			log.Errorf("Failed to validate payload: %v", err)
			common.WriteError(w, ErrInvalidPayload, http.StatusBadRequest)
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

		res := ClusterRegistrationResponse{
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

func UpdateCluster(db *gorm.DB, unmarshalFn Unmarshal, marshalFn MarshalIndent) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			cluster, err := getClusterFromRequest(r, tx)
			if err != nil {
				return err
			}
			if cluster == nil {
				common.WriteError(w, fmt.Errorf("cluster not found"), http.StatusNotFound)
				return nil
			}

			reqBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				common.WriteError(w, err, http.StatusBadRequest)
				return nil
			}

			clusterUpdates := &models.Cluster{}
			if err := unmarshalFn(reqBody, clusterUpdates); err != nil {
				return err
			}

			c := models.Cluster{}
			tx.First(&c, cluster.ID)
			c.IngressURL = clusterUpdates.IngressURL
			c.Name = clusterUpdates.Name
			// This particular variant of `Updates` (with a struct), does non-zero checking
			// and so don't update the Name or IngressURL if they are "zero"
			// https://gorm.io/docs/update.html#Update-Selected-Fields
			result := tx.Model(c).Updates(c)
			if result.Error != nil {
				common.WriteError(w, result.Error, http.StatusBadRequest)
				return nil
			}

			clusterView, err := getCluster(tx, c.ID)
			if err != nil {
				return err
			}
			respondWithJSON(w, http.StatusOK, clusterView, marshalFn)
			return nil
		})

		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
		}
	}
}

// types

type ClusterRegistrationRequest struct {
	Name       string `json:"name" validate:"required"`
	IngressURL string `json:"ingressUrl" validate:"omitempty,url"`
}

type ClusterRegistrationResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	IngressURL string `json:"ingressUrl"`
	Token      string `json:"token"`
}

type NodeView struct {
	Name           string `json:"name"`
	IsControlPlane bool   `json:"isControlPlane"`
	KubeletVersion string `json:"kubeletVersion"`
}

type ClusterView struct {
	ID    uint       `json:"id"`
	Name  string     `json:"name"`
	Type  string     `json:"type"`
	Token string     `json:"token"`
	Nodes []NodeView `json:"nodes,omitempty"`
}

type ClustersResponse struct {
	Clusters []ClusterView `json:"clusters"`
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
	w.Write(response)
}

func getClusterFromRequest(r *http.Request, db *gorm.DB) (*ClusterView, error) {
	vars := mux.Vars(r)
	idString := vars["id"]
	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse id %v: %v", idString, err)
	}

	return getCluster(db, uint(id))
}
