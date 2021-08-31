package pages

import (
	"fmt"
	"time"

	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

//Header webDriver elements
type CreateCluster struct {
	CreateHeader *agouti.Selection
	// TemplateName   *agouti.Selection
	Credentials     *agouti.Selection
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
	ErrorBar     *agouti.Selection
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
	Eventually(webDriver.FindByXPath(`//div[contains(., "Preview")]/following-sibling::textarea`)).Should(BeFound())
	Eventually(webDriver.FindByXPath(`//div[contains(., "Preview")]/parent::div/following-sibling::div/div[text()="GitOps"]`)).Should(BeFound())
}

//CreateCluster initialises the webDriver object
func GetCreateClusterPage(webDriver *agouti.Page) *CreateCluster {
	clusterPage := CreateCluster{
		CreateHeader: webDriver.Find(`.count-header`),
		// TemplateName:   webDriver.FindByXPath(`//*/div[text()="Create new cluster with template"]/following-sibling::text()`),
		Credentials:     webDriver.FindByXPath(`//div[@class="credentials"]//div[contains(@class, "dropdown-toggle")]`),
		TemplateSection: webDriver.AllByXPath(`//div[contains(@class, "form-group field field-object")]/child::div`),
		PreviewPR:       webDriver.FindByButton("Preview PR"),
	}

	return &clusterPage
}

// This function waits for Create emplate page to load completely
func (c CreateCluster) WaitForPageToLoad(webDriver *agouti.Page) {
	// Credentials dropdown takes a while to populate
	Eventually(webDriver.FindByXPath(`//div[@class="credentials"]//div[contains(@class, "dropdown-toggle")][@disabled]`),
		30*time.Second).ShouldNot(BeFound())
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
func GetCredentials(webDriver *agouti.Page) *agouti.MultiSelection {
	return webDriver.All(`div.dropdown-item`)
}

func GetCredential(webDriver *agouti.Page, value string) *agouti.Selection {
	return webDriver.Find(fmt.Sprintf(`div.dropdown-item[title*="%s"]`, value))
}

func GetParameterOption(webDriver *agouti.Page, value string) *agouti.Selection {
	return webDriver.Find(fmt.Sprintf(`li[data-value="%s"]`, value))
}

func GetPreview(webDriver *agouti.Page) Preview {
	return Preview{
		PreviewLabel: webDriver.FindByXPath(`//div[text()="Preview"]`),
		PreviewText:  webDriver.FindByXPath(`//div[contains(., "Preview")]/following-sibling::textarea`),
	}
}

func GetGitOps(webDriver *agouti.Page) GitOps {
	return GitOps{
		GitOpsLabel: webDriver.FindByXPath(`//div[contains(., "Preview")]/parent::div/following-sibling::div/div[text()="GitOps"]`),
		GitOpsFields: []FormField{
			{
				Label: webDriver.FindByLabel(`Create branch`),
				Field: webDriver.FindByID(`Create branch-input`),
			},
			{
				Label: webDriver.FindByLabel(`Pull request title`),
				Field: webDriver.FindByID(`Pull request title-input`),
			},
			{
				Label: webDriver.FindByLabel(`Commit message`),
				Field: webDriver.FindByID(`Commit message-input`),
			},
		},
		CreatePR: webDriver.FindByButton(`Create Pull Request`),
		ErrorBar: webDriver.FindByXPath(`//div[3]/div[3]/div/div`),
	}
}


