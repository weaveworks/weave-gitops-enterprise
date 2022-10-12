package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos/profiles"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/watcher/cache"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/metadata"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	OctetStreamType = "application/octet-stream"
	JSONType        = "application/json"
)

type ProfilesConfig struct {
	helmRepoNamespace string
	helmRepoName      string
	helmCache         cache.Cache
	clusterConfig     kube.ClusterConfig
}

func NewProfilesConfig(clusterConfig kube.ClusterConfig, helmCache cache.Cache, helmRepoNamespace, helmRepoName string) ProfilesConfig {
	return ProfilesConfig{
		helmRepoNamespace: helmRepoNamespace,
		helmRepoName:      helmRepoName,
		helmCache:         helmCache,
		clusterConfig:     clusterConfig,
	}
}

type ProfilesServer struct {
	pb.UnimplementedProfilesServer

	Log               logr.Logger
	HelmRepoName      string
	HelmRepoNamespace string
	HelmCache         cache.Cache
	ClientGetter      kube.ClientGetter
}

func NewProfilesServer(log logr.Logger, config ProfilesConfig) pb.ProfilesServer {
	configGetter := kube.NewImpersonatingConfigGetter(config.clusterConfig.DefaultConfig, false)
	clientGetter := kube.NewDefaultClientGetter(configGetter, config.clusterConfig.ClusterName)

	return &ProfilesServer{
		Log:               log.WithName("profiles-server"),
		HelmRepoNamespace: config.helmRepoNamespace,
		HelmRepoName:      config.helmRepoName,
		HelmCache:         config.helmCache,
		ClientGetter:      clientGetter,
	}
}

func (s *ProfilesServer) GetProfiles(ctx context.Context, msg *pb.GetProfilesRequest) (*pb.GetProfilesResponse, error) {
	kubeClient, err := s.ClientGetter.Client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get a Kubernetes client: %w", err)
	}

	helmRepoName := s.HelmRepoName
	helmRepoNamespace := s.HelmRepoNamespace
	clusterName := "management"
	clusterNamespace := "default"

	if msg.HelmRepoName != "" {
		helmRepoName = msg.HelmRepoName
	}

	if msg.HelmRepoNamespace != "" {
		helmRepoNamespace = msg.HelmRepoNamespace
	}

	if msg.ClusterName != "" {
		clusterName = msg.ClusterName
	}

	if msg.ClusterNamespace != "" {
		clusterNamespace = msg.ClusterNamespace
	}

	helmRepo := &sourcev1.HelmRepository{}
	err = kubeClient.Get(ctx, client.ObjectKey{
		Name:      helmRepoName,
		Namespace: helmRepoNamespace,
	}, helmRepo)

	if err != nil {
		if apierrors.IsNotFound(err) {
			errMsg := fmt.Sprintf("HelmRepository %q/%q does not exist", helmRepoNamespace, helmRepoName)
			s.Log.Error(err, errMsg)

			return &pb.GetProfilesResponse{
					Profiles: []*pb.Profile{},
				}, &grpcruntime.HTTPStatusError{
					Err:        errors.New(errMsg),
					HTTPStatus: http.StatusOK,
				}
		}

		return nil, fmt.Errorf("failed to get HelmRepository %q/%q: %w", helmRepoNamespace, helmRepoName, err)
	}

	log := s.Log.WithValues("repository", types.NamespacedName{
		Namespace: helmRepo.Namespace,
		Name:      helmRepo.Name,
	})

	ps, err := s.HelmCache.ListProfiles(
		logr.NewContext(ctx, log),
		types.NamespacedName{
			Namespace: clusterNamespace,
			Name:      clusterName,
		},
		types.NamespacedName{
			Namespace: helmRepo.Namespace,
			Name:      helmRepo.Name,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan HelmRepository %q/%q for charts: %w", s.HelmRepoNamespace, s.HelmRepoName, err)
	}

	return &pb.GetProfilesResponse{
		Profiles: ps,
	}, nil
}

func (s *ProfilesServer) GetProfileValues(ctx context.Context, msg *pb.GetProfileValuesRequest) (*httpbody.HttpBody, error) {
	kubeClient, err := s.ClientGetter.Client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get a Kubernetes client: %w", err)
	}

	helmRepoName := s.HelmRepoName
	helmRepoNamespace := s.HelmRepoNamespace
	clusterName := "management"
	clusterNamespace := "default"

	if msg.HelmRepoName != "" {
		helmRepoName = msg.HelmRepoName
	}

	if msg.HelmRepoNamespace != "" {
		helmRepoNamespace = msg.HelmRepoNamespace
	}

	helmRepo := &sourcev1.HelmRepository{}
	err = kubeClient.Get(ctx, client.ObjectKey{
		Name:      helmRepoName,
		Namespace: helmRepoNamespace,
	}, helmRepo)

	if err != nil {
		if apierrors.IsNotFound(err) {
			errMsg := fmt.Sprintf("HelmRepository %q/%q does not exist", helmRepoNamespace, helmRepoName)
			s.Log.Error(err, errMsg)

			return &httpbody.HttpBody{
					ContentType: "application/json",
					Data:        []byte{},
				}, &grpcruntime.HTTPStatusError{
					Err:        errors.New(errMsg),
					HTTPStatus: http.StatusOK,
				}
		}

		return nil, fmt.Errorf("failed to get HelmRepository %q/%q", helmRepoNamespace, helmRepoName)
	}

	log := s.Log.WithValues("repository", types.NamespacedName{
		Namespace: helmRepo.Namespace,
		Name:      helmRepo.Name,
	})

	data, err := s.HelmCache.GetProfileValues(
		logr.NewContext(ctx, log),
		types.NamespacedName{
			Namespace: clusterNamespace,
			Name:      clusterName,
		},
		types.NamespacedName{
			Namespace: helmRepo.Namespace,
			Name:      helmRepo.Name,
		},
		msg.ProfileName,
		msg.ProfileVersion,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve values file from Helm chart '%s' (%s): %w", msg.ProfileName, msg.ProfileVersion, err)
	}

	var acceptHeader string

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if accept, ok := md["accept"]; ok {
			acceptHeader = strings.Join(accept, ",")
		}
	}

	if strings.Contains(acceptHeader, OctetStreamType) {
		return &httpbody.HttpBody{
			ContentType: OctetStreamType,
			Data:        data,
		}, nil
	}

	res, err := json.Marshal(&pb.GetProfileValuesResponse{
		Values: base64.StdEncoding.EncodeToString(data),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response to JSON: %w", err)
	}

	return &httpbody.HttpBody{
		ContentType: JSONType,
		Data:        res,
	}, nil
}
