package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/weaveworks/wks/cmd/gitops-repo-broker/internal/common"
	"github.com/weaveworks/wks/common/database/models"
	"gorm.io/gorm"
)

var (
	ErrNilDB = errors.New("The database has not been initialised.")
)

// Signature for json.MarshalIndent accepted as a method parameter for unit tests
type MarshalIndent func(v interface{}, prefix, indent string) ([]byte, error)

func NewGetClusters(db *gorm.DB, fn MarshalIndent) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		var result []models.Cluster
		db.Find(&result)
		if result == nil {
			common.WriteError(w, ErrNilDB, http.StatusInternalServerError)
			return
		}

		data, err := fn(result, "", " ")
		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(data))
	}
}
