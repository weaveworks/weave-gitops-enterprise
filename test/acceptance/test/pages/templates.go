package pages

import (
	"fmt"

	"github.com/sclevine/agouti"
	"github.com/weaveworks/weave-gitops-enterprise/test/selectors"
)

// Header webDriver elements
type TemplatesPage struct {
	TemplateHeader         *agouti.Selection
	TemplateCount          *agouti.Selection
	TemplateTiles          *agouti.MultiSelection
	TemplatesList          *agouti.MultiSelection
	TemplateProvider       *agouti.Selection
	TemplateProviderPopup  *agouti.MultiSelection
	TemplateViewGridButton *agouti.Selection
	TemplateViewListButton *agouti.Selection
}

// TemplatesPage webdriver initialises the webDriver object
func GetTemplatesPage(webDriver *agouti.Page) *TemplatesPage {
	templatesPage := TemplatesPage{
		TemplateHeader:         selectors.Get(webDriver, "templates", "page", "header"),
		TemplateCount:          selectors.Get(webDriver, "templates", "page", "count"),
		TemplateTiles:          selectors.GetMulti(webDriver, "templates", "gridView", "tiles"),
		TemplatesList:          selectors.GetMulti(webDriver, "templates", "page", "listView"),
		TemplateProvider:       selectors.Get(webDriver, "templates", "gridView", "provider"),
		TemplateProviderPopup:  selectors.GetMulti(webDriver, "templates", "gridView", "providerPopup"),
		TemplateViewGridButton: selectors.Get(webDriver, "templates", "page", "gridViewButton"),
		TemplateViewListButton: selectors.Get(webDriver, "templates", "page", "listViewButton"),
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

func (t TemplatesPage) CountTemplateRows() int {
	count, _ := t.TemplatesList.Count()
	return count
}

func (t TemplatesPage) GetTemplateRow(webDriver *agouti.Page, templateName string) *TemplateRecord {
	rowCount, _ := t.TemplatesList.Count()
	for i := 0; i < rowCount; i++ {
		tileRow := t.TemplatesList.At(i).FindByXPath(fmt.Sprintf(`//td[1]//span[contains(text(), "%s")]/ancestor::tr`, templateName))
		if count, _ := tileRow.Count(); count == 1 {
			return &TemplateRecord{
				Name:             templateName,
				Provider:         tileRow.FindByXPath(`td[3]`),
				Description:      tileRow.FindByXPath(`td[4]`),
				CreateTemplate:   tileRow.FindByXPath(`td[5]//button[@id="create-cluster"]`),
				ErrorHeader:      tileRow.Find(`.template-error-header`),
				ErrorDescription: tileRow.Find(`.template-error-description`),
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
		return t.TemplateViewListButton
	case "grid":
		return t.TemplateViewGridButton
	}
	return nil
}

func GetEntitelment(webDriver *agouti.Page, typeEntitelment string) *agouti.Selection {

	switch typeEntitelment {
	case "expired":
		return webDriver.FindByXPath(`//*/div[contains(text(), "Your entitlement for Weave GitOps Enterprise has expired")]`)
	case "invalid", "missing":
		return webDriver.FindByXPath(`//div[@class="Toastify"]//div[@role="alert"]//div[contains(text(), "No entitlement was found for Weave GitOps Enterprise")]`)
	}
	return nil
}
