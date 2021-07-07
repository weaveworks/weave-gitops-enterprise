package credentials

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	capiv1_proto "github.com/weaveworks/wks/cmd/capi-server/pkg/protos"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

type IdentityParams struct {
	Group       string
	Version     string
	Kind        string
	ClusterKind string
}

var IdentityParamsList = []IdentityParams{
	// AWS
	// Version: v1alpha3
	{
		Group:       "infrastructure.cluster.x-k8s.io",
		Version:     "v1alpha3",
		Kind:        "AWSClusterStaticIdentity",
		ClusterKind: "AWSCluster",
	},
	{
		Group:       "infrastructure.cluster.x-k8s.io",
		Version:     "v1alpha3",
		Kind:        "AWSClusterRoleIdentity",
		ClusterKind: "AWSCluster",
	},
	// Version: v1alpha4
	{
		Group:       "infrastructure.cluster.x-k8s.io",
		Version:     "v1alpha4",
		Kind:        "AWSClusterStaticIdentity",
		ClusterKind: "AWSCluster",
	},
	{
		Group:       "infrastructure.cluster.x-k8s.io",
		Version:     "v1alpha4",
		Kind:        "AWSClusterRoleIdentity",
		ClusterKind: "AWSCluster",
	},
	// Azure
	// Version: v1alpha3
	{
		Group:       "infrastructure.cluster.x-k8s.io",
		Version:     "v1alpha3",
		Kind:        "AzureClusterIdentity",
		ClusterKind: "AzureCluster",
	},
	// Version: v1alpha4
	{
		Group:       "infrastructure.cluster.x-k8s.io",
		Version:     "v1alpha4",
		Kind:        "AzureClusterIdentity",
		ClusterKind: "AzureCluster",
	},
	// VSphere
	// Version: v1alpha3
	{
		Group:       "infrastructure.cluster.x-k8s.io",
		Version:     "v1alpha3",
		Kind:        "VSphereClusterIdentity",
		ClusterKind: "VSphereCluster",
	},
	// Version: v1alpha4
	{
		Group:       "infrastructure.cluster.x-k8s.io",
		Version:     "v1alpha4",
		Kind:        "VSphereClusterIdentity",
		ClusterKind: "VSphereCluster",
	},
}

func CheckAndInjectCredentials(c client.Client, tmplWithValues [][]byte, creds *capiv1_proto.Credential, tmpName string) ([]byte, error) {
	if creds == nil {
		return bytes.Join(tmplWithValues, []byte("\n---\n")), nil
	}
	exist, err := CheckCredentialsExist(c, creds)
	var tmpl [][]byte
	var result []byte
	if err != nil {
		return nil, fmt.Errorf("failed to check if credentials exist: %w", err)
	}
	if exist {
		tmpl, err = InjectCredentials(tmplWithValues, creds)
		if err != nil {
			return nil, fmt.Errorf("unable to inject credentials %q: %w", tmpName, err)
		}
		result = bytes.Join(tmpl, []byte("\n---\n"))
	} else if !exist {
		result = bytes.Join(tmplWithValues, []byte("\n---\n"))
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
		for _, identityParams := range IdentityParamsList {
			// see if we can find the capi type in the list here.
			clusterKind := identityParams.ClusterKind
			newBit, err := MaybeInjectCredentials(bit, clusterKind, creds)
			if err != nil {
				return nil, fmt.Errorf("unable to inject credentials %v %v %v: %v", creds, bit, clusterKind, err)
			}
			newBits = append(newBits, newBit)
		}
	}

	return newBits, nil
}

//
// Some good examples of kyaml at https://github.com/kubernetes-sigs/kustomize/blob/master/kyaml/yaml/example_test.go
//
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
		fmt.Printf("kind did not match! %v %v\n", kind, clusterKind)
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
