package poller

import (
	"context"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/common/messaging/handlers"
	"github.com/weaveworks/wks/common/messaging/payload"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

const (
	FluxDeploymentLabel string = "flux"
)

var (
	ErrNoFluxDeployments = errors.New("No flux deployments detected")
)

type FluxInfoPoller struct {
	token     string
	clientset kubernetes.Interface
	interval  time.Duration
	sender    *handlers.FluxInfoSender
}

func NewFluxInfoPoller(token string, clientset kubernetes.Interface, interval time.Duration, sender *handlers.FluxInfoSender) *FluxInfoPoller {
	return &FluxInfoPoller{
		token:     token,
		clientset: clientset,
		interval:  interval,
		sender:    sender,
	}
}

func (p *FluxInfoPoller) Run(stopCh <-chan struct{}) {
	log.Infof("Starting FluxInfo poller.")

	go wait.Until(p.runWorker, p.interval, stopCh)

	<-stopCh
	log.Infof("Stopping FluxInfo poller.")
}

func (p *FluxInfoPoller) runWorker() {
	potentialFluxDeployments := make([]appsv1.Deployment, 0)

	namespaces, err := p.clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("Unable to get list of namespaces from API server: %v", err)
		return
	}

	for _, ns := range namespaces.Items {
		deployments, err := p.clientset.AppsV1().Deployments(ns.ObjectMeta.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			log.Errorf("Unable to get list of deployments in ns %s from API server: %v", ns.ObjectMeta.Name, err)
			return
		}
		for _, deployment := range deployments.Items {
			if deployment.Spec.Selector != nil && deployment.Spec.Selector.MatchLabels != nil {
				if deployment.Spec.Selector.MatchLabels["name"] == FluxDeploymentLabel {
					potentialFluxDeployments = append(potentialFluxDeployments, deployment)
				}
			}
		}
	}

	FluxInfo, err := p.toFluxInfo(potentialFluxDeployments)
	if err != nil {
		log.Errorf("Unable to create FluxInfo object: %v", err)
		return
	}

	if err := p.sender.Send(context.Background(), *FluxInfo); err != nil {
		log.Errorf("Unable to send FluxInfo object: %v", err)
	}
}

func (p *FluxInfoPoller) toFluxInfo(deployments []appsv1.Deployment) (*payload.FluxInfo, error) {
	fluxDeployments := make([]payload.FluxDeploymentInfo, 0)

	if len(deployments) == 0 {
		return nil, ErrNoFluxDeployments
	}

	for _, deployment := range deployments {
		fdi := payload.FluxDeploymentInfo{
			Name:      deployment.ObjectMeta.Name,
			Namespace: deployment.ObjectMeta.Namespace,
			Args:      deployment.Spec.Template.Spec.Containers[0].Args,
			Image:     deployment.Spec.Template.Spec.Containers[0].Image,
		}

		fluxDeployments = append(fluxDeployments, fdi)
	}

	return &payload.FluxInfo{
		Token:       p.token,
		Deployments: fluxDeployments,
	}, nil
}
