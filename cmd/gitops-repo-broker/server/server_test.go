package server_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops-repo-broker/server"
	dbutils "github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
	"github.com/weaveworks/weave-gitops-enterprise/test/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var validEntitlement = `eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJsaWNlbmNlZFVudGlsIjoxNzg5MzgxMDE1LCJpYXQiOjE2MzE2MTQ2MTUsImlzcyI6InNhbGVzQHdlYXZlLndvcmtzIiwibmJmIjoxNjMxNjE0NjE1LCJzdWIiOiJ0ZWFtLXBlc3RvQHdlYXZlLndvcmtzIn0.klRpQQgbCtshC3PuuD4DdI3i-7Z0uSGQot23YpsETphFq4i3KK4NmgfnDg_WA3Pik-C2cJgG8WWYkWnemWQJAw`

func TestEntitlementMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		client   client.Client
		expected int
	}{
		{
			name:     "no entitlement",
			client:   utils.CreateFakeClient(t),
			expected: http.StatusInternalServerError,
		},
		{
			name:     "valid entitlement",
			client:   utils.CreateFakeClient(t, createSecret(validEntitlement)),
			expected: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			file, err := ioutil.TempFile("", "db")
			if err != nil {
				t.Fatalf("expected no errors but got: %v", err)
			}
			defer os.Remove(file.Name())

			db, err := dbutils.Open(file.Name(), "sqlite", "", "", "")
			if err != nil {
				t.Fatalf("expected no errors but got: %v", err)
			}
			err = dbutils.MigrateTables(db)
			if err != nil {
				t.Fatalf("expected no errors but got: %v", err)
			}

			s, err := server.NewServer(ctx, tt.client, client.ObjectKey{Name: "name", Namespace: "namespace"}, logr.Discard(), server.ParamSet{
				DbType: "sqlite",
				DbURI:  file.Name(),
				Port:   "8001",
			})
			if err != nil {
				t.Fatalf("expected no errors but got: %v", err)
			}
			defer s.Close()

			go func() {
				_ = s.ListenAndServe()
			}()

			time.Sleep(100 * time.Millisecond)
			res, err := http.Get("http://localhost:8001/gitops/api/clusters")
			if err != nil {
				t.Fatalf("expected no errors but got: %v", err)
			}
			if res.StatusCode != tt.expected {
				t.Fatalf("expected status code to be %d but got %d instead", tt.expected, res.StatusCode)
			}
		})
	}
}

func createSecret(s string) *corev1.Secret {
	// When reading a secret, only Data contains any data, StringData is empty
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "name",
			Namespace: "namespace",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{"entitlement": []byte(s)},
	}
}
