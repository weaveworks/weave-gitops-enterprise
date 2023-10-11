package connector

import (
	"context"
	"testing"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/stretchr/testify/assert"
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd"
)

func TestGetSecretNameFromCluster(t *testing.T) {
	scheme := newTestScheme(t)
	dynClient := dynamicfake.NewSimpleDynamicClient(scheme, &gitopsv1alpha1.GitopsCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "spoke",
			Namespace: corev1.NamespaceDefault,
		},
		Spec: gitopsv1alpha1.GitopsClusterSpec{
			SecretRef: &meta.LocalObjectReference{
				Name: "spoke-secret",
			},
		},
	},
	)

	secretName, err := getSecretNameFromCluster(context.TODO(), dynClient, scheme, types.NamespacedName{Name: "spoke", Namespace: corev1.NamespaceDefault})
	assert.NoError(t, err)
	assert.Equal(t, "spoke-secret", secretName)
}

func TestCreateOrUpdateGitOpsClusterSecret(t *testing.T) {
	client := fake.NewSimpleClientset()
	secretName := "spoke-secret"
	opts := clientcmd.NewDefaultPathOptions()
	opts.LoadingRules.ExplicitPath = "testdata/kube-config.yaml"

	restCfg, err := configForContext(context.TODO(), opts, "spoke")
	assert.NoError(t, err)
	config, err := kubeConfigWithToken(context.TODO(), restCfg, "spoke", []byte("testing-token"))
	assert.NoError(t, err)
	configBytes, err := clientcmd.Write(*config)
	assert.NoError(t, err)

	//serialize config
	expectedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: corev1.NamespaceDefault,
		},
		Data: map[string][]byte{
			"value": configBytes,
		},
	}

	secretCreated, err := createOrUpdateGitOpsClusterSecret(context.TODO(), client, "spoke-secret", "default", config)
	assert.NoError(t, err)
	assert.Equal(t, expectedSecret, secretCreated, "Secret created not equal expected")

	secretRetrieved, err := client.CoreV1().Secrets(corev1.NamespaceDefault).Get(context.TODO(), secretName, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.Equal(t, expectedSecret, secretRetrieved, "Secret retrieved from client not equal expected")
}

func newTestScheme(t *testing.T) *runtime.Scheme {
	scheme, err := NewGitopsClusterScheme()
	assert.NoError(t, err)

	return scheme
}
