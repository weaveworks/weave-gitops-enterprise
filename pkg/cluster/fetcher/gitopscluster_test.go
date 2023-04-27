package fetcher_test

import (
	"context"
	"strings"
	"testing"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	capiv1 "github.com/weaveworks/templates-controller/apis/capi/v1alpha2"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/fetcher"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster/clusterfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGitopsFetcher(t *testing.T) {
	g := NewGomegaWithT(t)

	clusterName := "gitops-cluster"
	secretName := "dev"

	testCases := []struct {
		context        string
		clusterObjects []runtime.Object
		expectedCount  int
	}{
		{
			context:        "when no gitops clusters found, returns nothing",
			clusterObjects: []runtime.Object{},
			expectedCount:  0,
		},
		{
			context: "fetches clusters with capi ref",
			clusterObjects: []runtime.Object{
				makeTestCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = clusterName
					o.Spec.CAPIClusterRef = &meta.LocalObjectReference{
						Name: secretName,
					}
				}),
				makeTestSecret(func(o *corev1.Secret) {
					o.ObjectMeta.Name = secretName + "-kubeconfig"
					o.Data = map[string][]byte{
						"value": secretData(clusterName),
					}
				}),
			},
			expectedCount: 1,
		},
		{
			context: "fetches clusters with secret ref",
			clusterObjects: []runtime.Object{
				makeTestCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = clusterName
					o.Spec.SecretRef = &meta.LocalObjectReference{
						Name: secretName,
					}
				}),
				makeTestSecret(func(o *corev1.Secret) {
					o.ObjectMeta.Name = secretName
					o.Data = map[string][]byte{
						"value": secretData(clusterName),
					}
				}),
			},
			expectedCount: 1,
		},
		{
			context: "if both capi ref and secret ref are configured for cluster, prioritizes capi ref",
			clusterObjects: []runtime.Object{
				makeTestCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = clusterName
					o.Spec.CAPIClusterRef = &meta.LocalObjectReference{
						Name: secretName,
					}
					o.Spec.SecretRef = &meta.LocalObjectReference{
						Name: secretName,
					}
				}),
				makeTestSecret(func(o *corev1.Secret) {
					o.ObjectMeta.Name = secretName + "-kubeconfig"
					o.Data = map[string][]byte{
						"value": secretData(clusterName),
					}
				}),
			},
			expectedCount: 1,
		},
		{
			context: "when no secret ref set for gitops cluster, does not fail, skips adding cluster to list",
			clusterObjects: []runtime.Object{
				makeTestCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = clusterName
				}),
			},
			expectedCount: 0,
		},
		{
			context: "when no secret found for gitops cluster, does not fail, skips adding cluster to list",
			clusterObjects: []runtime.Object{
				makeTestCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = clusterName
					o.Spec.CAPIClusterRef = &meta.LocalObjectReference{
						Name: secretName,
					}
				}),
			},
			expectedCount: 0,
		},
		{
			context: "fetches cluster when secret data key is value.yaml",
			clusterObjects: []runtime.Object{
				makeTestCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = clusterName
					o.Spec.SecretRef = &meta.LocalObjectReference{
						Name: secretName,
					}
				}),
				makeTestSecret(func(o *corev1.Secret) {
					o.ObjectMeta.Name = secretName
					o.Data = map[string][]byte{
						"value.yaml": secretData(clusterName),
					}
				}),
			},
			expectedCount: 1,
		},
		{
			context: "when secret does not contain data key, does not fail, skips adding cluster to list",
			clusterObjects: []runtime.Object{
				makeTestCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = clusterName
					o.Spec.SecretRef = &meta.LocalObjectReference{
						Name: secretName,
					}
				}),
				makeTestSecret(func(o *corev1.Secret) {
					o.ObjectMeta.Name = secretName
					o.Data = map[string][]byte{}
				}),
			},
			expectedCount: 0,
		},
		{
			context: "when secret data does not contain valid kubeconfig, does not fail, skips adding cluster to list",
			clusterObjects: []runtime.Object{
				makeTestCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = clusterName
					o.Spec.SecretRef = &meta.LocalObjectReference{
						Name: secretName,
					}
				}),
				makeTestSecret(func(o *corev1.Secret) {
					o.ObjectMeta.Name = secretName
					o.Data = map[string][]byte{
						"value": []byte("foo"),
					}
				}),
			},
			expectedCount: 0,
		},
		{
			context: "when cluster is not ready, it is not added",
			clusterObjects: []runtime.Object{
				makeTestCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = clusterName
					o.Spec.CAPIClusterRef = &meta.LocalObjectReference{
						Name: secretName,
					}

					// Remove ready status
					o.SetConditions(nil)
				}),
				makeTestSecret(func(o *corev1.Secret) {
					o.ObjectMeta.Name = secretName + "-kubeconfig"
					o.Data = map[string][]byte{
						"value": secretData(clusterName),
					}
				}),
			},
			expectedCount: 0,
		},
		{
			context: "when cluster has not connectivity, it is not added",
			clusterObjects: []runtime.Object{
				makeTestCluster(func(o *gitopsv1alpha1.GitopsCluster) {
					o.ObjectMeta.Name = clusterName
					o.Spec.CAPIClusterRef = &meta.LocalObjectReference{
						Name: secretName,
					}

					// Remove connectivity status
					o.SetConditions([]metav1.Condition{
						{Type: meta.ReadyCondition, Status: metav1.ConditionTrue},
					})
				}),
				makeTestSecret(func(o *corev1.Secret) {
					o.ObjectMeta.Name = secretName + "-kubeconfig"
					o.Data = map[string][]byte{
						"value": secretData(clusterName),
					}
				}),
			},
			expectedCount: 0,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.context, func(t *testing.T) {
			scheme, err := kube.CreateScheme()
			g.Expect(err).NotTo(HaveOccurred())
			client := makeClient(t, g, tt.clusterObjects...)
			cluster := new(clusterfakes.FakeCluster)
			cluster.GetNameReturns("management")
			cluster.GetServerClientReturns(client, nil)

			fetcher := fetcher.NewGitopsClusterFetcher(testr.New(t), cluster, "default", scheme, false, kube.UserPrefixes{})

			clusters, err := fetcher.Fetch(context.TODO())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(clusters).To(HaveLen(tt.expectedCount))
		})
	}
}

func makeClient(t *testing.T, g *GomegaWithT, clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		capiv1.AddToScheme,
		gitopsv1alpha1.AddToScheme,
	}
	g.Expect(schemeBuilder.AddToScheme(scheme)).To(Succeed())

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(clusterState...).
		Build()
}

func makeTestCluster(opts ...func(*gitopsv1alpha1.GitopsCluster)) *gitopsv1alpha1.GitopsCluster {
	c := &gitopsv1alpha1.GitopsCluster{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gitops.weave.works/v1alpha1",
			Kind:       "GitopsCluster",
		},
		Spec: gitopsv1alpha1.GitopsClusterSpec{},
		Status: gitopsv1alpha1.GitopsClusterStatus{
			Conditions: []metav1.Condition{
				{Type: meta.ReadyCondition, Status: metav1.ConditionTrue},
				{Type: gitopsv1alpha1.ClusterConnectivity, Status: metav1.ConditionTrue},
			},
		},
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

func makeTestSecret(opts ...func(*corev1.Secret)) *corev1.Secret {
	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
		},
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

func secretData(name string) []byte {
	cfg := `
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM1ekNDQWMrZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJeU1EUXdOakV4TXpJek9Wb1hEVE15TURRd016RXhNekl6T1Zvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTnQ4Cm5qRFBLREdwNU1QaDFYZG1GVk1CaDRLQUlKbVpyd0tTc0JVSkdaNTIxR2gzODh6TXpHVGZjZy9VQ0RmZFhTTWYKY1h0aTYxMGZqWVNzaUFMS1kva2NCMFdCb3g0ZnZic0o4S0E3L1A1UExGcDZmMlFhNUQyZGo2ODA5a3FCTVNHawpZU25sOWQ1SlRsLzB5NUUwdDdPVGk4am9hOXBuWVFrWGlkNWNFNlhNQ3JDUVVEMkxheVhpTDZmamlFb0IxRmZkCkliRnZTMmNNZlk2dU5sL2cwUHkraVl6L1pPYm95YVhoa3VzL0xjS1VWMEJHVDJzb2lXU3E5a1pySXVCSWJXTkQKUElpd3kxb0Y5Rjh0WEx4OFZGUFpRUmswcGwvR3JLM3p5Qnhlb0NrT3ZtQ0p1TldVMzBxYlN5MkRFd3lDUkZXSwpaZ1pib1AyemRIMC9STVFtWmVjQ0F3RUFBYU5DTUVBd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZCbkR3V25LV3FHVU9SNlJ3bE9LcUoydTFGU0JNQTBHQ1NxR1NJYjMKRFFFQkN3VUFBNElCQVFDZ2ZJZU8yNkxnVmZTNTdCblFRdEo1WjM1VDd2cjZKdWd4Q1ZDbDE4eCtQQkpuSVpNRQpuMllyNUpNY2QxL0xmdCtjbTk0NVJQTDZtWTZzMk1SU1JiZGZtbTJZdVNSSG1FY21makNwT1BxcXllVXFyZmtVCnFWK3B6STNST04zaEVPSW5EdVpOZzdhQmE0cytQZnJORjd4b3pDVFBBUDdpMThWZTc1Z0NsQmM5cy9nRVR0MW0KVDE3SE9zaElXY2l3WktGaWpyVmN5SXpnOTdZbzc4S0NNTEVvbnVyb2NyRmlyM3pEZWMvL0E3MFRlSzlQZHVaSAo5SS9kS1RFQVVQK1NvYUpzdVpSZFpBVldDb0lXdWxDRDZqZjJOWnc5VVplcWRlQ2FpYzkxUTdBMUNRQkg3VTYrCklYK3UwblUzdXlBWVU2WGkvZFFTUWF0UkxxNDIxak5HdllTWAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    server: https://1.2.3.4:5678
  name: NAME
contexts:
- context:
    cluster: NAME
    user: NAME
  name: NAME
current-context: NAME
kind: Config
preferences: {}
users:
- name: NAME
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURFekNDQWZ1Z0F3SUJBZ0lJTTFQQkVlUnAyNEF3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB5TWpBME1EWXhNVE15TXpsYUZ3MHlNekEwTURZeE1UTXlOREZhTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXNNK2RPRU9Wb0tmYW1rMWQKcXJkcVJLbGFZZVZseEUwVHpFVEY1N3dONk9TdGlpNmlINmhSYVVMWmpmc2djMlRKaGZST3ZURlZlclh3SVU1ZwpEOENaNGc3RWt0U3pFck5JUndUV2J1MUh1QUpNY2cvUkhDZlZlYUNOekluVmRndHJWbjl0UXVMa0tTb0JQVGdXCktQTHpvMGNEbGRXOTYzQ2dCdTh4clZTbHdyaTRGNzNTTFlhbUFaY09hOHJrbEcrd2dWVk1Hc3ZkNU1XcEZFQ28KTFlTem1YYVZ4UmxscjAvazhuTUFObzkydllWSGpoK1M0NnJpUDhCdGI4RStXSzFzMEUwV3lURFhmaUd6WjZ5bAp3d0xBZ3BlM0FJZ3gvTWpKSUkrZU13cDV4a3dTdUxxTGJhN28yd2J1SEZlcTVaMm03dnFxSnpHMFEyd0dzOGNZClh6dmVGUUlEQVFBQm8wZ3dSakFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0h3WURWUjBqQkJnd0ZvQVVHY1BCYWNwYW9aUTVIcEhDVTRxb25hN1VWSUV3RFFZSktvWklodmNOQVFFTApCUUFEZ2dFQkFDUjV2VTJ0N3g1MVZuZHRPT2hEbTlOa3pBSGlZUjhBR0xRekZiQ1hMQzN6RjNBeGxjelBYRGRUCnprODRBTUVvRjJyeU9RVmhHY2JVSnpqQWpSQ2Z4NG1CdzloMGFKNzhtZnpySjM0cU9CWE5IWmdrYkJiaGdzMVoKbUNMaUd6ZlNBNktvUngwaWVkc0pyMnhtNmt1UHJLSnp2dngxUCtrTDRDeXdJbDNnL2h0cUhFS1NUQllCNzJodgptclFteUZLSDNvRU5FSjgyUVp3SHlOVFNMbWNhV1pLR1hnWEFTVFZWOTcrWjgvbFFrekxJcHptdkRPSm1LZGZDCkpjclhSQjZLb3MwMFdZZGVpMVJmYWlmejQ2aUV6QXJUNFo0Mm1Xd0FreWxuK0pqWUxhdDBwVnRoeVVScURuQ1oKUmlZdmhqUTc4RktXM2pEZ3BVc0cySVNzZ044WER0dz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJQkFBS0NBUUVBc00rZE9FT1ZvS2ZhbWsxZHFyZHFSS2xhWWVWbHhFMFR6RVRGNTd3TjZPU3RpaTZpCkg2aFJhVUxaamZzZ2MyVEpoZlJPdlRGVmVyWHdJVTVnRDhDWjRnN0VrdFN6RXJOSVJ3VFdidTFIdUFKTWNnL1IKSENmVmVhQ056SW5WZGd0clZuOXRRdUxrS1NvQlBUZ1dLUEx6bzBjRGxkVzk2M0NnQnU4eHJWU2x3cmk0RjczUwpMWWFtQVpjT2E4cmtsRyt3Z1ZWTUdzdmQ1TVdwRkVDb0xZU3ptWGFWeFJsbHIwL2s4bk1BTm85MnZZVkhqaCtTCjQ2cmlQOEJ0YjhFK1dLMXMwRTBXeVREWGZpR3paNnlsd3dMQWdwZTNBSWd4L01qSklJK2VNd3A1eGt3U3VMcUwKYmE3bzJ3YnVIRmVxNVoybTd2cXFKekcwUTJ3R3M4Y1lYenZlRlFJREFRQUJBb0lCQUNydWJtVmYrNi9qc2UrdgpnMlBWWDBkR3U3eHpmKzlYSzh4NGtuay9MejF5Y1RUUk4rcHA2MEtjeWNod3hxTmVRSlIzQ3J0amhEYmtnR2NGCjZjdEpYOVFFOC9RWEUxZ2lFaFcwZGdDL09wL1NadzkzQ2JaRmNjOHpqZHF4U1JSOWwxV01ZVkpSVjBjcmZOdUoKaDgvdmxmcjZYa04rZjd2d1A5c1BMMGUvK3ZPNWtNVStMV1ZYN2pTZ05XRisrdnpOSHVGem44ZlRka1pab2Z2bgpyWktWRHpZczByN2RnS0FtWXF5dmF2VE5qUC9OTW5DSFhvdjBqVGRkTGxmVktDZStpUXQ3UDFyUFgvK3Z6TnoxCnExWThmZ201TXFhT3lHWUVGVHpTSFIxdHZLb2pOaWljUHRORTlnRjNtZ0VJc0YvWVVSU1pHLzNQTWVnMVBheVMKaU9KakpTMENnWUVBdzlMVmZXNEFmNjE5QkNDVXZJZm9zQzlxV1JZbDZUSjZjK09BM1dPMXRRdnpqcWlFY0JEWAoyUmdwNXp4VlpBUk5JY2VoMHNHWXcvaGlJRk1MUzV4c3c3Rk5ZN3ZCSm0rOW9yYm9PcTlZNGFVNDc4dlc1U05IClpXZjV3bDhReXNBVEpMZVMyblpjUFo5QzdqWSsrcE9ueXBpQURGaW04ZkVNa3NFZS9mNnJXNWNDZ1lFQTV5VVcKWFBYZzI3Q1Urd2dTbklhWTdZVkc1c3p1OVkxWG8xVXRPa0VXU0FFaHg3K0RCN1Q1b2p5OVNYNWt0VWdqSmtkSwpaZWpJSFBZQjhwNndnc1NqbmpyRzlpY0ZEUUxnSDNxUFcyU2VWTmpJRUkxMHo5Nklkc1VvL3V3VHBicE92Vk9oCkNEZGUzL1lZMm53MnIzRmVQTWladUZvNkFTY28xbnlXV2Voc09UTUNnWUVBaFF2SEQweGd2RjYwRk16S0lYbTUKcDVMZmo1MlRybWdrZUg1Mi9IUVZiZWVyMkI0NHRTZE1iK3lSODlDek41d1FoOFhwOVphaFkyeHJ4d2lGSVI4cgphcDRaTll6SVE0UWg5TjZPMCtoMDNBSjB0Ny9ueHBEOG5qSlJxRFVNNUtReG5YMjRJZ1BPMGZOVjl5RVdFd3VsCk1lb0EvZUp4c3VvU245YmtacS9UM3dzQ2dZQTczTm9PMTBzVitvU0xBd3MyNkpFQXFzeXpCNDQzb0JSN1k1cmsKQkdsTjJxVXlBMEpmSTVxblRzM0RFKzNuR1RpcE9EdG5hME13WlBJYU1Na01CUHRQQm0veTNpWXJ1WHZzQ3lUSAppYWFMMk56dmxJTVZOcy9tMnFjRVpvV3NIVFU1U1VoaVJWelg2ZmVEMWptZmRGL3dwQTlUdEdKalhBM3locSsxCnQwRVlDd0tCZ0ZLMFVvcytTVk5HcUVQUDRJN0oycGt3WEVFaUhHdEMxcVRzVldiN3VwcW9iSTRMTjNkc2crZFMKWDdIcmNkSzhpMk9nejRsOW9VMERmTDExdkpXOFhVSGFXclR6RUVXQlpGb3NFdEVRVitDLzZWekJ4UktqQ3gxNwpnNk4xZkRzWlVUQjF5VHhvOVZBT1REVjA0Q01mWHErckhaclhZUHo3cVhrS0JnSUN6REFyCi0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==
    token: afg
`

	return []byte(strings.Replace(cfg, "NAME", name, -1))
}
