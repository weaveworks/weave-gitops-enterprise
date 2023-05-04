package server

import (
	"errors"
	"fmt"
	"testing"
	"time"

	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster/clusterfakes"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"

	pacv2beta1 "github.com/weaveworks/policy-agent/api/v2beta1"
	pacv2beta2 "github.com/weaveworks/policy-agent/api/v2beta2"

	capiv1 "github.com/weaveworks/templates-controller/apis/capi/v1alpha2"
	apitemplates "github.com/weaveworks/templates-controller/apis/core"
	gapiv1 "github.com/weaveworks/templates-controller/apis/gitops/v1alpha2"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/mgmtfetcher"
	mgmtfetcherfake "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/mgmtfetcher/fake"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/estimation"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/helmfakes"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func newTestScheme(t *testing.T) *runtime.Scheme {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		capiv1.AddToScheme,
		sourcev1beta2.AddToScheme,
		pacv2beta2.AddToScheme,
		pacv2beta1.AddToScheme,
		gitopsv1alpha1.AddToScheme,
		gapiv1.AddToScheme,
		clusterv1.AddToScheme,
		rbacv1.AddToScheme,
		esv1beta1.AddToScheme,
		kustomizev1.AddToScheme,
	}
	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		t.Fatal(err)
	}

	return scheme
}

func createClient(t *testing.T, clusterState ...runtime.Object) client.Client {
	scheme := newTestScheme(t)

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(clusterState...).
		WithIndex(&corev1.Event{}, "type", client.IndexerFunc(func(o client.Object) []string {
			event := o.(*corev1.Event)
			return []string{event.Type}
		})).Build()
}

type serverOptions struct {
	clusterState          []runtime.Object
	client                client.Client
	namespace             string
	provider              git.Provider
	ns                    string
	profileHelmRepository *types.NamespacedName
	clustersManager       clustersmngr.ClustersManager
	capiEnabled           bool
	chartsCache           helm.ChartsCacheReader
	chartJobs             *helm.Jobs
	valuesFetcher         helm.ValuesFetcher
	cluster               string
	estimator             estimation.Estimator
}

func getServer(t *testing.T, clients map[string]client.Client, namespaces map[string][]corev1.Namespace) capiv1_protos.ClustersServiceServer {
	clientsPool := &clustersmngrfakes.FakeClientsPool{}
	clientsPool.ClientsReturns(clients)
	clientsPool.ClientStub = func(name string) (client.Client, error) {
		if c, found := clients[name]; found && c != nil {
			return c, nil
		}
		return nil, fmt.Errorf("cluster %s not found", name)
	}
	clustersClient := clustersmngr.NewClient(clientsPool, namespaces, logr.Discard())
	fakeFactory := &clustersmngrfakes.FakeClustersManager{}
	fakeFactory.GetImpersonatedClientForClusterReturns(clustersClient, nil)
	fakeFactory.GetImpersonatedClientReturns(clustersClient, nil)

	return createServer(t, serverOptions{
		clustersManager: fakeFactory,
	})
}

func createServer(t *testing.T, o serverOptions) capiv1_protos.ClustersServiceServer {
	c := o.client
	if c == nil {
		c = createClient(t, o.clusterState...)
	}
	dc := discovery.NewDiscoveryClient(fakeclientset.NewSimpleClientset().Discovery().RESTClient())

	mgmtFetcher := mgmtfetcher.NewManagementCrossNamespacesFetcher(&mgmtfetcherfake.FakeNamespaceCache{
		Namespaces: []*corev1.Namespace{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Namespace",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-ns",
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Namespace",
				},
			},
		},
	}, kubefakes.NewFakeClientGetter(c), &mgmtfetcherfake.FakeAuthClientGetter{})

	if o.estimator == nil {
		o.estimator = estimation.NilEstimator()
	}

	if o.profileHelmRepository == nil {
		o.profileHelmRepository = &types.NamespacedName{
			Name:      "weaveworks-charts",
			Namespace: "flux-system",
		}
	}

	if o.cluster == "" {
		o.cluster = "management"
	}

	return NewClusterServer(
		ServerOpts{
			Logger:                testr.New(t),
			ClustersManager:       o.clustersManager,
			GitProvider:           o.provider,
			ClientGetter:          kubefakes.NewFakeClientGetter(c),
			DiscoveryClient:       dc,
			ClustersNamespace:     o.ns,
			ProfileHelmRepository: *o.profileHelmRepository,
			CAPIEnabled:           o.capiEnabled,
			RestConfig:            &rest.Config{},
			ChartJobs:             o.chartJobs,
			ChartsCache:           o.chartsCache,
			ValuesFetcher:         o.valuesFetcher,
			ManagementFetcher:     mgmtFetcher,
			Cluster:               o.cluster,
			Estimator:             o.estimator,
		},
	)
}

func makeTestClustersManager(t *testing.T, clusterState ...runtime.Object) *clustersmngrfakes.FakeClustersManager {
	clientsPool := &clustersmngrfakes.FakeClientsPool{}
	fakeCl := createClient(t, clusterState...)
	clients := map[string]client.Client{"management": fakeCl}
	clientsPool.ClientsReturns(clients)
	clientsPool.ClientReturns(fakeCl, nil)
	clientsPool.ClientStub = func(name string) (client.Client, error) {
		if c, found := clients[name]; found && c != nil {
			return c, nil
		}
		return nil, fmt.Errorf("cluster %s not found", name)
	}
	clustersClient := clustersmngr.NewClient(clientsPool, map[string][]corev1.Namespace{}, logr.Discard())
	fakeFactory := &clustersmngrfakes.FakeClustersManager{}
	fakeFactory.GetImpersonatedClientReturns(clustersClient, nil)
	fakeFactory.GetImpersonatedClientForClusterReturns(clustersClient, nil)
	fakeCluster := &clusterfakes.FakeCluster{}
	fakeCluster.GetNameReturns("management")
	fakeFactory.GetClustersReturns([]cluster.Cluster{fakeCluster})
	return fakeFactory
}

func makeTestHelmRepository(base string, opts ...func(*sourcev1beta2.HelmRepository)) *sourcev1beta2.HelmRepository {
	hr := &sourcev1beta2.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1beta2.HelmRepositoryKind,
			APIVersion: sourcev1beta2.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testing",
			Namespace: "test-ns",
		},
		Spec: sourcev1beta2.HelmRepositorySpec{
			URL:      base + "/charts",
			Interval: metav1.Duration{Duration: time.Minute * 10},
		},
		Status: sourcev1beta2.HelmRepositoryStatus{
			URL: base + "/index.yaml",
		},
	}
	for _, o := range opts {
		o(hr)
	}
	return hr
}

func makeCAPITemplate(t *testing.T, opts ...func(*capiv1.CAPITemplate)) *capiv1.CAPITemplate {
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
			Kind:       capiv1.Kind,
			APIVersion: "capi.weave.works/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-template-1",
			Namespace: "default",
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
					Content: []apitemplates.ResourceTemplateContent{
						{
							RawExtension: rawExtension(basicRaw),
						},
					},
				},
			},
		},
	}
	for _, o := range opts {
		o(ct)
	}
	return ct
}

func makeClusterTemplates(t *testing.T, opts ...func(template *gapiv1.GitOpsTemplate)) *gapiv1.GitOpsTemplate {
	t.Helper()
	basicRaw := `
	{
		"apiVersion":"fooversion",
		"kind":"fookind",
		"metadata":{
		   "name":"${RESOURCE_NAME}",
		   "annotations":{
			  "templates.weave.works/display-name":"ClusterName"
		   }
		}
	 }`
	ct := &gapiv1.GitOpsTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       gapiv1.Kind,
			APIVersion: "templates.weave.works/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-template-1",
			Namespace: "default",
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
					Content: []apitemplates.ResourceTemplateContent{
						{
							RawExtension: rawExtension(basicRaw),
						},
					},
				},
			},
		},
	}
	for _, o := range opts {
		o(ct)
	}
	return ct
}

func rawExtension(s string) runtime.RawExtension {
	return runtime.RawExtension{
		Raw: []byte(s),
	}
}

func makePolicy(t *testing.T, opts ...func(p *pacv2beta2.Policy)) *pacv2beta2.Policy {
	t.Helper()
	policy := &pacv2beta2.Policy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Policy",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "weave.policies.missing-owner-label",
		},
		Spec: pacv2beta2.PolicySpec{
			Name:     "Missing Owner Label",
			Severity: "high",
			Code:     "foo",
			Targets: pacv2beta2.PolicyTargets{
				Labels:     []map[string]string{{"my-label": "my-value"}},
				Kinds:      []string{},
				Namespaces: []string{},
			},
			Standards: []pacv2beta2.PolicyStandard{},
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
				"policy_id":       "weave.policies.missing-app-label",
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

func testNewFakeChartCache(t *testing.T, clusterRef types.NamespacedName, repoRef helm.ObjectReference, charts []helm.Chart) helmfakes.FakeChartCache {
	fc := helmfakes.NewFakeChartCache(helmfakes.WithCharts(
		helmfakes.ClusterRefToString(
			repoRef,
			clusterRef,
		),
		charts,
	))

	return fc
}

func nsn(name, namespace string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
}

var notImplementedError = errors.New("not implemented")

type mockTracker struct {
	getImpl func(gvr schema.GroupVersionResource, ns, name string) (runtime.Object, error)
}

func (t *mockTracker) Add(obj runtime.Object) error {
	return notImplementedError
}

func (t *mockTracker) Get(gvr schema.GroupVersionResource, ns, name string) (runtime.Object, error) {
	return t.getImpl(gvr, ns, name)
}

func (t *mockTracker) Create(gvr schema.GroupVersionResource, obj runtime.Object, ns string) error {
	return notImplementedError
}

func (t *mockTracker) Update(gvr schema.GroupVersionResource, obj runtime.Object, ns string) error {
	return notImplementedError
}

func (t *mockTracker) List(gvr schema.GroupVersionResource, gvk schema.GroupVersionKind, ns string) (runtime.Object, error) {
	return nil, notImplementedError
}

func (t *mockTracker) Delete(gvr schema.GroupVersionResource, ns, name string) error {
	return notImplementedError
}

func (t *mockTracker) Watch(gvr schema.GroupVersionResource, ns string) (watch.Interface, error) {
	return nil, notImplementedError
}
