package server

import (
	"context"
	"fmt"

	helm "github.com/fluxcd/helm-controller/api/v2beta1"
	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/fetcher"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/pipelines/internal/convert"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *server) GetPipeline(ctx context.Context, msg *pb.GetPipelineRequest) (*pb.GetPipelineResponse, error) {
	c, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))

	if err != nil {
		return nil, fmt.Errorf("getting impersonated client: %w", err)
	}

	p := ctrl.Pipeline{
		ObjectMeta: v1.ObjectMeta{
			Name:      msg.Name,
			Namespace: msg.Namespace,
		},
	}

	if err := c.Get(ctx, fetcher.ManagementClusterName, client.ObjectKeyFromObject(&p), &p); err != nil {
		return nil, fmt.Errorf("failed to find pipeline=%s in namespace=%s in cluster=%s: %w", msg.Name, msg.Namespace, fetcher.ManagementClusterName, err)
	}

	pipelineResp := convert.PipelineToProto(p)
	pipelineResp.Status = &pb.PipelineStatus{
		Environments: map[string]*pb.PipelineStatus_TargetStatusList{},
	}

	for _, e := range p.Spec.Environments {
		for _, t := range e.Targets {
			app := &unstructured.Unstructured{}
			app.SetAPIVersion(p.Spec.AppRef.APIVersion)
			app.SetKind(p.Spec.AppRef.Kind)
			app.SetName(p.Spec.AppRef.Name)
			app.SetNamespace(t.Namespace)

			clusterName := fetcher.ManagementClusterName
			if t.ClusterRef.Name != "" {
				clusterName = t.ClusterRef.Name
			}

			if err := c.Get(ctx, clusterName, client.ObjectKeyFromObject(app), app); err != nil {
				return nil, fmt.Errorf("failed getting app=%s on cluster=%s: %w", app.GetName(), clusterName, err)
			}

			ws, err := getWorkloadStatus(app)
			if err != nil {
				return nil, err
			}

			if _, ok := pipelineResp.Status.Environments[e.Name]; !ok {
				pipelineResp.Status.Environments[e.Name] = &pb.PipelineStatus_TargetStatusList{
					TargetsStatuses: []*pb.PipelineTargetStatus{},
				}
			}

			targetsStatuses := pipelineResp.Status.Environments[e.Name].TargetsStatuses
			pipelineResp.Status.Environments[e.Name].TargetsStatuses = append(targetsStatuses, &pb.PipelineTargetStatus{
				ClusterRef: &pb.ClusterRef{
					Kind: t.ClusterRef.Kind,
					Name: t.ClusterRef.Name,
				},
				Namespace: p.Namespace,
				Workloads: []*pb.WorkloadStatus{ws},
			})
		}
	}

	return &pb.GetPipelineResponse{
		Pipeline: pipelineResp,
	}, nil
}

func getWorkloadStatus(obj *unstructured.Unstructured) (*pb.WorkloadStatus, error) {
	ws := &pb.WorkloadStatus{}

	switch obj.GetKind() {
	case "HelmRelease":
		hr := helm.HelmRelease{}

		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &hr); err != nil {
			return nil, fmt.Errorf("failed converting unstructured.Unstructured to HelmRelease: %w", err)
		}
		ws.Kind = hr.Kind
		ws.Name = hr.Name
		ws.Version = hr.Spec.Chart.Spec.Version
	}

	return ws, nil
}
