package tenants

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	pacv2beta1 "github.com/weaveworks/policy-agent/api/v2beta1"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/tenancy"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	apiextentionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	crdTypeMeta = metav1.TypeMeta{Kind: "CustomResourceDefinition", APIVersion: "v1"}
)

func Test_PreFlightChecks(t *testing.T) {
	tt := []struct {
		name         string
		tenants      *tenancy.Config
		clusterState []runtime.Object
		expectError  bool
	}{
		{
			name: "no policies applied to any tenants",
			tenants: &tenancy.Config{
				Tenants: []tenancy.Tenant{
					{
						Name: "test-tenant-01",
						Namespaces: []string{
							"test-ns-01", "test-ns-02",
						},
					},
					{
						Name: "test-tenant-02",
						Namespaces: []string{
							"test-ns-03", "test-ns-04",
						},
					},
				},
			},
			clusterState: []runtime.Object{},
			expectError:  false,
		},
		{
			name: "tenant contains policy and crd exists",
			tenants: &tenancy.Config{
				Tenants: []tenancy.Tenant{
					{
						Name:       "test-tenant-01",
						Namespaces: []string{"test-ns-01"},
						AllowedRepositories: []tenancy.AllowedRepository{
							{
								URL:  "https://github.com/testorg/testrepo",
								Kind: "GitRepository",
							},
						},
					},
				},
			},
			clusterState: []runtime.Object{
				setResourceVersion(newCrd(policyCRDName), 1),
			},
			expectError: false,
		},
		{
			name: "tenant contains allowed repo policy and crd does not exist",
			tenants: &tenancy.Config{
				Tenants: []tenancy.Tenant{
					{
						Name:       "test-tenant-01",
						Namespaces: []string{"test-ns-01"},
						AllowedRepositories: []tenancy.AllowedRepository{
							{
								URL:  "https://github.com/testorg/testrepo",
								Kind: "GitRepository",
							},
						},
					},
				},
			},
			clusterState: []runtime.Object{},
			expectError:  true,
		},
		{
			name: "tenant contains allowed cluster policy and crd does not exist",
			tenants: &tenancy.Config{
				Tenants: []tenancy.Tenant{
					{
						Name:       "test-tenant-01",
						Namespaces: []string{"test-ns-01"},
						AllowedClusters: []tenancy.AllowedCluster{
							{
								KubeConfig: "some-cluster-name",
							},
						},
					},
				},
			},
			clusterState: []runtime.Object{},
			expectError:  true,
		},
	}

	for _, tc := range tt {
		mockClient := newFakeClient(t, tc.clusterState...)
		kubeClient := &kube.KubeHTTP{Client: mockClient}

		err := preFlightCheck(context.TODO(), tc.tenants, kubeClient)
		switch {
		case tc.expectError:
			assert.Error(t, err)
			assert.True(t, apierrors.IsNotFound(err))
		default:
			assert.NoError(t, err)
		}
	}
}

func newFakeClient(t *testing.T, objs ...runtime.Object) client.Client {
	t.Helper()

	scheme := runtime.NewScheme()

	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	if err := pacv2beta1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	if err := apiextentionsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objs...).
		Build()
}

func newCrd(name string) client.Object {
	return &apiextentionsv1.CustomResourceDefinition{
		TypeMeta: crdTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func setResourceVersion[T client.Object](obj T, rv int) T {
	obj.SetResourceVersion(fmt.Sprintf("%v", rv))

	return obj
}
