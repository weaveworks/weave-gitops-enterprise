package poller

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/common/messaging/handlers"
	"github.com/weaveworks/wks/common/messaging/payload"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
)

type WorkspaceInfoPoller struct {
	token    string
	client   dynamic.Interface
	interval time.Duration
	sender   *handlers.WorkspaceInfoSender
}

func NewWorkspaceInfoPoller(token string, client dynamic.Interface, interval time.Duration, sender *handlers.WorkspaceInfoSender) *WorkspaceInfoPoller {
	return &WorkspaceInfoPoller{
		token:    token,
		client:   client,
		interval: interval,
		sender:   sender,
	}
}

func (p *WorkspaceInfoPoller) Run(stopCh <-chan struct{}) {
	log.Infof("Starting WorkspaceInfo poller.")

	go wait.Until(p.runWorker, p.interval, stopCh)

	<-stopCh
	log.Infof("Stopping WorkspaceInfo poller.")
}

func (p *WorkspaceInfoPoller) runWorker() {
	workspaceRes := schema.GroupVersionResource{Group: "wkp.weave.works", Version: "v1beta1", Resource: "workspaces"}
	result, err := p.client.Resource(workspaceRes).Namespace("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("Unable to list resources of type %s: %v", workspaceRes, err)
		return
	}

	info := payload.WorkspaceInfo{
		Token: p.token,
	}
	for _, i := range result.Items {
		info.Workspaces = append(info.Workspaces, payload.Workspace{
			Name:      i.GetName(),
			Namespace: i.GetNamespace(),
		})
	}

	if err := p.sender.Send(context.Background(), info); err != nil {
		log.Errorf("Unable to send WorkspaceInfo object: %v", err)
	}
}
