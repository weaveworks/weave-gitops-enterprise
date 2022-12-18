package server

import (
	"context"
	"errors"
	"fmt"

	esv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	"github.com/hashicorp/go-multierror"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
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

	var externalSecrets []*capiv1_proto.ExternalSecretItem
	var clusterExternalSecrets []*capiv1_proto.ExternalSecretItem

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		externalSecrets, err = s.listExternalSecrets(gctx, clustersClient)
		if err != nil {
			if !errors.As(err, &clustersmngr.ClusteredListError{}) {
				return err
			}
		}
		return nil
	})
	g.Go(func() error {
		clusterExternalSecrets, err = s.listClusterExternalSecrets(gctx, clustersClient)
		if err != nil {
			if !errors.As(err, &clustersmngr.ClusteredListError{}) {
				return err
			}
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	response := capiv1_proto.ListExternalSecretsResponse{
		Errors: respErrors,
	}

	response.Secrets = append(response.Secrets, externalSecrets...)
	response.Secrets = append(response.Secrets, clusterExternalSecrets...)
	response.Total = int32(len(response.Secrets))

	return &response, nil
}

func (s *server) listExternalSecrets(ctx context.Context, cl clustersmngr.Client) ([]*capiv1_proto.ExternalSecretItem, error) {
	list := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &esv1beta1.ExternalSecretList{}
	})

	if err := cl.ClusteredList(ctx, list, true); err != nil {
		return nil, fmt.Errorf("failed to list external secrets, error: %w", err)
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
					Timestamp:          item.CreationTimestamp.String(),
				}
				if item.Status.Conditions != nil {
					latest := item.Status.Conditions[len(item.Status.Conditions)-1]
					if latest.Type == esv1beta1.ExternalSecretReady &&
						latest.Status == v1.ConditionTrue {
						secret.Status = ExternalSecretStatusReady
					} else {
						secret.Status = ExternalSecretStatusNotReady
					}
				} else {
					secret.Status = ExternalSecretStatusNotReady
				}
				secrets = append(secrets, &secret)
			}

		}
	}
	return secrets, nil
}

func (s *server) listClusterExternalSecrets(ctx context.Context, cl clustersmngr.Client) ([]*capiv1_proto.ExternalSecretItem, error) {
	list := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &esv1beta1.ClusterExternalSecretList{}
	})

	if err := cl.ClusteredList(ctx, list, false); err != nil {
		return nil, fmt.Errorf("failed to list cluster external secrets, error: %w", err)
	}

	secretList := list.Lists()
	secrets := []*capiv1_proto.ExternalSecretItem{}
	for clusterName, objs := range secretList {
		for i := range objs {
			obj, ok := objs[i].(*esv1beta1.ClusterExternalSecretList)
			if !ok {
				continue
			}
			for _, item := range obj.Items {
				secret := capiv1_proto.ExternalSecretItem{
					ClusterName:        clusterName,
					SecretName:         item.Spec.ExternalSecretSpec.Target.Name,
					ExternalSecretName: item.GetName(),
					SecretStore:        item.Spec.ExternalSecretSpec.SecretStoreRef.Name,
					Timestamp:          item.CreationTimestamp.String(),
				}

				if item.Status.Conditions != nil {
					latest := item.Status.Conditions[len(item.Status.Conditions)-1]
					if latest.Type == esv1beta1.ClusterExternalSecretReady &&
						latest.Status == v1.ConditionTrue {
						secret.Status = ExternalSecretStatusReady
					} else {
						secret.Status = ExternalSecretStatusNotReady
					}
				} else {
					secret.Status = ExternalSecretStatusNotReady
				}

				secrets = append(secrets, &secret)
			}

		}
	}
	return secrets, nil
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
		var clusterExternalSecret esv1beta1.ClusterExternalSecret
		if err := clustersClient.Get(ctx, req.ClusterName, client.ObjectKey{Name: req.ExternalSecretName}, &clusterExternalSecret); err != nil {
			return nil, fmt.Errorf("error getting cluster external secret %s from cluster %s: %w", req.ExternalSecretName, req.ClusterName, err)
		}

		return &capiv1_proto.GetExternalSecretResponse{
			SecretName:         clusterExternalSecret.Spec.ExternalSecretSpec.Target.Name,
			ExternalSecretName: clusterExternalSecret.GetName(),
			ClusterName:        req.ClusterName,
			Namespace:          req.Namespace,
			SecretStore:        clusterExternalSecret.Spec.ExternalSecretSpec.SecretStoreRef.Name,
			SecretPath:         clusterExternalSecret.Spec.ExternalSecretSpec.Data[0].RemoteRef.Key,
			Property:           clusterExternalSecret.Spec.ExternalSecretSpec.Data[0].RemoteRef.Property,
			Version:            clusterExternalSecret.Spec.ExternalSecretSpec.Data[0].RemoteRef.Version,
		}, nil

	} else {
		var externalSecret esv1beta1.ExternalSecret
		if err := clustersClient.Get(ctx, req.ClusterName, client.ObjectKey{Name: req.ExternalSecretName, Namespace: req.Namespace}, &externalSecret); err != nil {
			return nil, fmt.Errorf("error getting external secret %s from cluster %s: %w", req.ExternalSecretName, req.ClusterName, err)
		}

		return &capiv1_proto.GetExternalSecretResponse{
			SecretName:         externalSecret.Spec.Target.Name,
			ExternalSecretName: externalSecret.GetName(),
			ClusterName:        req.ClusterName,
			Namespace:          req.Namespace,
			SecretStore:        externalSecret.Spec.SecretStoreRef.Name,
			SecretPath:         externalSecret.Spec.Data[0].RemoteRef.Key,
			Property:           externalSecret.Spec.Data[0].RemoteRef.Property,
			Version:            externalSecret.Spec.Data[0].RemoteRef.Version,
		}, nil
	}

}

func validateReq(req *capiv1_proto.GetExternalSecretRequest) error {
	if req.ClusterName == "" {
		return errors.New("cluster name is required")
	}
	if req.ExternalSecretName == "" {
		return errors.New("external secret name is required")
	}
	return nil
}
