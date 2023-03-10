package fakes

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewClient(log logr.Logger) FakeClient {
	return FakeClient{log: log}
}

type FakeClient struct {
	log logr.Logger
}

func (f FakeClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	//TODO implement me
	panic("implement me")
}

func (f FakeClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	//TODO implement me
	panic("implement me")
}

func (f FakeClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	//TODO implement me
	panic("implement me")
}

func (f FakeClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	//TODO implement me
	panic("implement me")
}

func (f FakeClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	//TODO implement me
	panic("implement me")
}

func (f FakeClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	//TODO implement me
	panic("implement me")
}

func (f FakeClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	//TODO implement me
	panic("implement me")
}

func (f FakeClient) Status() client.SubResourceWriter {
	//TODO implement me
	panic("implement me")
}

func (f FakeClient) SubResource(subResource string) client.SubResourceClient {
	//TODO implement me
	panic("implement me")
}

func (f FakeClient) Scheme() *runtime.Scheme {
	//TODO implement me
	panic("implement me")
}

func (f FakeClient) RESTMapper() meta.RESTMapper {
	//TODO implement me
	panic("implement me")
}
