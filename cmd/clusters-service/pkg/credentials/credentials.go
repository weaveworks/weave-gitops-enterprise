package credentials

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"

	// load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

var identityParamsList = []struct {
	Group        string
	Versions     []string
	Kind         string
	ClusterKinds []string
}{

	{
		Group:        "infrastructure.cluster.x-k8s.io",
		Versions:     []string{"v1alpha3", "v1alpha4", "v1beta1", "v1beta2"},
		Kind:         "AWSClusterStaticIdentity",
		ClusterKinds: []string{"AWSCluster", "AWSManagedControlPlane"},
	},
	{
		Group:        "infrastructure.cluster.x-k8s.io",
		Versions:     []string{"v1alpha3", "v1alpha4", "v1beta1"},
		Kind:         "AWSClusterRoleIdentity",
		ClusterKinds: []string{"AWSCluster", "AWSManagedControlPlane"},
	},
	{
		Group:        "infrastructure.cluster.x-k8s.io",
		Versions:     []string{"v1alpha3", "v1alpha4", "v1beta1"},
		Kind:         "AzureClusterIdentity",
		ClusterKinds: []string{"AzureCluster", "AzureManagedControlPlane"},
	},
	{
		Group:        "infrastructure.cluster.x-k8s.io",
		Versions:     []string{"v1alpha3", "v1alpha4", "v1beta1"},
		Kind:         "VSphereClusterIdentity",
		ClusterKinds: []string{"VSphereCluster"},
	},
}

func isEmptyCredentials(creds *capiv1_proto.Credential) bool {
	return creds.Name == "" &&
		creds.Namespace == "" &&
		creds.Kind == "" &&
		creds.Version == "" &&
		creds.Group == ""
}

// FindCredentials returns all the custom resources in the cluster that we think are CAPI identities. What we "think" are
// capi identities are hardcoded in this file in `identityParamsList`
func FindCredentials(ctx context.Context, c client.Client, dc discovery.DiscoveryInterface) ([]unstructured.Unstructured, error) {
	identities := []unstructured.Unstructured{}
	for _, identityParams := range identityParamsList {
		for _, v := range identityParams.Versions {
			gvk := schema.GroupVersionKind{
				Group:   identityParams.Group,
				Version: v,
				Kind:    identityParams.Kind,
			}

			// We can skip this checkCRDExists check and let k8s do it.
			// BUT if any of the above Identities are missing, client-go will purge its
			// CRD cache and try and find all the available CRDs again, for each missing identity.
			// This is a lot of requests, they get throttled, this func blows out to 10s+.
			//
			exists, err := checkCRDExists(dc, gvk)
			if err != nil {
				return nil, fmt.Errorf("failed to check if CRD exists, %v: %w", gvk, err)
			}
			if !exists {
				continue
			}

			identityList := &unstructured.UnstructuredList{}
			identityList.SetGroupVersionKind(gvk)
			err = c.List(context.Background(), identityList)
			if err != nil {
				return nil, fmt.Errorf("failed to list CRs of %v: %w", gvk, err)
			}
			identities = append(identities, identityList.Items...)
		}
	}

	// k8s doesn't internally differentiate between different apiVersions so we de-dup them
	// https://github.com/kubernetes/kubernetes/issues/58131#issuecomment-403829566
	//
	identityIndex := map[types.UID]unstructured.Unstructured{}
	for _, ident := range identities {
		identityIndex[ident.GetUID()] = ident
	}
	uniqueIdentities := []unstructured.Unstructured{}
	for _, v := range identityIndex {
		uniqueIdentities = append(uniqueIdentities, v)
	}

	return uniqueIdentities, nil
}

func checkCRDExists(dc discovery.DiscoveryInterface, gvk schema.GroupVersionKind) (bool, error) {
	gv := gvk.GroupVersion().String()
	apiResources, err := dc.ServerResourcesForGroupVersion(gv)
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to get resources for GV %v: %w", gv, err)
	}

	availableKindsIndex := map[string]bool{}
	for _, api := range apiResources.APIResources {
		availableKindsIndex[api.Kind] = true
	}
	_, ok := availableKindsIndex[gvk.Kind]

	return ok, nil
}

func CheckAndInjectCredentials(log logr.Logger, c client.Client, tmplWithValues [][]byte, creds *capiv1_proto.Credential, tmpName string) ([][]byte, error) {
	if creds == nil || isEmptyCredentials(creds) {
		log.Info("No credentials were passed or credentials are empty", "credentials", creds)
		return tmplWithValues, nil
	}
	exist, err := CheckCredentialsExist(c, creds)
	var result [][]byte
	if err != nil {
		return nil, fmt.Errorf("failed to check if credentials exist: %w", err)
	}
	if exist {
		log.Info("Found credentials", "credentials", creds)
		tmpl, err := InjectCredentials(tmplWithValues, creds)
		if err != nil {
			return nil, fmt.Errorf("unable to inject credentials %q: %w", tmpName, err)
		}
		result = append(result, tmpl...)
	} else if !exist {
		log.Info("Couldn't find credentials!", "credentials", creds)
		result = append(result, tmplWithValues...)
	}

	return result, nil
}

func CheckCredentialsExist(c client.Client, creds *capiv1_proto.Credential) (bool, error) {
	identity := &unstructured.Unstructured{}
	identity.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   creds.Group,
		Kind:    creds.Kind,
		Version: creds.Version,
	})
	err := c.Get(context.Background(), client.ObjectKey{
		Namespace: creds.Namespace,
		Name:      creds.Name,
	}, identity)

	if errors.IsNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func InjectCredentials(tmplWithValues [][]byte, creds *capiv1_proto.Credential) ([][]byte, error) {
	// Check if credentials are empty, if empty skip
	if creds == nil {
		return tmplWithValues, nil
	}

	newBits := [][]byte{}
	for _, bit := range tmplWithValues {
		var err error
		for _, identityParams := range identityParamsList {
			for _, v := range identityParams.Versions {
				// see if we can find the capi type in the list here.
				if creds.Group == identityParams.Group && creds.Kind == identityParams.Kind && creds.Version == v {
					for _, clusterKind := range identityParams.ClusterKinds {
						bit, err = MaybeInjectCredentials(bit, clusterKind, creds)
						if err != nil {
							return nil, fmt.Errorf("unable to inject credentials %v %v %v: %v", creds, bit, clusterKind, err)
						}
					}
				}
			}
		}
		newBits = append(newBits, bit)
	}

	return newBits, nil
}

// Some good examples of kyaml at https://github.com/kubernetes-sigs/kustomize/blob/master/kyaml/yaml/example_test.go
func MaybeInjectCredentials(templateObject []byte, clusterKind string, creds *capiv1_proto.Credential) ([]byte, error) {
	obj, err := kyaml.Parse(string(templateObject))
	if err != nil {
		return nil, fmt.Errorf("unable to parse object: %v %v", string(templateObject), err)
	}
	kindNode, err := obj.Pipe(kyaml.Get("kind"))
	if err != nil {
		return nil, fmt.Errorf("failed to get kind: %v", err)
	}
	kind, err := kindNode.String()
	if err != nil {
		return nil, fmt.Errorf("failed to stringify kind: %v", err)
	}
	if strings.TrimSpace(kind) != clusterKind {
		return templateObject, nil
	}

	credsNode, err := kyaml.Parse(fmt.Sprintf(`
kind: %s
name: %s
`, creds.Kind, creds.Name))
	if err != nil {
		return nil, fmt.Errorf("failed to create creds node: %v", err)
	}
	err = obj.PipeE(
		kyaml.LookupCreate(kyaml.MappingNode, "spec"),
		kyaml.SetField("identityRef", credsNode))
	if err != nil {
		return nil, fmt.Errorf("failed to add identityRef node: %v", err)
	}

	newTemplateObject, err := obj.String()
	if err != nil {
		return nil, fmt.Errorf("failed to remarshall obj: %v", err)
	}

	return []byte(newTemplateObject), nil
}
