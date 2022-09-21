package pages

import (
	"testing"

	"github.com/sclevine/agouti"
	"github.com/weaveworks/weave-gitops-enterprise/test/selectors"
)

func TestTemplates(t *testing.T) {
	// if any of the selectors are not found, the test will fail
	selectors.SetTestContext(t)

	fakePage := agouti.JoinPage("localhost")
	GetTemplatesPage(fakePage)
}
