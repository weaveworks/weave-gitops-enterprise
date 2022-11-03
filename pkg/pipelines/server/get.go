package server

import (
	"context"
	"fmt"
	"time"

	helm "github.com/fluxcd/helm-controller/api/v2beta1"
	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/pipelines/internal/convert"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

	if err := c.Get(ctx, s.cluster, client.ObjectKeyFromObject(&p), &p); err != nil {
		return nil, fmt.Errorf("failed to find pipeline=%s in namespace=%s in cluster=%s: %w", msg.Name, msg.Namespace, s.cluster, err)
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

			clusterName := s.cluster
			if t.ClusterRef != nil {
				ns := t.ClusterRef.Namespace
				if ns == "" {
					ns = p.Namespace
				}
				clusterName = types.NamespacedName{
					Name:      t.ClusterRef.Name,
					Namespace: ns,
				}.String()
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

			var clusterRef pb.ClusterRef

			if t.ClusterRef != nil {
				clusterRef = pb.ClusterRef{
					Kind:      t.ClusterRef.Kind,
					Name:      t.ClusterRef.Name,
					Namespace: t.ClusterRef.Namespace,
				}
			}
			pipelineResp.Status.Environments[e.Name].TargetsStatuses = append(targetsStatuses, &pb.PipelineTargetStatus{
				ClusterRef: &clusterRef,
				Namespace:  t.Namespace,
				Workloads:  []*pb.WorkloadStatus{ws},
			})
		}
	}
	pipelineYaml, err := yaml.Marshal(ctrl.Pipeline{})
	if err != nil {
		return nil, fmt.Errorf("error marshalling %s pipeline, %w", pipelineResp.Name, err)
	}
	pipelineResp.Yaml = string(pipelineYaml)

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
		ws.LastAppliedRevision = hr.Status.LastAppliedRevision
		ws.Conditions = []*pb.Condition{}
		for _, c := range hr.Status.Conditions {
			ws.Conditions = append(ws.Conditions, &pb.Condition{
				Type:      c.Type,
				Status:    string(c.Status),
				Reason:    c.Reason,
				Message:   c.Message,
				Timestamp: c.LastTransitionTime.Format(time.RFC3339),
			})
		}
	}

	return ws, nil
}
