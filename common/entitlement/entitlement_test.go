package entitlement

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/golang-jwt/jwt/v4"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	// This entitlement has been generated with the right private key for 1 day
	validEntitlement = `eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJsaWNlbmNlZFVudGlsIjoxNjMxMzYxMjg2LCJpYXQiOjE2MzEyNzQ4ODYsImlzcyI6InNhbGVzQHdlYXZlLndvcmtzIiwibmJmIjoxNjMxMjc0ODg2LCJzdWIiOiJ0ZXN0QHdlYXZlLndvcmtzIn0.EKGp89DFcRKZ_kGmC8FuLVPB0wiab2KddkQKAmVNC9UH459v63tCP13eFybx9dAmMuaC77SA8rp7ukN1qZM7DA`
	validTimestamp   = time.Unix(1631274886, 0)
)

func TestEntitlementHandler(t *testing.T) {
	tests := []struct {
		name     string
		state    []runtime.Object
		verified time.Time
		exists   bool
	}{
		{
			name:     "valid entitlement",
			state:    []runtime.Object{createSecret(validEntitlement)},
			verified: validTimestamp,
			exists:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				ent := ctx.Value(contextKeyEntitlement)
				exists := ent != nil
				if exists != tt.exists {
					if exists {
						t.Errorf("expected context value to not be present but was: %+v", ent)
					} else {
						t.Errorf("expected context value to be present but was not: %+v", ent)
					}
				}
			})

			at(tt.verified, func() {
				c := createFakeClient(tt.state)
				key := client.ObjectKey{Name: "name", Namespace: "namespace"}
				handler := EntitlementHandler(ctx, logr.Discard(), c, key, next)
				handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "http://test", nil))
			})

		})
	}
}

func TestCheckEntitlementHandler(t *testing.T) {
	tests := []struct {
		name        string
		ctxValue    interface{}
		status      int
		header      bool
		response    string
		headerValue string
	}{
		{
			name: "valid entitlement",
			ctxValue: &entitlement{
				LicencedUntil: time.Now().AddDate(1, 0, 0),
			},
			status: http.StatusOK,
			header: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			previous := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
					r = r.WithContext(context.WithValue(context.Background(), contextKeyEntitlement, tt.ctxValue))
					next.ServeHTTP(rw, r)
				})
			}

			next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusOK)
			})

			rec := httptest.NewRecorder()
			handler := previous(CheckEntitlementHandler(logr.Discard(), next, []string{}))
			handler.ServeHTTP(rec, httptest.NewRequest("GET", "http://test", nil))

			if rec.Code != tt.status {
				t.Errorf("expected response status code to equal %d but was not: %d", tt.status, rec.Code)
			}

			body, err := io.ReadAll(rec.Body)
			if err != nil {
				t.Errorf("unexpected error while reading response: %s", err)
			}

			if string(body) != tt.response {
				t.Errorf("expected response body to be %s but got: %s", tt.response, body)
			}

			h := rec.Header().Get(entitlementExpiredMessageHeader)
			if tt.header && h == "" {
				t.Errorf("expected response header to be present but was not: %+v", rec.Header())
				if h != tt.headerValue {
					t.Errorf("expected header to be %s but got: %s", tt.headerValue, h)
				}
			} else if !tt.header && h != "" {
				t.Errorf("expected response header to not be present but was: %+v", rec.Header())
			}
		})
	}
}

func createFakeClient(clusterState []runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
	}
	_ = schemeBuilder.AddToScheme(scheme)

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(clusterState...).
		Build()

	return c
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

func at(t time.Time, f func()) {
	jwt.TimeFunc = func() time.Time {
		return t
	}
	f()
	jwt.TimeFunc = time.Now
}
