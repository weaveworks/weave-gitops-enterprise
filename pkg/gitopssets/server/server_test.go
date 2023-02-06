package server_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	ctrl "github.com/weaveworks/gitopssets-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/internal/grpctesting"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/gitopssets"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/gitopssets/adapter"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/gitopssets/server"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

		obj.Status.SetLastHandledReconcileRequest(time.Now().Format(time.RFC3339Nano))
		return k.Status().Update(ctx, obj)
	}

	return errors.New("simulating reconcile: unsupported type")
}
