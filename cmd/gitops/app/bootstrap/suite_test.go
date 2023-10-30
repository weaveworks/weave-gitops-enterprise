//go:build acceptance

package bootstrap_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"

	testutils "github.com/weaveworks/weave-gitops-enterprise/test/utils"
	"github.com/weaveworks/weave-gitops/pkg/kube"

	"github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var k8sClient client.Client
var kubeconfigPath string

func TestMain(m *testing.M) {
	clusterName := "cli-bootstrap-acceptance"
	kindOutput, err := createKindCluster(clusterName)
	if err != nil {
		log.Fatalf("cannot create kind cluster: %s", string(kindOutput))
	}

	cmdOut, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	repoRoot := strings.TrimSpace(string(cmdOut))
	envTestPath := fmt.Sprintf("%s/tools/bin/envtest", repoRoot)
	os.Setenv("KUBEBUILDER_ASSETS", envTestPath)
	useExistingCluster := true
	testEnv := &envtest.Environment{
		UseExistingCluster: &useExistingCluster,
	}

	cfg, err := testEnv.Start()
	if err != nil {
		log.Fatalf("starting test env failed: %s", err)
	}
	log.Println("environment started")

	kubeconfigPath, err = testutils.CreateKubeconfigFileForRestConfig(*cfg)
	if err != nil {
		log.Fatalf("cannot create kubeconfig file: %v", err)
	}

	log.Println("kubeconfig created", kubeconfigPath)

	s, err := kube.CreateScheme()
	if err != nil {
		log.Fatalf("cannot create scheme: %v", err)
	}
	_, cancel := context.WithCancel(context.Background())

	k8sClient, err = client.New(cfg, client.Options{
		Scheme: s,
	})
	if err != nil {
		log.Fatalf("cannot create kubernetes client: %s", err)
	}
	log.Println("kube client created")

	gomega.RegisterFailHandler(func(message string, skip ...int) {
		log.Println(message)
	})

	retCode := m.Run()

	cancel()

	err = testEnv.Stop()
	if err != nil {
		log.Fatalf("stoping test env failed: %s", err)
	}
	log.Println("test environment stopped")

	kindOutput, err = deleteKindCluster(clusterName)
	if err != nil {
		log.Fatalf("cannot delete kind cluster: %s", string(kindOutput))
	}
	log.Println("cluster deleted")

	os.Exit(retCode)
}
