package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

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

func ListClusters(db *gorm.DB, marshalIndentFn MarshalIndent) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		type result struct {
			Name           string
			Type           string
			NodeName       string
			IsControlPlane bool
			KubeletVersion string
		}
		var results []result
		if err := db.Raw(`
			SELECT
				c.name AS Name, 
				ci.type AS Type, 
				ni.name AS NodeName, 
				ni.is_control_plane AS IsControlPlane, 
				ni.kubelet_version AS KubeletVersion
			FROM 
				clusters c 
				LEFT JOIN cluster_info ci ON c.token = ci.token 
				LEFT JOIN node_info ni ON c.token = ni.token
			WHERE c.deleted_at IS NULL
		`).Scan(&results).Error; err != nil {
			log.Errorf("Failed to query for clusters: %v", err)
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		res := &ClustersResponse{
			Clusters: []Cluster{},
		}

		if results == nil {
			data, err := marshalIndentFn(res, "", " ")
			if err != nil {
				common.WriteError(w, err, http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, string(data))
			return
		}

		clusters := map[string]*Cluster{}
		for _, r := range results {
			if cl, ok := clusters[r.Name]; !ok {
				// Add new cluster with node to map
				c := Cluster{
					Name: r.Name,
					Type: r.Type,
				}
				// Do not add nodes if they don't exist yet
				if r.NodeName != "" {
					c.Nodes = append(c.Nodes, Node{
						Name:           r.NodeName,
						IsControlPlane: r.IsControlPlane,
						KubeletVersion: r.KubeletVersion,
					})
				}
				clusters[r.Name] = &c
			} else {
				// Update existing cluster in map with node
				n := Node{
					Name:           r.NodeName,
					IsControlPlane: r.IsControlPlane,
					KubeletVersion: r.KubeletVersion,
				}
				cl.Nodes = append(cl.Nodes, n)
			}
		}

		for _, c := range clusters {
			res.Clusters = append(res.Clusters, *c)
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

type ClusterRegistrationRequest struct {
	Name       string `json:"name" validate:"required"`
	IngressURL string `json:"ingressUrl" validate:"omitempty,url"`
}

type ClusterRegistrationResponse struct {
	Name       string `json:"name"`
	IngressURL string `json:"ingressUrl"`
	Token      string `json:"token"`
}

type Node struct {
	Name           string `json:"name"`
	IsControlPlane bool   `json:"isControlPlane"`
	KubeletVersion string `json:"kubeletVersion"`
}

type Cluster struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Nodes []Node `json:"nodes,omitempty"`
}

type ClustersResponse struct {
	Clusters []Cluster `json:"clusters"`
}
