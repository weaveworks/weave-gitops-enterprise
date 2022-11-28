package pages

import (
	"fmt"

	"github.com/sclevine/agouti"
)

// Header webDriver elements
type TemplatesPage struct {
	TemplateHeader        *agouti.Selection
	TemplateTiles         *agouti.MultiSelection
	TemplatesList         *agouti.MultiSelection
	TemplateProvider      *agouti.Selection
	TemplateProviderPopup *agouti.MultiSelection
	TemplateView          *agouti.MultiSelection
}

// TemplatesPage webdriver initialises the webDriver object
func GetTemplatesPage(webDriver *agouti.Page) *TemplatesPage {
	templatesPage := TemplatesPage{
		TemplateHeader:        webDriver.Find(`div[role="heading"] a[href="/templates"]`),
		TemplateTiles:         webDriver.All(`[data-template-name]`),
		TemplatesList:         webDriver.All(`#templates-list tbody tr`),
		TemplateProvider:      webDriver.FindByID(`filter-by-provider`),
		TemplateProviderPopup: webDriver.All(`ul#filter-by-provider-popup li`),
		TemplateView:          webDriver.All(`#display-action > svg`),
	}

	return &templatesPage
}

type TemplateInformation struct {
	Name             string
	Type             *agouti.Selection
	Namespace        *agouti.Selection
	Provider         *agouti.Selection
	Description      *agouti.Selection
	CreateTemplate   *agouti.Selection
	ErrorHeader      *agouti.Selection
	ErrorDescription *agouti.Selection
}

func GetTemplateTile(webDriver *agouti.Page, templateName string) *TemplateInformation {
	tileNode := webDriver.Find(fmt.Sprintf(`[data-template-name="%s"]`, templateName))
	return &TemplateInformation{
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

func (t TemplatesPage) CountTemplateRows() int {
	count, _ := t.TemplatesList.Count()
	return count
}

func (t TemplatesPage) GetTemplateInformation(webDriver *agouti.Page, templateName string) *TemplateInformation {
	rowCount, _ := t.TemplatesList.Count()
	for i := 0; i < rowCount; i++ {
		tileRow := t.TemplatesList.At(i).FindByXPath(fmt.Sprintf(`//td[1]//span[contains(text(), "%s")]/ancestor::tr`, templateName))
		if count, _ := tileRow.Count(); count == 1 {
			return &TemplateInformation{
				Name:           templateName,
				Type:           tileRow.FindByXPath(`td[2]`),
				Namespace:      tileRow.FindByXPath(`td[3]`),
				Provider:       tileRow.FindByXPath(`td[4]`),
				Description:    tileRow.FindByXPath(`td[5]`),
				CreateTemplate: tileRow.FindByXPath(`td[6]//button[@id="create-cluster"]`),
			}
		}
	}
	return nil
}

func (t TemplatesPage) GetTemplateTableList() []string {
	rowCount, _ := t.TemplatesList.Count()
	rows := make([]string, rowCount)

	for i := 0; i < rowCount; i++ {
		rows[i], _ = t.TemplatesList.At(i).Text()
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
	case "table":
		return t.TemplateView.At(0)
	case "grid":
		return t.TemplateView.At(1)
	}
	return nil
}
