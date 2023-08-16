package connector

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd"
)

func getUnstructuredObj(obj *capiv1_protos.GitopsCluster) (*unstructured.Unstructured, error) {
	jsonObj, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	unstructuredObj := &unstructured.Unstructured{}
	mapInterface := make(map[string]interface{})
	err = json.Unmarshal(jsonObj, &mapInterface)
	if err != nil {
		return nil, err
	}

	unstructuredObj.SetUnstructuredContent(mapInterface)
	return unstructuredObj, nil

}

func TestGetSecretNameFromCluster(t *testing.T) {
	namespace := "default"
	opts := clientcmd.NewDefaultPathOptions()
	opts.LoadingRules.ExplicitPath = "testdata/kube-config.yaml"
	restCfg, err := ConfigForContext(opts, "spoke")
	// clusterConfig := clientcmd.GetConfigFromFileOrDie("gitopscluster-config.yaml")
	assert.NoError(t, err)

	dynClient, err := dynamic.NewForConfig(restCfg)

	// existing gitopscluster to be added
	existingGitOpsCluster := &capiv1_protos.GitopsCluster{
		Name:      "spoke",
		Namespace: namespace,
		SecretRef: &capiv1_protos.GitopsClusterRef{
			Name: "spoke-secret",
		},
	}
	unstructuredGitopsCluster, err := getUnstructuredObj(existingGitOpsCluster)
	resource := gitopsv1alpha1.GroupVersion.WithResource("gitopscluster")
	_, err = dynClient.Resource(resource).Create(context.Background(), unstructuredGitopsCluster, metav1.CreateOptions{})
	assert.NoError(t, err)

	secretName, err := getSecretNameFromCluster(dynClient, "spoke", namespace)
	assert.NoError(t, err)
	assert.Equal(t, "spoke-secret", secretName)

}

func TestSecretWithKubeconfig(t *testing.T) {
	client := fake.NewSimpleClientset()
	secretName := "spoke-secret"
	namespace := "default"

	opts := clientcmd.NewDefaultPathOptions()
	opts.LoadingRules.ExplicitPath = "testdata/kube-config.yaml"

	restCfg, err := ConfigForContext(opts, "spoke")
	assert.NoError(t, err)
	config, err := kubeConfigWithToken(restCfg, "spoke", []byte("testing-token"))
	assert.NoError(t, err)
	configBytes, err := json.Marshal(config)
	assert.NoError(t, err)

	//serialize config
	expectedSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"value": configBytes,
		},
	}

	secretCreated, err := secretWithKubeconfig(client, "spoke-secret", "default", config)
	assert.NoError(t, err)
	assert.Equal(t, expectedSecret, secretCreated, "Secret created not equal expected")

	secretRetrieved, err := client.CoreV1().Secrets(corev1.NamespaceDefault).Get(context.Background(), secretName, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.Equal(t, expectedSecret, secretRetrieved, "Secret retrieved from client not equal expected")

}
