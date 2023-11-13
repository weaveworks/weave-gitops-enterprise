package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"github.com/tonglil/buflogr"
	ctrl "github.com/weaveworks/gitopssets-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/internal/grpctesting"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/gitopssets"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/gitopssets/adapter"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/testing/protocmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

// var k8sEnv *K8sTestEnv

type K8sTestEnv struct {
	Env        *envtest.Environment
	Client     client.Client
	DynClient  dynamic.Interface
	RestMapper *restmapper.DeferredDiscoveryRESTMapper
	Rest       *rest.Config
	Stop       func()
}

const (
	MetadataUserKey   string = "test_principal_user"
	MetadataGroupsKey string = "test_principal_groups"
)

func TestListGitOpsSets(t *testing.T) {
	ctx := context.Background()

	obj := &ctrl.GitOpsSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsSet",
			APIVersion: "gitopssets.weave.works/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-obj",
			Namespace: "namespace-a-1",
		},
	}

	clusterClients := map[string]client.Client{
		"management": createClient(t, obj),
	}
	client := setup(t, clusterClients)

	res, err := client.ListGitOpsSets(ctx, &pb.ListGitOpsSetsRequest{})
	assert.NoError(t, err)

	expected := &pb.ListGitOpsSetsResponse{
		Gitopssets: []*pb.GitOpsSet{
			{
				Name:        obj.Name,
				Type:        "GitOpsSet",
				Namespace:   obj.Namespace,
				ClusterName: "management",
				ObjectRef: &pb.ObjectRef{
					Kind:        "GitOpsSet",
					Name:        obj.Name,
					Namespace:   obj.Namespace,
					ClusterName: "management",
				},
			},
		},
	}

	ignoreFields := protocmp.IgnoreFields(&pb.GitOpsSet{}, "yaml")
	if diff := cmp.Diff(expected, res, protocmp.Transform(), ignoreFields); diff != "" {
		t.Fatalf("expected %v, got %v, diff: %v", expected, res, diff)
	}
}

func TestListWithErrors(t *testing.T) {
	ctx := context.Background()

	obj := &ctrl.GitOpsSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsSet",
			APIVersion: "gitopssets.weave.works/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-obj",
			Namespace: "namespace-a-1",
		},
	}

	unusableClient := fake.NewClientBuilder().Build()

	clusterClients := map[string]client.Client{
		"management": createClient(t, obj),
		"leaf-1":     unusableClient,
	}
	client := setup(t, clusterClients)

	res, err := client.ListGitOpsSets(ctx, &pb.ListGitOpsSetsRequest{})
	assert.NoError(t, err)

	expected := &pb.ListGitOpsSetsResponse{
		Gitopssets: []*pb.GitOpsSet{
			{
				Name:        obj.Name,
				Type:        "GitOpsSet",
				Namespace:   obj.Namespace,
				ClusterName: "management",
				ObjectRef: &pb.ObjectRef{
					Kind:        "GitOpsSet",
					Name:        obj.Name,
					Namespace:   obj.Namespace,
					ClusterName: "management",
				},
			},
		},
		Errors: []*pb.GitOpsSetListError{
			{
				ClusterName: "leaf-1",
				Message:     "no kind is registered for the type v1alpha1.GitOpsSetList in scheme \"pkg/runtime/scheme.go:100\"",
				Namespace:   "namespace-x-1",
			},
		},
	}

	ignoreFields := protocmp.IgnoreFields(&pb.GitOpsSet{}, "yaml")
	if diff := cmp.Diff(expected, res, ignoreFields, protocmp.Transform()); diff != "" {
		t.Fatalf("expected %v, got %v, diff: %v", expected, res, diff)
	}
}

func TestToListErrors(t *testing.T) {
	// a multierror with 2 errors
	err := multierror.Append(
		&clustersmngr.ClientError{
			ClusterName: "cluster-1",
			Err:         errors.New("error 1"),
		},
		&clustersmngr.ClientError{
			ClusterName: "cluster-2",
			Err:         errors.New("error 2"),
		},
		// should be ignored
		errors.New("oh no"),
	).ErrorOrNil()

	errList, err := toListErrors(err)
	assert.NoError(t, err)

	expected := []*pb.GitOpsSetListError{
		{
			ClusterName: "cluster-1",
			Message:     "error 1",
		},
		{
			ClusterName: "cluster-2",
			Message:     "error 2",
		},
	}

	assert.Equal(t, expected, errList)

	if diff := cmp.Diff(expected, errList, protocmp.Transform()); diff != "" {
		t.Fatalf("expected %v, got %v, diff: %v", expected, errList, diff)
	}
}

func TestListWithMissingCRD(t *testing.T) {
	ctx := context.Background()

	obj := &ctrl.GitOpsSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsSet",
			APIVersion: "gitopssets.weave.works/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-obj",
			Namespace: "namespace-a-1",
		},
	}

	unusableClient := errorClient{
		err: &apimeta.NoResourceMatchError{
			PartialResource: schema.GroupVersionResource{
				Version:  ctrl.GroupVersion.Version,
				Group:    ctrl.GroupVersion.Group,
				Resource: "gitopssets",
			},
		},
	}

	clusterClients := map[string]client.Client{
		"management": createClient(t, obj),
		"leaf-1":     unusableClient,
	}

	buf := bytes.Buffer{}
	log := buflogr.NewWithBuffer(&buf)

	client := setup(t, clusterClients, func(opt ServerOpts) ServerOpts {
		opt.Logger = log
		return opt
	})

	res, err := client.ListGitOpsSets(ctx, &pb.ListGitOpsSetsRequest{})
	assert.NoError(t, err)

	expected := &pb.ListGitOpsSetsResponse{
		Gitopssets: []*pb.GitOpsSet{
			{
				Name:        obj.Name,
				Type:        "GitOpsSet",
				Namespace:   obj.Namespace,
				ClusterName: "management",
				ObjectRef: &pb.ObjectRef{
					Kind:        "GitOpsSet",
					Name:        obj.Name,
					Namespace:   obj.Namespace,
					ClusterName: "management",
				},
			},
		},
	}

	ignoreFields := protocmp.IgnoreFields(&pb.GitOpsSet{}, "yaml")
	if diff := cmp.Diff(expected, res, ignoreFields, protocmp.Transform()); diff != "" {
		t.Fatalf("expected %v, got %v, diff: %v", expected, res, diff)
	}

	// check for log message
	expectedLog := "INFO gitopssets crd not present on cluster, skipping error cluster leaf-1"
	if !strings.Contains(buf.String(), expectedLog) {
		t.Fatalf("expected log message %v, got %v", expectedLog, buf.String())
	}
}

func TestListWithNoRBAC(t *testing.T) {
	ctx := context.Background()

	obj := &ctrl.GitOpsSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsSet",
			APIVersion: "gitopssets.weave.works/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-obj",
			Namespace: "namespace-a-1",
		},
	}

	unusableClient := errorClient{
		err: &kerrors.StatusError{
			ErrStatus: metav1.Status{
				Reason: metav1.StatusReasonForbidden,
			},
		},
	}

	clusterClients := map[string]client.Client{
		"management": createClient(t, obj),
		"leaf-1":     unusableClient,
	}

	client := setup(t, clusterClients)

	res, err := client.ListGitOpsSets(ctx, &pb.ListGitOpsSetsRequest{})
	assert.NoError(t, err)

	expected := &pb.ListGitOpsSetsResponse{
		Gitopssets: []*pb.GitOpsSet{
			{
				Name:        obj.Name,
				Type:        "GitOpsSet",
				Namespace:   obj.Namespace,
				ClusterName: "management",
				ObjectRef: &pb.ObjectRef{
					Kind:        "GitOpsSet",
					Name:        obj.Name,
					Namespace:   obj.Namespace,
					ClusterName: "management",
				},
			},
		},
	}

	ignoreFields := protocmp.IgnoreFields(&pb.GitOpsSet{}, "yaml")
	if diff := cmp.Diff(expected, res, ignoreFields, protocmp.Transform()); diff != "" {
		t.Fatalf("expected %v, got %v, diff: %v", expected, res, diff)
	}
}

func TestSuspendGitOpsSet(t *testing.T) {
	ctx := context.Background()

	obj := &ctrl.GitOpsSet{}
	obj.Name = "my-obj"
	obj.Namespace = "namespace-a-1"
	obj.Spec = ctrl.GitOpsSetSpec{
		Suspend: false,
	}
	k8s := createClient(t, obj)
	clusterClients := map[string]client.Client{
		"management": k8s,
	}
	client := setup(t, clusterClients)

	// no kind is registered for the type v1alpha1.GitOpsSet in scheme "pkg/runtime/scheme.go:100" - how do we create objects in the absence of k8s?
	_, err := client.ToggleSuspendGitOpsSet(ctx, &pb.ToggleSuspendGitOpsSetRequest{
		Name:        obj.Name,
		Namespace:   obj.Namespace,
		ClusterName: "management",
		Suspend:     true,
	})
	assert.NoError(t, err)

	s := &ctrl.GitOpsSet{}
	key := types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}

	assert.NoError(t, k8s.Get(ctx, key, s))

	assert.True(t, s.Spec.Suspend, "expected Spec.Suspend to be true")
}

func TestGetReconciledObjects(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	gsName := "my-gs"
	ns1 := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "namespace-a-1",
		},
	}

	reconciledObjs := []runtime.Object{
		ns1,
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				UID:       "abc",
				Name:      "my-deployment",
				Namespace: ns1.Name,
				Labels: map[string]string{
					GitOpsSetNameKey:      gsName,
					GitOpsSetNamespaceKey: ns1.Name,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": gsName,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"app": gsName},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "nginx",
							Image: "nginx",
						}},
					},
				},
			},
		},
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				UID:       "efg",
				Name:      "my-configmap",
				Namespace: ns1.Name,
				Labels: map[string]string{
					GitOpsSetNameKey:      gsName,
					GitOpsSetNamespaceKey: ns1.Name,
				},
			},
		},
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				UID:       "hij",
				Name:      "my-configmap-2",
				Namespace: ns1.Name,
			},
		},
	}

	k8s := createClient(t, reconciledObjs...)
	clusterClients := map[string]client.Client{
		"management": k8s,
	}
	c := setup(t, clusterClients)

	type objectAssertion struct {
		kind string
		name string
	}

	tests := []struct {
		name            string
		user            string
		group           string
		expectedLen     int
		expectedObjects []objectAssertion
	}{
		{
			name:        "master user receives all objects",
			user:        "anne",
			group:       "system:masters",
			expectedLen: 2,
			expectedObjects: []objectAssertion{
				{
					kind: "Deployment",
					name: "my-deployment",
				},
				{
					kind: "ConfigMap",
					name: "my-configmap",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g = NewGomegaWithT(t)

			md := metadata.Pairs(MetadataUserKey, tt.user, MetadataGroupsKey, tt.group)
			outgoingCtx := metadata.NewOutgoingContext(ctx, md)
			res, err := c.GetReconciledObjects(outgoingCtx, &pb.GetReconciledObjectsRequest{
				Name:      gsName,
				Namespace: ns1.Name,
				Kinds: []*pb.GroupVersionKind{
					{Group: appsv1.SchemeGroupVersion.Group, Version: appsv1.SchemeGroupVersion.Version, Kind: "Deployment"},
					{Group: corev1.SchemeGroupVersion.Group, Version: corev1.SchemeGroupVersion.Version, Kind: "ConfigMap"},
				},
				ClusterName: "management",
			})

			fmt.Println(res.Objects)

			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(res.Objects).To(HaveLen(tt.expectedLen), "unexpected size of returned object list")

			actualObjs := make([]objectAssertion, len(res.Objects))

			for idx, actualObj := range res.Objects {
				var object map[string]interface{}

				g.Expect(json.Unmarshal([]byte(actualObj.Payload), &object)).To(Succeed(), "failed unmarshalling result object")
				metadata, ok := object["metadata"].(map[string]interface{})
				g.Expect(ok).To(BeTrue(), "object has unexpected metadata type")
				actualObjs[idx] = objectAssertion{
					kind: object["kind"].(string),
					name: metadata["name"].(string),
				}
			}
			g.Expect(actualObjs).To(ContainElements(tt.expectedObjects))
		})
	}
}

func setup(t *testing.T, clusterClients map[string]client.Client, opts ...func(opts ServerOpts) ServerOpts) pb.GitOpsSetsClient {
	clientsPool := &clustersmngrfakes.FakeClientsPool{}
	clientsPool.ClientsReturns(clusterClients)
	clientsPool.ClientStub = func(name string) (client.Client, error) {
		if c, found := clusterClients[name]; found && c != nil {
			return c, nil
		}
		return nil, fmt.Errorf("cluster %s not found", name)
	}
	namespaces := map[string][]corev1.Namespace{
		"management": {corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "namespace-a-1"}}},
		"leaf-1":     {corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "namespace-x-1"}}},
	}
	clustersClient := clustersmngr.NewClient(clientsPool, namespaces, logr.Discard())
	fakeFactory := &clustersmngrfakes.FakeClustersManager{}
	fakeFactory.GetImpersonatedClientForClusterReturns(clustersClient, nil)
	fakeFactory.GetImpersonatedClientReturns(clustersClient, nil)
	fakeFactory.GetUserNamespacesReturns(namespaces)

	log := testr.New(t)
	options := ServerOpts{
		ClientsFactory: fakeFactory,
		Logger:         log,
	}

	for _, o := range opts {
		options = o(options)
	}

	srv := NewGitOpsSetsServer(options)

	conn := grpctesting.Setup(t, func(s *grpc.Server) {
		pb.RegisterGitOpsSetsServer(s, srv)
	})

	return pb.NewGitOpsSetsClient(conn)
}

func createClient(t *testing.T, clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		appsv1.AddToScheme,
		corev1.AddToScheme,
		ctrl.AddToScheme,
	}
	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		t.Fatal(err)
	}

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(clusterState...).
		WithStatusSubresource(&ctrl.GitOpsSet{}).
		Build()

	return c
}

type errorClient struct {
	client.Client
	err error
}

func (s errorClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return s.err
}

func TestSyncGitOpsSet(t *testing.T) {
	ctx := context.Background()

	obj := &ctrl.GitOpsSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsSet",
			APIVersion: "gitopssets.weave.works/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-obj",
			Namespace: "default",
		},
	}
	cl := createClient(t, obj)
	clusterClients := map[string]client.Client{
		"management": cl,
	}
	gsClient := setup(t, clusterClients)

	key := types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}

	done := make(chan error)
	defer close(done)

	go func() {
		_, err := gsClient.SyncGitOpsSet(ctx, &pb.SyncGitOpsSetRequest{
			ClusterName: "management",
			Name:        obj.Name,
			Namespace:   obj.Namespace,
		})
		done <- err
	}()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:

			r := adapter.GitOpsSetAdapter{GitOpsSet: obj}

			if err := simulateReconcile(ctx, cl, key, r.AsClientObject()); err != nil {
				t.Fatalf("simulating reconcile: %s", err.Error())
			}

		case err := <-done:
			if err != nil {
				t.Errorf(err.Error())
			}
			return
		}
	}
}

func simulateReconcile(ctx context.Context, k client.Client, name types.NamespacedName, o client.Object) error {
	switch obj := o.(type) {
	case *ctrl.GitOpsSet:
		if err := k.Get(ctx, name, obj); err != nil {
			return err
		}

		obj.Status.SetLastHandledReconcileRequest(time.Now().Format(time.RFC3339Nano))
		return k.Status().Update(ctx, obj)
	}

	return errors.New("simulating reconcile: unsupported type")
}
