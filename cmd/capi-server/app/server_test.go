package app_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/app"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
)

func TestWeaveGitOpsHandlers(t *testing.T) {
	ctx := context.Background()
	defer ctx.Done()

	go func(ctx context.Context) {
		err := app.RunInProcessGateway(ctx, "0.0.0.0:8001", nil, nil, nil, nil, nil, "default", &kubefakes.FakeKube{})
		t.Logf("%v", err)
	}(ctx)

	time.Sleep(100 * time.Millisecond)
	res, err := http.Get("http://localhost:8001/v1/applications")
	if err != nil {
		t.Fatalf("expected no errors but got: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status code to be %d but got %d instead", http.StatusOK, res.StatusCode)
	}
	res, err = http.Get("http://localhost:8001/v1/pineapples")
	if err != nil {
		t.Fatalf("expected no errors but got: %v", err)
	}
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code to be %d but got %d instead", http.StatusNotFound, res.StatusCode)
	}
}
