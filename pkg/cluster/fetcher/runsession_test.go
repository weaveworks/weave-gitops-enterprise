package fetcher_test

import (
	"context"
	"testing"

	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/fetcher"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster/clusterfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestRunSessionFetcher(t *testing.T) {
	g := NewGomegaWithT(t)

	testCases := []struct {
		context        string
		clusterObjects []runtime.Object
		expectedCount  int
	}{
		{
			context: "fetches vcluster clusters with correct labels",
			clusterObjects: []runtime.Object{
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "run-head-somecommit",
						Labels: map[string]string{
							"app":                       "vcluster",
							"app.kubernetes.io/part-of": "gitops-run",
						},
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "vc-run-head-somecommit",
					},
					Data: map[string][]byte{
						"config": secretData("some-name"),
					},
				},
			},
			expectedCount: 1,
		},
		{
			context: "doesn't fetch vcluster that aren't part of gitops run",
			clusterObjects: []runtime.Object{
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "run-head-somecommit",
						Labels: map[string]string{
							"app": "vcluster",
						},
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "vc-run-head-somecommit",
					},
					Data: map[string][]byte{
						"config": secretData("some-name"),
					},
				},
			},
			expectedCount: 0,
		},
		{
			context: "ignores clusters with missing secret",
			clusterObjects: []runtime.Object{
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "run-head-somecommit",
						Labels: map[string]string{
							"app":                       "vcluster",
							"app.kubernetes.io/part-of": "gitops-run",
						},
					},
				},
			},
			expectedCount: 0,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.context, func(t *testing.T) {
			scheme, err := kube.CreateScheme()
			g.Expect(err).NotTo(HaveOccurred())
			client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(tt.clusterObjects...).Build()

			cluster := new(clusterfakes.FakeCluster)
			cluster.GetNameReturns("management")
			cluster.GetServerClientReturns(client, nil)

			fetcher := fetcher.NewRunSessionFetcher(testr.New(t), cluster, scheme, false, kube.UserPrefixes{})

			clusters, err := fetcher.Fetch(context.TODO())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(clusters).To(HaveLen(tt.expectedCount))
		})
	}
}
