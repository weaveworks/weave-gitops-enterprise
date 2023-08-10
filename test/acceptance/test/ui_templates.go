package acceptance

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

type TemplateField struct {
	Name   string
	Value  string
	Option string
}

func setParameterValues(createPage *pages.CreateCluster, parameters []TemplateField) {
	for i := 0; i < len(parameters); i++ {
		if parameters[i].Option != "" {
			gomega.Eventually(func(g gomega.Gomega) {
				g.Eventually(createPage.GetTemplateParameter(webDriver, parameters[i].Name).ListBox.Click).Should(gomega.Succeed())
				g.Eventually(pages.GetOption(webDriver, parameters[i].Option).Click).Should(gomega.Succeed())
				g.Expect(createPage.GetTemplateParameter(webDriver, parameters[i].Name).ListBox).Should(matchers.MatchText(parameters[i].Option))
			}, ASSERTION_30SECONDS_TIME_OUT).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to select %s parameter option: %s", parameters[i].Name, parameters[i].Option))
		} else {
			field := createPage.GetTemplateParameter(webDriver, parameters[i].Name).Field
			pages.ClearFieldValue(field)
			gomega.Expect(field.SendKeys(parameters[i].Value)).To(gomega.Succeed())
		}
	}
}

func installGitOpsTemplate(templateFiles map[string]string) {
	ginkgo.By("Installing GitOpsTemplate...", func() {
		for _, templateFile := range templateFiles {
			err := runCommandPassThrough("kubectl", "apply", "-f", templateFile)
			gomega.Expect(err).To(gomega.BeNil(), fmt.Sprintf("Failed to apply GitOpsTemplate template %s", templateFile))
		}
	})
}

func waitForTemplatesToAppear(templateCount int) {
	ginkgo.By("And wait for Templates to be rendered", func() {
		templatesPage := pages.GetTemplatesPage(webDriver)
		gomega.Eventually(func(g gomega.Gomega) {
			g.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
			pages.WaitForPageToLoad(webDriver)
			g.Eventually(templatesPage.TemplateHeader).Should(matchers.BeVisible())
			g.Eventually(templatesPage.CountTemplateRows).Should(gomega.BeNumerically(">=", templateCount))
		}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).ShouldNot(gomega.HaveOccurred(), "The number of template rows should be equal to number of templates created")
	})
}
