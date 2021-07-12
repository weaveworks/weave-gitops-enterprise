package pages

import (
	"fmt"

	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

//Header webDriver elements
type TemplatesPage struct {
	TemplateHeader *agouti.Selection
	TemplateCount  *agouti.Selection
	TemplateTiles  *agouti.MultiSelection
}

// This function waits for any template tile to appear (become visible)
func WaitForAnyTemplateToAppear(webDriver *agouti.Page) {
	Eventually(webDriver.All(`[data-template-name]`)).Should(BeVisible())
}

//TemplatesPage webdriver initialises the webDriver object
func GetTemplatesPage(webDriver *agouti.Page) *TemplatesPage {
	templatesPage := TemplatesPage{
		TemplateHeader: webDriver.Find(`div[role="heading"] a[href="/clusters/templates"]`),
		TemplateCount:  webDriver.FindByXPath(`//*[@href="/clusters/templates"]/parent::div[@role="heading"]/following-sibling::div`),
		TemplateTiles:  webDriver.All(`[data-template-name]`),
	}

	return &templatesPage
}

type TemplateTile struct {
	Name           string
	Description    *agouti.Selection
	CreateTemplate *agouti.Selection
}

func GetTemplateTile(webDriver *agouti.Page, templateName string) *TemplateTile {
	tileNode := webDriver.Find(fmt.Sprintf(`[data-template-name="%s"]`, templateName))
	return &TemplateTile{
		Name:           templateName,
		Description:    tileNode.Find(`button > div.MuiCardContent-root > p`),
		CreateTemplate: tileNode.Find(`#create-cluster`),
	}
}

func (t TemplatesPage) GetTemplateTileList() []string {
	tileCount, _ := t.TemplateTiles.Count()
	titles := make([]string, tileCount)

	for i := 0; i < tileCount; i++ {
		titles[i], _ = t.TemplateTiles.At(i).Text()
	}
	return titles
}
