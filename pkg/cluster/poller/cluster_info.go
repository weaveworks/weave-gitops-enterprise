package poller

import (
	"context"
	"errors"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/common/messaging/handlers"
	"github.com/weaveworks/wks/common/messaging/payload"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

const (
	NamespaceKubeSystem   string = "kube-system"
	LabelNodeControlPlane string = "node-role.kubernetes.io/master"
)

var (
	ErrNoNodesOrNamespaces = errors.New("No node or namespace information returned from API server.")
)

type ClusterInfoPoller struct {
	token     string
	clientset kubernetes.Interface
	interval  time.Duration
	sender    *handlers.ClusterInfoSender
}

func NewClusterInfoPoller(token string, clientset kubernetes.Interface, interval time.Duration, sender *handlers.ClusterInfoSender) *ClusterInfoPoller {
	return &ClusterInfoPoller{
		token:     token,
		clientset: clientset,
		interval:  interval,
		sender:    sender,
	}
}

func (p *ClusterInfoPoller) Run(stopCh <-chan struct{}) {
	log.Infof("Starting ClusterInfo poller.")

	go wait.Until(p.runWorker, p.interval, stopCh)

	<-stopCh
	log.Infof("Stopping ClusterInfo poller.")
}

func (p *ClusterInfoPoller) runWorker() {
	nodes, err := p.clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("Unable to get list of nodes from API server: %v", err)
		return
	}

	namespaces, err := p.clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("Unable to get list of namespaces from API server: %v", err)
		return
	}

	clusterInfo, err := p.toClusterInfo(nodes, namespaces)
	if err != nil {
		log.Errorf("Unable to create ClusterInfo object: %v", err)
		return
	}

	if err := p.sender.Send(context.Background(), *clusterInfo); err != nil {
		log.Errorf("Unable to send ClusterInfo object: %v", err)
	}
}

func (p *ClusterInfoPoller) toClusterInfo(nodeList *v1.NodeList, namespaceList *v1.NamespaceList) (*payload.ClusterInfo, error) {
	if len(nodeList.Items) == 0 || len(namespaceList.Items) == 0 {
		return nil, ErrNoNodesOrNamespaces
	}

	// Using the UID of the kube-system as a way to identify the cluster
	var id string
	for _, ns := range namespaceList.Items {
		if ns.Name == NamespaceKubeSystem {
			id = string(ns.UID)
			break
		}
	}

	// ID of the node assigned by the cloud provider in the format: <ProviderName>://<ProviderSpecificNodeID>
	providerID := nodeList.Items[0].Spec.ProviderID
	providerName := providerID[:strings.Index(providerID, "://")]

	nodes := make([]payload.Node, 0)
	for _, node := range nodeList.Items {
		n := payload.Node{
			MachineID:      node.Status.NodeInfo.MachineID,
			Name:           node.Name,
			KubeletVersion: node.Status.NodeInfo.KubeletVersion,
		}

		if _, ok := node.Labels[LabelNodeControlPlane]; ok {
			n.IsControlPlane = true
		}

		nodes = append(nodes, n)
	}

	cluster := payload.Cluster{
		ID:    id,
		Type:  providerName,
		Nodes: nodes,
	}

	return &payload.ClusterInfo{
		Token:   p.token,
		Cluster: cluster,
	}, nil
}
