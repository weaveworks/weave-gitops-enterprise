package server

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"sigs.k8s.io/controller-runtime/pkg/client"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// creating test for list external secrets
func TestListExternalSecrets(t *testing.T) {
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
				&esv1beta1.ExternalSecret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "external-secret-a-1",
						Namespace: "namespace-a-1",
					},
					Spec: esv1beta1.ExternalSecretSpec{
						SecretStoreRef: esv1beta1.SecretStoreRef{
							Name: "aws-secret-store",
						},
						Target: esv1beta1.ExternalSecretTarget{
							Name: "secret-a-1",
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
						Name: "namespace-x-1",
					},
				},
				&esv1beta1.ExternalSecret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "external-secret-x-1",
						Namespace: "namespace-x-1",
					},
					Spec: esv1beta1.ExternalSecretSpec{
						SecretStoreRef: esv1beta1.SecretStoreRef{
							Name: "aws-secret-store",
						},
						Target: esv1beta1.ExternalSecretTarget{
							Name: "secret-x-1",
						},
					},
				},
			},
		},
	}
	//creating tests for management cluster & leaf cluster
	//for both cluster external secrets & namespace external secrets
	tests := []struct {
		request  *capiv1_proto.ListExternalSecretsRequest
		response *capiv1_proto.ListExternalSecretsResponse
	}{
		{
			request: &capiv1_proto.ListExternalSecretsRequest{},
			response: &capiv1_proto.ListExternalSecretsResponse{
				Secrets: []*capiv1_proto.ExternalSecretItem{
					{
						SecretName:  "secret-a-1",
						Name:        "external-secret-a-1",
						Namespace:   "namespace-a-1",
						ClusterName: "management",
						SecretStore: "aws-secret-store",
					},
					{
						SecretName:  "secret-x-1",
						Name:        "external-secret-x-1",
						Namespace:   "namespace-x-1",
						ClusterName: "leaf-1",
						SecretStore: "aws-secret-store",
					},
				},
			},
		},
	}

	clustersClients := map[string]client.Client{}
	for _, cluster := range clusters {
		clustersClients[cluster.name] = createClient(t, cluster.state...)
	}

	namespaces := map[string][]v1.Namespace{
		"management": {v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "namespace-a-1"}}},
		"leaf-1":     {v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "namespace-x-1"}}},
	}
	ctx := context.Background()
	s := getServer(t, clustersClients, namespaces)

	for _, tt := range tests {
		res, err := s.ListExternalSecrets(ctx, tt.request)
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, len(tt.response.Secrets), len(res.Secrets), "External secrets count is not correct")

		expectedMap := map[string]*capiv1_proto.ExternalSecretItem{}
		for i := range tt.response.Secrets {
			expectedMap[tt.response.Secrets[i].Name] = tt.response.Secrets[i]
		}
		for i := range res.Secrets {
			actual := res.Secrets[i]
			expected, ok := expectedMap[actual.Name]
			if !ok {
				t.Errorf("found unexpected external secret %s", actual.Name)
			}
			assert.Equal(t, expected.SecretName, actual.SecretName, "secret name is not correct")
			assert.Equal(t, expected.ClusterName, actual.ClusterName, "cluster name is not correct")
			assert.Equal(t, expected.Namespace, actual.Namespace, "namespace are not correct")
			assert.Equal(t, expected.SecretStore, actual.SecretStore, "secret store is not correct")
		}
	}
}

func TestGetExternalSecret(t *testing.T) {
	clusters := []struct {
		name  string
		state []runtime.Object
	}{
		{
			name: "management",
			state: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-a",
					},
				},
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-b",
					},
				},
				&esv1beta1.SecretStore{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "aws-secret-store",
						Namespace: "namespace-a",
					},
					Spec: esv1beta1.SecretStoreSpec{
						Provider: &esv1beta1.SecretStoreProvider{
							AWS: &esv1beta1.AWSProvider{
								Region: "eu-north-1",
							},
						},
					},
				},
				&esv1beta1.ClusterSecretStore{
					ObjectMeta: metav1.ObjectMeta{
						Name: "valt-secret-store",
					},
					Spec: esv1beta1.SecretStoreSpec{
						Provider: &esv1beta1.SecretStoreProvider{
							Vault: &esv1beta1.VaultProvider{},
						},
					},
				},
				&esv1beta1.ExternalSecret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "external-secret-a",
						Namespace: "namespace-a",
					},
					Spec: esv1beta1.ExternalSecretSpec{
						SecretStoreRef: esv1beta1.SecretStoreRef{
							Name: "aws-secret-store",
						},
						Target: esv1beta1.ExternalSecretTarget{
							Name: "secret-a",
						},
						Data: []esv1beta1.ExternalSecretData{
							{
								SecretKey: "secret-a",
								RemoteRef: esv1beta1.ExternalSecretDataRemoteRef{
									Key:      "Data/key-a",
									Property: "property-a",
									Version:  "1.0.0",
								},
							},
						},
					},
				},
				&esv1beta1.ExternalSecret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "external-secret-b",
						Namespace: "namespace-b",
					},
					Spec: esv1beta1.ExternalSecretSpec{
						SecretStoreRef: esv1beta1.SecretStoreRef{
							Name: "valt-secret-store",
							Kind: esv1beta1.ClusterSecretStoreKind,
						},
						Target: esv1beta1.ExternalSecretTarget{
							Name: "secret-b",
						},
						Data: []esv1beta1.ExternalSecretData{
							{
								SecretKey: "secret-b",
								RemoteRef: esv1beta1.ExternalSecretDataRemoteRef{
									Key:      "Data/key-b",
									Property: "property-b",
									Version:  "1.0.0",
								},
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		request  *capiv1_proto.GetExternalSecretRequest
		response *capiv1_proto.GetExternalSecretResponse
		err      bool
	}{
		{
			request: &capiv1_proto.GetExternalSecretRequest{
				Name:        "external-secret-a",
				Namespace:   "namespace-a",
				ClusterName: "management",
			},
			response: &capiv1_proto.GetExternalSecretResponse{
				SecretName:      "secret-a",
				Name:            "external-secret-a",
				Namespace:       "namespace-a",
				ClusterName:     "management",
				SecretStore:     "aws-secret-store",
				SecretStoreType: "AWS Secrets Manager",
				SecretPath:      "Data/key-a",
				Properties: map[string]string{
					"property-a": "secret-a",
				},
				Version: "1.0.0",
			},
		},
		{
			request: &capiv1_proto.GetExternalSecretRequest{
				Name:        "external-secret-b",
				Namespace:   "namespace-b",
				ClusterName: "management",
			},
			response: &capiv1_proto.GetExternalSecretResponse{
				SecretName:      "secret-b",
				Name:            "external-secret-b",
				Namespace:       "namespace-b",
				ClusterName:     "management",
				SecretStore:     "valt-secret-store",
				SecretStoreType: "HashiCorp Vault",
				SecretPath:      "Data/key-b",
				Properties: map[string]string{
					"property-b": "secret-b",
				},
				Version: "1.0.0",
			},
		},
		{
			request: &capiv1_proto.GetExternalSecretRequest{
				Name:        uuid.NewString(),
				Namespace:   uuid.NewString(),
				ClusterName: uuid.NewString(),
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
		res, err := s.GetExternalSecret(context.Background(), tt.request)
		if err != nil {
			if tt.err {
				continue
			}
			t.Fatalf("got unexpected error when getting external secret, error: %v", err)
		}
		assert.Equal(t, tt.response.SecretName, res.SecretName, "secret name is not correct")
		assert.Equal(t, tt.response.Name, res.Name, "external secret name is not correct")
		assert.Equal(t, tt.response.Namespace, res.Namespace, "namespace is not correct")
		assert.Equal(t, tt.response.ClusterName, res.ClusterName, "cluster name is not correct")
		assert.Equal(t, tt.response.SecretStore, res.SecretStore, "secret store is not correct")
		assert.Equal(t, tt.response.SecretStoreType, res.SecretStoreType, "secret store type is not correct")
		assert.Equal(t, tt.response.SecretPath, res.SecretPath, "secret path is not correct")
		assert.Equal(t, tt.response.Properties, res.Properties, "property is not correct")
		assert.Equal(t, tt.response.Version, res.Version, "version is not correct")
	}
}

func TestListSecretStores(t *testing.T) {
	clusters := []struct {
		name  string
		state []runtime.Object
	}{
		{
			name: "management",
			state: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-a",
					},
				},
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-b",
					},
				},
				&esv1beta1.SecretStoreList{
					Items: []esv1beta1.SecretStore{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "secret-store-1",
								Namespace: "namespace-a",
							},
							Spec: esv1beta1.SecretStoreSpec{
								Provider: &esv1beta1.SecretStoreProvider{
									AWS: &esv1beta1.AWSProvider{
										Region: "eu-north-1",
									},
								},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "secret-store-2",
								Namespace: "namespace-b",
							},
							Spec: esv1beta1.SecretStoreSpec{
								Provider: &esv1beta1.SecretStoreProvider{
									AzureKV: &esv1beta1.AzureKVProvider{},
								},
							},
						},
					},
				},
				&esv1beta1.ClusterSecretStoreList{
					Items: []esv1beta1.ClusterSecretStore{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "cluster-secret-store-3",
								Namespace: "",
							},
							Spec: esv1beta1.SecretStoreSpec{
								Provider: &esv1beta1.SecretStoreProvider{
									GCPSM: &esv1beta1.GCPSMProvider{},
								},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "cluster-secret-store-4",
								Namespace: "",
							},
							Spec: esv1beta1.SecretStoreSpec{
								Provider: &esv1beta1.SecretStoreProvider{
									Vault: &esv1beta1.VaultProvider{},
								},
							},
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
						Name: "namespace-a",
					},
				},
				&esv1beta1.SecretStoreList{
					Items: []esv1beta1.SecretStore{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "secret-store-5",
								Namespace: "namespace-a",
							},
							Spec: esv1beta1.SecretStoreSpec{
								Provider: &esv1beta1.SecretStoreProvider{
									AWS: &esv1beta1.AWSProvider{
										Region: "eu-north-1",
									},
								},
							},
						},
					},
				},
				&esv1beta1.ClusterSecretStoreList{
					Items: []esv1beta1.ClusterSecretStore{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "cluster-secret-store-6",
								Namespace: "",
							},
							Spec: esv1beta1.SecretStoreSpec{
								Provider: &esv1beta1.SecretStoreProvider{
									AWS: &esv1beta1.AWSProvider{
										Region: "eu-north-1",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		request  *capiv1_proto.ListExternalSecretStoresRequest
		response *capiv1_proto.ListExternalSecretStoresResponse
		err      bool
	}{
		{
			request: &capiv1_proto.ListExternalSecretStoresRequest{
				ClusterName: "management",
			},
			response: &capiv1_proto.ListExternalSecretStoresResponse{
				Stores: []*capiv1_proto.ExternalSecretStore{
					{
						Kind:      esv1beta1.SecretStoreKind,
						Name:      "secret-store-1",
						Namespace: "namespace-a",
						Type:      "AWS Secrets Manager",
					},
					{
						Kind:      esv1beta1.SecretStoreKind,
						Name:      "secret-store-2",
						Namespace: "namespace-b",
						Type:      "Azure Key Vault",
					},
					{
						Kind:      esv1beta1.ClusterSecretStoreKind,
						Name:      "cluster-secret-store-3",
						Namespace: "",
						Type:      "Google Cloud Platform Secret Manager",
					},
					{
						Kind:      esv1beta1.ClusterSecretStoreKind,
						Name:      "cluster-secret-store-4",
						Namespace: "",
						Type:      "HashiCorp Vault",
					},
				},
				Total: 4,
			},
		},
		{
			request: &capiv1_proto.ListExternalSecretStoresRequest{
				ClusterName: "leaf-1",
			},
			response: &capiv1_proto.ListExternalSecretStoresResponse{
				Stores: []*capiv1_proto.ExternalSecretStore{
					{
						Kind:      esv1beta1.SecretStoreKind,
						Name:      "secret-store-5",
						Namespace: "namespace-a",
						Type:      "AWS Secrets Manager",
					},
					{
						Kind:      esv1beta1.ClusterSecretStoreKind,
						Name:      "cluster-secret-store-6",
						Namespace: "",
						Type:      "AWS Secrets Manager",
					},
				},
				Total: 2,
			},
		},
		{
			request: &capiv1_proto.ListExternalSecretStoresRequest{
				ClusterName: uuid.NewString(),
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
		res, err := s.ListExternalSecretStores(context.Background(), tt.request)
		if err != nil {
			if tt.err {
				continue
			}
			t.Fatalf("got unexpected error when getting external secret, error: %v", err)
		}
		assert.ElementsMatch(t, tt.response.Stores, res.Stores, "stores do not match expected stores")
		assert.Equal(t, tt.response.Total, res.Total, "total items number is not correct")

	}
}

// TestSyncExternalSecret exectue unitTest for SyncExternalSecret
func TestSyncExternalSecret(t *testing.T) {
	clusters := []struct {
		name  string
		state []runtime.Object
	}{
		{
			name: "management",
			state: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-a",
					},
				},
				&esv1beta1.SecretStore{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "aws-secret-store",
						Namespace: "namespace-a",
					},
					Spec: esv1beta1.SecretStoreSpec{
						Provider: &esv1beta1.SecretStoreProvider{
							AWS: &esv1beta1.AWSProvider{
								Region: "eu-north-1",
							},
						},
					},
				},
				&esv1beta1.ClusterSecretStore{
					ObjectMeta: metav1.ObjectMeta{
						Name: "valt-secret-store",
					},
					Spec: esv1beta1.SecretStoreSpec{
						Provider: &esv1beta1.SecretStoreProvider{
							Vault: &esv1beta1.VaultProvider{},
						},
					},
				},
				&esv1beta1.ExternalSecret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "external-secret-a",
						Namespace: "namespace-a",
					},
					Spec: esv1beta1.ExternalSecretSpec{
						SecretStoreRef: esv1beta1.SecretStoreRef{
							Name: "aws-secret-store",
						},
						Target: esv1beta1.ExternalSecretTarget{
							Name: "secret-a",
						},
						Data: []esv1beta1.ExternalSecretData{
							{
								RemoteRef: esv1beta1.ExternalSecretDataRemoteRef{
									Key:      "Data/key-a",
									Property: "property-a",
									Version:  "1.0.0",
								},
							},
						},
					},
				},
				&esv1beta1.ExternalSecret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "external-secret-b",
						Namespace: "namespace-a",
						Annotations: map[string]string{
							"force-sync": "2020-01-01T00:00:00Z",
						},
					},
					Spec: esv1beta1.ExternalSecretSpec{
						SecretStoreRef: esv1beta1.SecretStoreRef{
							Name: "aws-secret-store",
						},
						Target: esv1beta1.ExternalSecretTarget{
							Name: "secret-b",
						},
						Data: []esv1beta1.ExternalSecretData{
							{
								RemoteRef: esv1beta1.ExternalSecretDataRemoteRef{
									Key:      "Data/key-b",
									Property: "property-b",
									Version:  "1.0.0",
								},
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		request  *capiv1_proto.SyncExternalSecretsRequest
		response *capiv1_proto.SyncExternalSecretsResponse
		err      bool
	}{
		{
			request: &capiv1_proto.SyncExternalSecretsRequest{
				ClusterName: "management",
				Namespace:   "namespace-a",
				Name:        "external-secret-a",
			},
			response: nil,
		},
		{
			request: &capiv1_proto.SyncExternalSecretsRequest{
				ClusterName: "management",
				Namespace:   "namespace-a",
				Name:        "external-secret-b",
			},
			response: nil,
		},
		{
			request: &capiv1_proto.SyncExternalSecretsRequest{
				ClusterName: "management",
				Namespace:   "namespace-a",
				Name:        "external-secret-c",
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
		res, err := s.SyncExternalSecrets(context.Background(), tt.request)
		if tt.err {
			assert.NotNil(t, err)
			continue
		}
		assert.Nil(t, err)

		//get the CR from the cluster and check if the annotation has been added/updated
		externalSecret := &esv1beta1.ExternalSecret{}
		err = clustersClients[tt.request.ClusterName].Get(context.Background(), types.NamespacedName{
			Namespace: tt.request.Namespace,
			Name:      tt.request.Name,
		}, externalSecret)
		if err != nil {
			t.Fatalf("got unexpected error when getting external secret, error: %v", err)
		}
		_, ok := externalSecret.Annotations["force-sync"]
		if !ok {
			t.Fatalf("external secret has not been updated")
		}

		assert.Equal(t, tt.response, res, "stores do not match expected stores")
	}

}
