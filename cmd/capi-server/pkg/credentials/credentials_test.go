package credentials

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/api/v1alpha1"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/pkg/protos"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestMaybeInjectCredentials(t *testing.T) {
	result, _ := MaybeInjectCredentials(nil, "", nil)
	if diff := cmp.Diff(string(result), ""); diff != "" {
		t.Fatalf("result wasn't nil! %v", diff)
	}

	// Wrong kind
	templateBit := `
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4
kind: AWSCluster
`
	result, _ = MaybeInjectCredentials([]byte(templateBit), "FooKind", nil)
	if diff := cmp.Diff(string(result), templateBit); diff != "" {
		t.Fatalf("expected didn't match result! %v", diff)
	}

	// Right kind !
	expected := `apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4
kind: AWSCluster
spec:
  identityRef:
    kind: FooKind
    name: FooName
`
	result, _ = MaybeInjectCredentials([]byte(templateBit), "AWSCluster", &capiv1_protos.Credential{
		Kind: "FooKind",
		Name: "FooName",
	})

	if diff := cmp.Diff(string(result), expected); diff != "" {
		t.Fatalf("expected didn't match result! %v", diff)
	}

}

func TestCheckCredentialsExist(t *testing.T) {
	u := &unstructured.Unstructured{}
	u.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "test",
			"namespace": "test",
		},
		"spec": map[string]interface{}{
			"identityRef": map[string]interface{}{
				"kind": "FooKind",
				"name": "FooName",
			},
		},
	}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "AWSCluster",
		Version: "v1alpha4",
	})

	c := createFakeClient()
	_ = c.Create(context.Background(), u)

	creds := &capiv1_protos.Credential{
		Group:     "WrongGroup",
		Kind:      "WrongKind",
		Version:   "WrongVersion",
		Namespace: "WrongNamespace",
		Name:      "WrongName",
	}
	exist, err := CheckCredentialsExist(c, creds)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	if exist {
		t.Fatalf("Found credentials when they shouldn't exist: %v", creds)
	}

	creds = &capiv1_protos.Credential{
		Group:     "infrastructure.cluster.x-k8s.io",
		Kind:      "AWSCluster",
		Version:   "v1alpha4",
		Namespace: "test",
		Name:      "test",
	}
	exist, err = CheckCredentialsExist(c, creds)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	if !exist {
		t.Fatalf("Couldn't find credentials when they should exist: %v", creds)
	}
}

func TestInjectCredentials(t *testing.T) {
	result, _ := InjectCredentials(nil, nil)
	if diff := cmp.Diff(result, [][]uint8(nil)); diff != "" {
		t.Fatalf("result wasn't nil! %v", diff)
	}

	templateBits := [][]byte{
		[]byte(`
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4
kind: AWSCluster
`),
	}

	// no credentials
	result, _ = InjectCredentials(templateBits, nil)
	resultStr := convertToStringArray(result)
	if diff := cmp.Diff(resultStr[0], string(templateBits[0])); diff != "" {
		t.Fatalf("expected didn't match result! %v", diff)
	}

	for _, clusterKind := range []string{"AWSCluster", "AWSManagedCluster"} {
		t.Run(clusterKind, func(t *testing.T) {
			templateBits := [][]byte{
				[]byte(fmt.Sprintf(`
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4
kind: %s
`, clusterKind)),
			}

			// with creds
			result, err := InjectCredentials(templateBits, &capiv1_protos.Credential{
				Group:   "infrastructure.cluster.x-k8s.io",
				Version: "v1alpha4",
				Kind:    "AWSClusterStaticIdentity",
				Name:    "FooName",
			})
			if err != nil {
				t.Fatalf("unexpected err %v", err)
			}
			resultStr = convertToStringArray(result)

			expected := fmt.Sprintf(`apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4
kind: %s
spec:
  identityRef:
    kind: AWSClusterStaticIdentity
    name: FooName
`, clusterKind)
			if diff := cmp.Diff(resultStr[0], expected); diff != "" {
				t.Fatalf("expected didn't match result! %v", diff)
			}
		})
	}
}

func convertToStringArray(in [][]byte) []string {
	var result []string
	for _, i := range in {
		result = append(result, string(i))
	}
	return result
}

func createFakeClient() client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		capiv1.AddToScheme,
	}
	schemeBuilder.AddToScheme(scheme)

	return fake.NewClientBuilder().
		WithScheme(scheme).
		Build()
}
