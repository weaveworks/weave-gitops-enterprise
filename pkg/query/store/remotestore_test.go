package store

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/models"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func TestNewRemoteStore(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	tests := []struct {
		name       string
		opts       RemoteStoreOpts
		errPattern string
	}{
		{
			name:       "cannot create store without valid arguments",
			opts:       RemoteStoreOpts{},
			errPattern: "url cannot be empty",
		},
		{
			name: "cannot create store without valid arguments",
			opts: RemoteStoreOpts{
				Url: "https://test.com",
			},
			errPattern: "token cannot be empty",
		},
		{
			name: "can create url with valid arguments",
			opts: RemoteStoreOpts{
				Log:   log,
				Url:   "www.test.com",
				Token: "myToken",
			},
			errPattern: "",
		},
		{
			name: "can create remote store with writer",
			opts: RemoteStoreOpts{
				Log:               log,
				RemoteStoreWriter: NewFakeRemoteStoreWriter(log),
			},
			errPattern: "",
		},
		{
			name: "can create remote store with http writer",
			opts: RemoteStoreOpts{
				Log:   log,
				Url:   "http://localhost:8000",
				Token: "abc",
			},
			errPattern: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewRemoteStore(tt.opts)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(store).NotTo(BeNil())
		})
	}
}

func NewFakeRemoteStoreWriter(log logr.Logger) RemoteStoreWriter {
	return FakeRemoteStoreWriter{log: log}
}

// TODO move me to a better place
type FakeRemoteStoreWriter struct {
	log logr.Logger
}

func (f FakeRemoteStoreWriter) DeleteRoles(ctx context.Context, roles []models.Role) error {
	f.log.Info("faked DeleteRoles roles")
	return nil
}

func (f FakeRemoteStoreWriter) DeleteRoleBindings(ctx context.Context, roleBindings []models.RoleBinding) error {
	f.log.Info("faked DeleteRoleBindings roles")
	return nil
}

func (f FakeRemoteStoreWriter) StoreRoles(ctx context.Context, roles []models.Role) error {
	f.log.Info("faked store roles")
	return nil
}

func (f FakeRemoteStoreWriter) StoreRoleBindings(ctx context.Context, roleBindings []models.RoleBinding) error {
	f.log.Info("faked store role bindings")
	return nil
}

func (f FakeRemoteStoreWriter) StoreAccessRules(ctx context.Context, roles []models.AccessRule) error {
	f.log.Info("faked store access rules")
	return nil
}

func (f FakeRemoteStoreWriter) StoreObjects(ctx context.Context, objects []models.Object) error {
	f.log.Info("faked store objects")
	return nil
}

func (f FakeRemoteStoreWriter) DeleteObjects(ctx context.Context, object []models.Object) error {
	f.log.Info("faked delete objects")
	return nil
}

func (f FakeRemoteStoreWriter) GetUrl() string {
	f.log.Info("faked get url")
	return "https://magicworld.com"
}

func TestNewHttpRemoteStoreWriter(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	tests := []struct {
		name       string
		opts       RemoteStoreOpts
		errPattern string
	}{
		{
			name:       "cannot create store without valid arguments",
			opts:       RemoteStoreOpts{},
			errPattern: "url cannot be empty",
		},
		{
			name: "cannot create store without valid arguments",
			opts: RemoteStoreOpts{
				Url: "https://test.com",
			},
			errPattern: "token cannot be empty",
		},
		{
			name: "can create url with valid arguments",
			opts: RemoteStoreOpts{
				Log:   log,
				Url:   "https://test.com",
				Token: "myToken",
			},
			errPattern: "",
		},
		{
			name: "can create remote store with writer",
			opts: RemoteStoreOpts{
				Log:               log,
				RemoteStoreWriter: NewFakeRemoteStoreWriter(log),
			},
			errPattern: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewRemoteStore(tt.opts)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(store).NotTo(BeNil())
		})
	}
}

func TestRemoteStore_StoreRoles(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)

	opts := RemoteStoreOpts{
		Log:               log,
		RemoteStoreWriter: NewFakeRemoteStoreWriter(log),
	}
	store, err := NewRemoteStore(opts)
	g.Expect(err).To(BeNil())
	g.Expect(store).NotTo(BeNil())

	role := models.Role{
		Cluster:   "test-cluster",
		Namespace: "namespace",
		Name:      "someName",
		Kind:      "Role",
		PolicyRules: []models.PolicyRule{
			{
				APIGroups: strings.Join([]string{"example.com"}, ","),
				Resources: strings.Join([]string{"SomeKind"}, ","),
				Verbs:     strings.Join([]string{"get", "list"}, ","),
			},
		},
	}
	tests := []struct {
		name       string
		roles      []models.Role
		errPattern string
	}{
		{
			name: "can store roles",
			roles: []models.Role{
				role,
			},
			errPattern: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.StoreRoles(context.Background(), tt.roles)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
		})
	}
}

func TestRemoteStore_StoreRoleBindings(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)

	opts := RemoteStoreOpts{
		Log:               log,
		RemoteStoreWriter: NewFakeRemoteStoreWriter(log),
	}
	store, err := NewRemoteStore(opts)
	g.Expect(err).To(BeNil())
	g.Expect(store).NotTo(BeNil())

	role := models.Role{
		Cluster:   "test-cluster",
		Namespace: "namespace",
		Name:      "someName",
		Kind:      "Role",
		PolicyRules: []models.PolicyRule{
			{
				APIGroups: strings.Join([]string{"example.com"}, ","),
				Resources: strings.Join([]string{"SomeKind"}, ","),
				Verbs:     strings.Join([]string{"get", "list"}, ","),
			},
		},
	}

	rb := models.RoleBinding{
		Cluster:   "test-cluster",
		Namespace: "namespace",
		Name:      "someName",
		Kind:      "RoleBinding",
		Subjects: []models.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     role.Kind,
				Name:     role.Name,
			},
		},
		RoleRefName: role.Name,
		RoleRefKind: role.Kind,
	}

	tests := []struct {
		name         string
		roleBindings []models.RoleBinding
		errPattern   string
	}{
		{
			name: "can store role bindings",
			roleBindings: []models.RoleBinding{
				rb,
			},
			errPattern: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.StoreRoleBindings(context.Background(), tt.roleBindings)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
		})
	}
}

func TestRemoteStore_StoreObjects(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)

	opts := RemoteStoreOpts{
		Log:               log,
		RemoteStoreWriter: NewFakeRemoteStoreWriter(log),
	}
	store, err := NewRemoteStore(opts)
	g.Expect(err).To(BeNil())
	g.Expect(store).NotTo(BeNil())

	object := models.Object{
		Cluster:   "test-cluster",
		Name:      "someName",
		Namespace: "namespace",
		Kind:      "ValidKind",
	}

	tests := []struct {
		name       string
		objects    []models.Object
		errPattern string
	}{
		{
			name: "can store objects",
			objects: []models.Object{
				object,
			},
			errPattern: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.StoreObjects(context.Background(), tt.objects)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
		})
	}
}
