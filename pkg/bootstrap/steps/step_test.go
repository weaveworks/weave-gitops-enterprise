package steps

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/test/utils"
)

func Test_defaultInputStep(t *testing.T) {
	tests := []struct {
		name       string
		inputs     []StepInput
		config     Config
		userInputs []string
		want       []StepInput
		wantErr    assert.ErrorAssertionFunc
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
			config:     makeTestConfig(t, Config{}),
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
			config:     makeTestConfig(t, Config{}),
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
			config:     makeTestConfig(t, Config{}),
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
			config:     makeTestConfig(t, Config{}),
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
