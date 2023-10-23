//go:build integration

package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"

	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var cfg *rest.Config
var kubeconfigPath string

func TestMain(m *testing.M) {
	cmdOut, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	repoRoot := strings.TrimSpace(string(cmdOut))
	envTestPath := fmt.Sprintf("%s/tools/bin/envtest", repoRoot)
	os.Setenv("KUBEBUILDER_ASSETS", envTestPath)
	useExistingCluster := false
	testEnv := &envtest.Environment{
		UseExistingCluster: &useExistingCluster,
	}

	cfg, err = testEnv.Start()
	if err != nil {
		log.Fatalf("starting test env failed: %s", err)
	}
	log.Println("environment started")
	//
	//s, err := kube.CreateScheme()
	//if err != nil {
	//	log.Fatalf("cannot create scheme: %v", err)
	//}
	//_, cancel := context.WithCancel(context.Background())
	//
	//k8sClient, err = client.New(cfg, client.Options{
	//	Scheme: s,
	//})
	//if err != nil {
	//	log.Fatalf("cannot create kubernetes client: %s", err)
	//}
	//log.Println("kube client created")
	//
	//gomega.RegisterFailHandler(func(message string, skip ...int) {
	//	log.Println(message)
	//})

	retCode := m.Run()

	//cancel()

	err = testEnv.Stop()
	if err != nil {
		log.Fatalf("stoping test env failed: %s", err)
	}
	log.Println("test environment stopped")

	os.Exit(retCode)
}
