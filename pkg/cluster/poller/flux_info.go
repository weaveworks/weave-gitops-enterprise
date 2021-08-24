package poller

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/go-logfmt/logfmt"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/handlers"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/payload"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

const (
	FluxDeploymentLabel string = "flux"
	NamespaceFlux       string = "wkp-flux"
	NameFlux            string = "flux"
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

type FluxDeploymentLogs map[string]([]payload.FluxLogInfo)

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

	deployments, err := p.clientset.AppsV1().Deployments("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("Unable to get list of deployments from API server: %v", err)
		return
	}

	for _, deployment := range deployments.Items {
		if deployment.Spec.Selector != nil && deployment.Spec.Selector.MatchLabels != nil {
			if deployment.Spec.Selector.MatchLabels["name"] == FluxDeploymentLabel {
				potentialFluxDeployments = append(potentialFluxDeployments, deployment)
			}
		}
	}

	fluxDeploymentLogs := getAllFluxDeploymentLogs(p.clientset, potentialFluxDeployments)

	fluxInfo, err := p.toFluxInfo(potentialFluxDeployments, fluxDeploymentLogs)
	if err != nil {
		log.Errorf("Unable to create fluxInfo object: %v", err)
		return
	}

	if err != nil {
		log.Errorf("Failed to query Flux logs: %v", err)
		return
	}

	if err := p.sender.Send(context.Background(), *fluxInfo); err != nil {
		log.Errorf("Unable to send fluxInfo object: %v", err)
	}
}

func (p *FluxInfoPoller) toFluxInfo(deployments []appsv1.Deployment, logs FluxDeploymentLogs) (*payload.FluxInfo, error) {
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
			Syncs:     logs[deploymentKey(deployment)],
		}

		fluxDeployments = append(fluxDeployments, fdi)
	}

	return &payload.FluxInfo{
		Token:       p.token,
		Deployments: fluxDeployments,
	}, nil
}

func getAllFluxDeploymentLogs(clientset kubernetes.Interface, deployments []appsv1.Deployment) FluxDeploymentLogs {
	logs := FluxDeploymentLogs{}
	for _, d := range deployments {
		result, err := getFluxLogs(clientset, d)
		if err != nil {
			log.Errorf("Failed to query Flux logs: %v", err)
		} else {
			logs[deploymentKey(d)] = result
		}
	}
	return logs
}

// Scan all logs and return a sorted list of unique commit refreshes
func constructLogDetails(reader io.Reader) ([]payload.FluxLogInfo, error) {
	refreshIndex := map[string]payload.FluxLogInfo{}

	d := logfmt.NewDecoder(reader)
	for d.ScanRecord() {
		info := payload.FluxLogInfo{}
		for d.ScanKeyval() {
			k := string(d.Key())
			v := string(d.Value())
			switch k {
			case "ts":
				info.Timestamp = v
			case "url":
				info.URL = v
			case "branch":
				info.Branch = v
			case "HEAD":
				info.Head = v
			case "event":
				info.Event = v
			}
		}
		if info.Event == "refreshed" && info.Timestamp != "" && info.Head != "" {
			if _, found := refreshIndex[info.Head]; !found {
				refreshIndex[info.Head] = info
			}
		}
	}

	if d.Err() != nil {
		return nil, d.Err()
	}

	details := []payload.FluxLogInfo{}
	for _, v := range refreshIndex {
		details = append(details, v)
	}
	sort.Slice(details, func(i, j int) bool { return details[i].Timestamp < details[j].Timestamp })

	log.Infof("Found sync logs: %v", details)

	return details, nil
}

func getFluxLogs(clientset kubernetes.Interface, deployment appsv1.Deployment) ([]payload.FluxLogInfo, error) {
	// Get pod name
	namespace := deployment.ObjectMeta.Namespace
	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("Error reading flux label selector: %v", err)
	}
	options := metav1.ListOptions{LabelSelector: selector.String()}

	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), options)
	if err != nil {
		return nil, fmt.Errorf("Unable to get list of pods from API server: %v", err)
	}
	if len(podList.Items) == 0 {
		return nil, errors.New("No pods found for deployment")
	}

	podName := podList.Items[0].Name

	// Get logs from last 1 hour
	hrs := int64(3600)
	podLogOpts := corev1.PodLogOptions{SinceSeconds: &hrs}
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts)
	podLogs, err := req.Stream(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Error opening stream: %v", err)
	}
	defer podLogs.Close()

	return constructLogDetails(podLogs)
}

func deploymentKey(d appsv1.Deployment) string {
	return d.ObjectMeta.Namespace + d.ObjectMeta.Name
}
