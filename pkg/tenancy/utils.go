package tenancy

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func getObjectID(obj client.Object) string {
	objectID := fmt.Sprintf("%s/%s", strings.ToLower(obj.GetObjectKind().GroupVersionKind().GroupKind().String()), obj.GetName())
	if obj.GetNamespace() != "" {
		objectID = fmt.Sprintf("%s/%s", obj.GetNamespace(), objectID)
	}
	return objectID
}

func marshalOutput(out io.Writer, output runtime.Object) error {
	data, err := yaml.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	_, err = fmt.Fprintf(out, "%s", data)
	if err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}

	return nil
}

func outputResources(out io.Writer, resources []client.Object) error {
	for _, v := range resources {
		if err := marshalOutput(out, v); err != nil {
			return fmt.Errorf("failed outputting tenant: %w", err)
		}

		if _, err := out.Write([]byte("---\n")); err != nil {
			return err
		}
	}

	return nil
}

func typeMeta(kind, apiVersion string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       kind,
		APIVersion: apiVersion,
	}
}

func runtimeObjectFromObject(o client.Object) client.Object {
	return reflect.New(reflect.TypeOf(o).Elem()).Interface().(client.Object)
}
