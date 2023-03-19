package server

import (
	"context"

	"strings"
	"testing"

	goage "filippo.io/age"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	kustomizev1beta2 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/stretchr/testify/assert"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestEncryptSecret(t *testing.T) {
	ageKey, err := goage.GenerateX25519Identity()
	assert.Nil(t, err)

	pgpKey, err := generatePGPKey()
	assert.Nil(t, err)

	clusters := []struct {
		name  string
		state []runtime.Object
	}{
		{
			name: "management",
			state: []runtime.Object{
				&v1.Namespace{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: v1.SchemeGroupVersion.Identifier(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "default",
					},
				},
				&v1.Namespace{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: v1.SchemeGroupVersion.Identifier(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "flux-system",
					},
				},
				&v1.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: v1.SchemeGroupVersion.Identifier(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "sops-age",
						Namespace: "flux-system",
					},
					Data: map[string][]byte{
						"age.agekey": []byte(ageKey.String()),
					},
				},
				&v1.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: v1.SchemeGroupVersion.Identifier(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "sops-gpg",
						Namespace: "flux-system",
					},
					Data: map[string][]byte{
						"gpg.asc": []byte(pgpKey),
					},
				},
				&kustomizev1beta2.Kustomization{
					TypeMeta: metav1.TypeMeta{
						Kind:       kustomizev1beta2.KustomizationKind,
						APIVersion: v1.SchemeGroupVersion.Identifier(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-sops-age-secrets",
						Namespace: "flux-system",
					},
					Spec: kustomizev1beta2.KustomizationSpec{
						Path: "./secrets/age",
						Decryption: &kustomizev1beta2.Decryption{
							Provider: "sops",
							SecretRef: &meta.LocalObjectReference{
								Name: "sops-age",
							},
						},
					},
				},
				&kustomizev1beta2.Kustomization{
					TypeMeta: metav1.TypeMeta{
						Kind:       kustomizev1beta2.KustomizationKind,
						APIVersion: v1.SchemeGroupVersion.Identifier(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-sops-gpg-secrets",
						Namespace: "flux-system",
					},
					Spec: kustomizev1beta2.KustomizationSpec{
						Path: "./secrets/gpg",
						Decryption: &kustomizev1beta2.Decryption{
							Provider: "sops",
							SecretRef: &meta.LocalObjectReference{
								Name: "sops-gpg",
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		request *capiv1_proto.EncryptSopsSecretRequest
		path    string
		err     error
	}{
		{
			request: &capiv1_proto.EncryptSopsSecretRequest{
				ClusterName:            "management",
				Name:                   "my-secret",
				Namespace:              "default",
				KustomizationName:      "my-sops-age-secrets",
				KustomizationNamespace: "flux-system",
				Data: map[string]string{
					"username": "admin",
					"password": "password",
				},
			},
			path: "./secrets/age",
		},
		{
			request: &capiv1_proto.EncryptSopsSecretRequest{
				ClusterName:            "management",
				Name:                   "my-secret",
				Namespace:              "default",
				KustomizationName:      "my-sops-gpg-secrets",
				KustomizationNamespace: "flux-system",
				Data: map[string]string{
					"username": "admin",
					"password": "password",
				},
			},
			path: "./secrets/gpg",
		},
		{
			request: &capiv1_proto.EncryptSopsSecretRequest{
				ClusterName:            "management",
				Name:                   "my-secret",
				Namespace:              "default",
				KustomizationName:      "my-sops-age-secrets",
				KustomizationNamespace: "flux-system",
				StringData: map[string]string{
					"username": "admin",
					"password": "password",
				},
			},
			path: "./secrets/age",
		},
		{
			request: &capiv1_proto.EncryptSopsSecretRequest{
				ClusterName:            "management",
				Name:                   "my-secret",
				Namespace:              "default",
				KustomizationName:      "my-sops-gpg-secrets",
				KustomizationNamespace: "flux-system",
				StringData: map[string]string{
					"username": "admin",
					"password": "password",
				},
			},
			path: "./secrets/gpg",
		},
	}

	clustersClients := map[string]client.Client{}
	for _, cluster := range clusters {
		clustersClients[cluster.name] = createClient(t, cluster.state...)
	}

	namespaces := map[string][]v1.Namespace{
		"management": {
			v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
			v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "flux-system"}},
		},
	}

	ctx := context.Background()
	s := getServer(t, clustersClients, namespaces)

	for _, tt := range tests {

		res, err := s.EncryptSopsSecret(ctx, tt.request)
		if err != nil {
			t.Errorf(err.Error())
		}
		if tt.err != nil {
			continue
		}

		encryptedSecret := res.EncryptedSecret.GetStructValue().AsMap()
		if data, ok := encryptedSecret["data"].(map[string]interface{}); ok {
			for _, value := range data {
				if !strings.HasPrefix(value.(string), "ENC[") {
					t.Error("secret is not encrypted")
				}
			}
		}
		if data, ok := encryptedSecret["stringData"].(map[string]interface{}); ok {
			for _, value := range data {
				if !strings.HasPrefix(value.(string), "ENC[") {
					t.Error("secret is not encrypted")
				}
			}
		}

		assert.Equal(t, tt.path, res.Path)
	}
}

func generatePGPKey() (string, error) {
	k, err := crypto.GenerateKey("test", "test@test.com", "", 4096)
	if err != nil {
		panic(err)
	}
	return k.Armor()
}

func TestListKustomizations(t *testing.T) {

	//create clusters manager struct and add a cluster with kustomizations
	clusters := []struct {
		name  string
		state []runtime.Object
	}{
		{
			name: "management",
			state: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-a-1",
					},
				},
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-a-2",
					},
				},
				&kustomizev1beta2.Kustomization{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kustomization-a-1",
						Namespace: "namespace-a-1",
					},
					Spec: kustomizev1beta2.KustomizationSpec{
						Decryption: &kustomizev1beta2.Decryption{
							Provider: "sops",
						},
					},
				},
				&kustomizev1beta2.Kustomization{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kustomization-a-2",
						Namespace: "namespace-a-2",
					},
					Spec: kustomizev1beta2.KustomizationSpec{
						Decryption: &kustomizev1beta2.Decryption{
							Provider: "sops",
						},
					},
				},
				&kustomizev1beta2.Kustomization{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kustomization-a-3",
						Namespace: "namespace-a-1",
					},
					Spec: kustomizev1beta2.KustomizationSpec{
						Decryption: &kustomizev1beta2.Decryption{
							Provider: "decryption-provider-a-2",
						},
					},
				},
				&kustomizev1beta2.Kustomization{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kustomization-a-4",
						Namespace: "namespace-a-2",
					},
				},
			},
		},
		{
			name: "leaf-1",
			state: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-a-1",
					},
				},
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-a-2",
					},
				},
				&kustomizev1beta2.Kustomization{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kustomization-b-1",
						Namespace: "namespace-b-1",
					},
					Spec: kustomizev1beta2.KustomizationSpec{
						Decryption: &kustomizev1beta2.Decryption{
							Provider: "sops",
						},
					},
				},
				&kustomizev1beta2.Kustomization{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kustomization-b-2",
						Namespace: "namespace-a-2",
					},
				},
			},
		},
		{
			name: "leaf-2",
			state: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-a-1",
					},
				},
			},
		},
	}

	tests := []struct {
		request  *capiv1_proto.ListSopsKustomizationsRequest
		response *capiv1_proto.ListSopsKustomizationsResponse
		err      bool
	}{
		{
			request: &capiv1_proto.ListSopsKustomizationsRequest{
				ClusterName: "management",
			},
			response: &capiv1_proto.ListSopsKustomizationsResponse{
				Kustomizations: []*capiv1_proto.SopsKustomizations{
					{
						Name:      "kustomization-a-1",
						Namespace: "namespace-a-1",
					},
					{
						Name:      "kustomization-a-2",
						Namespace: "namespace-a-2",
					},
				},
				Total: 2,
			},
		},
		{
			request: &capiv1_proto.ListSopsKustomizationsRequest{
				ClusterName: "leaf-1",
			},
			response: &capiv1_proto.ListSopsKustomizationsResponse{
				Kustomizations: []*capiv1_proto.SopsKustomizations{
					{
						Name:      "kustomization-b-1",
						Namespace: "namespace-b-1",
					},
				},
				Total: 1,
			},
		},
		{
			request: &capiv1_proto.ListSopsKustomizationsRequest{
				ClusterName: "leaf-2",
			},
			response: &capiv1_proto.ListSopsKustomizationsResponse{
				Kustomizations: nil,
				Total:          0,
			},
		},
		{
			request: &capiv1_proto.ListSopsKustomizationsRequest{
				ClusterName: "leaf-3",
			},
			err: true,
		},
	}

	clustersClients := map[string]client.Client{}
	for _, cluster := range clusters {
		clustersClients[cluster.name] = createClient(t, cluster.state...)
	}

	s := getServer(t, clustersClients, nil)

	for _, tt := range tests {
		res, err := s.ListSopsKustomizations(context.Background(), tt.request)
		if tt.err {
			assert.NotNil(t, err)
			continue
		}
		assert.Nil(t, err)

		assert.Equal(t, tt.response.Total, res.Total, "total number of kustomizations not equal")
		for i, kustomization := range tt.response.Kustomizations {
			assert.Equal(t, kustomization.Name, res.Kustomizations[i].Name, "kustomization name not equal")
			assert.Equal(t, kustomization.Namespace, res.Kustomizations[i].Namespace, "kustomization namespace not equal")
		}

	}
}
