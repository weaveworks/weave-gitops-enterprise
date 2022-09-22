package terraform_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	tfctrl "github.com/weaveworks/tf-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/internal/grpctesting"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/terraform"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/terraform"
	"google.golang.org/grpc"
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
