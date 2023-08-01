package server

import (
	"context"

	"encoding/base64"
	"encoding/json"
	"os"
	"testing"

	goage "filippo.io/age"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/stretchr/testify/assert"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"go.mozilla.org/sops/v3/aes"
	"go.mozilla.org/sops/v3/cmd/sops/common"
	"go.mozilla.org/sops/v3/cmd/sops/formats"
	"go.mozilla.org/sops/v3/keyservice"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestEncryptSecret(t *testing.T) {
	ageKey, err := goage.GenerateX25519Identity()
	assert.Nil(t, err)

	gpgPrivateKey, gpgPublicKey, err := generatePGPKeyPairs()
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
						Name:      "sops-age-public-key",
						Namespace: "flux-system",
					},
					Data: map[string][]byte{
						"age.agekey": []byte(ageKey.Recipient().String()),
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
						"gpg.asc": []byte(gpgPrivateKey),
					},
				},
				&v1.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: v1.SchemeGroupVersion.Identifier(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "sops-gpg-public-key",
						Namespace: "flux-system",
					},
					Data: map[string][]byte{
						"gpg.asc": []byte(gpgPublicKey),
					},
				},
				&kustomizev1.Kustomization{
					TypeMeta: metav1.TypeMeta{
						Kind:       kustomizev1.KustomizationKind,
						APIVersion: v1.SchemeGroupVersion.Identifier(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-sops-age-secrets",
						Namespace: "flux-system",
						Annotations: map[string]string{
							SopsPublicKeyNameAnnotation:      "sops-age-public-key",
							SopsPublicKeyNamespaceAnnotation: "flux-system",
						},
					},
					Spec: kustomizev1.KustomizationSpec{
						Path: "./secrets/age",
						Decryption: &kustomizev1.Decryption{
							Provider: "sops",
							SecretRef: &meta.LocalObjectReference{
								Name: "sops-age",
							},
						},
					},
				},
				&kustomizev1.Kustomization{
					TypeMeta: metav1.TypeMeta{
						Kind:       kustomizev1.KustomizationKind,
						APIVersion: v1.SchemeGroupVersion.Identifier(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-sops-gpg-secrets",
						Namespace: "flux-system",
						Annotations: map[string]string{
							SopsPublicKeyNameAnnotation:      "sops-gpg-public-key",
							SopsPublicKeyNamespaceAnnotation: "flux-system",
						},
					},
					Spec: kustomizev1.KustomizationSpec{
						Path: "./secrets/gpg",
						Decryption: &kustomizev1.Decryption{
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
		key     string
		method  string
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
			path:   "./secrets/age",
			key:    ageKey.String(),
			method: "age",
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
			path:   "./secrets/gpg",
			key:    gpgPrivateKey,
			method: "gpg",
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
			path:   "./secrets/age",
			key:    ageKey.String(),
			method: "age",
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
			path:   "./secrets/gpg",
			key:    gpgPrivateKey,
			method: "gpg",
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

		rawSecret, err := res.EncryptedSecret.MarshalJSON()
		if err != nil {
			t.Errorf(err.Error())
		}

		decryptedValues, err := decryptSecretValues(rawSecret, tt.method, tt.key)
		if err != nil {
			t.Errorf(err.Error())
		}

		if tt.request.Data != nil {
			assert.EqualValues(t, tt.request.Data, decryptedValues)
		}

		if tt.request.StringData != nil {
			assert.EqualValues(t, tt.request.StringData, decryptedValues)
		}

		assert.Equal(t, tt.path, res.Path)
	}
}

func generatePGPKeyPairs() (string, string, error) {
	k, err := crypto.GenerateKey("test", "test@test.com", "", 4096)
	if err != nil {
		panic(err)
	}

	privateKey, err := k.Armor()
	if err != nil {
		return "", "", err
	}

	publicKey, err := k.GetArmoredPublicKey()
	if err != nil {
		return "", "", err
	}

	return privateKey, publicKey, nil
}

func decryptSecretValues(raw []byte, method, key string) (map[string]string, error) {
	store := common.StoreForFormat(formats.Json)
	tree, err := store.LoadEncryptedFile(raw)
	if err != nil {
		return nil, err
	}

	switch method {
	case "age":
		// set sops private key env var so the local client retrieve it
		os.Setenv("SOPS_AGE_KEY", key)
	case "gpg":
		// import gpg private key so the local client retrieve it
		if err := importPGPKey(key); err != nil {
			return nil, err
		}
	}

	// get the master key using the local client
	metadataKey, err := tree.Metadata.GetDataKeyWithKeyServices([]keyservice.KeyServiceClient{
		keyservice.NewLocalClient(),
	})

	if err != nil {
		return nil, err
	}

	cipher := aes.NewCipher()
	_, err = tree.Decrypt(metadataKey, cipher)
	if err != nil {
		return nil, err
	}

	out, err := store.EmitPlainFile(tree.Branches)
	if err != nil {
		return nil, err
	}

	var secretAsMap map[string]interface{}
	if err := json.Unmarshal(out, &secretAsMap); err != nil {
		return nil, err
	}

	secrets := map[string]string{}
	if data, ok := secretAsMap["data"].(map[string]interface{}); ok {
		for k, v := range data {
			val, err := base64.StdEncoding.DecodeString(v.(string))
			if err != nil {
				return nil, err
			}
			secrets[k] = string(val)
		}
	} else if data, ok := secretAsMap["stringData"].(map[string]interface{}); ok {
		for k, v := range data {
			secrets[k] = v.(string)
		}
	}
	return secrets, nil
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
				&kustomizev1.Kustomization{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kustomization-a-1",
						Namespace: "namespace-a-1",
						Annotations: map[string]string{
							"sops-public-key/name":      "sops-pgp",
							"sops-public-key/namespace": "flux-system",
						},
					},
					Spec: kustomizev1.KustomizationSpec{
						Decryption: &kustomizev1.Decryption{
							Provider: "sops",
						},
					},
				},
				&kustomizev1.Kustomization{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kustomization-a-2",
						Namespace: "namespace-a-2",
						Annotations: map[string]string{
							"sops-public-key/name":      "sops-pgp",
							"sops-public-key/namespace": "flux-system",
						},
					},
					Spec: kustomizev1.KustomizationSpec{
						Decryption: &kustomizev1.Decryption{
							Provider: "sops",
						},
					},
				},
				&kustomizev1.Kustomization{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kustomization-a-3",
						Namespace: "namespace-a-1",
					},
					Spec: kustomizev1.KustomizationSpec{
						Decryption: &kustomizev1.Decryption{
							Provider: "decryption-provider-a-2",
						},
					},
				},
				&kustomizev1.Kustomization{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kustomization-a-4",
						Namespace: "namespace-a-2",
					},
				},
				&kustomizev1.Kustomization{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kustomization-a-5",
						Namespace: "namespace-a-1",
					},
					Spec: kustomizev1.KustomizationSpec{
						Decryption: &kustomizev1.Decryption{
							Provider: "sops",
						},
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
				&kustomizev1.Kustomization{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kustomization-b-1",
						Namespace: "namespace-b-1",
						Annotations: map[string]string{
							"sops-public-key/name":      "sops-pgp",
							"sops-public-key/namespace": "flux-system",
						},
					},
					Spec: kustomizev1.KustomizationSpec{
						Decryption: &kustomizev1.Decryption{
							Provider: "sops",
						},
					},
				},
				&kustomizev1.Kustomization{
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
