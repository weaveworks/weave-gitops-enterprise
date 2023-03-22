package store

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/models"
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
				Address: "https://test.com",
			},
			errPattern: "token cannot be empty",
		},
		{
			name: "can create url with valid arguments",
			opts: RemoteStoreOpts{
				Log:     log,
				Address: "www.test.com",
				Token:   "myToken",
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
				Log:     log,
				Address: "http://localhost:8000",
				Token:   "abc",
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
				Address: "https://test.com",
			},
			errPattern: "token cannot be empty",
		},
		{
			name: "can create url with valid arguments",
			opts: RemoteStoreOpts{
				Address: "https://test.com",
				Token:   "myToken",
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

// TODO needs get token dynamically
func TestRemoteStore_StoreAccessRules(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)

	opts := RemoteStoreOpts{
		Log:     log,
		Address: "http://localhost:8000",
		Token:   "abc",
	}
	store, err := NewRemoteStore(opts)
	g.Expect(err).To(BeNil())
	g.Expect(store).NotTo(BeNil())

	accessRule := models.AccessRule{
		Cluster:         "test-cluster",
		Namespace:       "namespace",
		Principal:       "someuser",
		AccessibleKinds: []string{"example.com/v1beta2/SomeKind"},
	}
	tests := []struct {
		name       string
		rules      []models.AccessRule
		errPattern string
	}{
		{
			name: "can store access rules",
			rules: []models.AccessRule{
				accessRule,
			},
			errPattern: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.StoreAccessRules(context.Background(), tt.rules)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
		})
	}
}

// TODO needs get token dynamically
func TestRemoteStore_StoreObjects(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)

	opts := RemoteStoreOpts{
		Log:     log,
		Address: "http://localhost:8000",
		Token:   "abc",
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
