//go:build integration

package bootstrap_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubectl/pkg/scheme"

	"github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var k8sClient client.Client
var cfg *rest.Config
var kubeconfigPath string

func TestMain(m *testing.M) {
	// setup testEnvironment
	cmdOut, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	repoRoot := strings.TrimSpace(string(cmdOut))
	envTestPath := fmt.Sprintf("%s/tools/bin/envtest", repoRoot)
	os.Setenv("KUBEBUILDER_ASSETS", envTestPath)
	// enable me for development
	//useExistingCluster := true
	testEnv := &envtest.Environment{
		//UseExistingCluster: &useExistingCluster,
	}

	cfg, err = testEnv.Start()
	if err != nil {
		log.Fatalf("starting test env failed: %s", err)
	}
	log.Println("environment started")

	kubeconfigPath = CreateKubeconfigFileForRestConfig(*cfg)
	log.Println("kubeconfig created", kubeconfigPath)

	err = sourcev1beta2.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("add helm to schema failed: %s", err)
	}

	err = sourcev1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("add helm to schema failed: %s", err)
	}

	err = kustomizev1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("add helm to schema failed: %s", err)
	}

	err = helmv2beta1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("add helm to schema failed: %s", err)
	}

	_, cancel := context.WithCancel(context.Background())

	k8sClient, err = client.New(cfg, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		log.Fatalf("cannot create kubernetes client: %s", err)
	}

	log.Println("kube client created")

	gomega.RegisterFailHandler(func(message string, skip ...int) {
		log.Println(message)
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

func CreateKubeconfigFileForRestConfig(restConfig rest.Config) string {
	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters["default-cluster"] = &clientcmdapi.Cluster{
		Server:                   restConfig.Host,
		CertificateAuthorityData: restConfig.CAData,
	}
	contexts := make(map[string]*clientcmdapi.Context)
	contexts["default-context"] = &clientcmdapi.Context{
		Cluster:  "default-cluster",
		AuthInfo: "default-user",
	}
	authinfos := make(map[string]*clientcmdapi.AuthInfo)
	authinfos["default-user"] = &clientcmdapi.AuthInfo{
		ClientCertificateData: restConfig.CertData,
		ClientKeyData:         restConfig.KeyData,
	}
	clientConfig := clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: "default-context",
		AuthInfos:      authinfos,
	}
	kubeConfigFile, _ := os.CreateTemp("", "kubeconfig")
	_ = clientcmd.WriteToFile(clientConfig, kubeConfigFile.Name())
	return kubeConfigFile.Name()
}
