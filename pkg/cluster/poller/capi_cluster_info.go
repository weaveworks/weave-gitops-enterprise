package poller

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/handlers"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/payload"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
)

const CAPIGroup = "cluster.x-k8s.io"
const CAPIClusterResource = "clusters"

var ClusterGVRs = []schema.GroupVersionResource{
	{
		Group:    CAPIGroup,
		Version:  "v1alpha3",
		Resource: CAPIClusterResource,
	},
	{
		Group:    CAPIGroup,
		Version:  "v1alpha4",
		Resource: CAPIClusterResource,
	},
}

type CAPIClusterInfoPoller struct {
	token    string
	client   dynamic.Interface
	interval time.Duration
	sender   *handlers.CAPIClusterInfoSender
}

func NewCAPIClusterInfoPoller(token string, client dynamic.Interface, interval time.Duration, sender *handlers.CAPIClusterInfoSender) *CAPIClusterInfoPoller {
	return &CAPIClusterInfoPoller{
		token:    token,
		client:   client,
		interval: interval,
		sender:   sender,
	}
}

func (p *CAPIClusterInfoPoller) Run(stopCh <-chan struct{}) {
	log.Infof("Starting CAPIClusterInfo poller.")

	go wait.Until(p.runWorker, p.interval, stopCh)

	<-stopCh
	log.Infof("Stopping CAPIClusterInfo poller.")
}

func (p *CAPIClusterInfoPoller) runWorker() {
	info := payload.CAPIClusterInfo{
		Token: p.token,
	}
	for _, gvr := range ClusterGVRs {
		log.Infof("Scanning %v", gvr)
		result, err := p.client.Resource(gvr).Namespace("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Errorf("Unable to list resources of type %s: %v", gvr, err)
			continue
		}
		for _, i := range result.Items {
			encodedCluster, err := i.MarshalJSON()
			if err != nil {
				log.Errorf("Unable to marshall capi cluster %v %v/%v: %v",
					gvr, i.GetNamespace(), i.GetName(), err)
			}
			info.CAPIClusters = append(info.CAPIClusters, payload.CAPICluster{
				Name:          i.GetName(),
				Namespace:     i.GetNamespace(),
				CAPIVersion:   gvr.Version,
				EncodedObject: string(encodedCluster),
			})
		}
	}

	if err := p.sender.Send(context.Background(), info); err != nil {
		log.Errorf("Unable to send CAPIClusterInfo object: %v", err)
	}
}
