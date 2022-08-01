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
	gapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/gitopstemplate/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
)

var testdata1 = []byte(`
apiVersion: capi.weave.works/v1alpha1
kind: CAPITemplate
metadata:
  name: cluster-template
spec:
  description: this is test template 1
`)

var testdata2 = []byte(`
apiVersion: capi.weave.works/v1alpha1
kind: CAPITemplate
metadata:
  name: cluster-template2
spec:
  description: this is test template 2
`)

func makeClient(t *testing.T, clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		capiv1.AddToScheme,
		gapiv1.AddToScheme,
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
	t1 := parseCAPITemplateFromBytes(t, testdata1)
	t2 := parseCAPITemplateFromBytes(t, testdata2)
	lib := CRDLibrary{Log: logr.Discard(), ClientGetter: kubefakes.NewFakeClientGetter(makeClient(t, t1, t2))}
	result, err := lib.Get(context.Background(), "cluster-template2", capiv1.Kind)
	if err != nil {
		t.Fatalf("On no, error: %v", err)
	}
	if diff := cmp.Diff(t2, result); diff != "" {
		t.Fatalf("On no, diff templates: %v", diff)
	}
}

func TestListTemplateFromCAPICRDs(t *testing.T) {
	t1 := parseCAPITemplateFromBytes(t, testdata1)
	t2 := parseCAPITemplateFromBytes(t, testdata2)
	lib := CRDLibrary{Log: logr.Discard(), ClientGetter: kubefakes.NewFakeClientGetter(makeClient(t, t1, t2))}
	result, err := lib.List(context.Background(), capiv1.Kind)
	if err != nil {
		t.Fatalf("On no, error: %v", err)
	}
	want := map[string]templates.Template{
		"cluster-template":  t1,
		"cluster-template2": t2,
	}
	if diff := cmp.Diff(want, result); diff != "" {
		t.Fatalf("On no, diff templates: %v", diff)
	}
}
