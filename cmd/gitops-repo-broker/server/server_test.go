package server_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops-repo-broker/server"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var validEntitlement = `eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJsaWNlbmNlZFVudGlsIjoxNzg5MzgxMDE1LCJpYXQiOjE2MzE2MTQ2MTUsImlzcyI6InNhbGVzQHdlYXZlLndvcmtzIiwibmJmIjoxNjMxNjE0NjE1LCJzdWIiOiJ0ZWFtLXBlc3RvQHdlYXZlLndvcmtzIn0.klRpQQgbCtshC3PuuD4DdI3i-7Z0uSGQot23YpsETphFq4i3KK4NmgfnDg_WA3Pik-C2cJgG8WWYkWnemWQJAw`

func TestEntitlementMiddleware(t *testing.T) {
	ctx := context.Background()
	defer ctx.Done()

	c := createFakeClient(createSecret(validEntitlement))
	s, err := server.NewServer(ctx, c, client.ObjectKey{Name: "name", Namespace: "namespace"}, logr.Discard(), server.ParamSet{
		DbType: "sqlite",
		Port:   "8001",
	})
	if err != nil {
		t.Fatalf("expected no errors but got: %v", err)
	}

	go func(ctx context.Context) {
		s.ListenAndServe()
	}(ctx)

	res, err := http.Get("http://localhost:8001/gitops/healthz")
	if err != nil {
		t.Fatalf("expected no errors but got: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status code to be %d but got %d instead", http.StatusOK, res.StatusCode)
	}
}

func createFakeClient(clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
	}
	schemeBuilder.AddToScheme(scheme)

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
		Type: "Opaque",
		Data: map[string][]byte{"entitlement": []byte(s)},
	}
}
