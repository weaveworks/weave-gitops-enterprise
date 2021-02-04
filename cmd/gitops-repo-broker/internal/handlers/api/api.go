package api

import (
	"encoding/json"
	"fmt"
	"github.com/weaveworks/wks/cmd/gitops-repo-broker/internal/common"
	"github.com/weaveworks/wks/common/database/models"
	"github.com/weaveworks/wks/common/database/utils"
	"net/http"
)

func NewGetClusters(dbURI string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db, err := utils.Open(dbURI)
		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}
		var result []models.Cluster
		db.Find(&result)
		data, err := json.MarshalIndent(result, "", " ")
		if err != nil {
			common.WriteError(w, err, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(data))
	}
}
