package server

import (
	"context"
	"fmt"

	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (s *server) ListPolicyValidations(ctx context.Context, m *capiv1_proto.ListPolicyValidationsRequest) (*capiv1_proto.ListPolicyValidationsResponse, error) {
	config := ctrl.GetConfigOrDie()
	clientset := kubernetes.NewForConfigOrDie(config)

	namespace := "default"
	items, err := GetEvents(clientset, ctx, namespace)
	if err != nil {
		fmt.Println(err)
	} else {
		for _, item := range items {
			fmt.Printf("%+v\n", item)
		}
	}
	return &capiv1_proto.ListPolicyValidationsResponse{}, nil
}

func GetEvents(clientset *kubernetes.Clientset, ctx context.Context,
	namespace string) ([]corev1.Event, error) {

	list, err := clientset.CoreV1().Events(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	fmt.Printf("%+v\n", list.Items)
	return list.Items, nil
}
