package terraform_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	tfctrl "github.com/weaveworks/tf-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/internal/grpctesting"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/terraform"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/terraform"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/terraform/internal/adapter"
	fc "github.com/weaveworks/weave-gitops-enterprise/pkg/terraform/internal/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListTerraformObjects(t *testing.T) {
	ctx := context.Background()

	client, k8s := setup(t)

	obj := &tfctrl.Terraform{}
	obj.Name = "my-obj"
	obj.Namespace = "default"

	assert.NoError(t, k8s.Create(context.Background(), obj))

	res, err := client.ListTerraformObjects(ctx, &pb.ListTerraformObjectsRequest{})
	assert.NoError(t, err)

	assert.Len(t, res.Objects, 1)

	o := res.Objects[0]

	assert.Equal(t, o.ClusterName, "Default")
	assert.Equal(t, o.Name, obj.Name)
	assert.Equal(t, o.Namespace, obj.Namespace)
}

func TestListTerraformObjects_NoTFCRD(t *testing.T) {
	ctx := context.Background()

	client, k8s, fc := setupWithFakes(t)

	fc.ClusteredListReturns(clustersmngr.ClusteredListError{Errors: []clustersmngr.ListError{{
		Cluster: "some-cluster",
		Err: &apimeta.NoKindMatchError{
			GroupKind:        schema.ParseGroupKind("CoolObject.somegroup"),
			SearchedVersions: []string{"v1"},
		},
	}}})

	obj := &tfctrl.Terraform{}
	obj.Name = "my-obj"
	obj.Namespace = "default"

	assert.NoError(t, k8s.Create(context.Background(), obj))

	res, err := client.ListTerraformObjects(ctx, &pb.ListTerraformObjectsRequest{})
	assert.NoError(t, err)

	assert.Len(t, res.Errors, 0, "should not have had errors")
}

func TestGetTerraformObject(t *testing.T) {
	ctx := context.Background()
	client, k8s := setup(t)

	obj := &tfctrl.Terraform{}
	obj.Name = "my-obj"
	obj.Namespace = "default"

	assert.NoError(t, k8s.Create(context.Background(), obj))

	res, err := client.GetTerraformObject(ctx, &pb.GetTerraformObjectRequest{
		ClusterName: "Default",
		Name:        obj.Name,
		Namespace:   obj.Namespace,
	})
	assert.NoError(t, err)

	expectedYaml :=
		`apiVersion: infra.contrib.fluxcd.io/v1alpha1
kind: Terraform
metadata:
  creationTimestamp: null
  name: my-obj
  namespace: default
  resourceVersion: "1"
spec:
  interval: 0s
  runnerPodTemplate:
    metadata: {}
    spec: {}
  sourceRef:
    kind: ""
    name: ""
status:
  lock: {}
  plan: {}
`

	assert.Equal(t, res.Object.ClusterName, "Default")
	assert.Equal(t, res.Yaml, expectedYaml)
}

func TestGetTerraformObjectPlan(t *testing.T) {
	ctx := context.Background()
	client, k8s := setup(t)

	tfObj := &tfctrl.Terraform{}
	tfObj.Name = "my-obj"
	tfObj.Namespace = "default"

	tfObj.Spec.StoreReadablePlan = "human"

	assert.NoError(t, k8s.Create(context.Background(), tfObj))

	planObj := &corev1.ConfigMap{}
	planObj.Name = "tfplan-default-my-obj"
	planObj.Namespace = "default"
	planObj.Data = map[string]string{
		"tfplan": "terraform plan",
	}

	assert.NoError(t, k8s.Create(context.Background(), planObj))

	res, err := client.GetTerraformObjectPlan(ctx, &pb.GetTerraformObjectPlanRequest{
		ClusterName: "Default",
		Name:        tfObj.Name,
		Namespace:   tfObj.Namespace,
	})

	assert.NoError(t, err)

	expectedPlan := "terraform plan"

	assert.Equal(t, res.Plan, expectedPlan)
}

func TestSyncTerraformObject(t *testing.T) {
	ctx := context.Background()
	client, k8s := setup(t)

	obj := &tfctrl.Terraform{}
	obj.Name = "my-obj"
	obj.Namespace = "default"

	key := types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}

	assert.NoError(t, k8s.Create(context.Background(), obj))

	done := make(chan error)
	defer close(done)

	go func() {
		_, err := client.SyncTerraformObjects(ctx, &pb.SyncTerraformObjectsRequest{
			Objects: []*pb.ObjectRef{{ClusterName: "Default",
				Name:      obj.Name,
				Namespace: obj.Namespace,
				Kind:      "Terraform"}},
		})
		done <- err
	}()

	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-ticker.C:

			r := adapter.TerraformObjectAdapter{Terraform: obj}

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

func TestSuspendTerraformObject(t *testing.T) {
	ctx := context.Background()
	client, k8s := setup(t)

	obj := &tfctrl.Terraform{}
	obj.Name = "my-obj"
	obj.Namespace = "default"
	obj.Spec = tfctrl.TerraformSpec{
		Suspend: false,
	}

	assert.NoError(t, k8s.Create(ctx, obj))

	_, err := client.ToggleSuspendTerraformObjects(ctx, &pb.ToggleSuspendTerraformObjectsRequest{Objects: []*pb.ObjectRef{{
		Name:        obj.Name,
		Namespace:   obj.Namespace,
		ClusterName: "Default",
	}},
		Suspend: true,
	})
	assert.NoError(t, err)

	s := &tfctrl.Terraform{}
	key := types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}

	assert.NoError(t, k8s.Get(ctx, key, s))

	assert.True(t, s.Spec.Suspend, "expected Spec.Suspend to be true")

}

func TestReplanTerraformObject(t *testing.T) {
	ctx := context.Background()
	client, k8s := setup(t)

	obj := &tfctrl.Terraform{}
	obj.Name = "my-obj"
	obj.Namespace = "default"

	key := types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}

	assert.NoError(t, k8s.Create(context.Background(), obj))

	done := make(chan error)
	defer close(done)

	go func() {
		_, err := client.ReplanTerraformObject(ctx, &pb.ReplanTerraformObjectRequest{
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

			r := adapter.TerraformObjectAdapter{Terraform: obj}

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

func setup(t *testing.T) (pb.TerraformClient, client.Client) {
	k8s, factory := grpctesting.MakeFactoryWithObjects()
	opts := terraform.ServerOpts{
		ClientsFactory: factory,
	}
	srv := terraform.NewTerraformServer(opts)

	conn := grpctesting.Setup(t, func(s *grpc.Server) {
		pb.RegisterTerraformServer(s, srv)
	})

	return pb.NewTerraformClient(conn), k8s
}

// Use this function when you want to override the behavior of clustersmngr.Client.
// You must provide a stub or return for the FakeClient to see objects.
func setupWithFakes(t *testing.T) (pb.TerraformClient, client.Client, *fc.FakeClient) {
	k8s, factory := grpctesting.MakeFactoryWithObjects()
	fc := &fc.FakeClient{}

	pool := &clustersmngrfakes.FakeClientsPool{}
	pool.ClientsReturns(map[string]client.Client{"Default": k8s})
	fc.ClientsPoolReturns(pool)

	factory.GetServerClientReturns(fc, nil)
	factory.GetImpersonatedClientReturns(fc, nil)

	opts := terraform.ServerOpts{
		Logger:         logr.Discard(),
		ClientsFactory: factory,
	}
	srv := terraform.NewTerraformServer(opts)

	conn := grpctesting.Setup(t, func(s *grpc.Server) {
		pb.RegisterTerraformServer(s, srv)
	})

	return pb.NewTerraformClient(conn), k8s, fc
}

func simulateReconcile(ctx context.Context, k client.Client, name types.NamespacedName, o client.Object) error {
	switch obj := o.(type) {
	case *tfctrl.Terraform:
		if err := k.Get(ctx, name, obj); err != nil {
			return err
		}

		obj.Status.SetLastHandledReconcileRequest(time.Now().Format(time.RFC3339Nano))
		return k.Status().Update(ctx, obj)
	}

	return errors.New("simulating reconcile: unsupported type")
}
