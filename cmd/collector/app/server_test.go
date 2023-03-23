package collector_app

import (
	"context"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/storefakes"
	"testing"
)

func TestNewServer(t *testing.T) {

	g := NewWithT(t)
	log := testr.New(t)

	ctx := context.Background()
	tests := []struct {
		name       string
		opts       ServerOpts
		errPattern string
	}{
		{
			name: "cannot create server without remote store",
			opts: ServerOpts{
				Logger:            log,
				ClustersNamespace: "flux-system",
				Store:             nil,
			},
			errPattern: "invalid remote store",
		},
		{
			name: "can create server with remote store",
			opts: ServerOpts{
				Logger:            log,
				ClustersNamespace: "flux-system",
				Store:             &storefakes.FakeStore{},
			},
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, stopF, err := NewServer(ctx, tt.opts)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(server).NotTo(BeNil())
			g.Expect(stopF).NotTo(BeNil())
		})
	}
}

func TestCollectorServer_start(t *testing.T) {

	g := NewWithT(t)
	logr := testr.New(t)

	ctx := context.Background()
	tests := []struct {
		name       string
		opts       ServerOpts
		errPattern string
	}{
		{
			name: "can start Server from environment",
			opts: ServerOpts{
				Logger:            logr,
				ClustersNamespace: "flux-system",
				Store:             &storefakes.FakeStore{},
			},
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, stopF, err := NewServer(ctx, tt.opts)
			g.Expect(err).To(BeNil())
			g.Expect(s).NotTo(BeNil())
			g.Expect(stopF).NotTo(BeNil())
			err = s.Start(ctx)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
		})
	}

}

func TestNewServer_createClusterManager(t *testing.T) {
	g := NewWithT(t)
	log := testr.New(t)
	tests := []struct {
		name       string
		errPattern string
	}{
		{
			name:       "can create cluster manager",
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterManager, err := createClusterManager("", log)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(clusterManager).NotTo(BeNil())
		})
	}

}
