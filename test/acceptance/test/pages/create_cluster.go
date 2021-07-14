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
	Label   *agouti.Selection
	Field   *agouti.Selection
	ListBox *agouti.Selection
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

// scrolls the window
func ScrollWindow(webDriver *agouti.Page, xOffSet int, yOffSet int) {
	// script := fmt.Sprintf(`var elmnt = document.evaluate('%s', document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue; elmnt.scrollIntoView();`, xpath)

	script := fmt.Sprintf(`window.scrollTo(%d, %d)`, xOffSet, yOffSet)
	var result interface{}
	webDriver.RunScript(script, map[string]interface{}{}, &result)
}

// This function waits for previw and gitops to appear (become visible)
func WaitForDynamicSecToAppear(webDriver *agouti.Page) {
	Eventually(webDriver.FindByXPath(`//*/span[contains(., "Preview")]/parent::div/following-sibling::textarea`)).Should(BeFound())
	Eventually(webDriver.FindByXPath(`//*/span[text()="GitOps"]`)).Should(BeFound())
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
			Label:   fields.At(i).Find(`label`),
			Field:   fields.At(i).Find(`input`),
			ListBox: fields.At(i).Find(`div[role="button"][aria-haspopup="listbox"]`),
		})
	}

	return TemplateSection{
		Name:   name,
		Fields: formFields,
	}
}

func GetParameterOption(webDriver *agouti.Page, value string) *agouti.Selection {
	return webDriver.Find(fmt.Sprintf(`li[data-value="%s"]`, value))
}

func GetPreview(webDriver *agouti.Page) Preview {
	return Preview{
		PreviewLabel: webDriver.FindByXPath(`//*/span[text()="Preview & Commit"]`),
		PreviewText:  webDriver.FindByXPath(`//*/span[contains(., "Preview")]/parent::div/following-sibling::textarea`),
	}
}

func GetGitOps(webDriver *agouti.Page) GitOps {
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
