package server

import (
	"testing"
	"time"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"

	pacv2beta1 "github.com/weaveworks/policy-agent/api/v2beta1"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/capi/v1alpha1"
	gapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/gitopstemplate/v1alpha1"
	apitemplates "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"

	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/clusters"
)

func createClient(t *testing.T, clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		capiv1.AddToScheme,
		sourcev1.AddToScheme,
		pacv2beta1.AddToScheme,
		gitopsv1alpha1.AddToScheme,
		clusterv1.AddToScheme,
	}
	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		t.Fatal(err)
	}

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(clusterState...).
		Build()

	return c
}

type serverOptions struct {
	clusterState   []runtime.Object
	configMapName  string
	namespace      string
	provider       git.Provider
	ns             string
	hr             *sourcev1.HelmRepository
	clientsFactory clustersmngr.ClientsFactory
}

func createServer(t *testing.T, o serverOptions) capiv1_protos.ClustersServiceServer {
	c := createClient(t, o.clusterState...)
	dc := discovery.NewDiscoveryClient(fakeclientset.NewSimpleClientset().Discovery().RESTClient())

	return NewClusterServer(
		ServerOpts{
			Logger: logr.Discard(),
			TemplatesLibrary: &templates.ConfigMapLibrary{
				Log:           logr.Discard(),
				Client:        c,
				ConfigMapName: o.configMapName,
				CAPINamespace: o.namespace,
			},
			ClustersLibrary: &clusters.CRDLibrary{
				Log:          logr.Discard(),
				ClientGetter: kubefakes.NewFakeClientGetter(c),
				Namespace:    o.namespace,
			},
			ClientsFactory:            o.clientsFactory,
			GitProvider:               o.provider,
			ClientGetter:              kubefakes.NewFakeClientGetter(c),
			DiscoveryClient:           dc,
			ClustersNamespace:         o.ns,
			ProfileHelmRepositoryName: "weaveworks-charts",
			HelmRepositoryCacheDir:    t.TempDir(),
		},
	)
}

func makeTestHelmRepository(base string, opts ...func(*sourcev1.HelmRepository)) *sourcev1.HelmRepository {
	hr := &sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testing",
			Namespace: "test-ns",
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL:      base + "/charts",
			Interval: metav1.Duration{Duration: time.Minute * 10},
		},
		Status: sourcev1.HelmRepositoryStatus{
			URL: base + "/index.yaml",
		},
	}
	for _, o := range opts {
		o(hr)
	}
	return hr
}

func makeTemplateConfigMap(name string, s ...string) *corev1.ConfigMap {
	data := make(map[string]string)
	for i := 0; i < len(s); i += 2 {
		data[s[i]] = s[i+1]
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Data: data,
	}
}

func makeCAPITemplate(t *testing.T, opts ...func(*capiv1.CAPITemplate)) string {
	t.Helper()
	basicRaw := `
	{
		"apiVersion":"fooversion",
		"kind":"fookind",
		"metadata":{
		   "name":"${CLUSTER_NAME}",
		   "annotations":{
			  "capi.weave.works/display-name":"ClusterName"
		   }
		}
	 }`
	ct := &capiv1.CAPITemplate{
		Template: apitemplates.Template{
			TypeMeta: metav1.TypeMeta{
				Kind:       capiv1.Kind,
				APIVersion: "capi.weave.works/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "cluster-template-1",
			},
			Spec: apitemplates.TemplateSpec{
				Description: "this is test template 1",
				Params: []apitemplates.TemplateParam{
					{
						Name:        "CLUSTER_NAME",
						Description: "This is used for the cluster naming.",
					},
				},
				ResourceTemplates: []apitemplates.ResourceTemplate{
					{
						RawExtension: rawExtension(basicRaw),
					},
				},
			},
		},
	}
	for _, o := range opts {
		o(ct)
	}
	b, err := yaml.Marshal(ct)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func makeClusterTemplates(t *testing.T, opts ...func(template *gapiv1.GitOpsTemplate)) string {
	t.Helper()
	basicRaw := `
	{
		"apiVersion":"fooversion",
		"kind":"fookind",
		"metadata":{
		   "name":"${RESOURCE_NAME}",
		   "annotations":{
			  "clustertemplates.weave.works/display-name":"ClusterName"
		   }
		}
	 }`
	ct := &gapiv1.GitOpsTemplate{
		Template: apitemplates.Template{
			TypeMeta: metav1.TypeMeta{
				Kind:       gapiv1.Kind,
				APIVersion: "clustertemplates.weave.works/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "cluster-template-1",
			},
			Spec: apitemplates.TemplateSpec{
				Description: "this is test template 1",
				Params: []apitemplates.TemplateParam{
					{
						Name:        "RESOURCE_NAME",
						Description: "This is used for the resource naming.",
					},
				},
				ResourceTemplates: []apitemplates.ResourceTemplate{
					{
						RawExtension: rawExtension(basicRaw),
					},
				},
			},
		},
	}
	for _, o := range opts {
		o(ct)
	}
	b, err := yaml.Marshal(ct)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func rawExtension(s string) runtime.RawExtension {
	return runtime.RawExtension{
		Raw: []byte(s),
	}
}

func makePolicy(t *testing.T, opts ...func(p *pacv2beta1.Policy)) *pacv2beta1.Policy {
	t.Helper()
	policy := &pacv2beta1.Policy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Policy",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "weave.policies.missing-owner-label",
		},
		Spec: pacv2beta1.PolicySpec{
			Name:     "Missing Owner Label",
			Severity: "high",
			Code:     "foo",
			Targets: pacv2beta1.PolicyTargets{
				Labels:     []map[string]string{{"my-label": "my-value"}},
				Kinds:      []string{},
				Namespaces: []string{},
			},
			Standards: []pacv2beta1.PolicyStandard{},
		},
	}
	for _, o := range opts {
		o(policy)
	}
	return policy
}

func makeEvent(t *testing.T, opts ...func(e *corev1.Event)) *corev1.Event {
	t.Helper()
	event := &corev1.Event{
		InvolvedObject: corev1.ObjectReference{
			APIVersion:      "v1",
			Kind:            "Deployment",
			Name:            "my-deployment",
			Namespace:       "default",
			ResourceVersion: "1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"policy_name":     "Missing app Label",
				"cluster_id":      "cluster-1",
				"category":        "Access Control",
				"severity":        "high",
				"description":     "Missing app label",
				"how_to_solve":    "how_to_solve",
				"entity_manifest": `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"nginx-deployment","namespace":"default","uid":"af912668-957b-46d4-bc7a-51e6994cba56"},"spec":{"template":{"spec":{"containers":[{"image":"nginx:latest","imagePullPolicy":"Always","name":"nginx","ports":[{"containerPort":80,"protocol":"TCP"}]}]}}}}`,
				"occurrences":     `[{"message": "occurrence details"}]`,
			},
			Labels: map[string]string{
				"pac.weave.works/type": "Admission",
				"pac.weave.works/id":   "66101548-12c1-4f79-a09a-a12979903fba",
			},
			Name:      "Missing app Label - fake-event-1",
			Namespace: "default",
		},
		Message: "Policy event",
		Reason:  "PolicyViolation",
		Type:    "Warning",
	}
	for _, o := range opts {
		o(event)
	}
	return event
}
