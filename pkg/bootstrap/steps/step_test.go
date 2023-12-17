package steps

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	bootstrap_utils "github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"github.com/weaveworks/weave-gitops-enterprise/test/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_defaultInputStep(t *testing.T) {
	tests := []struct {
		name       string
		inputs     []StepInput
		config     Config
		userInputs []string
		want       []StepInput
	}{
		{
			name: "should return introduced value if does not exist",
			inputs: []StepInput{
				{
					Name:         "cluster_user_password",
					Msg:          "introduce cluster user password",
					Type:         stringInput,
					DefaultValue: "admin123",
				},
			},
			userInputs: []string{"admin124\n"},
			config:     MakeTestConfig(t, Config{}),
			want: []StepInput{
				{
					Name:         "cluster_user_password",
					Msg:          "introduce cluster user password",
					Type:         stringInput,
					DefaultValue: "admin123",
					Value:        "admin124",
				},
			},
		},
		{
			name: "should return nothing if exists and dont want to update",
			inputs: []StepInput{
				{
					Name:      "cluster_user_password",
					Msg:       "introduce cluster user password",
					IsUpdate:  true,
					UpdateMsg: "cluster user found. do you want to update",
				},
			},
			userInputs: []string{"n\n"},
			config:     MakeTestConfig(t, Config{}),
			want:       []StepInput{},
		},
		{
			name: "should return nothing if exists and udpate not supported",
			inputs: []StepInput{
				{
					Name:          "cluster_user_password",
					Msg:           "introduce cluster user password",
					IsUpdate:      true,
					SupportUpdate: false,
					UpdateMsg:     "cluster user found. do you want to update",
				},
			},
			userInputs: []string{"n\n"},
			config:     MakeTestConfig(t, Config{}),
			want:       []StepInput{},
		},
		{
			name: "should return new value if exists and want to update",
			inputs: []StepInput{
				{
					Name:          "cluster_user_password",
					Msg:           "introduce cluster user password",
					Type:          stringInput,
					IsUpdate:      true,
					SupportUpdate: true,
					DefaultValue:  "admin123",
					UpdateMsg:     "cluster user found. do you want to update",
				},
			},
			userInputs: []string{"y\n", "admin124\n"},
			config:     MakeTestConfig(t, Config{}),
			want: []StepInput{
				{
					Name:          "cluster_user_password",
					Msg:           "introduce cluster user password",
					Type:          stringInput,
					IsUpdate:      true,
					SupportUpdate: true,
					DefaultValue:  "admin123",
					Value:         "admin124",
					UpdateMsg:     "cluster user found. do you want to update",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdin := &utils.MockReader{Inputs: tt.userInputs}
			defer stdin.Close()

			got, err := defaultInputStep(tt.inputs, &tt.config, stdin)
			assert.NoError(t, err)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("not expected inputs:\n%s", diff)
			}
		})
	}
}

func Test_defaultOutputStep(t *testing.T) {
	tests := []struct {
		name       string
		outputs    []StepOutput
		config     Config
		assertFunc func(t *testing.T, outputs []StepOutput, config Config)
	}{
		{
			name: "should not create secret in cluster in export mode",
			outputs: []StepOutput{
				{
					Name: adminSecretName,
					Type: typeSecret,
					Value: v1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      adminSecretName,
							Namespace: WGEDefaultNamespace,
						},
						Data: map[string][]byte{
							"username": []byte(defaultAdminUsername),
						},
					},
				},
			},
			config: MakeTestConfig(t, Config{
				ModesConfig: ModesConfig{
					Export: true,
				},
			}),
			assertFunc: func(t *testing.T, outputs []StepOutput, config Config) {
				secret := outputs[0].Value.(v1.Secret)
				gotSecret, err := bootstrap_utils.GetSecret(config.KubernetesClient, secret.Name, secret.Namespace)
				assert.Error(t, err, "expected error")
				assert.Nil(t, gotSecret, "not expected secret")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := defaultOutputStep(tt.outputs, &tt.config)
			assert.NoError(t, err)
			tt.assertFunc(t, tt.outputs, tt.config)
		})
	}
}

var wantSecret = `---
apiVersion: v1
data:
  password: dGVzdC1wYXNzd29yZA==
  username: YWRtaW4=
kind: Secret
metadata:
  name: cluster-user-auth
  namespace: flux-system
type: Opaque
`

var wantHelmRepository = `---
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: weave-gitops-enterprise-charts
  namespace: flux-system
spec:
  interval: 1m0s
  secretRef:
    name: weave-gitops-enterprise-credentials
  url: https://charts.dev.wkp.weave.works/releases/charts-v3
`

func TestStepOutput_Export(t *testing.T) {
	tests := []struct {
		name    string
		output  StepOutput
		wantErr assert.ErrorAssertionFunc
		want    string
	}{
		{
			name: "can export secret",
			output: StepOutput{
				Name: adminSecretName,
				Type: typeSecret,
				Value: v1.Secret{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Secret",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster-user-auth",
						Namespace: WGEDefaultNamespace,
					},
					Type: "Opaque",
					Data: map[string][]byte{
						"username": []byte("admin"),
						"password": []byte("test-password"),
					},
				},
			},
			want: wantSecret,
		},
		{
			name: "can export file",
			output: StepOutput{
				Name: wgeHelmrepoFileName,
				Type: typeFile,
				Value: fileContent{
					Name:      wgeHelmrepoFileName,
					Content:   expectedHelmRepository,
					CommitMsg: wgeHelmRepoCommitMsg,
				},
			},
			want: wantHelmRepository,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := tt.output.Export(&buf)
			if tt.wantErr != nil {
				tt.wantErr(t, err, "error on export")
			}
			assert.NoError(t, err)
			got := buf.String()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("unexpected resource:\n%s", diff)
			}
		})
	}
}
