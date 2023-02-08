package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	ctrl "github.com/weaveworks/gitopssets-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/internal/grpctesting"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/gitopssets"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/gitopssets/adapter"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/gitopssets/server"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)


var k8sEnv *K8sTestEnv

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

	client, k8s := setup(t)

	obj := &ctrl.GitOpsSet{}
	obj.Name = "my-obj"
	obj.Namespace = "default"

	assert.NoError(t, k8s.Create(context.Background(), obj))

	res, err := client.ListGitOpsSets(ctx, &pb.ListGitOpsSetsRequest{})
	assert.NoError(t, err)

	assert.Len(t, res.Gitopssets, 1)

	o := res.Gitopssets[0]

	assert.Equal(t, o.ClusterName, "Default")
	assert.Equal(t, o.Name, obj.Name)
	assert.Equal(t, o.Namespace, obj.Namespace)

}

func TestSuspendGitOpsSet(t *testing.T) {
	ctx := context.Background()
	client, k8s := setup(t)

	obj := &ctrl.GitOpsSet{}
	obj.Name = "my-obj"
	obj.Namespace = "default"
	obj.Spec = ctrl.GitOpsSetSpec{
		Suspend: false,
	}

	assert.NoError(t, k8s.Create(ctx, obj))

	// no kind is registered for the type v1alpha1.GitOpsSet in scheme "pkg/runtime/scheme.go:100" - how do we create objects in the absence of k8s?
	_, err := client.ToggleSuspendGitOpsSet(ctx, &pb.ToggleSuspendGitOpsSetRequest{
		Name:        obj.Name,
		Namespace:   obj.Namespace,
		ClusterName: "Default",
		Suspend:     true,
	})
	assert.NoError(t, err)

	s := &ctrl.GitOpsSet{}
	key := types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}

	assert.NoError(t, k8s.Get(ctx, key, s))

	assert.True(t, s.Spec.Suspend, "expected Spec.Suspend to be true")
}

func TestSyncGitOpsSet(t *testing.T) {
	ctx := context.Background()
	client, k8s := setup(t)

	obj := &ctrl.GitOpsSet{}
	obj.Name = "my-obj"
	obj.Namespace = "default"

	key := types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}

	assert.NoError(t, k8s.Create(context.Background(), obj))

	done := make(chan error)
	defer close(done)

	go func() {
		_, err := client.SyncGitOpsSet(ctx, &pb.SyncGitOpsSetRequest{
			ClusterName: "Default",
			Name:        obj.Name,
			Namespace:   obj.Namespace,
		})
		done <- err
	}()

	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-ticker.C:

			r := adapter.GitOpsSetAdapter{GitOpsSet: obj}

			if err := simulateReconcile(ctx, k8s, key, r.AsClientObject()); err != nil {
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

func setup(t *testing.T) (pb.GitOpsSetsClient, client.Client) {
	k8s, factory := grpctesting.MakeFactoryWithObjects()
	opts := server.ServerOpts{
		ClientsFactory: factory,
	}
	srv := server.NewGitOpsSetsServer(opts)

	conn := grpctesting.Setup(t, func(s *grpc.Server) {
		pb.RegisterGitOpsSetsServer(s, srv)
	})

	return pb.NewGitOpsSetsClient(conn), k8s
}

func simulateReconcile(ctx context.Context, k client.Client, name types.NamespacedName, o client.Object) error {
	switch obj := o.(type) {
	case *ctrl.GitOpsSet:
		if err := k.Get(ctx, name, obj); err != nil {
			return err
		}

		// obj.Status.SetLastHandledReconcileRequest(time.Now().Format(time.RFC3339Nano))
		// return k.Status().Update(ctx, obj)
	}

	return errors.New("simulating reconcile: unsupported type")
}


func TestGetReconciledObjects(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	gsName := "my-gs"
	ns1 := newNamespace(ctx, k, g)
	ns2 := newNamespace(ctx, k, g)

	reconciledObjs := []client.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-deployment",
				Namespace: ns1.Name,
				Labels: map[string]string{
					server.GitOpsSetNameKey:      gsName,
					server.GitOpsSetNamespaceKey: ns1.Name,
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
				Name:      "my-configmap",
				Namespace: ns2.Name,
				Labels: map[string]string{
					server.GitOpsSetNameKey:     gsName,
					server.GitOpsSetNamespaceKey: ns1.Name,
				},
			},
		},
	}

	for _, obj := range reconciledObjs {
		g.Expect(k.Create(ctx, obj)).Should(Succeed())
	}

	crb := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns1.Name,
			Name:      "ns-admin",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.SchemeGroupVersion.Group,
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
		Subjects: []rbacv1.Subject{{
			APIGroup: rbacv1.SchemeGroupVersion.Group,
			Kind:     rbacv1.UserKind,
			Name:     "ns-admin",
		}},
	}
	g.Expect(k.Create(ctx, &crb)).Should((Succeed()))

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
			name:        "unknown user doesn't receive any objects",
			user:        "anne",
			expectedLen: 0,
		},
		{
			name:        "ns-admin sees only objects in their namespace",
			user:        "ns-admin",
			expectedLen: 1,
			expectedObjects: []objectAssertion{
				{
					kind: "Deployment",
					name: reconciledObjs[0].GetName(),
				},
			},
		},
		{
			name:        "master user receives all objects",
			user:        "anne",
			group:       "system:masters",
			expectedLen: 2,
			expectedObjects: []objectAssertion{
				{
					kind: "Deployment",
					name: reconciledObjs[0].GetName(),
				},
				{
					kind: "ConfigMap",
					name: reconciledObjs[1].GetName(),
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
				Name: gsName,
				Namespace:      ns1.Name,
				AutomationKind: kustomizev1.KustomizationKind,
				Kinds: []*pb.GroupVersionKind{
					{Group: appsv1.SchemeGroupVersion.Group, Version: appsv1.SchemeGroupVersion.Version, Kind: "Deployment"},
					{Group: corev1.SchemeGroupVersion.Group, Version: corev1.SchemeGroupVersion.Version, Kind: "ConfigMap"},
				},
				ClusterName: cluster.DefaultCluster,
			})

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

func TestGetChildObjects(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	automationName := "my-automation"

	ns := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-deployment",
			Namespace: ns.Name,
			UID:       "this-is-not-an-uid",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": automationName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": automationName},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "nginx",
						Image: "nginx",
					}},
				},
			},
		},
	}

	rs := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-123abcd", automationName),
			Namespace: ns.Name,
		},
		Spec: appsv1.ReplicaSetSpec{
			Template: deployment.Spec.Template,
			Selector: deployment.Spec.Selector,
		},
		Status: appsv1.ReplicaSetStatus{
			Replicas: 1,
		},
	}

	rs.SetOwnerReferences([]metav1.OwnerReference{{
		UID:        deployment.UID,
		APIVersion: appsv1.SchemeGroupVersion.String(),
		Kind:       "Deployment",
		Name:       deployment.Name,
	}})

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(&ns, deployment, rs).Build()
	cfg := makeServerConfig(client, t)
	c := makeServer(cfg, t)

	res, err := c.GetChildObjects(ctx, &pb.GetChildObjectsRequest{
		ParentUid: string(deployment.UID),
		Namespace: ns.Name,
		GroupVersionKind: &pb.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "ReplicaSet",
		},
		ClusterName: cluster.DefaultCluster,
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Objects).To(HaveLen(1))

	first := res.Objects[0]
	g.Expect(first.Payload).To(ContainSubstring("ReplicaSet"))
	g.Expect(first.Payload).To(ContainSubstring(rs.Name))
}
