package server

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"

	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *server) ApprovePromotion(ctx context.Context, msg *pb.ApprovePromotionRequest) (*pb.ApprovePromotionResponse, error) {
	c, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("getting impersonated client: %w", err)
	}

	p := ctrl.Pipeline{
		ObjectMeta: v1.ObjectMeta{
			Name:      msg.Name,
			Namespace: msg.Namespace,
		},
	}

	if err := c.Get(ctx, s.cluster, client.ObjectKeyFromObject(&p), &p); err != nil {
		return nil, fmt.Errorf("failed to find pipeline=%s in namespace=%s in cluster=%s: %w", msg.Name, msg.Namespace, s.cluster, err)
	}

	_, ok := p.Status.Environments[msg.Env]
	if !ok {
		return nil, fmt.Errorf("environment status is not available for pipeline=%s in namespace=%s in cluster=%s", msg.Name, msg.Namespace, s.cluster)
	}

	sc, err := s.clients.GetServerClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed getting server client: %w", err)
	}

	var hmacSecret *corev1.Secret

	if p.Spec.Promotion != nil && p.Spec.Promotion.Strategy.SecretRef != nil {
		err := sc.Get(ctx, s.cluster, client.ObjectKey{Namespace: msg.Namespace, Name: p.Spec.Promotion.Strategy.SecretRef.Name}, hmacSecret)
		if err != nil {
			return nil, fmt.Errorf("failed getting hmac secret for pipeline=%s in namespace=%s in cluster=%s: %w", msg.Name, msg.Namespace, s.cluster, err)
		}
	}

	prURL, err := s.postApproveRequest(s.pipelineControllerAddress, p, msg.Env, msg.Revision, hmacSecret)
	if err != nil {
		return nil, fmt.Errorf("failed sending approve request to pipeline controller for pipeline=%s in namespace=%s in cluster=%s: %w",
			msg.Name, msg.Namespace, s.cluster, err)
	}

	return &pb.ApprovePromotionResponse{
		PullRequestUrl: prURL,
	}, nil
}

func sign(payload, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(payload))
	return fmt.Sprintf("sha256=%x", h.Sum(nil))
}

func (s *server) postApproveRequest(controllerAddress string, p ctrl.Pipeline, env string, revision string, hmacSecret *corev1.Secret) (string, error) {
	headers := map[string][]string{
		"Content-Type": {"application/json"},
	}

	if hmacSecret != nil {
		// the approve endpoit does not require a body, so we sign a empty string
		// just to have a valid token for authentication
		headers["X-Signature"] = []string{sign("", string(hmacSecret.Data["hmac-key"]))}
	}

	s.log.Info("Sending POST request to pipeline controller", "url", controllerAddress)
	// Create the HTTP request
	address := fmt.Sprintf("%s/approval/%s/%s/%s/%s", controllerAddress, p.Namespace, p.Name, env, revision)

	httpReq, err := http.NewRequest("POST", address, bytes.NewBuffer([]byte{}))
	if err != nil {
		return "", fmt.Errorf("failed to create approve pipeline request: %w", err)
	}

	// Set the request headers
	for k, v := range headers {
		httpReq.Header[k] = v
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send approve pipeline request: %w", err)
	}
	defer resp.Body.Close()

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response body: %w", err)
		}

		return "", fmt.Errorf("request to pipeline controller failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	if val, ok := resp.Header["Location"]; ok {
		return val[0], nil
	}

	return "", nil
}
