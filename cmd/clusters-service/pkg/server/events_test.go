package server

import (
	"context"
	"fmt"
	"testing"

	cluster_services "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListEvents(t *testing.T) {
	ns := corev1.Namespace{}

	involvedObjectName := "someObject"

	evt := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s.16da7d2e2c5b0930", involvedObjectName),
			Namespace: ns.Name,
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Pod",
			Namespace: ns.Name,
			Name:      involvedObjectName,
		},
		Message: "this is a message",
	}

	unrelatedEvent := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "someotherthing.16da7d2e2c5b0930",
			Namespace: ns.Name,
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Deployment",
			Namespace: ns.Name,
			Name:      "someotherthing",
		},
		Message: "this is a message",
	}

	s := newServer(t, evt, unrelatedEvent)

	res, err := s.ListEvents(context.Background(), &cluster_services.ListEventsRequest{
		ClusterName: "Default",
		InvolvedObject: &cluster_services.ObjectRef{
			Name:      involvedObjectName,
			Namespace: ns.Name,
			Kind:      "Pod",
		},
	})
	if err != nil {
		t.Errorf("expected error not to have occured: %s", err)
	}

	if len(res.Events) != 1 {
		t.Errorf("expected events length to be %v, got %v", 1, len(res.Events))
	}

	first := res.Events[0]

	if first.Message != evt.Message {
		t.Errorf("expected %s to equal %s", first.Message, evt.Message)
	}
}

func newServer(t *testing.T, state ...runtime.Object) cluster_services.ClustersServiceServer {
	clientsPool := &clustersmngrfakes.FakeClientsPool{}
	fakeCl := createClient(t, state...)
	clientsPool.ClientsReturns(map[string]client.Client{"Default": fakeCl})
	clientsPool.ClientReturns(fakeCl, nil)
	clustersClient := clustersmngr.NewClient(clientsPool, map[string][]v1.Namespace{})

	fakeFactory := &clustersmngrfakes.FakeClientsFactory{}
	fakeFactory.GetImpersonatedClientReturns(clustersClient, nil)

	s := createServer(t, serverOptions{
		clientsFactory: fakeFactory,
	})

	return s
}
