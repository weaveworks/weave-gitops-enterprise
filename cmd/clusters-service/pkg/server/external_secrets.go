package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	"github.com/hashicorp/go-multierror"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	k8s_json "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ExternalSecretStatusReady    = "Ready"
	ExternalSecretStatusNotReady = "NotReady"
)

func (s *server) ListExternalSecrets(ctx context.Context, m *capiv1_proto.ListExternalSecretsRequest) (*capiv1_proto.ListExternalSecretsResponse, error) {
	respErrors := []*capiv1_proto.ListError{}
	clustersClient, err := s.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			for _, err := range merr.Errors {
				if cerr, ok := err.(*clustersmngr.ClientError); ok {
					respErrors = append(respErrors, &capiv1_proto.ListError{ClusterName: cerr.ClusterName, Message: cerr.Error()})
				}
			}
		} else {
			return nil, fmt.Errorf("unexpected error while getting clusters client, error: %w", err)
		}
	}

	externalSecrets, externalSecretsListErrors, err := s.listExternalSecrets(ctx, clustersClient)
	if err != nil {
		return nil, err
	}

	response := capiv1_proto.ListExternalSecretsResponse{
		Errors:  respErrors,
		Secrets: externalSecrets,
		Total:   int32(len(externalSecrets)),
	}

	response.Errors = append(response.Errors, externalSecretsListErrors...)
	return &response, nil
}

func (s *server) listExternalSecrets(ctx context.Context, cl clustersmngr.Client) ([]*capiv1_proto.ExternalSecretItem, []*capiv1_proto.ListError, error) {
	clusterListErrors := []*capiv1_proto.ListError{}

	list := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &esv1beta1.ExternalSecretList{}
	})

	if err := cl.ClusteredList(ctx, list, true); err != nil {
		if e, ok := err.(clustersmngr.ClusteredListError); ok {
			for i := range e.Errors {
				if !strings.Contains(e.Errors[i].Error(), "no matches for kind ") {
					clusterListErrors = append(clusterListErrors, &capiv1_proto.ListError{ClusterName: e.Errors[i].Cluster, Message: e.Errors[i].Error()})
				}

			}
		} else {
			if !strings.Contains(e.Error(), "no matches for kind ") {
				return nil, clusterListErrors, fmt.Errorf("failed to list external secrets, error: %w", err)
			}

		}
	}

	secretList := list.Lists()

	secrets := []*capiv1_proto.ExternalSecretItem{}
	for clusterName, objs := range secretList {
		for i := range objs {
			obj, ok := objs[i].(*esv1beta1.ExternalSecretList)
			if !ok {
				continue
			}
			for _, item := range obj.Items {
				secret := capiv1_proto.ExternalSecretItem{
					ClusterName:        clusterName,
					SecretName:         item.Spec.Target.Name,
					ExternalSecretName: item.GetName(),
					SecretStore:        item.Spec.SecretStoreRef.Name,
					Namespace:          item.GetNamespace(),
					Status:             getExternalSecretStatus(&item),
					Timestamp:          item.CreationTimestamp.Format(time.RFC3339),
				}

				secrets = append(secrets, &secret)
			}
		}
	}
	return secrets, clusterListErrors, nil
}

func (s *server) GetExternalSecret(ctx context.Context, req *capiv1_proto.GetExternalSecretRequest) (*capiv1_proto.GetExternalSecretResponse, error) {
	if err := validateReq(req); err != nil {
		return nil, err
	}

	clustersClient, err := s.clustersManager.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), req.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	if clustersClient == nil {
		return nil, fmt.Errorf("cluster %s not found", req.ClusterName)
	}

	if req.Namespace == "" {
		return nil, errors.New("cluster external secrets are not supported yet")

	} else {
		var externalSecret esv1beta1.ExternalSecret
		if err := clustersClient.Get(ctx, req.ClusterName, client.ObjectKey{Name: req.ExternalSecretName, Namespace: req.Namespace}, &externalSecret); err != nil {
			return nil, fmt.Errorf("error getting external secret %s from cluster %s: %w", req.ExternalSecretName, req.ClusterName, err)
		}

		externalSecret.GetObjectKind().SetGroupVersionKind(esv1beta1.SchemeGroupVersion.WithKind("ExternalSecret"))

		var buf bytes.Buffer

		serializer := k8s_json.NewSerializer(k8s_json.DefaultMetaFactory, nil, nil, false)
		if err := serializer.Encode(&externalSecret, &buf); err != nil {
			return nil, err
		}

		response := capiv1_proto.GetExternalSecretResponse{
			SecretName:         externalSecret.Spec.Target.Name,
			ExternalSecretName: externalSecret.GetName(),
			ClusterName:        req.ClusterName,
			Namespace:          req.Namespace,
			SecretStore:        externalSecret.Spec.SecretStoreRef.Name,
			Status:             getExternalSecretStatus(&externalSecret),
			Timestamp:          externalSecret.CreationTimestamp.Format(time.RFC3339),
			Yaml:               buf.String(),
		}

		if externalSecret.Spec.DataFrom != nil {
			response.SecretPath = externalSecret.Spec.DataFrom[0].Extract.Key
		}

		if externalSecret.Spec.Data != nil {
			response.SecretPath = externalSecret.Spec.Data[0].RemoteRef.Key
			response.Property = externalSecret.Spec.Data[0].RemoteRef.Property
			response.Version = externalSecret.Spec.Data[0].RemoteRef.Version
		}

		var secretStoreProvider *esv1beta1.SecretStoreProvider
		if externalSecret.Spec.SecretStoreRef.Kind == esv1beta1.ClusterSecretStoreKind {
			var clusterSecretStore esv1beta1.ClusterSecretStore
			if err := clustersClient.Get(ctx, req.ClusterName, client.ObjectKey{Name: externalSecret.Spec.SecretStoreRef.Name}, &clusterSecretStore); err == nil {
				secretStoreProvider = clusterSecretStore.Spec.Provider
			}

		} else {
			var secretStore esv1beta1.SecretStore
			if err := clustersClient.Get(ctx, req.ClusterName, client.ObjectKey{Name: externalSecret.Spec.SecretStoreRef.Name, Namespace: req.Namespace}, &secretStore); err == nil {
				secretStoreProvider = secretStore.Spec.Provider
			}
		}

		if secretStoreProvider != nil {
			response.SecretStoreType = getSecretStoreType(secretStoreProvider)
		}

		return &response, nil
	}

}

func validateReq(req *capiv1_proto.GetExternalSecretRequest) error {
	if req.ClusterName == "" {
		return errors.New("cluster name is required")
	}
	if req.ExternalSecretName == "" {
		return errors.New("external secret name is required")
	}
	if req.Namespace == "" {
		return errors.New("namespace is required")
	}
	return nil
}

func getExternalSecretStatus(item *esv1beta1.ExternalSecret) string {
	if item.Status.Conditions != nil {
		latest := item.Status.Conditions[len(item.Status.Conditions)-1]
		if latest.Type == esv1beta1.ExternalSecretReady &&
			latest.Status == v1.ConditionTrue {
			return ExternalSecretStatusReady
		} else {
			return ExternalSecretStatusNotReady
		}
	} else {
		return ExternalSecretStatusNotReady
	}
}

func (s *server) ListExternalSecretStores(ctx context.Context, req *capiv1_proto.ListExternalSecretStoresRequest) (*capiv1_proto.ListExternalSecretStoresResponse, error) {
	clustersClient, err := s.clustersManager.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), req.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	if clustersClient == nil {
		return nil, fmt.Errorf("cluster %s not found", req.ClusterName)
	}

	var secretStores esv1beta1.SecretStoreList
	var clusterSecretStores esv1beta1.ClusterSecretStoreList

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return clustersClient.List(gctx, req.ClusterName, &secretStores)
	})

	g.Go(func() error {
		return clustersClient.List(gctx, req.ClusterName, &clusterSecretStores)
	})

	if err := g.Wait(); err != nil {
		if strings.Contains(err.Error(), "no matches for kind") {
			return &capiv1_proto.ListExternalSecretStoresResponse{}, nil
		}
		return nil, fmt.Errorf("failed to list secret stores, error %w", err)
	}

	response := capiv1_proto.ListExternalSecretStoresResponse{}

	for _, item := range secretStores.Items {
		response.Stores = append(response.Stores, &capiv1_proto.ExternalSecretStore{
			Kind:      item.GetKind(),
			Name:      item.GetName(),
			Namespace: item.GetNamespace(),
			Type:      getSecretStoreType(item.Spec.Provider),
		})
	}

	for _, item := range clusterSecretStores.Items {
		response.Stores = append(response.Stores, &capiv1_proto.ExternalSecretStore{
			Kind:      item.GetKind(),
			Name:      item.GetName(),
			Namespace: item.GetNamespace(),
			Type:      getSecretStoreType(item.Spec.Provider),
		})
	}

	sort.Slice(response.Stores, func(i, j int) bool {
		return response.Stores[i].Name < response.Stores[j].Name
	})

	response.Total = int32(len(response.Stores))
	return &response, nil
}

// getSecretStoreType gets SecretStoreType from SecretStore object
func getSecretStoreType(provider *esv1beta1.SecretStoreProvider) string {
	if provider.AWS != nil {
		return "AWS Secrets Manager"
	} else if provider.AzureKV != nil {
		return "Azure Key Vault"
	} else if provider.GCPSM != nil {
		return "Google Cloud Platform Secret Manager"
	} else if provider.Vault != nil {
		return "HashiCorp Vault"
	} else {
		return "Unknown"
	}
}

// SyncExternalSecret triggers a sync of an externalSecret from externalSecret operators
func (s *server) SyncExternalSecrets(ctx context.Context, req *capiv1_proto.SyncExternalSecretsRequest) (*capiv1_proto.SyncExternalSecretsResponse, error) {

	clustersClient, err := s.clustersManager.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), req.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	if clustersClient == nil {
		return nil, fmt.Errorf("cluster %s not found", req.ClusterName)
	}

	var externalSecret esv1beta1.ExternalSecret
	if err := clustersClient.Get(ctx, req.ClusterName, client.ObjectKey{Name: req.ExternalSecretName, Namespace: req.Namespace}, &externalSecret); err != nil {
		return nil, fmt.Errorf("error getting external secret %s from cluster %s: %w", req.ExternalSecretName, req.ClusterName, err)
	} else {

		if _, ok := externalSecret.Annotations["force-sync"]; ok {
			externalSecret.Annotations["force-sync"] = time.Now().Format(time.RFC3339)
		} else {
			externalSecret.Annotations = map[string]string{"force-sync": time.Now().Format(time.RFC3339)}
		}

		if err := clustersClient.Update(ctx, req.ClusterName, &externalSecret); err != nil {
			return nil, fmt.Errorf("error updating external secret %s from cluster %s: %w", req.ExternalSecretName, req.ClusterName, err)
		}
	}

	return nil, nil
}
