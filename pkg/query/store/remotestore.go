package store

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/remotewriter"
)

type RemoteStoreOpts struct {
	Log               logr.Logger
	Url               string
	Token             string
	RemoteStoreWriter RemoteStoreWriter
}

type RemoteStore struct {
	log         logr.Logger
	storeWriter RemoteStoreWriter
}

func (r RemoteStore) StoreRoles(ctx context.Context, roles []models.Role) error {
	err := r.storeWriter.StoreRoles(ctx, roles)
	if err != nil {
		return fmt.Errorf("cannot remote store roles: %w", err)
	}
	return nil
}

func (r RemoteStore) StoreRoleBindings(ctx context.Context, roleBindings []models.RoleBinding) error {
	err := r.storeWriter.StoreRoleBindings(ctx, roleBindings)
	if err != nil {
		return fmt.Errorf("cannot remote store roles: %w", err)
	}
	return nil
}

func (r RemoteStore) GetAccessRules(ctx context.Context) ([]interface{}, error) {
	r.log.Info("in get access rules")
	return nil, fmt.Errorf("not implemented")
}

func (r RemoteStore) StoreObjects(ctx context.Context, objects []models.Object) error {
	err := r.storeWriter.StoreObjects(ctx, objects)
	if err != nil {
		return fmt.Errorf("cannot remote store objects: %w", err)
	}
	return nil
}

func (r RemoteStore) DeleteObjects(ctx context.Context, objects []models.Object) error {
	err := r.storeWriter.DeleteObjects(ctx, objects)
	if err != nil {
		return fmt.Errorf("cannot delete objects: %w", err)
	}
	return nil
}

func (r RemoteStore) GetObjects(ctx context.Context, q Query) ([]models.Object, error) {
	r.log.Info("in get objects")
	return nil, fmt.Errorf("not implemented")
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . RemoteStoreWriter
type RemoteStoreWriter interface {
	StoreWriter
	GetUrl() string
}

func NewRemoteStore(opts RemoteStoreOpts) (StoreWriter, error) {

	if err := validateOptions(opts); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	writer, err := NewRemoteStoreWriter(opts)
	if err != nil {
		return nil, fmt.Errorf("cannot create remote store writer: %w", err)
	}

	remoteStore := RemoteStore{
		log:         opts.Log,
		storeWriter: writer,
	}

	remoteStore.log.Info("remote store created", "url", opts.Url)

	return remoteStore, nil
}

func NewRemoteStoreWriter(opts RemoteStoreOpts) (RemoteStoreWriter, error) {
	if opts.RemoteStoreWriter != nil {
		return opts.RemoteStoreWriter, nil
	}

	writer, err := remotewriter.NewHttpRemoteStore(remotewriter.RemoteWriterOpts{
		Log:   opts.Log,
		Url:   opts.Url,
		Token: opts.Token,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create http remote writer: %w", err)
	}

	return writer, nil
}

func validateOptions(opts RemoteStoreOpts) error {
	//valid if already using a writer
	if opts.RemoteStoreWriter != nil {
		return nil
	}

	if opts.Url == "" {
		return fmt.Errorf("url cannot be empty")
	}

	if opts.Token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	return nil
}
