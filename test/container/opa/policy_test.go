package opa

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/wksctl/pkg/plan"
	"github.com/weaveworks/wksctl/pkg/plan/resource"
	"github.com/weaveworks/wksctl/pkg/utilities/object"
	"github.com/weaveworks/wksctl/test/container/images"
	"github.com/weaveworks/wksctl/test/container/testutils"
)

var sshPort = testutils.PortAllocator{Next: 2422}

func NewRunnerForTest(t *testing.T, image string) (*testutils.TestRunner, func()) {
	return testutils.MakeFootlooseTestRunner(t, image, sshPort.Allocate())
}

func TestPolicy(t *testing.T) {
	r, closer := NewRunnerForTest(t, images.CentOS7)
	defer closer()

	policyPath := "../../../pkg/opa/policy/rego/cac_check.rego"
	policyTestPath := "../../../pkg/opa/cac_check_test.rego"
	policyFile := &resource.File{
		Source:      policyPath,
		Destination: "/tmp/policy.rego",
	}
	policyTestFile := &resource.File{
		Source:      policyTestPath,
		Destination: "/tmp/policy_test.rego",
	}
	emptyDiff := plan.EmptyDiff()
	_, err := policyFile.Apply(r, emptyDiff)
	assert.NoError(t, err)
	_, err = policyTestFile.Apply(r, emptyDiff)
	assert.NoError(t, err)
	run := &resource.Run{
		Script: object.String("curl -L -o /opa https://github.com/open-policy-agent/opa/releases/download/v0.10.7/opa_linux_amd64"),
	}
	_, err = run.Apply(r, emptyDiff)
	assert.NoError(t, err)
	run = &resource.Run{
		Script: object.String("chmod 755 /opa"),
	}
	_, err = run.Apply(r, emptyDiff)
	assert.NoError(t, err)
	var result string
	run = &resource.Run{
		Script: object.String("/opa test /tmp/policy_test.rego /tmp/policy.rego"),
		Output: &result,
	}
	_, err = run.Apply(r, emptyDiff)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(result, "PASS:"))
}
