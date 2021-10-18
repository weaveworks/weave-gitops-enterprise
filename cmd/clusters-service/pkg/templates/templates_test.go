package templates

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/capi"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const testdata1 = `
apiVersion: capi.weave.works/v1alpha1
kind: CAPITemplate
metadata:
  name: cluster-template
spec:
  description: this is test template 1
`

const testdata2 = `
apiVersion: capi.weave.works/v1alpha1
kind: CAPITemplate
metadata:
  name: cluster-template2
spec:
  description: this is test template 2
`

func makeClient(clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		capiv1.AddToScheme,
	}
	schemeBuilder.AddToScheme(scheme)

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(clusterState...).
		Build()
}

func TestGetTemplateFromCRDs(t *testing.T) {
	t1 := mustParseTemplate(t, testdata1)
	t2 := mustParseTemplate(t, testdata2)
	lib := CRDLibrary{Log: logr.Discard(), Client: makeClient(t1, t2)}
	result, err := lib.Get(context.Background(), "cluster-template2")
	if err != nil {
		t.Fatalf("On no, error: %v", err)
	}
	if diff := cmp.Diff(t2, result); diff != "" {
		t.Fatalf("On no, diff templates: %v", diff)
	}
}

func TestListTemplateFromCRDs(t *testing.T) {
	t1 := mustParseTemplate(t, testdata1)
	t2 := mustParseTemplate(t, testdata2)
	lib := CRDLibrary{Log: logr.Discard(), Client: makeClient(t1, t2)}
	result, err := lib.List(context.Background())
	if err != nil {
		t.Fatalf("On no, error: %v", err)
	}
	want := map[string]*capiv1.CAPITemplate{
		"cluster-template":  t1,
		"cluster-template2": t2,
	}
	if diff := cmp.Diff(want, result); diff != "" {
		t.Fatalf("On no, diff templates: %v", diff)
	}
}

func mustParseTemplate(t *testing.T, data string) *capiv1.CAPITemplate {
	t.Helper()
	parsed, err := capi.ParseBytes([]byte(data), "a-key")
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
