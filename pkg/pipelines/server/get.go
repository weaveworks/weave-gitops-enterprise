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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type UnknownKind struct {
	kind   string
	source string
	name   string
}

func (e UnknownKind) Error() string {
	return fmt.Sprintf("unknown %s kind for %s: %s", e.source, e.name, e.kind)
}

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
	// client.Get does not always populate TypeMeta field, without this `kind` and
	// `apiVersion` are not returned in YAML representation.
	// https://github.com/kubernetes-sigs/controller-runtime/issues/1517#issuecomment-844703142
	p.SetGroupVersionKind(ctrl.GroupVersion.WithKind(ctrl.PipelineKind))

	pipelineErrors := []string{}
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
				// Do not throw an error, we want to return values we know,
				// and return with a list of errors in the response.
				pipelineErrors = append(pipelineErrors, err.Error())
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

			workloads := []*pb.WorkloadStatus{}
			if ws != nil {
				workloads = append(workloads, ws)
			}

			pipelineResp.Status.Environments[e.Name].TargetsStatuses = append(targetsStatuses, &pb.PipelineTargetStatus{
				ClusterRef: &clusterRef,
				Namespace:  t.Namespace,
				Workloads:  workloads,
			})
		}
	}

	pipelineYaml, err := yaml.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("error marshalling %s pipeline, %w", msg.Name, err)
	}
	pipelineResp.Yaml = string(pipelineYaml)

	return &pb.GetPipelineResponse{
		Pipeline: pipelineResp,
		Errors:   pipelineErrors,
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
		ws.Suspended = hr.Spec.Suspend
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
	default:
		return nil, UnknownKind{
			kind:   obj.GetKind(),
			source: "workload",
			name:   obj.GetName(),
		}
	}

	return ws, nil
}
