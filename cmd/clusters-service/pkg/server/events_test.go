//go:build integration
// +build integration

package server_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/internal/entesting"
	"github.com/weaveworks/weave-gitops-enterprise/test"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestListEvents(t *testing.T) {
	RegisterFailHandler(Fail)
	ctx := context.Background()

	os.Setenv("KUBEBUILDER_ASSETS", "../../../../tools/bin/envtest")
	k8sEnv, err := testutils.StartK8sTestEnvironment([]string{
		"../../../../tools/testcrds",
	})
	Expect(err).NotTo(HaveOccurred())

	c := entesting.MakeGRPCServer(t, k8sEnv.Rest, k8sEnv)

	t.Run("ListEvents returns only relevant events", func(t *testing.T) {
		ns := test.NewNamespace()

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
			Message: "this is another message",
		}
		test.Create(ctx, t, k8sEnv.Rest, ns, evt, unrelatedEvent)

		res, err := c.ListEvents(ctx, &pb.ListEventsRequest{
			ClusterName: "Default",
			InvolvedObject: &pb.ObjectRef{
				Name:      involvedObjectName,
				Namespace: ns.Name,
				Kind:      "Pod",
			},
		})
		if err != nil {
			t.Fatal("expected error not to have occurred: %w", err)
		}

		if len(res.Events) != 1 {
			t.Errorf("expected events length to be %v, got %v", 1, len(res.Events))
		}

		first := res.Events[0]

		if first.Message != evt.Message {
			t.Errorf("expected %s to equal %s", first.Message, evt.Message)
		}
	})
}
