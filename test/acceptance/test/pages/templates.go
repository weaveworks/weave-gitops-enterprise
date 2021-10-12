package pages

import (
	"fmt"

	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

//Header webDriver elements
type TemplatesPage struct {
	TemplateHeader        *agouti.Selection
	TemplateCount         *agouti.Selection
	TemplateTiles         *agouti.MultiSelection
	TemplatesTable        *agouti.MultiSelection
	TemplateProvider      *agouti.Selection
	TemplateProviderPopup *agouti.MultiSelection
	TemplateView          *agouti.MultiSelection
}

// This function waits for any template tile to appear (become visible)
func (t TemplatesPage) WaitForPageToLoad(webDriver *agouti.Page) {
	Eventually(webDriver.All(`[data-template-name]`)).Should(BeVisible())
}

//TemplatesPage webdriver initialises the webDriver object
func GetTemplatesPage(webDriver *agouti.Page) *TemplatesPage {
	templatesPage := TemplatesPage{
		TemplateHeader:        webDriver.Find(`div[role="heading"] a[href="/clusters/templates"]`),
		TemplateCount:         webDriver.FindByXPath(`//*[@href="/clusters/templates"]/parent::div[@role="heading"]/following-sibling::div`),
		TemplateTiles:         webDriver.All(`[data-template-name]`),
		TemplatesTable:        webDriver.All(`#templates-list tr[data-template-name]`),
		TemplateProvider:      webDriver.FindByID(`filter-by-provider`),
		TemplateProviderPopup: webDriver.All(`ul#filter-by-provider-popup li`),
		TemplateView:          webDriver.All(`main > div > div > div > svg`),
	}

	return &templatesPage
}

type TemplateRecord struct {
	Name             string
	Provider         *agouti.Selection
	Description      *agouti.Selection
	CreateTemplate   *agouti.Selection
	ErrorHeader      *agouti.Selection
	ErrorDescription *agouti.Selection
}

func GetTemplateTile(webDriver *agouti.Page, templateName string) *TemplateRecord {
	tileNode := webDriver.Find(fmt.Sprintf(`[data-template-name="%s"]`, templateName))
	return &TemplateRecord{
		Name:             templateName,
		Description:      tileNode.Find(`P[class^=MuiTypography-root]`),
		CreateTemplate:   tileNode.Find(`#create-cluster`),
		ErrorHeader:      tileNode.Find(`.template-error-header`),
		ErrorDescription: tileNode.Find(`.template-error-description`),
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

func GetTemplateRow(webDriver *agouti.Page, templateName string) *TemplateRecord {
	tileRow := webDriver.Find(fmt.Sprintf(`tr.summary[data-template-name="%s"]`, templateName))
	return &TemplateRecord{
		Name:             templateName,
		Provider:         tileRow.FindByXPath(`td[2]`),
		Description:      tileRow.FindByXPath(`td[3]`),
		CreateTemplate:   tileRow.FindByXPath(`td[4]`),
		ErrorHeader:      tileRow.Find(`.template-error-header`),
		ErrorDescription: tileRow.Find(`.template-error-description`),
	}
}

func (t TemplatesPage) GetTemplateTableList() []string {
	rowCount, _ := t.TemplatesTable.Count()
	rows := make([]string, rowCount)

	for i := 0; i < rowCount; i++ {
		rows[i], _ = t.TemplatesTable.At(i).Text()
	}
	return rows
}

func (t TemplatesPage) SelectProvider(providerName string) *agouti.Selection {
	pCount, _ := t.TemplateProviderPopup.Count()

	for i := 0; i < pCount; i++ {
		pName, _ := t.TemplateProviderPopup.At(i).Text()
		if providerName == pName {
			return t.TemplateProviderPopup.At(i)
		}
	}
	return nil
}

func (t TemplatesPage) SelectView(viewName string) *agouti.Selection {
	switch viewName {
	case "grid":
		return t.TemplateView.At(0)
	case "table":
		return t.TemplateView.At(1)
	}
	return nil
}
