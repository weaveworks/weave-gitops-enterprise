package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var errRequiredClusterName = errors.New("`clusterName` param is required")

func (s *server) ListEvents(ctx context.Context, msg *capiv1_proto.ListEventsRequest) (*capiv1_proto.ListEventsResponse, error) {
	c, err := s.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %s", err)
	}

	if msg.ClusterName == "" {
		return nil, errRequiredClusterName
	}

	scoped, err := c.Scoped(msg.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("scoping to cluster %q: %w", msg.ClusterName, err)
	}

	fields := client.MatchingFields{
		"involvedObject.kind":      msg.InvolvedObject.Kind,
		"involvedObject.name":      msg.InvolvedObject.Name,
		"involvedObject.namespace": msg.InvolvedObject.Namespace,
	}

	l := &corev1.EventList{}
	if err := scoped.List(ctx, l, fields); err != nil {
		return nil, fmt.Errorf("listing events: %w", err)
	}

	result := []*capiv1_proto.Event{}
	for _, e := range l.Items {
		result = append(result, &capiv1_proto.Event{
			Type:      e.Type,
			Component: e.Source.Component,
			Name:      e.ObjectMeta.Name,
			Reason:    e.Reason,
			Message:   e.Message,
			Timestamp: e.LastTimestamp.Format(time.RFC3339),
			Host:      e.Source.Host,
		})
	}

	return &capiv1_proto.ListEventsResponse{Events: result}, nil
}
