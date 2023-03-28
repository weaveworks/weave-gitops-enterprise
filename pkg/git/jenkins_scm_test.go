package git_test

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/git"
)

func TestJenkinsSCM_ParseURL(t *testing.T) {
	jscm := git.JenkinsSCM{}

	type testCase struct {
		caseName string
		input    string
		org      string
		project  string
		name     string
		err      error
	}

	testCases := []testCase{
		{
			caseName: "valid azure devops url",
			input:    "https://weaveworks@dev.azure.com/weaveworks/weave-gitops-integration/_git/sunglow-test-repo",
			org:      "weaveworks",
			project:  "weave-gitops-integration",
			name:     "sunglow-test-repo",
			err:      nil,
		},
		{
			caseName: "invalid azure devops url",
			input:    "https://weaveworks@dev.azure.com/weaveworks",
			org:      "",
			project:  "",
			name:     "",
			err:      errors.New("unbale to parse url https://weaveworks@dev.azure.com/weaveworks"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.caseName, func(t *testing.T) {
			u, err := url.Parse(tt.input)
			assert.NoError(t, err)

			org, project, repoName, err := jscm.ParseURL(u)
			assert.Equal(t, tt.org, org)
			assert.Equal(t, tt.project, project)
			assert.Equal(t, tt.name, repoName)
			assert.Equal(t, tt.err, err)
		})
	}
}
