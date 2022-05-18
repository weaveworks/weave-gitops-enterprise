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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"

	policiesv1 "github.com/weaveworks/policy-agent/api/v1"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
)

func createClient(t *testing.T, clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		capiv1.AddToScheme,
		sourcev1.AddToScheme,
		policiesv1.AddToScheme,
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

func createServer(t *testing.T, clusterState []runtime.Object, configMapName, namespace string, provider git.Provider, ns string, hr *sourcev1.HelmRepository) capiv1_protos.ClustersServiceServer {
	c := createClient(t, clusterState...)
	dc := discovery.NewDiscoveryClient(fakeclientset.NewSimpleClientset().Discovery().RESTClient())

	return NewClusterServer(
		logr.Discard(),
		nil,
		&templates.ConfigMapLibrary{
			Log:           logr.Discard(),
			Client:        c,
			ConfigMapName: configMapName,
			Namespace:     namespace,
		},
		provider,
		kubefakes.NewFakeClientGetter(c),
		dc,
		ns,
		"weaveworks-charts", t.TempDir(),
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

func makeTemplateConfigMap(s ...string) *corev1.ConfigMap {
	data := make(map[string]string)
	for i := 0; i < len(s); i += 2 {
		data[s[i]] = s[i+1]
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "capi-templates",
			Namespace: "default",
		},
		Data: data,
	}
}

func makeTemplate(t *testing.T, opts ...func(*capiv1.CAPITemplate)) string {
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
		TypeMeta: metav1.TypeMeta{
			Kind:       "CAPITemplate",
			APIVersion: "capi.weave.works/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-template-1",
		},
		Spec: capiv1.CAPITemplateSpec{
			Description: "this is test template 1",
			Params: []capiv1.TemplateParam{
				{
					Name:        "CLUSTER_NAME",
					Description: "This is used for the cluster naming.",
				},
			},
			ResourceTemplates: []capiv1.CAPIResourceTemplate{
				{
					RawExtension: rawExtension(basicRaw),
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

func makePolicy(t *testing.T, opts ...func(p *policiesv1.Policy)) *policiesv1.Policy {
	t.Helper()
	policy := &policiesv1.Policy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Policy",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "weave.policies.missing-owner-label",
		},
		Spec: policiesv1.PolicySpec{
			Name:     "Missing Owner Label",
			Severity: "high",
			Code:     "foo",
			Targets: policiesv1.PolicyTargets{
				Labels: []map[string]string{{"my-label": "my-value"}},
			},
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
			},
			Labels: map[string]string{
				"pac.weave.works/type": "Admission",
				"pac.weave.works/id":   "weave.policies.missing-app-label",
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
