package mgmtfetcher

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	gapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/gitopstemplate/v1alpha2"
	mgmtfake "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/mgmtfetcher/fake"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestFetch(t *testing.T) {
	testCases := []struct {
		name            string
		clusterState    []runtime.Object
		namespacesCache NamespacesCache
		resourceKind    string
		fn              returnListFactory
		want            []NamespacedList
		err             error
	}{
		{
			name: "get gitopsclusters",
			clusterState: []runtime.Object{
				&gitopsv1alpha1.GitopsCluster{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1alpha1",
						Kind:       "GitopsCluster",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:            "cluster-test",
						Namespace:       "ns-test",
						ResourceVersion: "0",
					},
				},
			},
			namespacesCache: &mgmtfake.FakeNamespaceCache{
				Namespaces: []*v1.Namespace{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "ns-test",
						},
						TypeMeta: metav1.TypeMeta{
							APIVersion: "v1",
							Kind:       "Namespace",
						},
					},
				},
			},
			resourceKind: "GitopsCluster",
			fn: func() client.ObjectList {
				return &gitopsv1alpha1.GitopsClusterList{}
			},
			want: []NamespacedList{
				{
					Namespace: "ns-test",
					List: &gitopsv1alpha1.GitopsClusterList{
						TypeMeta: metav1.TypeMeta{
							Kind:       "GitopsClusterList",
							APIVersion: "gitops.weave.works/v1alpha1",
						},
						Items: []gitopsv1alpha1.GitopsCluster{
							{
								TypeMeta: metav1.TypeMeta{
									APIVersion: "v1alpha1",
									Kind:       "GitopsCluster",
								},
								ObjectMeta: metav1.ObjectMeta{
									Name:            "cluster-test",
									Namespace:       "ns-test",
									ResourceVersion: "0",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "failed to list namespaces from cache",
			namespacesCache: &mgmtfake.FakeNamespaceCache{
				Err: errors.New("cache error"),
			},
			resourceKind: "GitopsCluster",
			fn: func() client.ObjectList {
				return &gitopsv1alpha1.GitopsClusterList{}
			},
			err: errors.New("cache error"),
		},
		{
			name:            "unsupported resource",
			resourceKind:    "invalid",
			namespacesCache: &mgmtfake.FakeNamespaceCache{},
			fn: func() client.ObjectList {
				return &gitopsv1alpha1.GitopsClusterList{}
			},
			err: errors.New("unsupported resource kind: invalid"),
		},
	}

	cliGetter := &mgmtfake.FakeAuthClientGetter{}
	ctx := auth.WithPrincipal(context.Background(), &auth.UserPrincipal{ID: "userID"})
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			c := createClient(t, tt.clusterState...)

			mgmtFetcher := NewManagementCrossNamespacesFetcher(tt.namespacesCache, kubefakes.NewFakeClientGetter(c), cliGetter)
			got, err := mgmtFetcher.Fetch(ctx, tt.resourceKind, tt.fn)
			if err != nil {
				assert.Equal(t, tt.err.Error(), err.Error())
			}
			assert.Equal(t, tt.want, got)
		})
	}

}

func createClient(t *testing.T, clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		gitopsv1alpha1.AddToScheme,
		gapiv1.AddToScheme,
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
