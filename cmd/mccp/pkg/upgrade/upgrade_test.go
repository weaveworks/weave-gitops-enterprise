package upgrade_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/mccp/pkg/upgrade"
)

func TestUpgrade(t *testing.T) {
	tests := []struct {
		name             string
		result           string
		entitlement      string
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:             "error returned",
			err:              errors.New("failed to get entitlement: exit status 1"),
			expectedErrorStr: "failed to get entitlement: exit status 1",
		},
		{
			name:             "error returned",
			entitlement:      "foo",
			err:              errors.New("failed to get entitlement: exit status 1"),
			expectedErrorStr: "failed to get entitlement: exit status 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := new(bytes.Buffer)
			err := upgrade.Upgrade(w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

// func createFakeClient(clusterState []runtime.Object) client.Client {
// 	scheme := runtime.NewScheme()
// 	schemeBuilder := runtime.SchemeBuilder{
// 		corev1.AddToScheme,
// 	}
// 	schemeBuilder.AddToScheme(scheme)

// 	c := fake.NewClientBuilder().
// 		WithScheme(scheme).
// 		WithRuntimeObjects(clusterState...).
// 		Build()

// 	return c
// }

// func createSecret(name string) *corev1.Secret {
// 	return &corev1.Secret{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      name,
// 			Namespace: "default",
// 		},
// 		Type: "Opaque",
// 		Data: map[string][]byte{"entitlement": []byte("foo")},
// 	}
// }
