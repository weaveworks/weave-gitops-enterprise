//go:build integration
// +build integration

package objectscollector_test

import (
	"context"
	"fmt"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"k8s.io/client-go/rest"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/onsi/gomega"
	clusterctrlv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var cfg *rest.Config
var ctx context.Context
var cancel context.CancelFunc

func TestMain(m *testing.M) {
	// setup testEnvironment
	cmdOut, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	repoRoot := strings.TrimSpace(string(cmdOut))
	envTestPath := fmt.Sprintf("%s/tools/bin/envtest", repoRoot)
	os.Setenv("KUBEBUILDER_ASSETS", envTestPath)
	useExistingCluster := true
	testEnv := &envtest.Environment{
		UseExistingCluster: &useExistingCluster,
	}

	cfg, err = testEnv.Start()
	if err != nil {
		log.Fatalf("starting test env failed: %s", err)
	}

	defer testEnv.Stop()

	log.Println("environment started")

	err = sourcev1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("add helm to schema failed: %s", err)
	}

	err = v2beta1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("add helm to schema failed: %s", err)
	}

	err = clusterctrlv1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("add GitopsCluster to schema failed: %s", err)
	}
	ctx, cancel = context.WithCancel(context.Background())

	gomega.RegisterFailHandler(func(message string, skip ...int) {
		log.Fatalf(message)
	})

	retCode := m.Run()
	log.Printf("suite ran with return code: %d", retCode)

	cancel()

	err = testEnv.Stop()
	if err != nil {
		log.Fatalf("stoping test env failed: %s", err)
	}

	log.Println("test environment stopped")
	os.Exit(retCode)
}
