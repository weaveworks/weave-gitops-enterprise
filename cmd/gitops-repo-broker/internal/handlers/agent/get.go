package agent

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/weaveworks/wks/cmd/gitops-repo-broker/internal/common"
	"github.com/weaveworks/wks/common/database/models"
	"github.com/weaveworks/wks/pkg/version"
	"gorm.io/gorm"
)

// ErrNilDB is a database error
var ErrNilDB = errors.New("The database has not been initialised")

const agentManifestTemplate = `
---
apiVersion: v1
kind: Namespace
metadata:
  name: wkp-agent

---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    name: wkp-agent
  name: wkp-agent
  namespace: wkp-agent

---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    name: wkp-agent
  name: wkp-agent
rules:
  - apiGroups: ['*']
    resources: ['*']
    verbs: ["get", "watch", "list"]
  - nonResourceURLs: ['*']
    verbs: ["get"]

---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    name: wkp-agent
  name: wkp-agent
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: wkp-agent
subjects:
  - kind: ServiceAccount
    name: wkp-agent
    namespace: wkp-agent

---
apiVersion: v1
kind: Secret
metadata:
  name: wkp-agent-token
  namespace: wkp-agent
type: Opaque
data:
  token: {{ .Token }}

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: wkp-agent
  namespace: wkp-agent
spec:
  selector:
    matchLabels:
      app: wkp-agent
  replicas: 1
  template:
    metadata:
      labels:
        app: wkp-agent
    spec:
      serviceAccountName: wkp-agent
      containers:
      - name: agent
        image: weaveworks/wkp-agent:{{ .ImageTag }}
        args:
        - watch
        - --nats-url={{ .NatsURL }}
        env:
        - name: WKP_AGENT_TOKEN
          valueFrom:
            secretKeyRef:
              name: wkp-agent-token
              key: token
      - name: alertmanager-agent
        image: weaveworks/wkp-agent:{{ .ImageTag }}
        args:
        - agent-server
        - --nats-url={{ .NatsURL }}
        - --alertmanager-url={{ .AlertmanagerURL }}
        env:
        - name: WKP_AGENT_TOKEN
          valueFrom:
            secretKeyRef:
              name: wkp-agent-token
              key: token
`

func renderTemplate(token, imageTag, natsURL, alertmanagerURL string) (string, error) {
	t, err := template.New("agent-manifest").Parse(agentManifestTemplate)
	if err != nil {
		return "", err
	}
	encodedToken := base64.StdEncoding.EncodeToString([]byte(token))

	var populated bytes.Buffer
	err = t.Execute(&populated, struct {
		Token           string
		ImageTag        string
		NatsURL         string
		AlertmanagerURL string
	}{encodedToken, imageTag, natsURL, alertmanagerURL})
	if err != nil {
		return "", err
	}
	return populated.String(), nil
}

// NewGetHandler returns a YAMLStream given a token
func NewGetHandler(db *gorm.DB, defaultNatsURL, defaultAlertmanagerURL string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")

		token := r.URL.Query().Get("token")
		if token == "" {
			common.WriteError(w, errors.New("Missing token param"), http.StatusUnauthorized)
			return
		}

		checkToken, err := isTokenPresent(db, token)
		if err != nil {
			common.WriteError(w, err, http.StatusBadRequest)
			return
		}
		if !checkToken {
			common.WriteError(w, errors.New("Cluster not found"), http.StatusNotFound)
			return
		}

		// TODO: Maybe support reading natsURL from query params?
		// natsURL := r.URL.Query().Get("natsURL")
		alertmanagerURL := r.URL.Query().Get("alertmanagerURL")
		if alertmanagerURL == "" {
			alertmanagerURL = defaultAlertmanagerURL
		}
		stream, err := renderTemplate(token, version.ImageTag, defaultNatsURL, alertmanagerURL)
		if err != nil {
			common.WriteError(w, err, 500)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(stream))
	}
}

func isTokenPresent(db *gorm.DB, token string) (bool, error) {
	if db == nil {
		return false, ErrNilDB
	}

	var cluster models.Cluster
	result := db.Where("token = ?", token).First(&cluster)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}

	return true, result.Error
}
