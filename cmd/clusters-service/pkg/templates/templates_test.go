package templates

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/capi/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
	tapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/tfcontroller/v1alpha1"
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

func makeClient(t *testing.T, clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		capiv1.AddToScheme,
		tapiv1.AddToScheme,
	}
	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		t.Fatal(err)
	}

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(clusterState...).
		Build()
}

func TestGetTemplateFromCAPICRDs(t *testing.T) {
	t1 := mustParseCAPITemplate(t, testdata1)
	t2 := mustParseCAPITemplate(t, testdata2)
	lib := CRDLibrary{Log: logr.Discard(), ClientGetter: kubefakes.NewFakeClientGetter(makeClient(t, t1, t2))}
	result, err := lib.Get(context.Background(), "cluster-template2", capiv1.Kind)
	if err != nil {
		t.Fatalf("On no, error: %v", err)
	}
	if diff := cmp.Diff(&t2.Template, result); diff != "" {
		t.Fatalf("On no, diff templates: %v", diff)
	}
}

func TestListTemplateFromCAPICRDs(t *testing.T) {
	t1 := mustParseCAPITemplate(t, testdata1)
	t2 := mustParseCAPITemplate(t, testdata2)
	lib := CRDLibrary{Log: logr.Discard(), ClientGetter: kubefakes.NewFakeClientGetter(makeClient(t, t1, t2))}
	result, err := lib.List(context.Background(), capiv1.Kind)
	if err != nil {
		t.Fatalf("On no, error: %v", err)
	}
	want := map[string]*templates.Template{
		"cluster-template":  &t1.Template,
		"cluster-template2": &t2.Template,
	}
	if diff := cmp.Diff(want, result); diff != "" {
		t.Fatalf("On no, diff templates: %v", diff)
	}
}

func mustParseCAPITemplate(t *testing.T, data string) *capiv1.CAPITemplate {
	t.Helper()
	parsed, err := ParseBytes([]byte(data), "a-key")
	if err != nil {
		t.Fatal(err)
	}
	return &capiv1.CAPITemplate{Template: *parsed}
}
