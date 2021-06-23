package pages

import (
	"fmt"

	"github.com/sclevine/agouti"
)

//Header webDriver elements
type TemplatesPage struct {
	TemplateHeader *agouti.Selection
	TemplateCount  *agouti.Selection
	TemplateTiles  *agouti.MultiSelection
}

//TemplatesPage webdriver initialises the webDriver object
func GetTemplatesPage(webDriver *agouti.Page) *TemplatesPage {
	templatesPage := TemplatesPage{
		TemplateHeader: webDriver.Find(`div[role="heading"] a[href="/clusters/templates"]`),
		TemplateCount:  webDriver.FindByXPath(`//*[@href="/clusters/templates"]/parent::div[@role="heading"]/following-sibling::div`),
		TemplateTiles:  webDriver.All(`.MuiPaper-root.MuiCard-root`),
	}

	return &templatesPage
}

type TemplateTile struct {
	Name           string
	Description    *agouti.Selection
	CreateTemplate *agouti.Selection
}

// Find required template tile in TemplateTiles
func GetTemplateTile(webDriver *agouti.Page, templateName string) *TemplateTile {
	tileNode := webDriver.Find(fmt.Sprintf(`[data-template-name="%s"]`, templateName))
	return &TemplateTile{
		Name:           templateName,
		Description:    tileNode.Find(`button > div.MuiCardContent-root > p`),
		CreateTemplate: tileNode.Find(`#create-cluster`),
	}
}
