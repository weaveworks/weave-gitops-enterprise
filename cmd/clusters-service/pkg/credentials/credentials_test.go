package credentials

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	capiv1 "github.com/weaveworks/templates-controller/apis/capi/v1alpha2"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	fakediscovery "k8s.io/client-go/discovery/fake"
	k8stesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestMaybeInjectCredentials(t *testing.T) {
	result, _ := MaybeInjectCredentials(nil, "", nil)
	if diff := cmp.Diff(string(result), ""); diff != "" {
		t.Fatalf("result wasn't nil! %v", diff)
	}

	// Wrong kind
	templateBit := `apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4
kind: AWSCluster
`
	result, _ = MaybeInjectCredentials([]byte(templateBit), "FooKind", nil)
	if diff := cmp.Diff(templateBit, string(result)); diff != "" {
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

	if diff := cmp.Diff(expected, string(result)); diff != "" {
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

	c := newFakeClient(t)
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

	for _, clusterKind := range []string{"AWSCluster", "AWSManagedControlPlane"} {
		t.Run(clusterKind, func(t *testing.T) {
			templateBits := [][]byte{
				[]byte(fmt.Sprintf(`apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4
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
			if diff := cmp.Diff(expected, resultStr[0]); diff != "" {
				t.Fatalf("expected didn't match result! %v", diff)
			}
		})
	}
}

func TestInjectCredentials_ignores_types(t *testing.T) {
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

	for _, clusterKind := range []string{"AWSManagedCluster"} {
		t.Run(clusterKind, func(t *testing.T) {
			templateBits := [][]byte{
				[]byte(fmt.Sprintf(`apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4
kind: %s
spec: {}
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
spec: {}
`, clusterKind)
			if diff := cmp.Diff(expected, resultStr[0]); diff != "" {
				t.Fatalf("expected didn't match result! %v", diff)
			}
		})
	}
}

func TestFindCredentials(t *testing.T) {
	apiResources := []*metav1.APIResourceList{
		{
			GroupVersion: "infrastructure.cluster.x-k8s.io/v1alpha4",
			APIResources: []metav1.APIResource{
				{Name: "awsclusterroleidentities", SingularName: "awsclusterroleidentity", Kind: "AWSClusterRoleIdentity", Namespaced: true},
				{Name: "azureclusteridentities", SingularName: "azureclusteridentity", Kind: "AzureClusterIdentity", Namespaced: true},
			},
		},
	}
	fakeDiscovery := &fakediscovery.FakeDiscovery{Fake: &k8stesting.Fake{Resources: apiResources}}

	findTests := []struct {
		name        string
		clusterObjs []runtime.Object
		want        []unstructured.Unstructured
	}{
		{
			"no credentials",
			[]runtime.Object{},
			[]unstructured.Unstructured{},
		},
		{
			"single credential",
			[]runtime.Object{newUnstructured("infrastructure.cluster.x-k8s.io/v1alpha4", "AWSClusterRoleIdentity", "test-creds", "test-ns", "uid1")},
			[]unstructured.Unstructured{
				*newUnstructured("infrastructure.cluster.x-k8s.io/v1alpha4", "AWSClusterRoleIdentity", "test-creds", "test-ns", "uid1")},
		},
		{
			"multi credentials",
			[]runtime.Object{
				newUnstructured("infrastructure.cluster.x-k8s.io/v1alpha4", "AWSClusterRoleIdentity", "test-creds1", "test-ns", "uid1"),
				newUnstructured("infrastructure.cluster.x-k8s.io/v1alpha4", "AWSClusterRoleIdentity", "test-creds2", "test-ns", "uid2"),
			},
			[]unstructured.Unstructured{
				*newUnstructured("infrastructure.cluster.x-k8s.io/v1alpha4", "AWSClusterRoleIdentity", "test-creds1", "test-ns", "uid1"),
				*newUnstructured("infrastructure.cluster.x-k8s.io/v1alpha4", "AWSClusterRoleIdentity", "test-creds2", "test-ns", "uid2"),
			},
		},
		{
			"multi credentials - returned for different versions",
			[]runtime.Object{
				newUnstructured("infrastructure.cluster.x-k8s.io/v1alpha3", "AWSClusterRoleIdentity", "test-creds1", "test-ns", "uid1"),
				newUnstructured("infrastructure.cluster.x-k8s.io/v1alpha4", "AWSClusterRoleIdentity", "test-creds1", "test-ns", "uid1"),
			},
			[]unstructured.Unstructured{
				*newUnstructured("infrastructure.cluster.x-k8s.io/v1alpha4", "AWSClusterRoleIdentity", "test-creds1", "test-ns", "uid1"),
			},
		},
		{
			"multi different kind & identities",
			[]runtime.Object{
				newUnstructured("infrastructure.cluster.x-k8s.io/v1alpha4", "AzureClusterIdentity", "test-creds1", "test-ns", "uid1"),
				newUnstructured("infrastructure.cluster.x-k8s.io/v1alpha4", "AWSClusterRoleIdentity", "test-creds2", "test-ns", "uid2"),
			},
			[]unstructured.Unstructured{
				*newUnstructured("infrastructure.cluster.x-k8s.io/v1alpha4", "AzureClusterIdentity", "test-creds1", "test-ns", "uid1"),
				*newUnstructured("infrastructure.cluster.x-k8s.io/v1alpha4", "AWSClusterRoleIdentity", "test-creds2", "test-ns", "uid2"),
			},
		},
	}

	for _, tt := range findTests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := newFakeClient(t, tt.clusterObjs...)

			found, err := FindCredentials(context.TODO(), fakeClient, fakeDiscovery)
			if err != nil {
				t.Fatal(err)
			}

			credSorter := func(a, b unstructured.Unstructured) bool {
				return strings.Compare(string(a.GetUID()), string(b.GetUID())) < 0
			}
			resourceVersion := func(k string, _ interface{}) bool {
				return k == "resourceVersion"
			}
			if diff := cmp.Diff(tt.want, found, cmpopts.SortSlices(credSorter), cmpopts.IgnoreMapEntries(resourceVersion)); diff != "" {
				t.Fatalf("FindCredentials() failed:\n%s", diff)
			}
		})
	}
}

func newFakeClient(t *testing.T, objs ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		capiv1.AddToScheme,
	}
	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		t.Fatal(err)
	}

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objs...).
		Build()
}

func newUnstructured(apiVersion, kind, name, namespace, uid string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetName(name)
	u.SetNamespace(namespace)
	u.SetKind(kind)
	u.SetAPIVersion(apiVersion)
	u.SetUID(types.UID(uid))
	return u
}

func convertToStringArray(in [][]byte) []string {
	var result []string
	for _, i := range in {
		result = append(result, string(i))
	}
	return result
}
