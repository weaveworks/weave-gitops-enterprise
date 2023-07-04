package collector

import (
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"testing"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"k8s.io/client-go/rest"

	. "github.com/onsi/gomega"
)

func TestWatcher_NewWatcher(t *testing.T) {
	g := NewGomegaWithT(t)
	fakeObjectsChannel := make(chan []models.ObjectTransaction)
	defer close(fakeObjectsChannel)

	errPattern := "invalid service account name"
	_, err := NewWatcher("cluster-A", &rest.Config{
		Host: "http://idontexist",
	}, configuration.SupportedObjectKinds, fakeObjectsChannel, log)
	if err != nil {
		return
	}
	g.Expect(err).To(MatchError(MatchRegexp(errPattern)))
}
