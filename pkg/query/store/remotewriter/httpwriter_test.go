//go:build integration
// +build integration

package remotewriter

import (
	"context"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/models"
	"os"
	"strings"
	"testing"
)

func TestRemoteStore_StoreObjects(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)

	writerUrl := os.Getenv("QUERY_SERVER_URL")
	g.Expect(writerUrl == "").To(BeFalse())
	writerToken := os.Getenv("QUERY_SERVER_TOKEN")
	g.Expect(writerToken == "").To(BeFalse())

	opts := RemoteWriterOpts{
		Log:   log,
		Url:   writerUrl,
		Token: writerToken,
	}
	store, err := NewHttpRemoteStore(opts)
	g.Expect(err).To(BeNil())
	g.Expect(store).NotTo(BeNil())

	object := models.Object{
		Cluster:    "cluster-a",
		Namespace:  "ns-a",
		APIGroup:   "helm.toolkit.fluxcd.io",
		APIVersion: "v2beta1",
		Kind:       "HelmRelease",
		Name:       "podinfo",
		Status:     "status",
		Message:    "message",
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

func TestRemoteStore_DeleteObjects(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)

	writerUrl := os.Getenv("QUERY_SERVER_URL")
	g.Expect(writerUrl == "").To(BeFalse())
	writerToken := os.Getenv("QUERY_SERVER_TOKEN")
	g.Expect(writerToken == "").To(BeFalse())

	opts := RemoteWriterOpts{
		Log:   log,
		Url:   writerUrl,
		Token: writerToken,
	}
	store, err := NewHttpRemoteStore(opts)
	g.Expect(err).To(BeNil())
	g.Expect(store).NotTo(BeNil())

	object := models.Object{
		Cluster:    "cluster-a",
		Namespace:  "ns-a",
		APIGroup:   "helm.toolkit.fluxcd.io",
		APIVersion: "v2beta1",
		Kind:       "HelmRelease",
		Name:       "podinfo",
		Status:     "status",
		Message:    "message",
	}

	tests := []struct {
		name       string
		objects    []models.Object
		errPattern string
	}{
		{
			name: "can delete objects",
			objects: []models.Object{
				object,
			},
			errPattern: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.DeleteObjects(context.Background(), tt.objects)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
		})
	}
}

func TestHttpRemoteStore_StoreRoles(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)

	writerUrl := os.Getenv("QUERY_SERVER_URL")
	g.Expect(writerUrl == "").To(BeFalse())
	writerToken := os.Getenv("QUERY_SERVER_TOKEN")
	g.Expect(writerToken == "").To(BeFalse())

	opts := RemoteWriterOpts{
		Log:   log,
		Url:   writerUrl,
		Token: writerToken,
	}
	store, err := NewHttpRemoteStore(opts)
	g.Expect(err).To(BeNil())
	g.Expect(store).NotTo(BeNil())

	roles := []models.Role{{
		Name:      "role-a",
		Cluster:   "cluster-a",
		Namespace: "ns-a",
		Kind:      "Role",
		PolicyRules: []models.PolicyRule{{
			APIGroups: strings.Join([]string{"example.com/v1"}, ","),
			Resources: strings.Join([]string{"somekind"}, ","),
			Verbs:     strings.Join([]string{"get", "list", "watch"}, ","),
		}},
	}}

	tests := []struct {
		name       string
		roles      []models.Role
		errPattern string
	}{
		{
			name:       "can store roles",
			roles:      roles,
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

func TestHttpRemoteStore_StoreRoleBindings(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)

	writerUrl := os.Getenv("QUERY_SERVER_URL")
	g.Expect(writerUrl == "").To(BeFalse())
	writerToken := os.Getenv("QUERY_SERVER_TOKEN")
	g.Expect(writerToken == "").To(BeFalse())

	opts := RemoteWriterOpts{
		Log:   log,
		Url:   writerUrl,
		Token: writerToken,
	}
	store, err := NewHttpRemoteStore(opts)
	g.Expect(err).To(BeNil())
	g.Expect(store).NotTo(BeNil())

	bindings := []models.RoleBinding{{
		Cluster:   "cluster-a",
		Name:      "binding-a",
		Namespace: "",
		Kind:      "ClusterRoleBinding",
		Subjects: []models.Subject{{
			Kind: "User",
			Name: "some-user",
		}},
		RoleRefName: "role-a",
		RoleRefKind: "ClusterRole",
	}}
	tests := []struct {
		name         string
		roleBindings []models.RoleBinding
		errPattern   string
	}{
		{
			name:         "can store binidings",
			roleBindings: bindings,
			errPattern:   "",
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
