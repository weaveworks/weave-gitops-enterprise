package bootstrap

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	. "github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/steps"
)

func Test_executeSteps(t *testing.T) {

	var buf bytes.Buffer

	config := MakeTestConfig(t, Config{
		Output: &buf,
		ClusterUserAuth: ClusterUserAuthConfig{
			Password: "password123",
		},
		ModesConfig: ModesConfig{
			Silent: true,
			Export: true,
		},
	})
	clusterUserAuthStep, err := NewAskAdminCredsSecretStep(config.ClusterUserAuth, config.ModesConfig)
	assert.NoError(t, err)

	steps := []BootstrapStep{
		clusterUserAuthStep,
	}

	t.Run("should support export", func(t *testing.T) {
		err := execute(config, steps)
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "cluster-user-auth")
	})

}
