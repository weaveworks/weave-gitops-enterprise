package poller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	httppayload "github.com/weaveworks/weave-gitops-enterprise/common/http/payload"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/handlers"
	"github.com/weaveworks/weave-gitops-enterprise/common/messaging/payload"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

const (
	NamespaceGitopsRepoBroker string = "wkp-gitops-repo-broker"
	NameGitopsRepoBroker      string = "gitops-repo-broker"
)

var (
	ErrNoBranches = errors.New("No branch information available")
)

type GitCommitInfoPoller struct {
	token      string
	clientset  kubernetes.Interface
	interval   time.Duration
	httpClient *resty.Client
	sender     *handlers.GitCommitInfoSender
}

func NewGitCommitInfoPoller(token string, clientset kubernetes.Interface, interval time.Duration, httpClient *resty.Client, sender *handlers.GitCommitInfoSender) *GitCommitInfoPoller {
	return &GitCommitInfoPoller{
		token:      token,
		clientset:  clientset,
		interval:   interval,
		httpClient: httpClient,
		sender:     sender,
	}
}

func (p *GitCommitInfoPoller) Run(stopCh <-chan struct{}) {
	log.Infof("Starting GitCommitInfo poller.")

	go wait.Until(p.runWorker, p.interval, stopCh)

	<-stopCh
	log.Infof("Stopping GitCommitInfo poller.")
}

func (p *GitCommitInfoPoller) runWorker() {
	service, err := p.clientset.CoreV1().Services(NamespaceGitopsRepoBroker).Get(context.Background(), NameGitopsRepoBroker, metav1.GetOptions{})
	if err != nil {
		log.Errorf("Unable to get service '%s' in namespace '%s' from API server: %v", NameGitopsRepoBroker, NamespaceGitopsRepoBroker, err)
		return
	}
	port := service.Spec.Ports[0].Port
	url := fmt.Sprintf("http://%s.%s:%d%s", NameGitopsRepoBroker, NamespaceGitopsRepoBroker, port, "/gitops/repo/branches")

	result := httppayload.BranchesView{}
	resp, err := p.httpClient.R().
		SetHeader("Accept", "application/json").
		SetResult(&result).
		Get(url)

	if err != nil {
		log.Errorf("Failed to query '%s': %v", url, err)
		return
	}

	if resp.StatusCode() != http.StatusOK {
		log.Errorf("Failed to query '%s': %v", url, resp)
		return
	}

	gitCommitInfo, err := p.toGitCommitInfo(result)
	if err != nil {
		log.Errorf("Unable to create GitCommitInfo object: %v", err)
		return
	}

	if err := p.sender.Send(context.Background(), *gitCommitInfo); err != nil {
		log.Errorf("Unable to send GitCommitInfo object: %v", err)
	}
}

func (p *GitCommitInfoPoller) toGitCommitInfo(view httppayload.BranchesView) (*payload.GitCommitInfo, error) {

	if len(view.Branches) == 0 {
		return nil, ErrNoBranches
	}

	head := view.Branches[0].Head

	info := &payload.GitCommitInfo{
		Token: p.token,
		Commit: payload.CommitView{
			Sha: head.Hash,
			Author: payload.UserView{
				Name:  head.Author.Name,
				Email: head.Author.Email,
				Date:  head.Author.When,
			},
			Committer: payload.UserView{
				Name:  head.Committer.Name,
				Email: head.Committer.Email,
				Date:  head.Committer.When,
			},
			Message: head.Message,
		},
	}

	return info, nil
}
