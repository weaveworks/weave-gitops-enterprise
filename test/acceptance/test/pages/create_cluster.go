package pages

import (
	"fmt"

	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

//Header webDriver elements
type CreateCluster struct {
	CreateHeader *agouti.Selection
	// TemplateName   *agouti.Selection
	TemplateSection *agouti.MultiSelection
	PreviewPR       *agouti.Selection
}

type FormField struct {
	Label *agouti.Selection
	Field *agouti.Selection
}
type TemplateSection struct {
	Name   *agouti.Selection
	Fields []FormField
}

type Preview struct {
	PreviewLabel *agouti.Selection
	PreviewText  *agouti.Selection
}

type GitOps struct {
	GitOpsLabel  *agouti.Selection
	GitOpsFields []FormField
	CreatePR     *agouti.Selection
}

// scrolls the element into view
func ScrollIntoView(webDriver *agouti.Page, xpath string) {
	script := fmt.Sprintf(`var elmnt = document.evaluate('%s', document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue; elmnt.scrollIntoView();`, xpath)
	var result interface{}
	webDriver.RunScript(script, map[string]interface{}{"xpath": xpath}, &result)
}

func (g GitOps) ScrollTo(webDriver *agouti.Page, selection *agouti.Selection) {
	xpath := ""

	if selection == g.GitOpsLabel {
		xpath = `//*/span[text()="GitOps"]`
	}

	ScrollIntoView(webDriver, xpath)
}

func (p Preview) ScrollTo(webDriver *agouti.Page, selection *agouti.Selection) {
	xpath := ""

	if selection == p.PreviewText {
		xpath = `//*/span[contains(., "Preview")]/parent::div/following-sibling::textarea`
	}

	ScrollIntoView(webDriver, xpath)
}

//CreateCluster initialises the webDriver object
func GetCreateClusterPage(webDriver *agouti.Page) *CreateCluster {
	clusterPage := CreateCluster{
		CreateHeader: webDriver.Find(`.count-header`),
		// TemplateName:   webDriver.FindByXPath(`//*/div[text()="Create new cluster with template"]/following-sibling::text()`),
		TemplateSection: webDriver.AllByXPath(`//div[contains(@class, "form-group field field-object")]/child::div`),
		PreviewPR:       webDriver.FindByButton("Preview PR"),
	}

	return &clusterPage
}

func (c CreateCluster) GetTemplateSection(webdriver *agouti.Page, sectionName string) TemplateSection {

	name := webdriver.FindByXPath(fmt.Sprintf(`//div[contains(@class, "form-group field field-object")]/child::div//div[.="%s"]`, sectionName))
	fields := webdriver.AllByXPath(fmt.Sprintf(`//div[contains(@class, "form-group field field-object")]/child::div//div[.="%s"]/following-sibling::div[1]/div`, sectionName))

	formFields := []FormField{}
	fCnt, _ := fields.Count()

	for i := 0; i < fCnt; i++ {
		formFields = append(formFields, FormField{
			Label: fields.At(i).Find(`label`),
			Field: fields.At(i).Find(`input`),
		})
	}

	return TemplateSection{
		Name:   name,
		Fields: formFields,
	}
}

func GetPreview(webDriver *agouti.Page) Preview {
	Eventually(webDriver.FindByXPath(`//*/span[contains(., "Preview")]/parent::div/following-sibling::textarea`)).Should(BeFound())
	return Preview{
		PreviewLabel: webDriver.FindByXPath(`//*/span[text()="Preview & Commit"]`),
		PreviewText:  webDriver.FindByXPath(`//*/span[contains(., "Preview")]/parent::div/following-sibling::textarea`),
	}
}

func GetGitOps(webDriver *agouti.Page) GitOps {
	Eventually(webDriver.FindByXPath(`//*/span[text()="GitOps"]`)).Should(BeFound())
	return GitOps{
		GitOpsLabel: webDriver.FindByXPath(`//*/span[text()="GitOps"]`),
		GitOpsFields: []FormField{
			{
				Label: webDriver.FindByLabel(`Create branch`),
				Field: webDriver.FindByID(`Create branch-input`),
			},
			{
				Label: webDriver.FindByLabel(`Title pull request`),
				Field: webDriver.FindByID(`Title pull request-input`),
			},
			{
				Label: webDriver.FindByLabel(`Commit message`),
				Field: webDriver.FindByID(`Commit message-input`),
			},
		},
		CreatePR: webDriver.FindByButton(`Create Pull Request`),
	}
}
